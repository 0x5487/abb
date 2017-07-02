package abb

import (
	"github.com/docker/docker/api/types/swarm"
	"github.com/jasonsoft/abb/types"
)

func newPortInfo(source swarm.PortConfig) *types.PortInfo {
	port := &types.PortInfo{
		Mode:      source.PublishMode,
		Target:    source.TargetPort,
		Published: source.PublishedPort,
		Protocol:  source.Protocol,
	}

	return port
}

func convertToService(source *swarm.Service) *types.Service {

	ports := []*types.PortInfo{}
	for _, val := range source.Endpoint.Ports {
		port := newPortInfo(val)
		ports = append(ports, port)
	}

	result := types.Service{
		ID:    source.ID,
		Name:  source.Spec.Name,
		Image: source.Spec.TaskTemplate.ContainerSpec.Image,
		Ports: ports,
	}
	return &result
}

func convertToDockerService(source *types.Service) *swarm.Service {
	result := swarm.Service{}
	return &result
}
