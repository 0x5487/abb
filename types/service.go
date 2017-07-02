package types

import "github.com/docker/docker/api/types/swarm"

type Service struct {
	swarm.Service
	Status ServiceStatus `json:"status"`
}

// ServiceStatus stores the information about mode and replicas to be used by template
type ServiceStatus struct {
	Mode              string `json:"mode"`
	AvailableReplicas int    `json:"available_replicas"`
	Replicas          int    `json:"replicas"`
}
