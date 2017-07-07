package dkstat

import (
	"github.com/fsouza/go-dockerclient"
	"github.com/itchenyi/eye-depend/gpool"
	"fmt"
	"os/exec"
)

// MonitorDockerClient represents restricted interface for docker client
// that is used in monitor, docker.Client is a subset of this interface
type MonitorClient interface {
	InspectContainer(id string) (*docker.Container, error)
	Stats(opts docker.StatsOptions) error
}

type Monitor struct {
	client   MonitorClient
	id       string
	ip       string
	app      string
	task     string
}

func NewMonitor(client MonitorClient, id string) (*Monitor, error) {
	container, err := client.InspectContainer(id)
	if err != nil {
		return nil, err
	}

	getContainerIP := func () string {
		for key := range container.NetworkSettings.Networks {
			if obj, exists := container.NetworkSettings.Networks[key]; exists {
				return obj.IPAddress
			}
		}

		cmd := exec.Command(
			fmt.Sprintf("docker exec %s ifconfig ", container.ID) +
				"eth0|grep -oP '\\d.+(?=  (Bcast:|netmask))'")

		out, _ := cmd.CombinedOutput()
		return string(out)
	}

	return &Monitor{
		client:   client,
		id:       container.ID,
		ip:		  getContainerIP(),
		app:      container.Config.Labels["SRV_NAME"],
		task:     container.Config.Labels["MESOS_TASK_ID"],
	}, nil
}

func (monitor *Monitor) handle(pool *gpool.Pool, ch chan<- Stats) error {
	statsChan := make(chan *docker.Stats)

	go func() {
		pool.JobDone()
		stats := <-statsChan

		ch <- Stats{
			ID:    monitor.id,
			IP:	   monitor.ip,
			App:   monitor.app,
			Task:  monitor.task,
			Stats: *stats,
		}
	}()

	return monitor.client.Stats(docker.StatsOptions{
		ID:     monitor.id,
		Stats:  statsChan,
		Stream: false,
	})
}
