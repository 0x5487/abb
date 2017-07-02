package abb

import (
	"github.com/docker/docker/api/types/swarm"
	"github.com/jasonsoft/abb/types"
)

// GetServicesStatus returns a map of mode and replicas
func GetServicesStatus(services []swarm.Service, nodes []swarm.Node, tasks []swarm.Task) map[string]types.ServiceStatus {
	running := map[string]int{}
	tasksNoShutdown := map[string]int{}

	activeNodes := make(map[string]struct{})
	for _, n := range nodes {
		if n.Status.State != swarm.NodeStateDown {
			activeNodes[n.ID] = struct{}{}
		}
	}

	for _, task := range tasks {
		if task.DesiredState != swarm.TaskStateShutdown {
			tasksNoShutdown[task.ServiceID]++
		}

		if _, nodeActive := activeNodes[task.NodeID]; nodeActive && task.Status.State == swarm.TaskStateRunning {
			running[task.ServiceID]++
		}
	}

	info := map[string]types.ServiceStatus{}
	for _, service := range services {
		info[service.ID] = types.ServiceStatus{}
		if service.Spec.Mode.Replicated != nil && service.Spec.Mode.Replicated.Replicas != nil {
			info[service.ID] = types.ServiceStatus{
				Mode:              "replicated",
				AvailableReplicas: running[service.ID],
				Replicas:          (int)(*service.Spec.Mode.Replicated.Replicas),
			}
		} else if service.Spec.Mode.Global != nil {
			info[service.ID] = types.ServiceStatus{
				Mode:              "global",
				AvailableReplicas: running[service.ID],
				Replicas:          tasksNoShutdown[service.ID],
			}
		}
	}
	return info
}
