package dkstat

import (
	"github.com/fsouza/go-dockerclient"
	"github.com/itchenyi/eye-depend/gpool"
	"fmt"
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
			fmt.Println(key)
			if obj, isExist := container.NetworkSettings.Networks[key]; isExist {
				if obj.IPAddress != "" {
					return obj.IPAddress
				}

				continue
			}
		}

		return ""
	}

	return &Monitor{
		client:   client,
		id:       container.ID,
		ip:		  getContainerIP(),
		app:      container.Config.Labels["SRV_NAME"],
		task:     container.Config.Labels["MESOS_TASK_ID"],
	}, nil
}

func (monitor *Monitor) handle(pool *gpool.Pool, statsChan chan<- Stats) error {
	dockerStatsChan := make(chan *docker.Stats)

	go func() {
		pool.JobDone()
		stats := <-dockerStatsChan

		statsChan <- Stats{
			ID:    monitor.id,
			IP:	   monitor.ip,
			App:   monitor.app,
			Task:  monitor.task,
			Stats: *stats,
		}
	}()

	return monitor.client.Stats(docker.StatsOptions{
		ID:     monitor.id,
		Stats:  dockerStatsChan,
		Stream: false,
	})
}
