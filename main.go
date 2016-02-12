package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/sec51/honeymail/smtpd"
	"github.com/sec51/honeymail/storage"
)

func main() {

	// DB STORAGE for emails
	db, err := bolt.Open("mail.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	envelopeChannel := make(chan smtpd.Envelope)

	// ============================
	storageService := storage.NewStorageService(db, envelopeChannel)
	storageService.Start()
	if todayEmails, err := storageService.ViewTodayEnvelopes(); err == nil {
		for _, envelope := range todayEmails {
			log.Infof("Id: %s => From: %s => To: %s\n%s", envelope.Id, envelope.From.String(), envelope.To.String(), envelope.Message)
		}
	}
	// ============================

	ip := "localhost"
	port := "10025"
	serverName := ip
	withTLS := false
	server, err := smtpd.NewTCPServer(ip, port, serverName, withTLS, envelopeChannel)
	if err != nil {
		log.Fatal(err)
	}

	server.Start()
}
