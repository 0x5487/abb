package abb

import (
	"context"
	"time"

	"fmt"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/jasonsoft/abb/app"
	"github.com/jasonsoft/log"
)

type ClusterManager struct {
}

func NewClusterManager() *ClusterManager {
	return nil
}

func (manager *ClusterManager) ClusterList(ctx context.Context) ([]*Cluster, error) {
	return nil, nil
}

func (manager *ClusterManager) ClusterByName(ctx context.Context, name string) (*Cluster, error) {
	cluster, err := NewCluster("tcp://10.200.252.123:2376")
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

type Cluster struct {
	Client *client.Client

	Name      string    `json:"name"`
	Host      string    `json:"host"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewCluster(host string) (*Cluster, error) {
	client, err := client.NewClient(host, "v1.30", nil, nil)
	if err != nil {
		return nil, err
	}

	return &Cluster{
		Client: client,
	}, nil
}

// type GetServiceListOptions struct {
// 	Page    int
// 	PerPage int
// }

// func (c *Cluster) ServiceList(ctx context.Context, opt GetServiceListOptions) ([]*types.Service, error) {
// 	svcOpts := dockerTypes.ServiceListOptions{}
// 	dockerServiceList, err := m.client.ServiceList(ctx, svcOpts)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var services []*types.Service
// 	for _, dockerService := range dockerServiceList {
// 		svc := convertToService(&dockerService)
// 		services = append(services, svc)
// 	}

// 	return services, nil
// }

type ServiceGetOptions struct {
	ServiceID string
}

func (c *Cluster) ServiceGet(ctx context.Context, opt ServiceGetOptions) (*swarm.Service, error) {
	serviceInspectOptions := dockerTypes.ServiceInspectOptions{}
	svc, _, err := c.Client.ServiceInspectWithRaw(ctx, opt.ServiceID, serviceInspectOptions)
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

// func (c *Cluster) ServiceCreate(ctx context.Context, svc *types.Service) (*types.Service, error) {
// 	dockerSvc := convertToDockerService(svc)
// 	createOptions := dockerTypes.ServiceCreateOptions{}
// 	svcResp, err := m.client.ServiceCreate(ctx, dockerSvc.Spec, createOptions)
// 	if err != nil {
// 		return nil, err
// 	}

// 	result, err := m.ServiceGet(ctx, svcResp.ID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return result, nil
// }

// func (c *Cluster) ServiceDelete(id int) {

// }

// func (c *Cluster) ServiceUpdate() {

// }
