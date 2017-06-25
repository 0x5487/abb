package abb

import (
	"time"
)

type PortInfo struct {
	Target    uint
	Published uint
	Protocol  string
	Mode      string
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
	Version      string
	ID           int          `json:"id"`
	Name         string       `json:"name"`
	Image        string       `json:"image"`
	Ports        []PortInfo   `json:"ports"`
	Volumes      []VolumeInfo `json:"volumes"`
	Environments []string     `json:"environments"`
	Networks     []string     `json:"networks"`
	Deploy       Deploy       `json:"deploy"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}
