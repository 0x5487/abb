package types

import (
	"context"
	"time"
)

type Cluster struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Host      string    `json:"host" db:"host"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type ClusterRepository interface {
	ClusterList(ctx context.Context) ([]*Cluster, error)
	ClusterByName(ctx context.Context, name string) (*Cluster, error)
}

type ClusterService interface {
	ClusterList(ctx context.Context) ([]*Cluster, error)
	ClusterByName(ctx context.Context, name string) (*Cluster, error)
}
