package abb

import (
	"context"
	"fmt"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/jasonsoft/abb/app"
	"github.com/jasonsoft/abb/types"
	"github.com/jasonsoft/log"
)

type ServiceManager struct {
	client  *client.Client
	cluster *types.Cluster
}

func NewServiceManager(cluster *types.Cluster) (types.ServiceService, error) {
	client, err := client.NewClient(cluster.Host, "v1.30", nil, nil)
	if err != nil {
		return nil, err
	}

	return &ServiceManager{
		client:  client,
		cluster: cluster,
	}, nil
}

func (m *ServiceManager) DockerClient() *client.Client {
	return m.client
}

func (m *ServiceManager) ServiceGet(ctx context.Context, opt types.ServiceGetOptions) (*swarm.Service, error) {
	serviceInspectOptions := dockerTypes.ServiceInspectOptions{}
	svc, _, err := m.client.ServiceInspectWithRaw(ctx, opt.ServiceID, serviceInspectOptions)
	if err != nil {
		if client.IsErrServiceNotFound(err) {
			return nil, app.AppError{
				ErrorCode: "not_found",
				Message:   fmt.Sprintf("service:'%s' was not found", opt.ServiceID),
			}
		}
		log.Errorf("abb: get service error: %v", err)
		return nil, err
	}
	return &svc, nil
}
