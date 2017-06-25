package abb

import (
	"github.com/docker/docker/client"
)

type Manager struct {
	client *client.Client
}

func NewManager(address string) (*Manager, error) {
	client, err := client.NewClient(address, "v1.27", nil, nil)
	if err != nil {
		return nil, err
	}

	return &Manager{
		client: client,
	}, nil
}

func (m *Manager) ServiceList() *[]Service {
	return nil
}

func (m *Manager) ServiceGet() {

}

func (m *Manager) ServiceCreate(svc *Service) *Service {
	// save service to database
	return nil
}

func (m *Manager) ServiceDelete(id int) {

}

func (m *Manager) ServiceUpdate() {

}
