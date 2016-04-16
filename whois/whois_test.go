package whois

import (
	"testing"
)

func TestWhois(t *testing.T) {
	resultGeneral, err := Whois("google.ch")
	if err != nil {
		t.Fatal(err)
	}

	if resultGeneral.WhoisServer != "whois.nic.ch" {
		t.Errorf("WhoisServer property mismatch - expected `whois.nic.ch` got s\n", resultGeneral.WhoisServer)
	}

	if resultGeneral.Refer == "" {
		t.Errorf("Refer property mismatch - expected `whois.nic.ch` got s\n", resultGeneral.Refer)
	}

	if resultGeneral.Changed == "" {
		t.Error("Changed property is empty")
	}

	if resultGeneral.Created == "" {
		t.Errorf("Created property mismatch - expected `1987-05-20` got s\n", resultGeneral.Created)
	}
}
