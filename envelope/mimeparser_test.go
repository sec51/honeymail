package envelope

import (
	"bytes"
	"io/ioutil"
	"net/mail"
	"testing"
)

func TestEmailParsing(t *testing.T) {
	data, err := ioutil.ReadFile("virus_email_do_not_open.eml")
	if err != nil {
		t.Fatal(err)
	}

	reader := bytes.NewReader(data)
	message, err := mail.ReadMessage(reader)
	if err != nil {
		t.Fatal(err)
	}

	parts, err := parseEmailParts(*message)
	if err != nil {
		t.Fatal(err)
	}

	if len(parts) != 2 {
		t.Fatal("Failed to parse the email message")
	}

}
