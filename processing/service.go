package processing

import (
	"net/http"
	"net/url"

	"github.com/sec51/honeymail/config"
	log "github.com/sec51/honeymail/logging"
	"github.com/sec51/honeymail/models"
	"github.com/sec51/honeymail/utils"
)

var (
	client = http.Client{}
)

type ProcessingService struct {
	bruteforce chan models.BruteforceAttack
	emails     chan models.Email
}

func NewProcessingService(bruteforce chan models.BruteforceAttack, emails chan models.Email) ProcessingService {
	s := ProcessingService{
		bruteforce: bruteforce,
		emails:     emails,
	}

	return s
}

func (s ProcessingService) Start() {
	go processIp(s.bruteforce)
	go processEmails(s.emails)
}

func processIp(attacks chan models.BruteforceAttack) {
	for bf := range attacks {
		params := url.Values{
			"ip":                {bf.Ip},
			"service":           {config.HONEYPOT_SERVICE},
			"type":              {"bruteforce"},
			"honeypot_ip":       {config.HONEYPOT_IP},
			"honeypot_location": {config.HONEYPOT_LOCATION},
			"honeypot_provider": {config.HONEYPOT_PROVIDER},
		}

		req, err := utils.MakeRequest(config.PROCESS_IP_URL, "POST", params)
		if err != nil {
			log.Error.Println("processIp:", err)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Error.Println("processIp:", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			log.Error.Printf("processIp - Got invalid HTTP response: %d\n", resp.StatusCode)
			continue
		}

	}
}

func processEmails(emails chan models.Email) {
	for cmd := range emails {
		params := url.Values{
			"ip":   {cmd.Ip},
			"emai": {cmd.Data},
		}

		req, err := utils.MakeRequest(config.PROCESS_EMAILS_URL, "POST", params)
		if err != nil {
			log.Error.Println("processEmails:", err)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Error.Println("processEmails:", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			log.Error.Printf("processEmails - Got invalid HTTP response: %d\n", resp.StatusCode)
			continue
		}

	}
}
