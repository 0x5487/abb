package types

import (
	"time"

	"github.com/docker/docker/api/types/swarm"
)

type PortInfo struct {
	Target    uint32                      `json:"target"`
	Published uint32                      `json:"published"`
	Protocol  swarm.PortConfigProtocol    `json:"protocol"`
	Mode      swarm.PortConfigPublishMode `json:"mode"`
}

type VolumeInfo struct {
	Type   string
	Source string
	Target string
}

type Placement struct {
	Constraints map[string]string
}

type RestartPolicy struct {
	Condition   string
	Delay       time.Duration
	MaxAttempts uint
	Window      time.Duration
}

type Deploy struct {
	Mode      string
	Replicas  uint
	Placement Placement
}

type Service struct {
	Version      string        `json:"version"`
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Image        string        `json:"image"`
	Ports        []*PortInfo   `json:"ports"`
	Volumes      []*VolumeInfo `json:"volumes"`
	Environments []string      `json:"environments"`
	Networks     []string      `json:"networks"`
	Deploy       Deploy        `json:"deploy"`
	Age          time.Duration `json:"age"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}
