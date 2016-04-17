package envelope

import (
	"bufio"
	"bytes"
	"encoding/gob"
	//"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/mvdan/xurls"
	"github.com/oschwald/geoip2-golang"
	"io/ioutil"
	"net"
	"net/mail"
	"time"
)

func init() {
	gob.Register(Envelope{})
	gob.Register(EnvelopeStats{})
	gob.Register(mail.Address{})
	gob.Register(geoip2.City{})
	gob.Register(emailPart{})
	gob.Register(mail.Message{})
	gob.Register(bufio.Reader{})
	gob.Register(mail.Header{})
}

type Envelope struct {
	Id                string          // unique envelope ID
	RemoteIp          string          // remote ip of the connection
	RemotePort        string          // remote port of the connection
	From              *mail.Address   // From email info
	To                *mail.Address   // To Email info
	Forward           []*mail.Address // Forward array of email info
	Timestamp         time.Time       // Timestamp when the message was received
	SecurelyDelivered bool            // If it was securely delivered
	Stats             *EnvelopeStats  // Additional statistics about the message
	Body              []byte          // Body of the message
	Headers           mail.Header     // headers of the message
	Message           []byte          // contains the full original message converted to bytes

	// this is the parse mail message
	// needed to extract additional info but not exported because it's a copy of Message field
	mailMessage mail.Message
}

type MailMessage struct {
	Body    []byte
	Headers mail.Header
}

func NewEnvelope(clientId string, isConnectionSecure bool) *Envelope {

	// set the clientId as remote ip string
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
	md.SecurelyDelivered = isConnectionSecure
	md.Timestamp = time.Now().UTC()
	return &md
}

func (md *Envelope) AddForward(mail *mail.Address) {
	md.Forward = append(md.Forward, mail)
}

// Serialize the envelope into bytes
func (md *Envelope) Serialize() ([]byte, error) {
	var buffer bytes.Buffer
	e := gob.NewEncoder(&buffer)
	err := e.Encode(md)
	return buffer.Bytes(), err
	// return json.Marshal(md)
}

// De-serialize back the envelope from bytes
func EnvelopeFromBytes(data []byte) (*Envelope, error) {
	md := Envelope{}
	var buffer bytes.Buffer
	buffer.Write(data)
	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(&md)
	// err := json.Unmarshal(data, &md)
	return &md, err
}

func (md *Envelope) generateMailMessage() error {
	reader := bytes.NewReader(md.Message)
	message, err := mail.ReadMessage(reader)
	if err != nil {
		log.Errorf("Could not parse the mail message id %s with error: %s", md.Id, err)
		return err
	}
	md.mailMessage = *message
	return nil
}

// process the envelope and generate its statistics
func (md *Envelope) CalculateStats() {

	md.Stats = NewEnvelopeStats(md.RemoteIp)

	// parse the mail message
	if err := md.generateMailMessage(); err == nil {
		// TODO: parse the Bcc header and add all the info to the stats

		// extract the message parts
		parts, _ := parseEmailParts(md.mailMessage)
		for _, part := range parts {
			if part.IsAttachment {
				md.Stats.Attachments = append(md.Stats.Attachments, part)
			} else {
				md.Stats.EmailParts = append(md.Stats.EmailParts, part)
			}
		}

		// extract the Subject
		md.Stats.Subject = md.mailMessage.Header.Get("Subject")

		// IMPORTANT !!!!
		// This is needed again, otherwise the reader has reached the end and no data is extracted !
		// There is probably a better way of doing it...
		md.generateMailMessage()
		// ==========

		// extract the body
		if body, err := ioutil.ReadAll(md.mailMessage.Body); err == nil {
			// convert to a string
			bodyString := string(body)
			log.Infoln("Body", bodyString)

			// extract the URLs
			md.Stats.URLs = append(md.Stats.URLs, xurls.Strict.FindAllString(bodyString, -1)...)

			// assign the body to the OriginalMessage
			md.Body = body
			md.Headers = md.mailMessage.Header

		} else {
			log.Errorf("Error parsgin the body of the message %s with error: %s", md.Id, err)
		}
	}

	// Message hash
	md.Stats.MessageHash = hash(md.Message)

	// From hash
	md.Stats.FromHash = hash([]byte(md.From.Address))

	// To hash
	md.Stats.ToHash = hash([]byte(md.To.Address))

	// source domain
	md.Stats.SourceDomain = addressDomain(md.From)
	md.Stats.FromInfo = getAddressGeoInfo(md.Stats.SourceDomain)

	// destination domain
	md.Stats.DestinationDomain = addressDomain(md.To)
	md.Stats.DestinationInfo = getAddressGeoInfo(md.Stats.DestinationDomain)

	// all info about forward
	for _, forward := range md.Forward {
		// calculate the hash for each forward
		md.Stats.ForwardHash = append(md.Stats.ForwardHash, hash([]byte(forward.Address)))

		if domain := addressDomain(forward); domain != "" {
			md.Stats.ForwardDomains = append(md.Stats.ForwardDomains, domain)
			if info := getAddressGeoInfo(domain); len(info) > 0 {
				md.Stats.ForwardInfo = append(md.Stats.ForwardInfo, info)
			}
		}
	}

}
