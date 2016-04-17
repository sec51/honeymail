package goconf

import (
	"github.com/astaxie/beego/config"
	"log"
)

var (
	AppConf           config.Configer
	confFileLocations = []string{
		"conf/development.conf",
		"development.conf",
		"../conf/development.conf",
		"../../conf/development.conf",
		"../../../conf/development.conf",
		"conf/production.conf",
		"production.conf",
		"../conf/production.conf",
		"../../conf/production.conf",
		"../../../conf/production.conf",
	}
)

func init() {
	var err error
	for _, file := range confFileLocations {
		AppConf, err = config.NewConfig("ini", file)
		// if there is no error then we found it
		if err == nil {
			return
		}
	}

	if err != nil {
		log.Fatal("Could not find any config file.")
	}
}
