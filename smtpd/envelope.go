package smtpd

import (
	log "github.com/Sirupsen/logrus"
	"net"
	"net/mail"
)

type envelope struct {
	Id         string // unique envelope ID
	RemoteIp   string
	RemotePort string
	From       *mail.Address
	To         *mail.Address
	Forward    []*mail.Address
	state      conState // the status of the envelope
	Message    []byte
}

func (e *envelope) isInDataMode() bool {
	return e.state == sData
}

func (e *envelope) MarkInDataMode() {
	e.state = sData
}

func (e *envelope) MarkInPostDataMode() {
	e.state = sPostData
}

func NewEnvelope(clientId string) *envelope {

	// set the agentId as remote ip string
	host, port, errSplit := net.SplitHostPort(clientId)
	if errSplit != nil {
		log.Println("Failed to split IP and port", errSplit)
		host = clientId
		port = ""
	}

	md := envelope{}
	md.Id = <-idGenerator // get a unique ID from the id.go generator
	md.RemoteIp = host
	md.RemotePort = port
	md.Forward = []*mail.Address{}
	md.state = sInitial
	return &md
}

func (md *envelope) addForward(mail *mail.Address) {
	md.Forward = append(md.Forward, mail)
}

// this is used internally to reset the data once the client send a RSET
func (md *envelope) reset() {
	md.Message = []byte{}
	md.From = nil
	md.To = nil
	md.Forward = []*mail.Address{}
	md.state = sInitial
}
