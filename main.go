package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/sec51/honeymail/smtpd"
)

func main() {
	ip := "localhost"
	port := "10025"
	serverName := ip
	withTLS := false
	server, err := smtpd.NewTCPServer(ip, port, serverName, withTLS)
	if err != nil {
		log.Fatal(err)
	}
	server.Start()
}
