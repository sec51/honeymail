package smtpd

import (
	"bufio"
	"crypto/tls"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"net"
	"net/mail"
	"net/textproto"
	"sync"
	"time"
)

var (
	// map which contains a mapping between the connection and the conversation between the server and the client
	clientConnections = make(map[string]*clientSession)
	// mutex needed to moodify the map
	clientMutex sync.Mutex

	// current amount of clients connected
	totalClientConnections = 0

	// max amount of clients
	maxClientConnections = 64000
)

type tcpServer struct {
	stopMutex sync.Mutex
	localAddr string
	localPort string
	name      string
	withTLS   bool
	tlsConfig *tls.Config
	//readerChannel chan models.AgentMessage // from this channel we can read all the data coming from the clients
	//writerChannel chan []byte              // from this channel we can write all the data to the clients
	conn *net.TCPListener
}

// this is the module responsible for setting up a communication channel (TCP or UDP)
// where the data (protobuf, or JSON) can be exchanged

func NewTCPServer(ip, port, serverName string, withTLS bool) (*tcpServer, error) {

	server := tcpServer{
		localAddr: ip,
		localPort: port,
		name:      serverName,
		withTLS:   withTLS,
	}

	if withTLS {
		cert, err := tls.LoadX509KeyPair("/path/to/cert", "/path/to/key")
		if err != nil {
			return nil, err
		}
		server.tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.VerifyClientCertIfGiven,
			ServerName:   serverName}

	}

	return &server, nil

}

// this is a blocking call
func (s *tcpServer) Start() {

	addr, err := net.ResolveTCPAddr("tcp", s.localAddr+":"+s.localPort)

	if err != nil {
		log.Fatalln(err)
	}

	ln, err := net.ListenTCP("tcp", addr)

	if err != nil {
		log.Fatalln(err)
	}

	// assign the conn to stop the server
	s.stopMutex.Lock()
	s.conn = ln
	s.stopMutex.Unlock()

	log.Infoln("Mail server is listening on", s.localAddr, ":", s.localPort)

	for {
		if conn, err := ln.AcceptTCP(); err == nil {
			// we accept a maximum of 6400 concurrent connections
			// each agent creates 1 connection, therefore it should be enough for handling up to 6400 agents
			clientMutex.Lock()
			if totalClientConnections >= maxClientConnections {
				log.Errorln("Too many connections from mail clients. Stopped accepting new connections.")
				continue
			}
			clientMutex.Unlock()

			// otherwise accept the connection
			log.Infoln("Amount of mail client connections:", totalClientConnections)

			// set a read timeout
			conn.SetReadDeadline(time.Now().Add(DefaultLimits.CmdInput))
			go s.handleTCPConnection(NewClientSession(conn))

		}
	}
}

func (s *tcpServer) Stop() error {
	s.stopMutex.Lock()
	defer s.stopMutex.Unlock()
	return s.conn.Close()
}

