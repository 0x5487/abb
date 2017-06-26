package abb

import "os"

var (
	_manager *Manager
)

func init() {
	var err error
	dockerHost := os.Getenv("ABB_DOCKER_HOST")
	_manager, err = NewManager(dockerHost)
	if err != nil {
		panic(err)
	}
}
