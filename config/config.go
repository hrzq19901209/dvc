package config

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/client"
	"io"
	"os"
	"time"
)

var config *Config

type configInfo struct {
	EtcdEndpoints []string
	IP            string
	Hostname      string
	Cluster       string
}

type Config struct {
	*configInfo
	Etcd client.Client
}

func GetConfig() *Config {
	if config == nil {
		log.Errorf("Error: config file not load")
	}
	return config
}

func LoadConfig(filename string) *Config {
	file, err := os.Open(filename)
	if err != nil {
		log.Errorf("Error: open config file %s %s\n", filename, err)
	}
	info := readconfigInfo(file)

	cfg := client.Config{
		Endpoints:               info.EtcdEndpoints,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}

	etcdClient, err := client.New(cfg)
	if err != nil {
		log.Errorf("Error: cannot connect to etcd: %s", err)
	}

	config = &Config{
		configInfo: info,
		Etcd:       etcdClient,
	}

	return config
}

func readconfigInfo(reader io.Reader) *configInfo {
	configInfo := &configInfo{}
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(configInfo)
	if err != nil {
		log.Errorf("Error: decode config from reader %s", err)
	}

	return configInfo
}
