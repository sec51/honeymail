package api

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type APIController struct {
	SimpleController
}

func (c *APIController) TodayEmail(w http.ResponseWriter, r *http.Request) {

	envs, err := storageService.ViewTodayEnvelopes()

	if err != nil {
		fmt.Printf("%s\n", err)
		c.Error500(err.Error(), w, r)
		return
	}

	c.JSON(&envs, w, r)

}

func (c *APIController) Email(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	id := r.Form.Get("id")

	envs, err := storageService.GetEnvelope(id)

	if err != nil {
		fmt.Printf("%s\n", err)
		c.Error500(err.Error(), w, r)
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
