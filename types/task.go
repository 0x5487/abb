package types

import (
	"context"
	"time"

	"github.com/docker/docker/client"
)

type TaskStatus struct {
	TimeStamp time.Time `json:"timestamp"`
	State     string    `json:"state"`
	Message   string    `json:"message"`
}

type Task struct {
	ID     string     `json:"id"`
	Node   string     `json:"node"`
	Slot   int        `json:"slot"`
	Status TaskStatus `json:"status"`
}

type TaskListOption struct {
	ServiceID    string
	DesiredState string
}

type TaskService interface {
	DockerClient() *client.Client
	List(ctx context.Context, opts TaskListOption) ([]Task, error)
	Close(ctx context.Context) error
}
