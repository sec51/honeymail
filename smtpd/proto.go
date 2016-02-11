package smtpd

// States of the SMTP conversation. These are bits and can be masked
// together.
type conState int

const (
	sStartup conState = iota // Must be zero value

	sInitial conState = 1 << iota
	sHelo
	sAuth // during SASL dialog
	sMail
	sRcpt
	sData
	sReceivingData
	sQuit // QUIT received and ack'd, we're exiting.
	sPostData
	sAbort

	// Synthetic state

)

func (c conState) String() string {
	switch c {
	case sHelo:
		return "HELLO"
	case sAuth:
		return "AUTH"
	case sMail:
		return "MAIL"
	case sRcpt:
		return "RCPT"
	case sData:
		return "DATA"
	case sPostData:
		return "POST_DATA"
	case sQuit:
		return "QUIT"
	default:
		return "unknown"
	}
}

// A command not in the states map is handled in all states (probably to
// be rejected).
var states = map[Command]struct {
	validin, next conState
}{
	HELO:     {sInitial | sHelo, sHelo},
	EHLO:     {sInitial | sHelo, sHelo},
	AUTH:     {sHelo, sHelo},
	MAILFROM: {sHelo, sMail},
	RCPTTO:   {sMail | sRcpt, sRcpt},
	DATA:     {sRcpt, sData},
}
