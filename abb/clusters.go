package abb

import (
	"context"
	"strings"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/jasonsoft/abb/app"
	"github.com/jasonsoft/abb/types"
	"github.com/jasonsoft/log"
	uuid "github.com/satori/go.uuid"
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

func (manager *ClusterManager) ClusterCreate(ctx context.Context, target *types.Cluster) error {
	err := manager.repo.ClusterCreate(ctx, target)
	if err != nil {
		return err
	}
	return nil
}

func (manager *ClusterManager) ClusterUpdate(ctx context.Context, target *types.Cluster) error {
	err := manager.repo.ClusterUpdate(ctx, target)
	if err != nil {
		return err
	}
	return nil
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

// type ClusterDatabase struct {
// 	db *sqlx.DB
// }

// func NewClusterDatabase(db *sqlx.DB) types.ClusterRepository {
// 	return &ClusterDatabase{
// 		db: db,
// 	}
// }

// const clusterListSQL = "SELECT `id`, `name`, `host`, `created_at`, `updated_at` FROM clusters"

// func (c *ClusterDatabase) ClusterList(ctx context.Context) ([]*types.Cluster, error) {
// 	logger := log.FromContext(ctx)

// 	var clusters []*types.Cluster
// 	err := c.db.Select(&clusters, clusterListSQL)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, nil
// 		}
// 		logger.Errorf("abb: get all clusters fail: %v", err)
// 		return nil, err
// 	}

// 	return clusters, nil
// }

// const clusterByNameSQL = "SELECT `id`, `name`, `host`, `created_at`, `updated_at` FROM clusters where (`name`= :name);"

// func (c *ClusterDatabase) ClusterByName(ctx context.Context, name string) (*types.Cluster, error) {
// 	logger := log.FromContext(ctx)

// 	clusterByNameSQLStmt, err := c.db.PrepareNamed(clusterByNameSQL)
// 	if err != nil {
// 		logger.Errorf("abb: prepare sql fail: %v", err)
// 		return nil, err
// 	}
// 	defer clusterByNameSQLStmt.Close()

// 	m := map[string]interface{}{
// 		"name": name,
// 	}

// 	cluster := types.Cluster{}

// 	for i := 0; i < 10; i++ {
// 		err = clusterByNameSQLStmt.Get(&cluster, m)
// 		if err == driver.ErrBadConn {
// 			continue
// 		}
// 		if err != nil {
// 			if err == sql.ErrNoRows {
// 				return nil, nil
// 			}
// 			logger.Errorf("abb: get request fail: %v", err)
// 			return nil, err
// 		}
// 		break
// 	}

// 	return &cluster, nil
// }

// ************************
// MongoDB
// ************************

type ClusterMongo struct {
}

func NewClusterMongo() (types.ClusterRepository, error) {
	session := _mongoSession.Clone()
	defer session.Close()
	c := session.DB("abb").C("clusters")

	// create index
	nameIdx := mgo.Index{
		Name:       "idx_cluster_name",
		Key:        []string{"name"},
		Background: true,
		Sparse:     true,
		Unique:     true,
	}
	err := c.EnsureIndex(nameIdx)
	if err != nil {
		return nil, err
	}

	return &ClusterMongo{}, nil
}

func (c *ClusterMongo) ClusterCreate(ctx context.Context, target *types.Cluster) error {
	logger := log.FromContext(ctx)

	session := _mongoSession.Clone()
	defer session.Close()

	col := session.DB("abb").C("clusters")
	target.ID = uuid.NewV4().String()
	nowUTC := time.Now().UTC()
	target.CreatedAt = nowUTC
	target.UpdatedAt = nowUTC
	err := col.Insert(target)

	if err != nil {
		if strings.HasPrefix(err.Error(), "E11000") {
			return app.AppError{ErrorCode: "duplicate", Message: "the cluster name already exits"}
		}
		logger.Errorf("abb: cluster create error: %v", err)
		return err
	}
	return nil
}

func (c *ClusterMongo) ClusterUpdate(ctx context.Context, target *types.Cluster) error {
	logger := log.FromContext(ctx)

	if len(target.ID) == 0 {
		return app.AppError{ErrorCode: "invalid_input", Message: "id can't be empty or null."}
	}
	nowUTC := time.Now().UTC()
	target.UpdatedAt = nowUTC

	session := _mongoSession.Clone()
	defer session.Close()

	col := session.DB("abb").C("clusters")
	colQuerier := bson.M{"_id": target.ID}
	err := col.Update(colQuerier, target)
	if err != nil {
		logger.Errorf("abb: cluster update error: %v", err)
		return err
	}
	return nil
}

func (c *ClusterMongo) ClusterList(ctx context.Context) ([]*types.Cluster, error) {
	logger := log.FromContext(ctx)

	session := _mongoSession.Clone()
	defer session.Close()

	var clusters []*types.Cluster
	col := session.DB("abb").C("clusters")

	err := col.Find(bson.M{}).Sort("+created_at").All(&clusters)
	if err != nil {
		if err.Error() == "not found" {
			return nil, nil
		}
		logger.Errorf("abb: clusterlist error: %v", err)
		return nil, err
	}
	return clusters, nil
}

func (c *ClusterMongo) ClusterByName(ctx context.Context, name string) (*types.Cluster, error) {
	logger := log.FromContext(ctx)

	session := _mongoSession.Clone()
	defer session.Close()

	cluster := types.Cluster{}
	col := session.DB("abb").C("clusters")
	err := col.Find(bson.M{"name": name}).One(&cluster)
	if err != nil {
		if err.Error() == "not found" {
			return nil, nil
		}
		logger.Errorf("abb: clusterbyName error: %v", err)
		return nil, err
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
