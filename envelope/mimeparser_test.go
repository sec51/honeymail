package envelope

import (
	"bytes"
	"io/ioutil"
	"net/mail"
	"testing"
)

func TestEmailParsing(t *testing.T) {
	data, err := ioutil.ReadFile("data/virus_email_do_not_open.eml")
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

func TestSignedEmailParsing(t *testing.T) {
	data, err := ioutil.ReadFile("data/test_signed_email.eml")
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

	if !bytes.Equal(parts[0].Data, []byte{118, 118, 118, 118, 118, 118, 10}) {
		t.Fatal("Failed to parse the body of the message")
	}

	if len(parts) != 2 {
		t.Fatal("Failed to parse the signed email message")
	}
}

func TestSeveralMultipartEmailParsing(t *testing.T) {
	data, err := ioutil.ReadFile("data/several_multipart.eml")
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

	if len(parts) != 3 {
		t.Fatal("Failed to parse the signed email message")
	}
}
