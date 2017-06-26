package abb

import "github.com/jasonsoft/napnap"

func NewServiceRouter() *napnap.Router {
	router := napnap.NewRouter()

	router.Get("/v1/services", serviceListEndpoint)
	router.Post("/v1/services", serviceCreateEndpoint)

	return router
}

func serviceCreateEndpoint(c *napnap.Context) {
	ctx := c.StdContext()

	svcList, err := _manager.ServiceList(ctx)
	if err != nil {
		panic(err)
	}

	c.JSON(200, svcList)
}

func serviceListEndpoint(c *napnap.Context) {
	ctx := c.StdContext()

	var svc Service
	c.BindJSON(&svc)

	_manager.ServiceCreate(ctx, &svc)

	c.JSON(200, svc)
}
