package api

import (
	"bytes"
	"encoding/json"
	"github.com/sec51/goconf"
	"io"
	"log"
	"net"
	"net/http"
)

var null = []byte("null")
var empty = []byte("{}")
var validIPs = goconf.AppConf.DefaultStrings("http.allowed_hosts", []string{"127.0.0.1"})

type output interface {
	JSON(data interface{}, w http.ResponseWriter, r *http.Request)
	Error500(errorString string, w http.ResponseWriter, r *http.Request)
}

type SimpleController struct {
	output
}

func arrayContains(ip string, validationSet []string) bool {

	for _, entry := range validationSet {
		if ip == entry {
			return true
		}
	}

	return false
}

func isValidIP(ip string) bool {

	host, _, err := net.SplitHostPort(ip)
	if err != nil {
		return false
	}

	return arrayContains(host, validIPs)

}

func (c SimpleController) JSON(data interface{}, w http.ResponseWriter, r *http.Request) {

	// if the ip is invalid then return 404
	if !isValidIP(r.RemoteAddr) {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var content []byte
	var err error

	content, err = json.Marshal(data)

	if err != nil {
		log.Printf("JSON marshalling error: %s", err)
		c.Error500(err, w, r)
		return
	}

	if bytes.Equal(content, null) {
		_, err = w.Write(empty)
		return
	}

	_, err = w.Write(content)
}

func (c SimpleController) Error500(err error, w http.ResponseWriter, r *http.Request) {

	if !isValidIP(r.RemoteAddr) {
		http.NotFound(w, r)
		return
	}

	if err == io.EOF {
		http.Error(w, "Item not found", http.StatusInternalServerError)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
