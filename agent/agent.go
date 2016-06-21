package main

import (
	"bughunter.com/dvc/config"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/content"
	"time"
)

type AgentInfo struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
}

type Agent struct{}

func HeartBeat(config *config.Config) {
	api := client.NewKeysAPI(config.Etcd)

	for {
		key := fmt.Sprintf("agents/%s/%s", config.Cluster, config.Hostname)

		info := &Agent{
			IP:       config.IP,
			Hostname: config.Hostname,
		}
		value, _ := json.Marshal(info)
		_, err = api.Set(context.Background(), key, string(value), &client.SetOptions{
			TTL: time.Second * 30,
		})

		if err != nil {
			log.Errorf("Warning: update workerInfo &s", err)
		}
		time.Sleep(time.Second * 3)
	}
}

func NewAgent() *Agent {
	config := config.GetConfig()
	agent := &Agent{}
	go HeartBeat(config)
	return agent
}
