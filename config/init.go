package config

import (
	"os"

	"github.com/jasonsoft/log"
	"github.com/jasonsoft/log/handlers/console"
)

var (
	_config *Configuration
)

func init() {
	// setup log
	clog := console.New()
	levels := log.GetLevelsFromMinLevel("debug")
	log.RegisterHandler(clog, levels...)

	_config = &Configuration{
		DockerHost: os.Getenv("ABB_DOCKER_HOST"),
	}

}
