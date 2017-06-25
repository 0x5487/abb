package abb

import "github.com/jasonsoft/napnap"

func NewServiceRouter() *napnap.Router {
	router := napnap.NewRouter()

	router.Get("/v1/services", serviceListEndpoint)
	router.Post("/v1/services", serviceCreateEndpoint)

	return router
}

func serviceCreateEndpoint(c *napnap.Context) {

}

func serviceListEndpoint(c *napnap.Context) {

}
