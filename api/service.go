package api

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/sec51/honeymail/storage"
	"log"
	"net/http"
)

var storageService storage.StorageInterface

type apiService struct {
	port           string
	host           string
	storageService storage.StorageInterface
}

func NewAPIService(host, port string, storageSVC storage.StorageInterface) *apiService {

	storageService = storageSVC

	a := new(apiService)
	a.host = host
	a.port = port
	a.storageService = storageSVC
	return a
}

func (s *apiService) Start() {

	apiController := new(APIController)
	emailHandler := http.HandlerFunc(apiController.Email)

	// ROUTER
	router := httprouter.New()
	router.GET("/api/emails/:date", apiController.DateEmails)
	router.Handler("GET", "/api/email", emailHandler)

	// static file
	router.ServeFiles("/public/*filepath", http.Dir("public"))

	// listen
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", s.host, s.port), router))
}
