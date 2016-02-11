package smtpd

import "time"

// Limits has the time and message limits for a Conn, as well as some
// additional options.
//
// A Conn always accepts 'BODY=[7BIT|8BITMIME]' as the sole MAIL FROM
// parameter, since it advertises support for 8BITMIME.
type Limits struct {
	CmdInput time.Duration // client commands, eg MAIL FROM
	MsgInput time.Duration // total time to get the email message itself
	ReplyOut time.Duration // server replies to client commands
	TLSSetup time.Duration // time limit to finish STARTTLS TLS setup
	MsgSize  int64         // total size of an email message
	BadCmds  int           // how many unknown commands before abort
	NoParams bool          // reject MAIL FROM/RCPT TO with parameters
}

// The default limits that are applied if you do not specify anything.
// Two minutes for command input and command replies, ten minutes for
// receiving messages, and 5 Mbytes of message size.
//
// Note that these limits are not necessarily RFC compliant, although
// they should be enough for real email clients.
var DefaultLimits = Limits{
	CmdInput: 2 * time.Minute,
	MsgInput: 10 * time.Minute,
	ReplyOut: 2 * time.Minute,
	TLSSetup: 4 * time.Minute,
	MsgSize:  5 * 1024 * 1024,
	BadCmds:  5,
	NoParams: true,
}
