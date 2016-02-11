package smtpd

import (
	"time"
)

// AuthConfig specifies the authentication mechanisms that
// the server announces as supported.
type AuthConfig struct {
	// Both slices should contain uppercase SASL mechanism names,
	// e.g. PLAIN, LOGIN, EXTERNAL.
	Mechanisms    []string // Supported mechanisms before STARTTLS
	TLSMechanisms []string // Supported mechanisms after STARTTLS
}

// Config represents the configuration for a Conn. If unset, Limits is
// DefaultLimits, LocalName is 'localhost', and SftName is 'go-smtpd'.
type Config struct {
	//TLSConfig *tls.Config   // TLS configuration if TLS is to be enabled
	Limits    *Limits       // The limits applied to the connection
	Auth      *AuthConfig   // If non-nil, client must authenticate before MAIL FROM
	Delay     time.Duration // Delay every character in replies by this much.
	SayTime   bool          // report the time and date in the server banner
	LocalName string        // The local hostname to use in messages
	SftName   string        // The software name to use in messages
	Announce  string        // extra stuff to announce in greeting banner
}
