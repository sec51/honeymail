package smtpd

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

var (
	kCommandNotRecognized  = "500 Command not recognized."
	kSyntaxError           = "501 Syntax error in parameters or arguments"
	kCommandNotImplemented = "502 Command not implemented"
	kBadCommandSequence    = "503 Bad command sequence"
	kGreeting              = "220 foo.com Simple Mail Transfer Service Ready"
	kMessageAccepted       = "250 Thanks message will be delivered ASAP"
	kRecipientAccepted     = "250 Okay, I'll believe you for now"
	kClosingConnection     = "221 foo.com Service closing transmission channel"
	kRequestAborted        = "451 Requested action aborted: error in processing"
)

// Command represents known SMTP commands in encoded form.
type Command int

// Recognized SMTP commands. Not all of them do anything
// (e.g. VRFY and EXPN are just refused).
const (
	noCmd  Command = iota // artificial zero value
	BadCmd Command = iota
	HELO
	EHLO
	MAILFROM
	RCPTTO
	DATA
	QUIT
	RSET
	NOOP
	VRFY
	EXPN
	HELP
	AUTH
	STARTTLS
)

// ParsedLine represents a parsed SMTP command line.  Err is set if
// there was an error, empty otherwise. Cmd may be BadCmd or a
// command, even if there was an error.
type ParsedLine struct {
	Cmd Command
	Arg string
	// Params is K=V for ESMTP MAIL FROM and RCPT TO
	// or the initial SASL response for AUTH
	Params   string
	Err      error
	Response string // this is the response to give in case of error parsing the command
}

// See http://www.ietf.org/rfc/rfc1869.txt for the general discussion of
// params. We do not parse them.

type cmdArgs int

const (
	noArg cmdArgs = iota
	canArg
	mustArg
	oneOrTwoArgs
	colonAddress // for ':<addr>[ options...]'
)

// Our ideal of what requires an argument is slightly relaxed from the
// RFCs, ie we will accept argumentless HELO/EHLO.
var smtpCommand = []struct {
	cmd     Command
	text    string
	argtype cmdArgs
}{
	{HELO, "HELO", canArg},
	{EHLO, "EHLO", canArg},
	{MAILFROM, "MAIL FROM", colonAddress},
	{RCPTTO, "RCPT TO", colonAddress},
	{DATA, "DATA", noArg},
	{QUIT, "QUIT", noArg},
	{RSET, "RSET", noArg},
	{NOOP, "NOOP", noArg},
	{VRFY, "VRFY", mustArg},
	{EXPN, "EXPN", mustArg},
	{HELP, "HELP", canArg},
	{STARTTLS, "STARTTLS", noArg},
	{AUTH, "AUTH", oneOrTwoArgs},
}

func (v Command) String() string {
	switch v {
	case noCmd:
		return "<zero Command value>"
	case BadCmd:
		return "<bad SMTP command>"
	default:
		for _, c := range smtpCommand {
			if c.cmd == v {
				return fmt.Sprintf("<SMTP '%s'>", c.text)
			}
		}
		// ... because someday I may screw this one up.
		return fmt.Sprintf("<Command cmd val %d>", v)
	}
}

// Returns True if the argument is all 7-bit ASCII. This is what all SMTP
// commands are supposed to be, and later things are going to screw up if
// some joker hands us UTF-8 or any other equivalent.
func isall7bit(b []byte) bool {
	for _, c := range b {
		if c > 127 {
			return false
		}
	}
	return true
}

