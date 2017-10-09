package abb

import (
	"strings"

	"github.com/jasonsoft/abb/app"
	"github.com/jasonsoft/abb/config"
	"github.com/jasonsoft/abb/types"
	mgo "gopkg.in/mgo.v2"
)

var (
	_config         *config.Configuration
	_clusterManager types.ClusterService

	// repository
	_serviceRepo types.ServiceRepository

	_mongoSession *mgo.Session
)

func init() {
	_config = config.Config()
	dbx := app.DBX

	var err error
	_mongoSession, err = mgo.Dial(_config.Database.ConnectionString)
	if err != nil {
		panic(err)
	}

	switch strings.ToLower(_config.Database.Type) {
	case "mysql":
		clusterRepo := NewClusterDatabase(dbx)
		if err != nil {
			panic(err)
		}
		_clusterManager = NewClusterManager(clusterRepo)

		_serviceRepo = newServiceDAO(dbx)

	case "mongo":
		clusterRepo, err := NewClusterMongo()
		if err != nil {
			panic(err)
		}
		_clusterManager = NewClusterManager(clusterRepo)

		_serviceRepo, err = NewServiceMongo()
		if err != nil {
			panic(err)
		}
	}

}
