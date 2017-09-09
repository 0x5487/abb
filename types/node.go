package types

import (
	"context"
	"time"
)

type Node struct {
	ID        string    `json:"id" bson:"_id"`
	ClusterID string    `json:"cluster_id" bson:"cluster_id"`
	Name      string    `json:"name" bson:"name"`
	Host      string    `json:"host" bson:"host"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type NodeListOptions struct {
}

type NodeRepository interface {
	Insert(ctx context.Context, target *Node) error
	Update(ctx context.Context, target *Node) error
	Delete(ctx context.Context, id string) error
	Find(ctx context.Context, opts NodeListOptions) ([]*Node, error)
}
