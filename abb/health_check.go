package abb

import (
	"context"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jasonsoft/abb/app"
	"github.com/jasonsoft/abb/types"
	"github.com/jasonsoft/log"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"
)

// ************************
// Business
// ************************

func NewHealthCheckerManager(cluster *types.Cluster, repo types.HealthCheckerRepository) (types.HealthChecker, error) {
	return &HealthCheckManager{
		repo: repo,
	}, nil
}

type HealthCheckManager struct {
	repo types.HealthCheckerRepository
}

func (m *HealthCheckManager) Create(ctx context.Context, entity *types.HealthCheck) error {
	entity.ID = uuid.NewV4().String()
	return m.repo.Insert(ctx, entity)
}

func (m *HealthCheckManager) List(ctx context.Context, opts types.HealthCheckFilterOptions) ([]*types.HealthCheck, error) {
	list, err := m.repo.Find(ctx, opts)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// ************************
// Database
// ************************

func newHealthChecker(db *sqlx.DB) *HealthCheckDAO {
	return &HealthCheckDAO{
		db: db,
	}
}

type HealthCheckDAO struct {
	db *sqlx.DB
}

const insertHealthCheckSQL = "INSERT INTO `healthcheck` (`id`, `cluster_id`, `name`, `url`, `interval`, `timeout`, `retries`, `is_enabled`, `created_at`, `updated_at`) VALUES (UNHEX(:id), UNHEX(:cluster_id), :name, :url, :interval, :timeout, :retries, :is_enabled, :created_at, :updated_at);"

func (repo *HealthCheckDAO) Insert(ctx context.Context, entity *types.HealthCheck) error {
	logger := log.FromContext(ctx)

	nowUTC := time.Now().UTC()
	entity.ID = strings.Replace(entity.ID, "-", "", -1)
	entity.CreatedAt = &nowUTC
	entity.UpdatedAt = &nowUTC

	_, err := repo.db.NamedExec(insertHealthCheckSQL, entity)
	if err != nil {
		mysqlerr, ok := err.(*mysql.MySQLError)
		if ok && mysqlerr.Number == 1062 {
			return app.AppError{ErrorCode: "healthcheck_name_exists", Message: "healthcheck name already exists"}
		}
		logger.Errorf("abb: insert healthcheck fail: %v", err)
		return err
	}

	return nil
}

func (repo *HealthCheckDAO) Update(ctx context.Context, target *types.HealthCheck) error {

	return nil
}

const findHealthcheckSQL = "SELECT LOWER(HEX(id)) as `id`, LOWER(HEX(cluster_id)) as `cluster_id`, `name`,  `url`, `interval`, `timeout`, `retries`, `is_enabled`, `created_at`, `updated_at` FROM healthcheck WHERE 1=1"

func (repo *HealthCheckDAO) Find(ctx context.Context, opts types.HealthCheckFilterOptions) ([]*types.HealthCheck, error) {
	logger := log.FromContext(ctx)

	findSQL := findHealthcheckSQL
	param := map[string]interface{}{}
	if len(opts.ClusterID) > 0 {
		findSQL += " AND cluster_id = UNHEX(:cluster_id)"
		logger.Debugf("abb: find healthcheck: cluster_id: %s", opts.ClusterID)
		param["cluster_id"] = opts.ClusterID
	}

	if len(opts.Name) > 0 {
		findSQL += " And name = :name"
		param["name"] = opts.Name
		logger.Debugf("abb: find healthcheck: name: %s", opts.Name)
	}

	if opts.IsEnabled > -1 {
		findSQL += " And is_enabled = :is_enabled"
		param["is_enabled"] = opts.IsEnabled
		logger.Debugf("abb: find healthcheck: isEnabled: %s", opts.IsEnabled)
	}

	healthCheckList := []*types.HealthCheck{}
	findSQLStmt, err := repo.db.PrepareNamed(findSQL)
	if err != nil {
		log.Errorf("abb: prepare sql fail: %v", err)
		return nil, err
	}
	defer findSQLStmt.Close()

	err = findSQLStmt.Select(&healthCheckList, param)
	if err != nil {
		logger.Errorf("abb: list healthcheck fail: %v", err)
	}

	log.Debugf("abb: healthcheck count: %d", len(healthCheckList))
	return healthCheckList, nil
}

func (repo *HealthCheckDAO) Delete(ctx context.Context, id string) error {

	return nil
}

func (repo *HealthCheckDAO) FindOne(ctx context.Context, opts types.HealthCheckFilterOptions) (*types.HealthCheck, error) {

	return nil, nil
}

func EnableHealthCheck() {
	// get all healthcheck rules

}
