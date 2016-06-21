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
var configManager *ConfigManager

type configInfo struct {
	EtcdEndpoints []string
	IP            string
	Hostname      string
	Cluster       string
}

type configManagerInfo struct {
	EtcdEndpoints []string
}

type Config struct {
	*configInfo
	Etcd client.Client
}

type ConfigManager struct {
	Etcd client.Client
}

func GetConfig() *Config {
	if config == nil {
		log.Errorf("Error: config file not load")
	}
	return config
}

func GetConfigManager() *ConfigManager {
	if configManager == nil {
		log.Errorf("Error: config file not load")
	}
	return configManager
}

func LoadConfigManager(filename string) *ConfigManager {
	file, err := os.Open(filename)
	if err != nil {
		log.Errorf("Error: open config file %s %s\n", filename, err)
	}
	info := readConfigInfo(file)

	cfg := client.Config{
		Endpoints:               info.EtcdEndpoints,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}

	etcdClient, err := client.New(cfg)
	if err != nil {
		log.Errorf("Error: cannot connect to etcd: %s", err)
	}

	configManager = &ConfigManager{
		Etcd: etcdClient,
	}

	return configManager
}

func LoadConfig(filename string) *Config {
	file, err := os.Open(filename)
	if err != nil {
		log.Errorf("Error: open config file %s %s\n", filename, err)
	}
	info := readConfigInfo(file)

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

func readConfigInfo(reader io.Reader) *configInfo {
	configInfo := &configInfo{}
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(configInfo)
	if err != nil {
		log.Errorf("Error: decode config from reader %s", err)
	}

	return configInfo
}
