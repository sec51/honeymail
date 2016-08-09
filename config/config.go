package config

import (
	"github.com/sec51/goconf"
)

var (
	HONEYMASTER_KEY    = goconf.AppConf.String("honeymaster.key")
	HONEYMASTER_SECRET = goconf.AppConf.String("honeymaster.secret")

	HONEYPOT_SERVICE  = goconf.AppConf.String("honeypot.service")
	HONEYPOT_IP       = goconf.AppConf.String("honeypot.ip")
	HONEYPOT_LOCATION = goconf.AppConf.String("honeypot.location")
	HONEYPOT_PROVIDER = goconf.AppConf.String("honeypot.provider")

	PROCESS_IP_URL     = goconf.AppConf.String("url.process.ip")
	PROCESS_EMAILS_URL = goconf.AppConf.String("url.process.email")
)
