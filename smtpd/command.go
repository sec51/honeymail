package smtpd

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/sec51/goconf"
	"strconv"
	"strings"
	"unicode"
)

var (
	domainName             = goconf.AppConf.DefaultString("smtp.domain", "google.com")
	kCommandNotRecognized  = fmt.Sprintf("500 %s", goconf.AppConf.DefaultString("smtp.cmd_not_recognized", "Command not recognized"))
	kSyntaxError           = fmt.Sprintf("501 %s", goconf.AppConf.DefaultString("smtp.syntax_error", "Syntax error in parameters or arguments"))
	kCommandNotImplemented = fmt.Sprintf("502 %s", goconf.AppConf.DefaultString("smtp.cmd_not_implemented", "Command not implemented"))
	kBadCommandSequence    = fmt.Sprintf("503 %s", goconf.AppConf.DefaultString("smtp.bad_cmd_sequence", "Bad command sequence"))
	// here the domain will be substituted if the string contains %s
	kGreeting = fmt.Sprintf("220 %s", goconf.AppConf.DefaultString("smtp.greetings", "%s Simple Mail Transfer Service Ready"))
	// here the automatically generated unique message id will be substituted if the string contains %s
	kMessageAccepted     = fmt.Sprintf("250 %s", goconf.AppConf.DefaultString("smtp.message_accepted", "Message %s accepted for delivery"))
	kRecipientAccepted   = fmt.Sprintf("250 %s", goconf.AppConf.DefaultString("smtp.recipient_accepted", "Okay, I'll believe you for now"))
	kClosingConnection   = fmt.Sprintf("221 %s", goconf.AppConf.DefaultString("smtp.closing_connection", "Closing transmission channel"))
	kRequestAborted      = fmt.Sprintf("451 %s", goconf.AppConf.DefaultString("smtp.request_aborted", "Requested action aborted: error in processing"))
	kLineTooLong         = fmt.Sprintf("500 %s", goconf.AppConf.DefaultString("smtp.line_too_long", "Line too long"))
	kPathTooLong         = fmt.Sprintf("501 %s", goconf.AppConf.DefaultString("smtp.path_too_long", "Path too long"))
	kTooManyRecipients   = fmt.Sprintf("452 %s", goconf.AppConf.DefaultString("smtp.too_many_recipients", "Too many recipients"))
	kTooMuchMailData     = fmt.Sprintf("552 %s", goconf.AppConf.DefaultString("smtp.mail_data_exceeded", "Mail data exceeded"))
	kInsufficientStorage = fmt.Sprintf("452 %s", goconf.AppConf.DefaultString("smtp.insufficient_storage", "Insufficient storage"))
	kSendData            = fmt.Sprintf("354 %s", goconf.AppConf.DefaultString("smtp.send_data_now", "Send away"))
	kVerifyAddress       = fmt.Sprintf("252 %s", goconf.AppConf.DefaultString("smtp.verify_addr_response", "Send some mail, I'll try my best"))
	kNoopCommand         = fmt.Sprintf("250 %s", goconf.AppConf.DefaultString("smtp.noop_response", "Yes I am still here"))
	kFixedSize           = 26214400
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
