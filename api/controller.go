package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

var null = []byte("null")
var empty = []byte("{}")

type output interface {
	JSON(data interface{}, w http.ResponseWriter, r *http.Request)
	Error500(errorString string, w http.ResponseWriter, r *http.Request)
}

type SimpleController struct {
	output
}

func (c SimpleController) JSON(data interface{}, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var content []byte
	var err error

	content, err = json.Marshal(data)

	if err != nil {
		log.Printf("JSON marshalling error: %s", err)
		c.Error500(err.Error(), w, r)
		return
	}

	if bytes.Equal(content, null) {
		_, err = w.Write(empty)
		return
	}

	_, err = w.Write(content)
}

func (c SimpleController) Error500(errorString string, w http.ResponseWriter, r *http.Request) {
	http.Error(w, errorString, http.StatusInternalServerError)
}
