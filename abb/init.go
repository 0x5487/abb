package abb

import (
	"github.com/jasonsoft/abb/app"
	"github.com/jasonsoft/abb/config"
	"github.com/jasonsoft/abb/types"
)

var (
	_config         *config.Configuration
	_clusterManager types.ClusterService
)

func init() {
	_config = config.Config()

	dbx := app.DBX()
	clusterRepo := NewClusterDatabase(dbx)
	_clusterManager = NewClusterManager(clusterRepo)
}
