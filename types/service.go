package types

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	sqlxTypes "github.com/jmoiron/sqlx/types"
)

type ServiceGetOptions struct {
	ID   int
	Name string
}

type ServiceService interface {
	DockerClient() *client.Client
	ServiceCreate(ctx context.Context, target *Service) error
	ServiceGetByID(ctx context.Context, id string) (*Service, error)
	ServiceRawByID(ctx context.Context, id string) (*swarm.Service, error)
	ServiceLogsByID(ctx context.Context, id string) (string, error)
	ServiceGetByName(ctx context.Context, name string) (*Service, error)
	ServiceDelete(ctx context.Context, id string) error
	ServiceUpdate(ctx context.Context, target *Service) error
	ServiceStop(ctx context.Context, id string) error
	Redeploy(ctx context.Context, serviceName string) error
	List(ctx context.Context, opts ServiceFilterOptions) ([]*Service, error)
}

type ServiceRepository interface {
	Insert(ctx context.Context, target *Service) error
	Update(ctx context.Context, target *Service) error
	Delete(ctx context.Context, id string) error
	FindOne(ctx context.Context, opts ServiceFilterOptions) (*Service, error)
	Find(ctx context.Context, opts ServiceFilterOptions) ([]*Service, error)
}

type ServiceSpec struct {
	Image        string          `json:"image" db:"-" bson:"image"`
	Ports        []PortInfo      `json:"ports" db:"-" bson:"ports"`
	Volumes      []VolumeInfo    `json:"volumes" db:"-" bson:"volumes"`
	Environments []string        `json:"environments" db:"-" bson:"environments"`
	Configs      []ServiceConfig `json:"configs" db:"-" bson:"configs"`
	Networks     []string        `json:"networks" db:"-" bson:"networks"`
	Deploy       Deploy          `json:"deploy" db:"-" bson:"deploy"`
}

type Service struct {
	ID               string             `json:"id" db:"id" bson:"_id"`
	ClusterID        string             `json:"cluster_id" db:"cluster_id" bson:"cluster_id"`
	Name             string             `json:"name" db:"name" bson:"name"`
	Spec             ServiceSpec        `json:"spec" db:"-" bson:"spec"`
	SpecJSON         sqlxTypes.JSONText `json:"-" db:"specJSON" bson:"-"`
	DeploymentStatus DeploymentStatus   `json:"deployment_status" db:"-" bson:"-"`
	CreatedAt        *time.Time         `json:"created_at" db:"created_at" bson:"created_at"`
	UpdatedAt        *time.Time         `json:"updated_at" db:"updated_at" bson:"updated_at"`
}

// DeploymentStatus stores the information about mode and replicas to be used by template
type DeploymentStatus struct {
	ServiceName       string `json:"-"`
	Image             string `json:"image"`
	Mode              string `json:"mode"`
	AvailableReplicas int    `json:"available_replicas"`
	Replicas          int    `json:"replicas"`
	UpdateState       string `json:"update_state"`
}

type PortInfo struct {
	Target    uint32 `json:"target" bson:"target"`
	Published uint32 `json:"published" bson:"plblished"`
	Protocol  string `json:"protocol" bson:"protocol"`
	Mode      string `json:"mode" bson:"mode"`
}

type VolumeInfo struct {
	Type     string `json:"type" bson:"type"`
	Source   string `json:"source" bson:"source"`
	Target   string `json:"target" bson:"target"`
	ReadOnly bool   `json:"read_only" bson:"read_only"`
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

type UpdateConfig struct {
	Order string `json:"order"`
}

type Deploy struct {
	Mode          string        `json:"mode" bson:"mode"`
	Replicas      uint64        `json:"replicas" bson:"replicas"`
	UpdateConfig  UpdateConfig  `json:"update_config"`
	RestartPolicy RestartPolicy `json:"restart_policy" bson:"restart_policy"`
	Constraints   []string      `json:"constraints" bson:"constraints"`
}

type ServiceConfig struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type ServiceFilterOptions struct {
	ClusterID   string
	ServiceID   string
	ServiceName string
}

type ServiceLogResult struct {
	Logs string `json:"logs"`
}
