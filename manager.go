package abb

import (
	"context"

	"github.com/docker/docker/client"
)

type Manager struct {
	client *client.Client
}

func NewManager(host string) (*Manager, error) {
	client, err := client.NewClient(host, "v1.27", nil, nil)
	if err != nil {
		return nil, err
	}

	return &Manager{
		client: client,
	}, nil
}

func (m *Manager) ServiceList(ctx context.Context) (*[]Service, error) {
	return nil, nil
}

func (m *Manager) ServiceGet() {

}

func (m *Manager) ServiceCreate(ctx context.Context, svc *Service) *Service {
	// save service to database
	return nil
}

func (m *Manager) ServiceDelete(id int) {

}

func (m *Manager) ServiceUpdate() {

}
