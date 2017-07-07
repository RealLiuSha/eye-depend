package dkstat

import "github.com/fsouza/go-dockerclient"

// Stats represents singe stat from docker stats api for specific task
type Stats struct {
	ID    string
	IP    string
	App   string
	Task  string
	Stats docker.Stats
}

