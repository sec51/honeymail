package smtpd

import (
	"bytes"
	"encoding/gob"
	log "github.com/Sirupsen/logrus"
	"net"
	"net/mail"
	"time"
)

type Envelope struct {
	Id         string // unique envelope ID
	RemoteIp   string
	RemotePort string
	From       *mail.Address
	To         *mail.Address
	Forward    []*mail.Address
	state      clientSessionState // the status of the envelope
	Message    []byte
	Timestamp  time.Time
}

func (e *Envelope) isInDataMode() bool {
	return e.state == sData
}

func (e *Envelope) MarkInDataMode() {
	e.state = sData
}

func (e *Envelope) MarkInPostDataMode() {
	e.state = sPostData
}

func NewEnvelope(clientId string) *Envelope {

	// set the agentId as remote ip string
	host, port, errSplit := net.SplitHostPort(clientId)
	if errSplit != nil {
		log.Println("Failed to split IP and port", errSplit)
		host = clientId
		port = ""
	}

	md := Envelope{}
	md.Id = <-idGenerator // get a unique ID from the id.go generator
	md.RemoteIp = host
	md.RemotePort = port
	md.Forward = []*mail.Address{}
	md.state = sInitial
	md.Timestamp = time.Now().UTC()
	return &md
}

func (md *Envelope) addForward(mail *mail.Address) {
	md.Forward = append(md.Forward, mail)
}

// this is used internally to reset the data once the client send a RSET
func (md *Envelope) reset() {
	md.Message = []byte{}
	md.From = nil
	md.To = nil
	md.Forward = []*mail.Address{}
	md.state = sInitial
}

func (md *Envelope) Serialize() ([]byte, error) {
	var buffer bytes.Buffer
	e := gob.NewEncoder(&buffer)
	err := e.Encode(md)
	return buffer.Bytes(), err
}

func EnvelopeFromBytes(data []byte) (*Envelope, error) {
	md := Envelope{}
	var buffer bytes.Buffer
	buffer.Write(data)
	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(&md)
	return &md, err
}
