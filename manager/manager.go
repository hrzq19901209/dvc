package manager

import (
	"bughunter.com/dvc/agent"
	"bughunter.com/dvc/config"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/content"
	"time"
)

type Status int

const (
	Active Status = iota
	Down
)

var statusMap = map[Status]string{
	Active: "Active",
	Down:   "Down",
}

func (s Status) String() string {
	return statusMap[s]
}

type Member struct {
	IP       string
	Hostname string
	Cluster  string
	Status
}
type Manager struct {
	members map[string]*Member
}

func (m *Manager) addWorker(info *agent.AgentInfo) {
	member := &Member{
		IP:       info.IP,
		Hostname: info.Hostname,
		Cluster:  info.Cluster,
		Status:   Active,
	}

	m.members[member.Hostname] = member
}

func (m *Manager) updateAgent(info *agent.AgentInfo) {
	member := m.members[m.Hostname]
	member.IP = info.IP
	member.Hostname = info.Hostname
	member.Cluster = info.Cluster
	member.Status = Active
}

func (m *Manager) nodeToAgentInfo(node *client.Node) *agent.AgentInfo {
	info := &agent.AgentInfo{}
	err := json.Unmarshal([]byte(node.Value), info)
	if err != nil {
		log.Errorf(err)
	}
	return info
}

func (m *Manager) watchAgents() {
	config := config.GetConfigManager()
	KAPI := client.NewKeysAPI(config.Etcd)
	watcher := KAPI.Watcher("agents/", &client.WatcherOptions{
		Recursive: true,
	})
	for {
		res, err := watcher.Next(context.Background())
		if err != nil {
			log.Errorf("Warn: watch agents:%s", err.Error())
			time.Sleep(time.Second * 3)
			continue
		}
		if res.Action == "expire" {
			info := m.nodeToAgentInfo(res.PrevNode)
			member, ok := m.members[info.Hostname]
			if ok {
				member.Status = Down
			}
		} else if res.Action == "set" {
			info := m.nodeToAgentInfo(res.Node)
			if _, ok := m.members[info.Hostname]; ok {
				m.upadteAgent(info)
			} else {
				m.addAgent(info)
			}
		}
	}
}