// ParseCmd parses a SMTP command line and returns the result.
// The line should have the ending CR-NL already removed.
func ParseCmd(line string) ParsedLine {
	var res ParsedLine
	res.Cmd = BadCmd

	// We're going to upper-case this, which may explode on us if this
	// is UTF-8 or anything that smells like it.
	if !isall7bit([]byte(line)) {
		res.Err = errors.New("command contains non 7-bit ASCII")
		res.Response = kCommandNotRecognized
		return res
	}

	// Trim trailing space from the line, because some confused people
	// send eg 'RSET ' or 'QUIT '. Probably other people put trailing
	// spaces on other commands. This is probably not completely okay
	// by the RFCs, but my view is 'real clients trump RFCs'.
	line = strings.TrimRightFunc(line, unicode.IsSpace)

	// Search in the command table for the prefix that matches. If
	// it's not found, this is definitely not a good command.
	// We search on an upper-case version of the line to make my life
	// much easier.
	found := -1
	upper := strings.ToUpper(line)
	for i := range smtpCommand {
		if strings.HasPrefix(upper, smtpCommand[i].text) {
			found = i
			break
		}
	}
	if found == -1 {
		res.Err = errors.New("unrecognized command")
		res.Response = kCommandNotRecognized
		return res
	}

	// Validate that we've ended at a word boundary, either a space or
	// ':'. If we don't, this is not a valid match. Note that we now
	// work with the original-case line, not the upper-case version.
	cmd := smtpCommand[found]
	llen := len(line)
	clen := len(cmd.text)
	if !(llen == clen || line[clen] == ' ' || line[clen] == ':') {
		res.Err = errors.New("unrecognized command")
		res.Response = kCommandNotRecognized
		return res
	}

	// This is a real command, so we must now perform real argument
	// extraction and validation. At this point any remaining errors
	// are command argument errors, so we set the command type in our
	// result.
	res.Cmd = cmd.cmd
	switch cmd.argtype {
	case noArg:
		if llen != clen {
			res.Err = errors.New("SMTP command does not take an argument")
			res.Response = kSyntaxError
			return res
		}
	case mustArg:
		if llen <= clen+1 {
			res.Err = errors.New("SMTP command requires an argument")
			res.Response = kSyntaxError
			return res
		}
		// Even if there are nominal characters they could be
		// all whitespace. Although we've trimmed trailing
		// whitespace before now, there could be whitespace
		// *before* the argument and we want to trim it too.
		t := strings.TrimSpace(line[clen+1:])
		if len(t) == 0 {
			res.Err = errors.New("SMTP command requires an argument")
			res.Response = kSyntaxError
			return res
		}
		res.Arg = t
	case oneOrTwoArgs:
		// This implicitly allows 'a b c', with 'b c' becoming
		// the Params value.
		// TODO: is this desirable? Is this allowed by the AUTH RFC?
		parts := strings.SplitN(line, " ", 3)
		switch len(parts) {
		case 1:
			res.Err = errors.New("SMTP command requires at least one argument")
			res.Response = kSyntaxError
		case 2:
			res.Arg = parts[1]
		case 3:
			res.Arg = parts[1]
			res.Params = parts[2]
		}
	case canArg:
		// get rid of whitespace between command and the argument.
		if llen > clen+1 {
			res.Arg = strings.TrimSpace(line[clen+1:])
		}
	case colonAddress:
		var idx int
		// Minimum llen is clen + ':<>', three characters
		if llen < clen+3 {
			res.Err = errors.New("SMTP command requires an address")
			res.Response = kSyntaxError
			return res
		}
		// We explicitly check for '>' at the end of the string
		// to accept (at this point) 'MAIL FROM:<<...>>'. This will
		// fail if people also supply ESMTP parameters, of course.
		// Such is life.
		// TODO: reject them here? Maybe it's simpler.
		// BUG: this is imperfect because in theory I think you
		// can embed a quoted '>' inside a valid address and so
		// fool us. But I'm not putting a full RFC whatever address
		// parser in here, thanks, so we'll reject those.
		if line[llen-1] == '>' {
			idx = llen - 1
		} else {
			idx = strings.IndexByte(line, '>')
			if idx != -1 && line[idx+1] != ' ' {
				res.Err = errors.New("improper argument formatting")
				res.Response = kSyntaxError
				return res
			}
		}
		// NOTE: the RFC is explicit that eg 'MAIL FROM: <addr...>'
		// is not valid, ie there cannot be a space between the : and
		// the '<'. Normally we'd refuse to accept it, but a few too
		// many things invalidly generate it.
		if line[clen] != ':' || idx == -1 {
			res.Err = errors.New("improper argument formatting")
			res.Response = kSyntaxError
			return res
		}
		spos := clen + 1
		if line[spos] == ' ' {
			spos++
		}
		if line[spos] != '<' {
			res.Err = errors.New("improper argument formatting")
			res.Response = kSyntaxError
			return res
		}
		res.Arg = line[spos+1 : idx]
		// As a side effect of this we generously allow trailing
		// whitespace after RCPT TO and MAIL FROM. You're welcome.
		res.Params = strings.TrimSpace(line[idx+1 : llen])
	}
	return res
}
