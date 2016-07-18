package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gogo/protobuf/proto"

	mesos "github.com/mesos/mesos-go/mesosproto"
	util "github.com/mesos/mesos-go/mesosutil"
	sched "github.com/mesos/mesos-go/scheduler"
)

// The docker images that we launch
const (
	NginxImage = "nginx"
)

// Resource usage of the tasks
const (
	MemPerDaemonTask = 128 // mining shouldn't be memory-intensive
	MemPerServerTask = 256 // I'm just guessing
	CPUPerServerTask = 1   // a miner server does not use much CPU
)

// minerScheduler implements the Scheduler interface and stores the state
// needed to scheduler tasks.
type minerScheduler struct {
	tasksLaunched int
}

func newMinerScheduler() *minerScheduler {
	return &minerScheduler{
		tasksLaunched: 0,
	}
}

func (s *minerScheduler) Registered(_ sched.SchedulerDriver, frameworkID *mesos.FrameworkID, masterInfo *mesos.MasterInfo) {
	log.Infoln("Framework registered with Master ", masterInfo)
}

func (s *minerScheduler) Reregistered(_ sched.SchedulerDriver, masterInfo *mesos.MasterInfo) {
	log.Infoln("Framework Re-Registered with Master ", masterInfo)
}

func (s *minerScheduler) Disconnected(sched.SchedulerDriver) {
	log.Infoln("Framework disconnected with Master")
}

func (s *minerScheduler) ResourceOffers(driver sched.SchedulerDriver, offers []*mesos.Offer) {
	if s.tasksLaunched > 6 {
		log.Infoln("completed........")
		return
	}
	var offerIds []*mesos.OfferID
	var tasks []*mesos.TaskInfo
	for _, offer := range offers {
		memResources := util.FilterResources(offer.Resources, func(res *mesos.Resource) bool {
			return res.GetName() == "mem"
		})
		mems := 0.0
		for _, res := range memResources {
			mems += res.GetScalar().GetValue()
		}

		cpuResources := util.FilterResources(offer.Resources, func(res *mesos.Resource) bool {
			return res.GetName() == "cpus"
		})
		cpus := 0.0
		for _, res := range cpuResources {
			cpus += res.GetScalar().GetValue()
		}

		portsResources := util.FilterResources(offer.Resources, func(res *mesos.Resource) bool {
			return res.GetName() == "ports"
		})

		var ports uint64
		for _, res := range portsResources {
			portRanges := res.GetRanges().GetRange()
			for _, portRange := range portRanges {
				ports += portRange.GetEnd() - portRange.GetBegin()
			}
		}

		cluster := offer.GetHostname()

		log.Infof("cluster:%s, mems: %f, cpus: %f\n, ports: %v", cluster, mems, cpus, ports)
		// If a miner server is running, we start a new miner daemon.  Otherwise, we start a new miner server.
		total := 0
		for total < 3 && strings.Compare(cluster, "nanjing") == 0 && mems >= MemPerServerTask && cpus >= CPUPerServerTask {
			var taskID *mesos.TaskID
			var task *mesos.TaskInfo

			s.tasksLaunched++
			taskID = &mesos.TaskID{
				Value: proto.String("miner-server-" + strconv.Itoa(s.tasksLaunched)),
			}

			containerType := mesos.ContainerInfo_DOCKER
			network := mesos.ContainerInfo_DockerInfo_BRIDGE

			task = &mesos.TaskInfo{
				Name:    proto.String("task-" + taskID.GetValue()),
				TaskId:  taskID,
				SlaveId: offer.SlaveId,
				Container: &mesos.ContainerInfo{
					Type: &containerType,
					Docker: &mesos.ContainerInfo_DockerInfo{
						Image:   proto.String(NginxImage),
						Network: &network,
					},
				},
				Command: &mesos.CommandInfo{
					Shell: proto.Bool(false),
				},
				Resources: []*mesos.Resource{
					util.NewScalarResource("cpus", CPUPerServerTask),
					util.NewScalarResource("mem", MemPerServerTask),
				},
			}

			cpus -= CPUPerServerTask
			mems -= MemPerServerTask
			total++
			// update state
			tasks = append(tasks, task)
		}
		offerIds = append(offerIds, offer.Id)

	}
	driver.LaunchTasks(offerIds, tasks, &mesos.Filters{RefuseSeconds: proto.Float64(1)})
}

func (s *minerScheduler) StatusUpdate(driver sched.SchedulerDriver, status *mesos.TaskStatus) {
	log.Infoln("Status update: task", status.TaskId.GetValue(), " is in state ", status.State.Enum().String())
	log.Infoln(status.GetMessage())
	// If the mining server failed for any reason, kill all daemons, since they will be trying to talk to the failed mining server
	if strings.Contains(status.GetTaskId().GetValue(), "server") &&
		(status.GetState() == mesos.TaskState_TASK_LOST ||
			status.GetState() == mesos.TaskState_TASK_KILLED ||
			status.GetState() == mesos.TaskState_TASK_FINISHED ||
			status.GetState() == mesos.TaskState_TASK_ERROR ||
			status.GetState() == mesos.TaskState_TASK_FAILED) {

	}
}

func (s *minerScheduler) OfferRescinded(_ sched.SchedulerDriver, offerID *mesos.OfferID) {
	log.Printf("Offer rescinded: %s", offerID)
}

func (s *minerScheduler) FrameworkMessage(_ sched.SchedulerDriver, executorID *mesos.ExecutorID, slaveID *mesos.SlaveID, message string) {
	log.Printf("Received framework message from %s %s: %s", executorID, slaveID, message)
}

func (s *minerScheduler) SlaveLost(_ sched.SchedulerDriver, slaveID *mesos.SlaveID) {
	log.Printf("Slave lost: %s", slaveID)
}

func (s *minerScheduler) ExecutorLost(_ sched.SchedulerDriver, executorID *mesos.ExecutorID, slaveID *mesos.SlaveID, _ int) {
	log.Printf("Executor lost: %s %s", executorID, slaveID)
}

func (s *minerScheduler) Error(driver sched.SchedulerDriver, err string) {
	log.Printf("Error: %s", err)
}

func printUsage() {
	fmt.Println(`
Usage: scheduler [--FLAGS] [RPC username] [RPC password]
Your RPC username and password can be found in your bitcoin.conf file.
To see a detailed description of the flags available, type "scheduler --help"
`)
}

func main() {

	// flags
	master := flag.String("master", "10.16.51.127:5050", "Master address <ip:port>")
	// auth
	flag.Parse()

	fwinfo := &mesos.FrameworkInfo{
		User: proto.String(""),
		Name: proto.String("BTC Mining Framework (Go)"),
	}

	config := sched.DriverConfig{
		Scheduler:  newMinerScheduler(),
		Framework:  fwinfo,
		Master:     *master,
		Credential: (*mesos.Credential)(nil),
	}

	driver, err := sched.NewMesosSchedulerDriver(config)
	if err != nil {
		log.Errorln("Unable to create a SchedulerDriver ", err.Error())
	}

	if stat, err := driver.Run(); err != nil {
		log.Infof("Framework stopped with status %s and error: %s", stat.String(), err.Error())
	}
}
