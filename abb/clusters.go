package abb

import (
	"context"
	"database/sql"
	"database/sql/driver"

	"github.com/jasonsoft/abb/types"
	xlog "github.com/jasonsoft/log"
	"github.com/jmoiron/sqlx"
)

// ************************
// Business
// ************************

type ClusterManager struct {
	repo types.ClusterRepository
}

func NewClusterManager(repo types.ClusterRepository) types.ClusterService {
	return &ClusterManager{
		repo: repo,
	}
}

func (manager *ClusterManager) ClusterList(ctx context.Context) ([]*types.Cluster, error) {
	clusters, err := manager.repo.ClusterList(ctx)
	if err != nil {
		return nil, err
	}

	return clusters, nil
}

func (manager *ClusterManager) ClusterByName(ctx context.Context, name string) (*types.Cluster, error) {
	cluster, err := manager.repo.ClusterByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return cluster, nil
}

// ************************
// Database
// ************************

type ClusterDatabase struct {
	db *sqlx.DB
}

func NewClusterDatabase(db *sqlx.DB) types.ClusterRepository {
	return &ClusterDatabase{
		db: db,
	}
}

const clusterListSQL = "SELECT `id`, `name`, `host`, `created_at`, `updated_at` FROM clusters"

func (c *ClusterDatabase) ClusterList(ctx context.Context) ([]*types.Cluster, error) {
	log := xlog.FromContext(ctx)

	var clusters []*types.Cluster
	err := c.db.Select(&clusters, clusterListSQL)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Errorf("abb: get all clusters fail: %v", err)
		return nil, err
	}

	return clusters, nil
}

const clusterByNameSQL = "SELECT `id`, `name`, `host`, `created_at`, `updated_at` FROM clusters where (`name`= :name);"

func (c *ClusterDatabase) ClusterByName(ctx context.Context, name string) (*types.Cluster, error) {
	log := xlog.FromContext(ctx)

	clusterByNameSQLStmt, err := c.db.PrepareNamed(clusterByNameSQL)
	if err != nil {
		log.Errorf("abb: prepare sql fail: %v", err)
		return nil, err
	}
	defer clusterByNameSQLStmt.Close()

	m := map[string]interface{}{
		"name": name,
	}

	cluster := types.Cluster{}

	for i := 0; i < 10; i++ {
		err = clusterByNameSQLStmt.Get(&cluster, m)
		if err == driver.ErrBadConn {
			continue
		}
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			log.Errorf("abb: get request fail: %v", err)
			return nil, err
		}
		break
	}

	return &cluster, nil
}

// type GetServiceListOptions struct {
// 	Page    int
// 	PerPage int
// }

// func (c *Cluster) ServiceList(ctx context.Context, opt GetServiceListOptions) ([]*types.Service, error) {
// 	svcOpts := dockerTypes.ServiceListOptions{}
// 	dockerServiceList, err := m.client.ServiceList(ctx, svcOpts)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var services []*types.Service
// 	for _, dockerService := range dockerServiceList {
// 		svc := convertToService(&dockerService)
// 		services = append(services, svc)
// 	}

// 	return services, nil
// }

// func (c *Cluster) ServiceCreate(ctx context.Context, svc *types.Service) (*types.Service, error) {
// 	dockerSvc := convertToDockerService(svc)
// 	createOptions := dockerTypes.ServiceCreateOptions{}
// 	svcResp, err := m.client.ServiceCreate(ctx, dockerSvc.Spec, createOptions)
// 	if err != nil {
// 		return nil, err
// 	}

// 	result, err := m.ServiceGet(ctx, svcResp.ID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return result, nil
// }

// func (c *Cluster) ServiceDelete(id int) {

// }

// func (c *Cluster) ServiceUpdate() {

// }
