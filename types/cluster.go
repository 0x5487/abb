package types

import (
	"context"
	"time"
)

type Cluster struct {
	ID        string    `json:"id" bson:"_id"`
	Name      string    `json:"name" bson:"name"`
	Host      string    `json:"host" bson:"host"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type ClusterRepository interface {
	ClusterCreate(ctx context.Context, target *Cluster) error
	ClusterUpdate(ctx context.Context, target *Cluster) error
	ClusterList(ctx context.Context) ([]*Cluster, error)
	ClusterByName(ctx context.Context, name string) (*Cluster, error)
}

type ClusterService interface {
	ClusterCreate(ctx context.Context, target *Cluster) error
	ClusterUpdate(ctx context.Context, target *Cluster) error
	ClusterList(ctx context.Context) ([]*Cluster, error)
	ClusterByName(ctx context.Context, name string) (*Cluster, error)
}
