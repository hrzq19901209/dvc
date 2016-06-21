package api

import (
	"bughunter.com/manager"
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

func HandleListMembers(manager *Manager) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc(ApiUrlPrefix+"list", HandleList()).Methods("GET")
	log.Fatal(http.ListenAndServe(":1234", r))
}
