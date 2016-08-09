package utils

import (
	"bytes"
	"github.com/sec51/honeyssh/config"
	"net/http"
	"net/url"
	"strconv"
)

func MakeRequest(url, method string, params url.Values) (*http.Request, error) {

	keyIsPresent := params.Get("key")
	if keyIsPresent == "" {
		key := config.HONEYMASTER_KEY
		params.Add("key", key)
	}

	secretIsPresent := params.Get("secret")
	if secretIsPresent == "" {
		secret := config.HONEYMASTER_SECRET
		params.Add("secret", secret)
	}

	encodedData := params.Encode()

	req, err := http.NewRequest(method, url, bytes.NewBufferString(encodedData))
	if err != nil {
		return req, err
	}

	// isAuthEnabled := goconf.AppConf.DefaultBool("falcon.auth", false)

	// if isAuthEnabled {
	// 	user := goconf.AppConf.DefaultString("falcon.user", "")
	// 	pass := goconf.AppConf.DefaultString("falcon.password", "")
	// 	//fmt.Println(user, pass)
	// 	req.SetBasicAuth(user, pass)
	// }

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(encodedData)))

	return req, nil

}
