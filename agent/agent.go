package main

import (
	"bughunter.com/dvc/config"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/client"
	dockerclient "github.com/docker/engine-api/client"
	"github.com/mikespook/golib/signal"
	"golang.org/x/net/context"
	"net/rpc"
	"os"
	"time"
)

type AgentInfo struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	Cluster  string `json:"cluster"`
}

type Agent struct {
	Cli dockerclient
}

type Reply struct {
	OK  string
	Msg string
}

func HeartBeat(config *config.Config) {
	api := client.NewKeysAPI(config.Etcd)

	for {
		key := fmt.Sprintf("agents/%s", config.Hostname)

		info := &AgentInfo{
			IP:       config.IP,
			Hostname: config.Hostname,
			Cluster:  config.Cluster,
		}
		value, _ := json.Marshal(info)
		_, err := api.Set(context.Background(), key, string(value), &client.SetOptions{
			TTL: time.Second * 30,
		})

		if err != nil {
			log.Errorf("Warning: update workerInfo %s", err)
		}
		time.Sleep(time.Second * 3)
	}
}

func NewAgent() *Agent {
	config := config.GetConfig()
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.12", nil, defaultHeaders)
	if err != nil {
		panic(err)
	}
	agent := &Agent{
		Cli: cli,
	}
	go HeartBeat(config)
	return agent
}

func (agent *Agent) ListenRPC() {
	rpc.Register(NewAgent())
	rpc.HandleHTTP()
	conn, err := net.Listen("tcp", ":4200")
	if err != nil {
		log.Errorf("Error: listen 4200 error", err)
	}
	go http.Serve(conn, nil)
}

func (agent *Agent) BuildImage(buildContext io.Reader, reply *Reply) {

}

func main() {
	config.LoadConfig("agent.conf")
	agent := NewAgent()
	agent.ListenRPC()
	signal.Bind(os.Interrupt, func() uint { return signal.BreakExit })
	signal.Wait()
}
