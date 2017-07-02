package abb

import "github.com/jasonsoft/abb/config"

var (
	_config         *config.Configuration
	_clusterManager *ClusterManager
)

func init() {
	_config = config.GetConfig()

}
