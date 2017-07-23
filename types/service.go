package types

import (
	"context"
	"time"

	"github.com/docker/docker/client"
)

type ServiceGetOptions struct {
	ID   int
	Name string
}

type ServiceService interface {
	DockerClient() *client.Client
	ServiceCreate(ctx context.Context, target *Service) error
	ServiceGetByID(ctx context.Context, id string) (*Service, error)
	ServiceGetByName(ctx context.Context, name string) (*Service, error)
	ServiceDelete(ctx context.Context, id string) error
	ServiceUpdate(ctx context.Context, target *Service) error
	ServiceStop(ctx context.Context, id string) error
	Redeploy(ctx context.Context, serviceName string) error
	List(ctx context.Context, opts ServiceListOptions) ([]*Service, error)
}

type ServiceRepository interface {
	FindByID(ctx context.Context, id string) (*Service, error)
	FindByName(ctx context.Context, name string) (*Service, error)
	Insert(ctx context.Context, target *Service) error
	Update(ctx context.Context, target *Service) error
	Delete(ctx context.Context, id string) error
	Find(ctx context.Context, opts ServiceListOptions) ([]*Service, error)
}

type ServiceSpec struct {
	Image        string       `json:"image" db:"-" bson:"image"`
	Ports        []PortInfo   `json:"ports" db:"-" bson:"ports"`
	Volumes      []VolumeInfo `json:"volumes" db:"-" bson:"volumes"`
	Environments []string     `json:"environments" db:"-" bson:"environments"`
	Networks     []string     `json:"networks" db:"-" bson:"networks"`
	Deploy       Deploy       `json:"deploy" db:"-" bson:"deploy"`
}

type Service struct {
	ID        string        `json:"id" db:"id" bson:"_id"`
	ClusterID string        `json:"cluster_id" db:"cluster_id" bson:"cluster_id"`
	Name      string        `json:"name" db:"name" bson:"name"`
	Spec      ServiceSpec   `json:"spec" db:"-" bson:"spec"`
	Status    ServiceStatus `json:"status" db:"-" bson:"-"`
	CreatedAt time.Time     `json:"created_at" db:"created_at" bson:"created_at"`
	UpdatedAt time.Time     `json:"updated_at" db:"updated_at" bson:"updated_at"`
}

// ServiceStatus stores the information about mode and replicas to be used by template
type ServiceStatus struct {
	ServiceName       string `json:"-"`
	Mode              string `json:"mode"`
	AvailableReplicas int    `json:"available_replicas"`
	Replicas          int    `json:"replicas"`
}

type PortInfo struct {
	Target    uint32 `json:"target" bson:"target"`
	Published uint32 `json:"published" bson:"plblished"`
	Protocol  string `json:"protocol" bson:"protocol"`
	Mode      string `json:"mode" bson:"mode"`
}

type VolumeInfo struct {
	Type   string `json:"type" bson:"type"`
	Source string `json:"source" bson:"source"`
	Target string `json:"target" bson:"target"`
}

type Placement struct {
	Constraints map[string]string `json:"name" bson:"constraints"`
}

type RestartPolicy struct {
	Condition   string        `json:"condition" bson:"condition"`
	Delay       time.Duration `json:"delay" bson:"delay"`
	MaxAttempts uint64        `json:"max_attempts" bson:"max_attempts"`
	Window      time.Duration `json:"window" bson:"window"`
}

type Deploy struct {
	Mode          string        `json:"mode" bson:"mode"`
	Replicas      uint64        `json:"replicas" bson:"replicas"`
	RestartPolicy RestartPolicy `json:"restart_policy" bson:"restart_policy"`
	Constraints   []string      `json:"constraints" bson:"constraints"`
}

type ServiceListOptions struct {
}
