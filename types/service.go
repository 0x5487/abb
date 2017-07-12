package types

import (
	"context"

	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
)

type ServiceGetOptions struct {
	ServiceID string
}

type ServiceService interface {
	DockerClient() *client.Client
	ServiceGet(ctx context.Context, opt ServiceGetOptions) (*swarm.Service, error)
}

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
