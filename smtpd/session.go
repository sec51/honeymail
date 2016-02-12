package smtpd

import (
	"crypto/tls"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"net"
)

type clientSession struct {
	hostname      string
	remoteAddress string
	remotePort    string
	conn          *net.TCPConn // the connection with the client
	tlsConn       *tls.Conn
	state         clientSessionState
	isTLS         bool //indicate whether the client upgraded successfully to TLS
	calledMail    bool
}

func NewClientSession(conn *net.TCPConn) *clientSession {

	host, port, errSplit := net.SplitHostPort(conn.RemoteAddr().String())
	if errSplit != nil {
		log.Errorln("Failed to split IP and port", errSplit)
		host = ""
		port = ""
	}

	s := clientSession{}
	s.conn = conn

	s.remoteAddress = host
	s.remotePort = port
	s.state = sInitial
	s.calledMail = false
	return &s
}

// verify the clinet provided hostname
// we need to verify:
// 1- MX presence
// 2- SPF (TXT)
// 3- DKIM (TXT)
// 4- DMARC (TXT)
// see 4.1.3 Address Literals - the remoteHostname can also be an IP address both ipv4 and ipv6
// in this case we need to verify a PTR record presence
func (s *clientSession) verifyHost(remoteHostname string) error {

	// if the string passed is localhost
	if remoteHostname == "localhost" {
		// assign it to the client session
		remoteHostname = s.remoteAddress
	}

	// try to parse it as IP address
	// check if it was a LoopBack IP address, if so check the remote ip of the connection
	if ip := net.ParseIP(remoteHostname); ip != nil && ip.IsLoopback() {
		remoteHostname = s.remoteAddress
	}

	// try with IP resolution first
	hosts, err := net.LookupAddr(remoteHostname)
	// this means it was an ip address
	if err == nil {

		for _, host := range hosts {
			err = verifyMX(host)
			// if it did not fail it means the host is a valid MX
			if err == nil {
				return err
			}
		}

	}

	// if we reached this point then we need to verify the MX via the hostname
	return verifyMX(remoteHostname)

}

func verifyMX(host string) error {

	if host == "" {
		return errors.New("Remote host verification failed got an empty hostname!")
	}

	mxs, err := net.LookupMX(host)
	if err != nil {
		return err
	}
	if len(mxs) == 0 {
		return errors.New(fmt.Sprintf("Could not find any MX record for the host %s", host))
	}

	return nil

}

// reste the session to initial state
func (s *clientSession) reset() {
	s.state = sInitial
	s.calledMail = false
}

func (s *clientSession) writeData(data string) error {
	if s.isTLS && s.tlsConn != nil {
		_, err := s.tlsConn.Write([]byte(data + "\r\n"))
		return err
	}
	_, err := s.conn.Write([]byte(data + "\r\n"))
	return err
}

func (s *clientSession) needsToQuit() bool {
	return s.state == sQuit
}

// this is used to verify that the current command is executed in the proper order
func (s *clientSession) verifyState(cmd Command) error {

	switch cmd {
	case HELO:
		if s.state != sInitial {
			return errors.New("Got HELO command at the wrong stage")
		}
	case MAILFROM:
		if s.state != sHelo {
			return errors.New("Got MAIL command at the wrong stage")
		}
	case RCPTTO:
		// if mail was never called and the state is different from mail
		// means it;s the first time after the mail command
		if !s.calledMail && s.state != sMail {
			return errors.New("Got RCPT command at the wrong stage")
		}
	case DATA:
		if s.state != sRcpt {
			return errors.New("Got DATA command at the wrong stage")
		}
	}

	return nil
}

// advance the session to the next stage
func (s *clientSession) markState(cmd Command) {

	switch cmd {
	case HELO, EHLO:
		s.state = sHelo
	case MAILFROM:
		s.state = sMail
		s.calledMail = true
	case RCPTTO:
		s.state = sRcpt
	case DATA:
		s.state = sData
	case RSET:
		s.state = sInitial
	case QUIT:
		s.state = sQuit
	}
}
