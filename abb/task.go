package abb

import (
	"context"
	"fmt"
	"strings"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/jasonsoft/abb/app"
	"github.com/jasonsoft/abb/types"
	"github.com/jasonsoft/log"
)

type TaskManager struct {
	client  *client.Client
	cluster *types.Cluster
}

func newTaskManager(cluster *types.Cluster) (types.TaskService, error) {
	client, err := client.NewClient(cluster.Host, "v1.30", nil, nil)
	if err != nil {
		return nil, err
	}

	return &TaskManager{
		client:  client,
		cluster: cluster,
	}, nil
}

func newTaskFromSwarmTask(target swarm.Task) types.Task {
	return types.Task{
		ID:   target.ID,
		Name: target.Name,
		Status: types.TaskStatus{
			TimeStamp: target.Status.Timestamp,
			Message:   target.Status.Message,
			State:     string(target.Status.State),
		},
	}
}

func (m *TaskManager) DockerClient() *client.Client {
	return m.client
}

func (m *TaskManager) List(ctx context.Context, opts types.TaskListOption) ([]types.Task, error) {
	logger := log.FromContext(ctx)

	// get all docker services
	dockerServiceListOpts := dockerTypes.ServiceListOptions{}
	dockerSvcList, err := m.client.ServiceList(ctx, dockerServiceListOpts)
	if err != nil {
		panic(err)
	}

	// get task per service
	filterArgs := filters.NewArgs()
	filterArgs.Add("desired-state", "running")

	if len(opts.ServiceID) > 0 {
		filterArgs.Add("service", opts.ServiceID)
	}

	taskListOpt := dockerTypes.TaskListOptions{
		Filters: filterArgs,
	}
	taskList, err := m.client.TaskList(ctx, taskListOpt)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, app.AppError{
				ErrorCode: "not_found",
				Message:   "task was not found",
			}
		}
		logger.Errorf("abb: list task fail: %#v", err)
		return nil, err
	}

	result := []types.Task{}

	for _, task := range taskList {
		newTask := newTaskFromSwarmTask(task)

		for _, svc := range dockerSvcList {
			if svc.ID == task.ServiceID {
				if task.Slot != 0 {
					newTask.Name = fmt.Sprintf("%v.%v", svc.Spec.Name, task.Slot)
				} else {
					newTask.Name = fmt.Sprintf("%v.%v", svc.Spec.Name, task.NodeID) // Global mode
				}
			}
		}

		result = append(result, newTask)
	}

	return result, nil
}

func (m *TaskManager) Close(ctx context.Context) error {
	return m.client.Close()
}
