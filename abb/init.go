package abb

import (
	"github.com/jasonsoft/abb/config"
	"github.com/jasonsoft/abb/types"
	mgo "gopkg.in/mgo.v2"
)

var (
	_config         *config.Configuration
	_clusterManager types.ClusterService
	_serviceRepo    types.ServiceRepository

	_mongoSession *mgo.Session
)

func init() {
	_config = config.Config()

	var err error
	_mongoSession, err = mgo.Dial(_config.Database.ConnectionString)
	if err != nil {
		panic(err)
	}

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
