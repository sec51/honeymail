package smtpd

// States of the SMTP conversation. These are bits and can be masked
// together.
type clientSessionState int

const (
	sInitial clientSessionState = 1 << iota
	sHelo
	sAuth
	sMail
	sRcpt
	sData
	sReceivingData
	sQuit
	sPostData
)

func (c clientSessionState) String() string {
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
