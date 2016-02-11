package smtpd

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"strconv"
	"strings"
	"unicode"
)

var (
	kCommandNotRecognized  = "500 Command not recognized."
	kSyntaxError           = "501 Syntax error in parameters or arguments"
	kCommandNotImplemented = "502 Command not implemented"
	kBadCommandSequence    = "503 Bad command sequence"
	kGreeting              = "220 foo.com Simple Mail Transfer Service Ready"
	kMessageAccepted       = "250 Message %s accepted for delivery"
	kRecipientAccepted     = "250 Okay, I'll believe you for now"
	kClosingConnection     = "221 foo.com Service closing transmission channel"
	kRequestAborted        = "451 Requested action aborted: error in processing"
	kLineTooLong           = "500 Line too long"
	kPathTooLong           = "501 Path too long"
	kTooManyRecipients     = "452 Too many recipients"
	kTooMuchMailData       = "552 Too much mail data"
	kFixedSize             = 26214400
	kInsufficientStorage   = "452 Insufficient channel storage"
)

// Define the SMTP command
type Command string

const (
	HELO     = "HELO"
	EHLO     = "EHLO"
	MAILFROM = "MAIL FROM"
	RCPTTO   = "RCPT TO"
	DATA     = "DATA"
	QUIT     = "QUIT"
	RSET     = "RSET"
	NOOP     = "NOOP"
	VRFY     = "VRFY"
	EXPN     = "EXPN"
	HELP     = "HELP"
	AUTH     = "AUTH"
	STARTTLS = "STARTTLS"
)

type ParsedCommand struct {
	Cmd        Command
	Argument   string
	Parameters map[string]string
	Response   string // response to give the client in case of an error parsing the command
}

func NewParsedCommand() *ParsedCommand {
	command := new(ParsedCommand)
	command.Cmd = ""
	command.Parameters = make(map[string]string)
	return command
}

func (pc *ParsedCommand) AddParameter(key, value string) {
	if key != "" {
		pc.Parameters[strings.ToUpper(key)] = value
	}
}

// Check that the command string contains only ASCII chars printable
func IsAsciiPrintable(s string) bool {
	for _, c := range s {
		if c > unicode.MaxASCII || !unicode.IsPrint(c) {
			return false
		}
	}
	return true
}

// this function interpret the command and extract the parameters if needed
func interpretCommandArguments(command *ParsedCommand, originalLine string) {

	// determine the command size
	commandSize := len(command.Cmd)

	// create a substring without the command
	lineWithNoCommand := originalLine[commandSize:]

	// remove leading and trailing space
	trimmedLine := strings.TrimSpace(lineWithNoCommand)

	// check if now the trimmedLine starts with a `:` if so remove it and trim it again
	if strings.HasPrefix(trimmedLine, ":") {
		trimmedLine = strings.Replace(trimmedLine, ":", "", 1) // replace just once
		trimmedLine = strings.TrimSpace(trimmedLine)
	}

	// divied the trimmed line in parts
	parts := strings.Fields(trimmedLine)

	// store the length of the command parts
	partsLength := len(parts)

	var argument string

	switch command.Cmd {
	case HELO, EHLO, MAILFROM, RCPTTO, VRFY:

		// expect valid domain or ip
		if partsLength == 0 {
			command.Response = kSyntaxError
			log.Errorln(fmt.Sprintf("%s expects an argument", command.Cmd))
			return
		}
		// assign the argument
		argument = parts[0]

		// make sure MAIL FROM and RCP TO contains <>
		if command.Cmd == MAILFROM || command.Cmd == RCPTTO {
			if !strings.HasPrefix(argument, "<") && !strings.HasSuffix(argument, ">") {
				command.Response = kSyntaxError
				log.Errorln(fmt.Sprintf("%s argument needs to have: <>", command.Cmd))
				return
			}
		}

		command.Argument = argument

		// means the client is sending the SIZE of the message
		if partsLength > 1 {
			for i := 1; i < partsLength; i++ {
				key, value, err := parseParameter(parts[i])
				if err != nil {
					log.Errorln(err)
					continue
				}
				command.AddParameter(key, value)
			}
		}

		break
	case DATA, RSET, STARTTLS, NOOP, QUIT:
		// this do not expect parameters
		if partsLength > 0 {
			command.Response = kSyntaxError
			log.Errorln(fmt.Sprintf("%s does NOT expect an argument", command.Cmd))
			return
		}
		break
	case AUTH:
		// expect one or two args
		// TODO: needs to be implemented
		command.Response = kCommandNotImplemented
		log.Errorln(fmt.Sprintf("%s not implemented", command.Cmd))
		break
	default:
		command.Response = kCommandNotRecognized
		log.Errorln("Unknown command")
	}
}

func parseParameter(arg string) (string, string, error) {
	splitted := strings.Split(arg, "=")
	if len(splitted) != 2 {
		return "", "", errors.New("Could not parse the additional command parameter")
	}
	return splitted[0], splitted[1], nil
}

// Parses the SMPT command sent from the client
// Line parameter expects to have CRLF already stripped out from textproto package
func ParseCmd(line string) *ParsedCommand {

	command := NewParsedCommand()

	// if it's an empty line then return: nothing to do here
	if line == "" {
		log.Errorln("got an empty command")
		command.Response = kCommandNotRecognized
		return command
	}

	if len(line) > 256 {
		command.Response = kLineTooLong
		return command
	}

	// make sure the command contains only printable ASCII chars
	if !IsAsciiPrintable(line) {
		log.Errorln("command contains non 7-bit printable ASCII")
		command.Response = kCommandNotRecognized
		return command
	}

	// Trim leading and trailing space in case the client does not conform with RFC
	line = strings.TrimSpace(line)

	// Make sure the command sent from the client is a supported one
	lineUpperCase := strings.ToUpper(line)

	switch {
	case strings.HasPrefix(lineUpperCase, HELO):
		command.Cmd = HELO
		break
	case strings.HasPrefix(lineUpperCase, EHLO):
		command.Cmd = EHLO
		break
	case strings.HasPrefix(lineUpperCase, MAILFROM):
		command.Cmd = MAILFROM
		break
	case strings.HasPrefix(lineUpperCase, RCPTTO):
		command.Cmd = RCPTTO
		break
	case strings.HasPrefix(lineUpperCase, VRFY):
		command.Cmd = VRFY
		break
	case strings.HasPrefix(lineUpperCase, EXPN):
		command.Cmd = EXPN
		break
	case strings.HasPrefix(lineUpperCase, STARTTLS):
		command.Cmd = STARTTLS
		break
	case strings.HasPrefix(lineUpperCase, NOOP):
		command.Cmd = NOOP
		break
	case strings.HasPrefix(lineUpperCase, RSET):
		command.Cmd = RSET
		break
	case strings.HasPrefix(lineUpperCase, QUIT):
		command.Cmd = QUIT
		break
	case strings.HasPrefix(lineUpperCase, DATA):
		command.Cmd = DATA
		break
	case strings.HasPrefix(lineUpperCase, AUTH):
		command.Cmd = AUTH
	}

	// parse the command
	interpretCommandArguments(command, line)

	// if the response is empty, means the parsing was successful, therefore one last check
	// check the size
	if command.Cmd == MAILFROM {
		if size, ok := command.Parameters["SIZE"]; ok {
			// convert it to an int and if successful check whether the size is bigger than advertised
			if sizeInt, err := strconv.Atoi(size); err == nil && sizeInt > kFixedSize {
				command.Response = kInsufficientStorage
			}
		}
	}

	return command
}
