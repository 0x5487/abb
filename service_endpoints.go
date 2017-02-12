package octopus

import "github.com/jasonsoft/napnap"

func NewServiceRouter() *napnap.Router {
	router := napnap.NewRouter()
	router.Post("/v1/services", ServiceCreateEndpoint)

	return router
}

func ServiceCreateEndpoint(c *napnap.Context) {

}
