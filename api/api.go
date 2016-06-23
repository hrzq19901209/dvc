package main

import (
	"bughunter.com/dvc/config"
	"bughunter.com/dvc/manager"
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

type ListMembersResponse struct {
	Members map[string]*manager.Member `json:"members"`
}

func HandleListMembers(manager *manager.Manager) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		members := manager.GetMembers()
		resp := &ListMembersResponse{
			Members: members,
		}
		w.Header().Set(contentType, jsonContentType)
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Errorf("Error occured when marshalling response: %s", err)
			return
		}
	}
}

func main() {
	config.LoadConfigManager("test.config")
	manager := manager.NewManager()
	r := mux.NewRouter()
	r.HandleFunc(ApiUrlPrefix+"list", HandleListMembers(manager)).Methods("GET")
	log.Fatal(http.ListenAndServe(":1234", r))
}
