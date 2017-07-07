package dkstat

import (
	"sync"
	"github.com/fsouza/go-dockerclient"
	"github.com/itchenyi/eye-depend/gpool"
	log "github.com/Sirupsen/logrus"
)

type Collector struct {
	StatsChan  chan Stats
	Length     int
	mutex      sync.Mutex
	client     *docker.Client
	registered map[string]struct{}
}

func (collector *Collector) register(id string) bool {
	collector.mutex.Lock()
	defer collector.mutex.Unlock()

	if _, ok := collector.registered[id]; ok {
		return false
	}

	collector.registered[id] = struct{}{}
	return true
}

func (collector *Collector) unregister(id string) {
	collector.mutex.Lock()
	delete(collector.registered, id)
	collector.mutex.Unlock()
}

func (collector *Collector) handle(pool *gpool.Pool, id string) {
	m, err := NewMonitor(collector.client, id)
	if err != nil {
		log.Errorf("error handling %s: %s\n", id, err)
		return
	}

	go func() {
		if !collector.register(id) {
			return
		}

		err := m.handle(pool, collector.StatsChan)
		if err != nil {
			log.Errorf("error handling container for app %s: %s\n", m.app, err)
		}

		collector.unregister(id)
	}()
}

func NewCollector(client *docker.Client) *Collector {
	statsChan := make(chan Stats)

	return &Collector{
		StatsChan:  statsChan,
		mutex:      sync.Mutex{},
		client:     client,
		registered: map[string]struct{}{},
	}
}

func (collector *Collector) Run() error {
	ch := make(chan *docker.APIEvents)
	err := collector.client.AddEventListener(ch)

	if err != nil {
		return err
	}

	defer func() {
		err = collector.client.RemoveEventListener(ch)
		if err != nil {
			log.Errorf("error handling container for listener: %s", err)
		}

	}()

	containers, err := collector.client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		return err
	}

	collector.Length = len(containers)

	pool := gpool.NewPool(collector.Length, collector.Length)
	defer pool.Release()

	pool.WaitCount(collector.Length)
	for _, container := range containers {
		id := container.ID

		pool.JobQueue <- func() {
			collector.handle(pool, id)
		}
	}

	pool.WaitAll()
	return nil
}
