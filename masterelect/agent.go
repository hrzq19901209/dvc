package main

import (
	"bughunter.com/dvc/config"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"time"
)

var conf *config.Config

func init() {
	log.SetLevel(log.InfoLevel)
	conf = config.LoadConfig("agent.conf")
}

type AgentInfo struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	Cluster  string `json:"cluster"`
}

func GetMaster(info *AgentInfo) {
	api := client.NewKeysAPI(conf.Etcd)
	value, _ := json.Marshal(info)
	key := "mangers/master"
	for {
		_, err := api.Set(context.Background(), key, string(value), &client.SetOptions{ //尝试去创建master,试图成为主节点
			PrevExist: client.PrevNoExist,
			TTL:       time.Second * 30,
		})

		if err != nil {
			log.Errorf("Faile to be master%s", err) //master节点可能健康
		} else {
			log.Info("Change to master!")
			break //成为master，退出此循环，进入heartbeat状态
		}
		time.Sleep(time.Second * 15) //尝试失败，15秒后重新尝试
	}
	for {
		_, err := api.Set(context.Background(), key, string(value), &client.SetOptions{ //尝试去创建master,试图成为主节点
			PrevValue: string(value),
			TTL:       time.Second * 30,
		})

		if err != nil {
			panic(err)
		}
		log.Info("Update successfully!I am %s", info.Hostname)
		time.Sleep(time.Second * 15) //尝试失败，15秒后重新尝试
	}
}

func main() {
	info := &AgentInfo{
		IP:       conf.IP,
		Hostname: conf.Hostname,
		Cluster:  conf.Cluster,
	}
	GetMaster(info)
}
