package abb

import (
	"strings"

	"github.com/jasonsoft/abb/app"
	"github.com/jasonsoft/abb/config"
	"github.com/jasonsoft/abb/types"
	"github.com/nlopes/slack"
	mgo "gopkg.in/mgo.v2"
)

var (
	_config         *config.Configuration
	_clusterManager types.ClusterService
	_slack          *slack.RTM

	// repository
	_serviceRepo     types.ServiceRepository
	_healthCheckRepo types.HealthCheckerRepository

	_mongoSession *mgo.Session
)

func init() {
	_config = config.Config()
	dbx := app.DBX

	// setup slack
	api := slack.New(_config.Slack.Token)
	_slack = api.NewRTM()

	go _slack.ManageConnection()

	var err error

	switch strings.ToLower(_config.Database.Type) {
	case "mysql":
		clusterRepo := NewClusterDatabase(dbx)
		if err != nil {
			panic(err)
		}
		_clusterManager = NewClusterManager(clusterRepo)

		_serviceRepo = newServiceDAO(dbx)
		_healthCheckRepo = newHealthChecker(dbx)
	case "mongo":
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
}
