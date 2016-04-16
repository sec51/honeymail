// this package gathers the whois info from whois.iana.org
// this is sufficient for getting the creation date of a domain, which is what we are interested in
// when checking against phishing attacks

package whois

import (
	"bufio"
	log "github.com/Sirupsen/logrus"
	"io"
	"net"
	"net/textproto"
	"strings"
)

var whoisServer = "whois.iana.org"
var whoisEndPoint = whoisServer + ":43"

type whoisGeneralResponse struct {
	WhoisServer string
	Refer       string
	Status      string
	Created     string
	Changed     string
}

func Whois(domain string) (whoisGeneralResponse, error) {
	response := whoisGeneralResponse{}
	conn, err := net.Dial("tcp", whoisEndPoint)
	if err != nil {
		return response, err
	}
	defer conn.Close()
	conn.Write([]byte(domain + "\r\n"))

	buffer := bufio.NewReader(conn)
	reader := textproto.NewReader(buffer)

	for {
		line, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorln(err)
		}

		key, value := parseTokens(line)
		switch key {
		case "whois":
			response.WhoisServer = value
			continue
		case "refer":
			response.Refer = value
			continue
		case "status":
			response.Status = value
			continue
		case "changed":
			response.Changed = value
			continue
		case "created":
			response.Created = value
			continue
		}

	}

	return response, nil
}

func parseTokens(line string) (string, string) {

	tokens := strings.Fields(line)
	if len(tokens) > 0 {
		var key string
		switch tokens[0] {
		case "whois:":
			key = "whois"
			break
		case "created:":
			key = "created"
			break
		case "changed:":
			key = "changed"
			break
		case "status:":
			key = "status"
			break
		case "refer:":
			key = "refer"
			break
		}
		return key, tokens[1]
	}

	return "", ""
}
