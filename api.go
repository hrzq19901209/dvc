package main

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"net/http"
)

const (
	ApiUrlPrefix = "/api/v1/"

	contentType     = "Content-Type"
	jsonContentType = "application/json;charset=UTF-8"
)

type People struct {
	Name string `json:"name"`
}

func HandleList() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		p := &People{
			Name: "bughunter",
		}

		w.Header().Set(contentType, jsonContentType)
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(p); err != nil {
			log.Errorf("Error occuered when mashalling response: %s", err)
			return
		}
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc(ApiUrlPrefix+"list", HandleList()).Methods("GET")
	log.Fatal(http.ListenAndServe(":1234", r))
}
