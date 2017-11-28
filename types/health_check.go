package types

import (
	"context"
	"time"
)

type HealthCheck struct {
	ID        string     `json:"id" db:"id" bson:"_id"`
	ClusterID string     `json:"cluster_id" db:"cluster_id" bson:"cluster_id"`
	Name      string     `json:"name" db:"name" bson:"name"`
	URL       string     `json:"url" db:"url"`
	Interval  int        `json:"interval" db:"interval"`
	Timeout   int        `json:"timeout" db:"timeout"`
	Retries   int        `json:"retries" db:"retries"`
	IsEnabled int        `json:"is_enabled" db:"is_enabled"`
	IsHealth  bool       `json:"-" db:"-"`
	CreatedAt *time.Time `json:"created_at" db:"created_at" bson:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at" bson:"updated_at"`
}

type HealthChecker interface {
	Create(ctx context.Context, entity *HealthCheck) error
	List(ctx context.Context, opts HealthCheckFilterOptions) ([]*HealthCheck, error)
}

type HealthCheckFilterOptions struct {
	ID        string
	ClusterID string
	Name      string
	IsEnabled int
}

type HealthCheckerRepository interface {
	Insert(ctx context.Context, target *HealthCheck) error
	Update(ctx context.Context, target *HealthCheck) error
	Delete(ctx context.Context, id string) error
	FindOne(ctx context.Context, opts HealthCheckFilterOptions) (*HealthCheck, error)
	Find(ctx context.Context, opts HealthCheckFilterOptions) ([]*HealthCheck, error)
}
