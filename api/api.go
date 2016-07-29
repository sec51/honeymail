package api

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"time"
)

const dateFormat = "2006-01-02"

type APIController struct {
	SimpleController
}

func (c *APIController) DateEmails(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	date := p.ByName("date")
	log.Println(date)

	switch date {
	case "today", "":
		envs, err := storageService.ViewTodayEnvelopes()
		if err != nil {
			c.Error500(err, w, r)
			return
		}
		log.Printf("%v\n", envs)
		c.JSON(&envs, w, r)
		return

	default:
		parsedTime, err := time.Parse(dateFormat, date)
		if err != nil {
			c.Error500(err, w, r)
			return
		}

		envs, err := storageService.ViewDateEnvelopes(parsedTime)
		if err != nil {
			c.Error500(err, w, r)
			return
		}
		log.Printf("%v\n", envs)
		c.JSON(&envs, w, r)
		return
	}

}

func (c *APIController) Email(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	id := r.Form.Get("id")

	envs, err := storageService.GetEnvelope(id)

	if err != nil {
		c.Error500(err, w, r)
		return
	}

	c.JSON(&envs, w, r)

}

func (c *APIController) Alias(w http.ResponseWriter, r *http.Request) {

}

func (c *APIController) Malboxes(w http.ResponseWriter, r *http.Request) {

}

func TodayEmails(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}
