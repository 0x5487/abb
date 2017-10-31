package abb

import (
	"context"
	"encoding/base64"
	"strings"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/jasonsoft/abb/types"
	"github.com/jasonsoft/log"
)

type ConfigManager struct {
	client  *client.Client
	cluster *types.Cluster
}

func newConfigManager(cluster *types.Cluster) (*ConfigManager, error) {
	client, err := client.NewClient(cluster.Host, "1.30", nil, nil)
	if err != nil {
		return nil, err
	}

	return &ConfigManager{
		client:  client,
		cluster: cluster,
	}, nil
}

func newConfigFromSwarmConfig(config swarm.Config) *types.Config {
	data := base64.StdEncoding.EncodeToString(config.Spec.Data[:])
	return &types.Config{
		ID:        config.ID,
		Name:      config.Spec.Name,
		Data:      data,
		CreatedAt: config.CreatedAt,
	}
}

func (m *ConfigManager) DockerClient() *client.Client {
	return m.client
}

func (m *ConfigManager) Get(ctx context.Context, configID string) (*types.Config, error) {
	logger := log.FromContext(ctx)

	dockerConfig, _, err := m.client.ConfigInspectWithRaw(ctx, configID)
	if err != nil {
		logger.Errorf("abb: get config err: %v", err)
		return nil, err
	}

	cfg := newConfigFromSwarmConfig(dockerConfig)
	return cfg, nil
}

func (m *ConfigManager) List(ctx context.Context, opts types.ConfigListOption) ([]*types.Config, error) {
	logger := log.FromContext(ctx)

	dockerOpts := dockerTypes.ConfigListOptions{}
	dockerConfigs, err := m.client.ConfigList(ctx, dockerOpts)

	if err != nil {
		logger.Errorf("abb: list config err: %v", err)
		return nil, err
	}

	result := []*types.Config{}
	for _, val := range dockerConfigs {
		cfg := newConfigFromSwarmConfig(val)
		result = append(result, cfg)
	}

	return result, nil
}

func (m *ConfigManager) Create(ctx context.Context, config *types.Config) error {
	logger := log.FromContext(ctx)

	config.Name = strings.TrimSpace(config.Name)
	config.Data = strings.TrimSpace(config.Data)

	data, err := base64.StdEncoding.DecodeString(config.Data)
	if err != nil {
		logger.Errorf("abb: decode config err: %v", err)
		return err
	}

	configSpec := swarm.ConfigSpec{
		Data: data,
	}
	configSpec.Name = config.Name

	createResp, err := m.client.ConfigCreate(ctx, configSpec)
	if err != nil {
		logger.Errorf("abb: create config err: %v", err)
		return err
	}

	config.ID = createResp.ID
	return nil
}

func (m *ConfigManager) Delete(ctx context.Context, configID string) error {
	logger := log.FromContext(ctx)

	err := m.client.ConfigRemove(ctx, configID)
	if err != nil {
		logger.Errorf("abb: delete config err: %v", err)
		return err
	}

	return nil
}

func (m *ConfigManager) Close(ctx context.Context) error {
	return m.client.Close()
}
