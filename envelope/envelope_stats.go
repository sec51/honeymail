package envelope

import (
	"crypto/sha256"
	"encoding/hex"
	log "github.com/Sirupsen/logrus"
	"github.com/oschwald/geoip2-golang"
	"github.com/sec51/honeymail/geoip"
	"net"
	"net/mail"
	"strings"
)

type EnvelopeStats struct {
	Subject           string           // the message subject
	SourceDomain      string           // the domain of the mail "From" header
	DestinationDomain string           // the domain of the rcpt to header (first)
	ForwardDomains    []string         // the domains of the additional rcp to
	RemoteInfo        *geoip2.City     // the resolved geoip information, from the client IP connections
	MessageHash       string           // this is the sha256 hash of the message body
	FromHash          string           // this is the sha256 hash of the from
	ToHash            string           // this is the sha256 hash of the to
	ForwardHash       []string         // this is the sha256 hash of all the forward addresses, one hash each
	FromInfo          []*geoip2.City   // this is based on the MX ip of the domain info of the From field (this is usually spoofed)
	DestinationInfo   []*geoip2.City   // this is based on the MX ip of the domain info of the TO field (The first only)
	ForwardInfo       [][]*geoip2.City // this is based on the MX ip of the domain info of the Forward field (the following ones)
	SPFPass           bool             // TODO: whether ot not this passed the SPF
	DKIMPass          bool             // TODO: whether ot not this was properly signed
	URLs              []string         // list of URLs in the email's body
	Attachments       []emailPart      // list of attachments
	EmailParts        []emailPart      // list of the email parts (for example there could be an HTML and a TXT representation of the message)
}

func NewEnvelopeStats(remoteIp string) *EnvelopeStats {
	stats := new(EnvelopeStats)
	remoteInfo, err := geoip.Resolver().ResolveIp(net.ParseIP(remoteIp))
	if err != nil {
		log.Infof("Could not resolve geographical info from remote peer: %s - %s", remoteIp, err)
	}
	stats.RemoteInfo = remoteInfo
	stats.ForwardDomains = []string{}
	stats.ForwardHash = []string{}
	stats.FromInfo = []*geoip2.City{}
	stats.DestinationInfo = []*geoip2.City{}
	stats.ForwardInfo = [][]*geoip2.City{}
	stats.URLs = []string{}
	stats.Attachments = []emailPart{}
	stats.EmailParts = []emailPart{}

	// TODO implement these checks
	stats.SPFPass = false
	stats.DKIMPass = false

	return stats
}

func hash(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}

func addressDomain(address *mail.Address) string {
	if address != nil && address.Address != "" {
		parts := strings.Split(address.Address, "@")
		if len(parts) > 1 {
			return parts[1]
		}
	}
	return ""
}

func getAddressGeoInfo(domain string) []*geoip2.City {
	var allInfo []*geoip2.City
	if domain != "" {
		if mxs, err := net.LookupMX(domain); err == nil {
			for _, mx := range mxs {
				if addresses, err := net.LookupHost(mx.Host); err == nil {
					for _, host := range addresses {
						if ip := net.ParseIP(host); ip != nil {
							if city, err := geoip.Resolver().ResolveIp(ip); err == nil {
								allInfo = append(allInfo, city)
							} else {
								log.Errorln("getAddressGeoInfo:", err)
							}
						}
					}
				}
			}
		}
	}
	return allInfo

}
