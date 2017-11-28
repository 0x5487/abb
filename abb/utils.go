package abb

import (
	"fmt"

	"github.com/docker/docker/api/types/swarm"
	"github.com/jasonsoft/abb/types"
	"github.com/nlopes/slack"
)

func getServicesStatus(services []swarm.Service, nodes []swarm.Node, tasks []swarm.Task) map[string]types.DeploymentStatus {
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

	info := map[string]types.DeploymentStatus{}
	for _, service := range services {
		if service.Spec.Mode.Replicated != nil && service.Spec.Mode.Replicated.Replicas != nil {
			deploymentStatus := types.DeploymentStatus{
				ServiceName:       service.Spec.Name,
				Image:             service.Spec.TaskTemplate.ContainerSpec.Image,
				Mode:              "replicated",
				AvailableReplicas: running[service.ID],
				Replicas:          (int)(*service.Spec.Mode.Replicated.Replicas),
			}
			if service.UpdateStatus != nil {
				deploymentStatus.UpdateState = string(service.UpdateStatus.State)
			}
			info[service.ID] = deploymentStatus
		} else if service.Spec.Mode.Global != nil {
			deploymentStatus := types.DeploymentStatus{
				ServiceName:       service.Spec.Name,
				Image:             service.Spec.TaskTemplate.ContainerSpec.Image,
				Mode:              "global",
				AvailableReplicas: running[service.ID],
				Replicas:          tasksNoShutdown[service.ID],
			}
			if service.UpdateStatus != nil {
				deploymentStatus.UpdateState = string(service.UpdateStatus.State)
			}
			info[service.ID] = deploymentStatus
		}
	}
	return info
}

func GetGroupIDByName(api *slack.RTM) map[string]string {
	result := map[string]string{}
	groups, err := api.GetGroups(false)
	if err != nil {
		fmt.Printf("%s\n", err)
		return nil
	}
	for _, group := range groups {
		fmt.Printf("ID: %s, Name: %s\n", group.ID, group.Name)
		result[group.Name] = group.ID
	}

	return result
}
