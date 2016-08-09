package main

import (
	//"flag"
	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/sec51/goconf"
	"github.com/sec51/honeymail/api"
	"github.com/sec51/honeymail/envelope"
	"github.com/sec51/honeymail/geoip"
	"github.com/sec51/honeymail/models"
	"github.com/sec51/honeymail/processing"
	"github.com/sec51/honeymail/processor"
	"github.com/sec51/honeymail/smtpd"
	"github.com/sec51/honeymail/storage"
)

func main() {

	// define configurations
	dbPath := goconf.AppConf.DefaultString("maxmind.db.path", "GeoLite2-City.mmdb")
	ip := goconf.AppConf.DefaultString("smtp.listen_to", "0.0.0.0")
	serverName := goconf.AppConf.DefaultString("smtp.server_name", "localhost")
	smtpPort := goconf.AppConf.DefaultString("smtp.port", "10025")
	smtpSecurePort := goconf.AppConf.DefaultString("smtp.secure_port", "10026")
	certificate := goconf.AppConf.DefaultString("smtp.tls.public_key", "")
	privateKey := goconf.AppConf.DefaultString("smtp.tls.private_key", "")

	apiHost := goconf.AppConf.DefaultString("http.listen_to", "0.0.0.0")
	apiPort := goconf.AppConf.DefaultString("http.port", "8080")

	// ===========================

	// DB STORAGE for emails
	db, err := bolt.Open("mail.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// GeoIP Resolution
	err = geoip.InitGeoDb(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	defer geoip.Resolver().Close()

	// ==========================================

	// channel for processing the envelopes
	envelopeChannel := make(chan envelope.Envelope)

	// channel for storing the envelopes
	storageChannel := make(chan envelope.Envelope)

	// channel for sending the emails to the honeymaster
	emailChan := make(chan models.Email)

	// channel for sending the bruteforce attacksd to the honeymaster
	bruteChan := make(chan models.BruteforceAttack)

	// ============================
	// Honeymaster processing service
	honeymasterService := processing.NewProcessingService(bruteChan, emailChan)
	honeymasterService.Start()

	// ============================
	// Storage service
	storageService := storage.NewStorageService(db, storageChannel)
	storageService.Start()

	// ============================
	// Processing service - caluclates the stats and extract additional info from each envelope
	// the passes it onto the storage channel for storing the results
	processorService := processor.NewProcessorService(envelopeChannel, storageChannel, emailChan)
	processorService.Start()

	// DEBUG ONLY
	if todayEmails, err := storageService.ViewTodayEnvelopes(); err == nil {
		for _, envelope := range todayEmails {
			log.Infof("Id: %s => From: %s => To: %s\n%s", envelope.Id, envelope.From.String(), envelope.To.String(), envelope.Message)
		}
	}
	// ============================

	withTLS := certificate != "" && privateKey != ""
	smtpServer, err := smtpd.NewTCPServer(ip, smtpPort, smtpSecurePort, serverName, certificate, privateKey, withTLS, envelopeChannel, bruteChan)
	if err != nil {
		log.Fatal(err)
	}

	smtpServer.Start()

	// API
	apiService := api.NewAPIService(apiHost, apiPort, storageService)
	apiService.Start()
}
