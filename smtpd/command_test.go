package smtpd

import (
	"net/mail"
	"testing"
)

func TestMailParsing(t *testing.T) {
	address := "<ned@thor.innosoft.com>"
	email, err := mail.ParseAddress(address)
	if err != nil {
		t.Error(err)
	}
	if email.Address != "ned@thor.innosoft.com" {
		t.Error("Could not parse email")
	}
}

func TestParseCommand(t *testing.T) {

	// check the size
	mailFromSize := "MAIL FROM:<ned@thor.innosoft.com> SIZE=50000000"
	cmd := ParseCmd(mailFromSize)

	if cmd.Cmd != MAILFROM {
		t.Error("Could not detect the right SMTP command")
	}

	if cmd.Response == "" {
		t.Error("Should have complained about the big message size")
	}

	if cmd.Argument != "<ned@thor.innosoft.com>" {
		t.Error("ParseCommand messes up the Argument")
	}

	// check the line length
	mailFromSize = "MAIL FROM:<ned@thor.innosoft.com> SIZE=50000000"
	for i := 0; i < 100; i++ {
		mailFromSize += "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	}
	cmd = ParseCmd(mailFromSize)

	if cmd.Response == "" {
		t.Error("Should have complained about the line length")
	}

}

func TestParseCommandASCII(t *testing.T) {

	// check the size
	mailFromSize := "MAIL FROM:<ned@åthor.innosoft.com> SIZE=50000000"
	cmd := ParseCmd(mailFromSize)
	if cmd.Response == "" {
		t.Error("Should have caught UTF-8 char")
	}

}

func TestInterpretCommandArgumentsHELO(t *testing.T) {

	cmd := NewParsedCommand()
	cmd.Cmd = HELO

	mailFromSize := "HeLo localhost"
	interpretCommandArguments(cmd, mailFromSize)

	if cmd.Response != "" {
		t.Error(cmd.Response)
	}

	if cmd.Argument != "localhost" {
		t.Errorf("HELO argument should be localhost - got: %s\n", cmd.Argument)
	}

	mailFromSize = "HELO"
	interpretCommandArguments(cmd, mailFromSize)

	if cmd.Response == "" {
		t.Error("HELO did not detect the `address` argument was not passed. This is an error")
	}

}

func TestInterpretCommandArgumentsEHLO(t *testing.T) {

	cmd := NewParsedCommand()
	cmd.Cmd = EHLO

	mailFromSize := "eHlO localhost"
	interpretCommandArguments(cmd, mailFromSize)

	if cmd.Response != "" {
		t.Error(cmd.Response)
	}

	if cmd.Argument != "localhost" {
		t.Errorf("EHLO argument should be localhost - got: %s\n", cmd.Argument)
	}

	mailFromSize = "EHLO"
	interpretCommandArguments(cmd, mailFromSize)

	if cmd.Response == "" {
		t.Error("EHLO did not detect the `address` argument was not passed. This is an error")
	}

}

func TestInterpretCommandArgumentsMAILFROM(t *testing.T) {

	cmd := NewParsedCommand()
	cmd.Cmd = MAILFROM

	mailFromSize := "MAIL FROM:<ned@thor.innosoft.com> SIZE=500000"
	interpretCommandArguments(cmd, mailFromSize)
	testMailFrom(cmd, t)

	mailFromSize = "MAIL FROM: <ned@thor.innosoft.com> SIZE=500000"
	interpretCommandArguments(cmd, mailFromSize)
	testMailFrom(cmd, t)

	mailFromSize = "MAIL FROM : <ned@thor.innosoft.com> SIZE=500000"
	interpretCommandArguments(cmd, mailFromSize)
	testMailFrom(cmd, t)

	mailFromSize = "mail from : <ned@thor.innosoft.com> SIZE=500000"
	interpretCommandArguments(cmd, mailFromSize)
	testMailFrom(cmd, t)

	mailFromSize = "mail from : ned@thor.innosoft.com SIZE=500000"
	interpretCommandArguments(cmd, mailFromSize)
	if cmd.Response == "" {
		t.Error("interpretCommandArguments cannot detect missing <> in MAIL FROM command")
	}

}

func TestInterpretCommandArgumentsRCPTO(t *testing.T) {

	cmd := NewParsedCommand()
	cmd.Cmd = RCPTTO

	rcpTo := "RCPT TO:<ned@thor.innosoft.com>"
	interpretCommandArguments(cmd, rcpTo)
	testRcptTo(cmd, t)

	rcpTo = "RCPT TO :<ned@thor.innosoft.com>"
	interpretCommandArguments(cmd, rcpTo)
	testRcptTo(cmd, t)

	rcpTo = "RCPT TO : <ned@thor.innosoft.com>"
	interpretCommandArguments(cmd, rcpTo)
	testRcptTo(cmd, t)

	rcpTo = "RCPT TO: <ned@thor.innosoft.com>"
	interpretCommandArguments(cmd, rcpTo)
	testRcptTo(cmd, t)

	rcpTo = "rcpt to : ned@thor.innosoft.com SIZE=500000"
	interpretCommandArguments(cmd, rcpTo)
	if cmd.Response == "" {
		t.Error("interpretCommandArguments cannot detect missing <> in RCPT TO command")
	}

}

func testMailFrom(cmd *ParsedCommand, t *testing.T) {

	if cmd.Response != "" {
		t.Error(cmd.Response)
	}

	if len(cmd.Parameters) == 0 {
		t.Error("There should be at least one parameter SIZE")
	}

	size, _ := cmd.Parameters["SIZE"]
	if size != "500000" {
		t.Error("Could not parse correctly SIZE parameter")
	}

	if cmd.Argument != "<ned@thor.innosoft.com>" {
		t.Error("Could not parse the argument (return-path) of MAIL FROM")
	}
}

func testRcptTo(cmd *ParsedCommand, t *testing.T) {

	if cmd.Response != "" {
		t.Error(cmd.Response)
	}

	if cmd.Argument != "<ned@thor.innosoft.com>" {
		t.Error("Could not parse the argument (return-path) of RCPT TO")
	}
}