// Handles incoming requests.
// withTLS means the client connected directly with TLS
// this means you need to create two TCP server objects.
// one which listen to the TLS port wanted
func (s *tcpServer) handleTCPConnection(client *clientSession) {

	// close the connection in case this exists
	defer client.conn.Close()

	// get the client remote address
	clientId := client.conn.RemoteAddr().String()

	// write the welcome message to the client
	if err := client.writeData(kGreeting); err != nil {
		log.Println("Error writing greeting message to mail client", err)
		return
	}

	// increment connection counter
	s.incrementConnectionCounter(clientId)

	// new mail client connection was successfully created
	// create a new envelope because we expect the client to send the HELO/EHLO command
	envelope := NewEnvelope(clientId)

	// new buffered reader
	bufferedReader := bufio.NewReader(client.conn)
	reader := textproto.NewReader(bufferedReader)

	// parsed command
	var command ParsedLine

command_loop:
	for {

		// we are receiving data
		// so we need to keep reading it and we cannot read the input as a command
		// therefore it needs to happen before the ReadLine
		if envelope.isInDataMode() {
			// check if the message ends and read all buffer
			// if the message does not end, continue reading
			dotBytes, err := reader.ReadDotBytes()

			if err != nil {
				break
			}
			// means the message ends
			if err == nil && len(dotBytes) > 0 {

				// assign the data read to the mailData struct
				envelope.Message = dotBytes

				// write back to the client
				client.writeData(kMessageAccepted)

				// set the state as post data, so during the loop it does not eneter here again
				envelope.MarkInPostDataMode()

				// queue the envelope for processing
				// at this stage the client is allowed only to RSET or to QUIT
				// dereference the envelope and send it
				s.queueForDelivery(*envelope)

				// continue the loop
				continue
			}

			// continue reading
			continue
		}

		// read the command sent from the client, which is in the buffer
		line, err := reader.ReadLine()

		// parse the command line
		command = ParseCmd(line)
		if command.Err != nil {
			log.Println("CAUGHT error while parsing the command", err, line)

			// write the error response
			response := command.Response
			if response == "" {
				response = kCommandNotRecognized
			}
			client.writeData(response)
			continue
		}

		// verify that it's a valid command in the sequence
		// if it's not valid then answer and wait for a different command (continue)
		if err := client.verifyState(command.Cmd); err != nil {
			client.writeData(kBadCommandSequence)
			continue
		}

		// mark the state
		client.markState(command.Cmd)

		switch command.Cmd {
		case EXPN:
			client.writeData(kCommandNotImplemented)
			continue
		case HELP:
			client.writeData("214 SMTP servers help those who help themselves.")
			client.writeData("214 Go read http://cr.yp.to/smtp.html.")
			break
		case NOOP:
			client.writeData("250 Yes I am still here")
			break
		case VRFY:
			client.writeData("252 Send some mail, I'll try my best")
			break
		case RSET:
			// reset the envelope
			envelope = nil
			envelope = NewEnvelope(clientId)

			// resent the client state for the sequence of commands
			client.reset()
			break
		case STARTTLS:
			// Init a new TLS connection. I need a *tls.Conn type
			// so that I can do the Handshake()
			var tlsConn *tls.Conn
			tlsConn = tls.Server(client.conn, s.tlsConfig)

			// run a handshake
			// Verify on the RFC what the server is supposed to do when the TLS handshake fails
			err := tlsConn.Handshake()
			if err != nil {
				log.Errorln("Failed to handshake with the client a valid SSL connection")
				client.writeData(kClosingConnection)
				break command_loop
			}

			client.isTLS = true
			// Here is the trick. Since I do not need to access
			// any of the TLS functions anymore,
			// I can convert tlsConn back in to a net.Conn type
			client.tlsConn = tlsConn

			// defer closing of the connection
			defer client.tlsConn.Close()

			break
		case HELO:
			if err := client.verifyHost(command.Arg); err != nil {
				log.Errorln("Suspicious connection...continuing nonetheless")
			}
			client.writeData(fmt.Sprintf("250 %s Hello %v", s.name, client.remoteAddress))
			break
		case EHLO:
			if err := client.verifyHost(command.Arg); err != nil {
				log.Errorln("Suspicious connection...continuing nonetheless")
			}

			client.writeData(fmt.Sprintf("250-%s Hello %v", s.name, client.remoteAddress))
			// We advertise 8BITMIME per
			// http://cr.yp.to/smtp/8bitmime.html
			client.writeData("250-8BITMIME")
			client.writeData("250-VRFY")
			client.writeData("250-HELP")
			client.writeData("250-PIPELINING")
			// STARTTLS RFC says: MUST NOT advertise STARTTLS
			// after TLS is on.
			if !s.withTLS && !client.isTLS {
				client.writeData("250-STARTTLS")
			}
			// RFC4954 notes: A server implementation MUST
			// implement a configuration in which it does NOT
			// permit any plaintext password mechanisms, unless
			// either the STARTTLS [SMTP-TLS] command has been
			// negotiated...
			// if c.Config.Auth != nil {
			// 	c.replyMore("250-AUTH " + strings.Join(c.authMechanisms(), " "))
			// }
			// We do not advertise SIZE because our size limits
			// are different from the size limits that RFC 1870
			// wants us to use. We impose a flat byte limit while
			// RFC 1870 wants us to not count quoted dots.
			// Advertising SIZE would also require us to parse
			// SIZE=... on MAIL FROM in order to 552 any too-large
			// sizes.
			// On the whole: pass. Cannot implement.
			// (In general SIZE is hella annoying if you read the
			// RFC religiously.)
			//c.replyMore("250 HELP")
		case AUTH:
			//c.authDone(true)
			//c.reply("235 Authentication successful")
			break
		case MAILFROM:

			// parse the mail address and make sure it'a a valid one
			fromAddress, err := verifyEmailAddress(command.Arg)
			if err != nil {
				log.Println("Error parsing FROM address", err)
				client.writeData(kRequestAborted)
				continue
			}
			envelope.From = fromAddress
			client.writeData(kRecipientAccepted)
			break
		case RCPTTO:

			// parse the mail address and make sure it'a a valid one
			toAddress, err := verifyEmailAddress(command.Arg)
			if err != nil {
				log.Println("Error parsing TO address", err)
				client.writeData(kRequestAborted)
				continue
			}

			// the first add it to the TO the following to the forward
			if envelope.To == nil {
				envelope.To = toAddress
			} else {
				envelope.addForward(toAddress)
			}

			client.writeData("250 Okay, I'll believe you for now")
			break
		case DATA:
			client.writeData("354 Send away")
			break
			//}
		case QUIT:
			client.writeData(kClosingConnection)
			break command_loop
		default:
			if client.needsToQuit() {
				break command_loop
			}
		}

	}

	// at this point the connection will be closed therefore decrease the counter
	s.decrementConnectionCounter(clientId)
	log.Infoln("Client", clientId, "disconnected")

}

func (s *tcpServer) incrementConnectionCounter(clientId string) {

	// update the map and the total connections
	clientMutex.Lock()
	totalClientConnections++
	clientConnections[clientId] = client
	clientMutex.Unlock()

}

func (s *tcpServer) decrementConnectionCounter(clientId string) {
	clientMutex.Lock()
	totalClientConnections--
	delete(clientConnections, clientId)
	clientMutex.Unlock()
}

func (s *tcpServer) queueForDelivery(e envelope) {

}

func verifyEmailAddress(email string) (*mail.Address, error) {
	return mail.ParseAddress(email)
}
