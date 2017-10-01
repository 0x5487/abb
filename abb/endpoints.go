package abb

import (
	"github.com/jasonsoft/abb/app"
	"github.com/jasonsoft/abb/types"
	"github.com/jasonsoft/napnap"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
)

func NewAbbRouter() *napnap.Router {
	router := napnap.NewRouter()

	// clusters
	router.Post("/v1/clusters", clusterCreateEndpoint)
	router.Get("/v1/clusters", clusterListEndpoint)

	// nodes
	router.Get("/v1/clusters/:cluster_name/nodes", nodeListEndpoint)
	router.Get("/v1/clusters/:cluster_name/nodes/:node_id", nodeGetEndpoint)
	router.Post("/v1/clusters/:cluster_name/nodes/:node_id", nodeUpdateEndpoint)

	// network
	router.Get("/v1/clusters/:cluster_name/networks", networkListEndpoint)

	// service
	router.Post("/v1/clusters/:cluster_name/services/:service_id/redeploy", serviceRedeployEndpoint)
	router.Post("/v1/clusters/:cluster_name/services/:service_id/rollback", serviceRollbackEndpoint)
	router.Post("/v1/clusters/:cluster_name/services/:service_id/stop", serviceStopEndpoint)
	router.Get("/v1/clusters/:cluster_name/services/:service_id/raw", serviceRawEndpoint)
	router.Get("/v1/clusters/:cluster_name/services/:service_id/logs", serviceLogsEndpoint)
	router.Get("/v1/clusters/:cluster_name/services/:service_id", serviceGetEndpoint)
	router.Put("/v1/clusters/:cluster_name/services/:service_id", serviceUpdateEndpoint)
	router.Delete("/v1/clusters/:cluster_name/services/:service_id", serviceDeleteEndpoint)
	router.Get("/v1/clusters/:cluster_name/services", serviceListEndpoint)
	router.Post("/v1/clusters/:cluster_name/services", serviceCreateEndpoint)

	// task
	router.Get("/v1/clusters/:cluster_name/tasks", taskListEndpoint)

	return router
}

func taskListEndpoint(c *napnap.Context) {
	ctx := c.StdContext()
	pagination := app.GetPaginationFromContext(c)

	clusterName := c.Param("cluster_name")
	if len(clusterName) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster_name parameter was invalid"})
	}

	cluster, err := _clusterManager.ClusterByName(ctx, clusterName)
	if err != nil {
		panic(err)
	}

	taskManager, err := newTaskManager(cluster)
	if err != nil {
		panic(err)
	}
	defer taskManager.Close(ctx)

	serviceID := c.Query("service_id")

	opts := types.TaskListOption{
		ServiceID: serviceID,
	}

	taskList, err := taskManager.List(ctx, opts)
	if err != nil {
		panic(err)
	}

	pagination.SetTotalCount(len(taskList))
	apiResult := app.ApiPagiationResult{
		Pagination: pagination,
		Data:       taskList,
	}

	c.JSON(200, apiResult)
}

func clusterCreateEndpoint(c *napnap.Context) {
	ctx := c.StdContext()

	cluster := types.Cluster{}
	err := c.BindJSON(&cluster)
	if err != nil {
		panic(err)
	}

	err = _clusterManager.ClusterCreate(ctx, &cluster)
	if err != nil {
		panic(err)
	}

	c.JSON(200, cluster)

}

func clusterListEndpoint(c *napnap.Context) {
	ctx := c.StdContext()
	pagination := app.GetPaginationFromContext(c)

	clusters, err := _clusterManager.ClusterList(ctx)
	if err != nil {
		panic(err)
	}
	if clusters == nil {
		clusters = []*types.Cluster{}
	}

	pagination.SetTotalCount(len(clusters))
	apiResult := app.ApiPagiationResult{
		Pagination: pagination,
		Data:       clusters,
	}

	c.JSON(200, apiResult)
}

func networkListEndpoint(c *napnap.Context) {
	ctx := c.StdContext()
	pagination := app.GetPaginationFromContext(c)

	clusterName := c.Param("cluster_name")
	if len(clusterName) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster_name parameter was invalid"})
	}

	cluster, err := _clusterManager.ClusterByName(ctx, clusterName)
	if err != nil {
		panic(err)
	}

	serviceManager, err := NewServiceManager(cluster, _serviceRepo)
	if err != nil {
		panic(err)
	}

	dockerClient := serviceManager.DockerClient()
	defer dockerClient.Close()

	opt := dockerTypes.NetworkListOptions{}
	networkList, err := dockerClient.NetworkList(ctx, opt)
	if err != nil {
		panic(err)
	}

	pagination.SetTotalCount(len(networkList))
	apiResult := app.ApiPagiationResult{
		Pagination: pagination,
		Data:       networkList,
	}

	c.JSON(200, apiResult)
}

func nodeGetEndpoint(c *napnap.Context) {
	ctx := c.StdContext()

	clusterName := c.Param("cluster_name")
	if len(clusterName) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster_name parameter was invalid"})
	}

	cluster, err := _clusterManager.ClusterByName(ctx, clusterName)
	if err != nil {
		panic(err)
	}

	serviceManager, err := NewServiceManager(cluster, _serviceRepo)
	if err != nil {
		panic(err)
	}

	dockerClient := serviceManager.DockerClient()
	defer dockerClient.Close()

	nodeID := c.Param("node_id")
	if len(nodeID) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "node_id parameter was invalid"})
	}

	node, _, err := dockerClient.NodeInspectWithRaw(ctx, nodeID)
	if err != nil {
		panic(err)
	}

	c.JSON(200, node)
}

func nodeUpdateEndpoint(c *napnap.Context) {
	ctx := c.StdContext()

	clusterName := c.Param("cluster_name")
	if len(clusterName) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster_name parameter was invalid"})
	}

	cluster, err := _clusterManager.ClusterByName(ctx, clusterName)
	if err != nil {
		panic(err)
	}

	serviceManager, err := NewServiceManager(cluster, _serviceRepo)
	if err != nil {
		panic(err)
	}

	dockerClient := serviceManager.DockerClient()
	defer dockerClient.Close()

	nodeID := c.Param("node_id")
	if len(nodeID) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "node_id parameter was invalid"})
	}

	nodeSpec := swarm.NodeSpec{}
	err = c.BindJSON(&nodeSpec)
	if err != nil {
		panic(err)
	}

	node, _, err := dockerClient.NodeInspectWithRaw(ctx, nodeID)
	if err != nil {
		panic(err)
	}

	err = dockerClient.NodeUpdate(ctx, nodeID, node.Version, nodeSpec)
	if err != nil {
		panic(err)
	}

	// refresh
	node, _, err = dockerClient.NodeInspectWithRaw(ctx, nodeID)
	if err != nil {
		panic(err)
	}

	c.JSON(200, node)
}

func nodeListEndpoint(c *napnap.Context) {
	ctx := c.StdContext()
	pagination := app.GetPaginationFromContext(c)

	clusterName := c.Param("cluster_name")
	if len(clusterName) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster_name parameter was invalid"})
	}

	cluster, err := _clusterManager.ClusterByName(ctx, clusterName)
	if err != nil {
		panic(err)
	}

	serviceManager, err := NewServiceManager(cluster, _serviceRepo)
	if err != nil {
		panic(err)
	}

	dockerClient := serviceManager.DockerClient()
	defer dockerClient.Close()

	opt := dockerTypes.NodeListOptions{}
	nodeList, err := dockerClient.NodeList(ctx, opt)
	if err != nil {
		panic(err)
	}

	pagination.SetTotalCount(len(nodeList))
	apiResult := app.ApiPagiationResult{
		Pagination: pagination,
		Data:       nodeList,
	}

	c.JSON(200, apiResult)
}

func serviceRollbackEndpoint(c *napnap.Context) {
	// ctx := c.StdContext()

	// clusterName := c.Param("cluster_name")
	// if len(clusterName) <= 0 {
	// 	panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster_name parameter was invalid"})
	// }

	// cluster, err := _clusterManager.ClusterByName(ctx, clusterName)
	// if err != nil {
	// 	panic(err)
	// }

	// serviceManager, err := NewServiceManager(cluster)
	// if err != nil {
	// 	panic(err)
	// }

	// dockerClient := serviceManager.DockerClient()
	// defer dockerClient.Close()

	// serviceID := c.Param("service_id")
	// if len(serviceID) <= 0 {
	// 	panic(app.AppError{ErrorCode: "invalid_input", Message: "service_id parameter was invalid"})
	// }

	// // get service and rollback
	// getServiceOpt := types.ServiceGetOptions{
	// 	ServiceID: serviceID,
	// }
	// svc, err := serviceManager.ServiceGet(ctx, getServiceOpt)
	// if err != nil {
	// 	panic(err)
	// }

	// updateOpt := dockerTypes.ServiceUpdateOptions{
	// 	Rollback: "previous",
	// }
	// _, err = dockerClient.ServiceUpdate(ctx, serviceID, svc.Version, svc.Spec, updateOpt)
	// if err != nil {
	// 	log.Panicf("abb: rollback service fail: %s", err.Error())
	// }

	// c.SetStatus(200)
}

func serviceStopEndpoint(c *napnap.Context) {
	ctx := c.StdContext()

	clusterName := c.Param("cluster_name")
	if len(clusterName) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster_name parameter was invalid"})
	}

	cluster, err := _clusterManager.ClusterByName(ctx, clusterName)
	if err != nil {
		panic(err)
	}

	serviceManager, err := NewServiceManager(cluster, _serviceRepo)
	if err != nil {
		panic(err)
	}

	serviceID := c.Param("service_id")
	if len(serviceID) == 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "service_id parameter was invalid"})
	}

	svc, err := serviceManager.ServiceGetByID(ctx, serviceID)
	if err != nil {
		panic(err)
	}

	if svc == nil {
		panic(app.AppError{ErrorCode: "not_found", Message: "service was not found"})
	}

	err = serviceManager.ServiceStop(ctx, serviceID)
	if err != nil {
		panic(err)
	}

	c.SetStatus(200)
}

func serviceRawEndpoint(c *napnap.Context) {
	ctx := c.StdContext()

	clusterName := c.Param("cluster_name")
	if len(clusterName) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster_name parameter was invalid"})
	}

	cluster, err := _clusterManager.ClusterByName(ctx, clusterName)
	if err != nil {
		panic(err)
	}

	serviceManager, err := NewServiceManager(cluster, _serviceRepo)
	if err != nil {
		panic(err)
	}

	serviceID := c.Param("service_id")
	if len(serviceID) == 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "service_id parameter was invalid"})
	}

	svc, err := serviceManager.ServiceRawByID(ctx, serviceID)
	if err != nil {
		panic(err)
	}

	if svc == nil {
		panic(app.AppError{ErrorCode: "not_found", Message: "service was not found"})
	}

	c.JSON(200, svc)
}

func serviceLogsEndpoint(c *napnap.Context) {
	ctx := c.StdContext()

	clusterName := c.Param("cluster_name")
	if len(clusterName) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster_name parameter was invalid"})
	}

	cluster, err := _clusterManager.ClusterByName(ctx, clusterName)
	if err != nil {
		panic(err)
	}

	serviceManager, err := NewServiceManager(cluster, _serviceRepo)
	if err != nil {
		panic(err)
	}

	serviceID := c.Param("service_id")
	if len(serviceID) == 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "service_id parameter was invalid"})
	}

	logs, err := serviceManager.ServiceLogsByID(ctx, serviceID)
	if err != nil {
		panic(err)
	}

	// result := types.ServiceLogResult{
	// 	Logs: logs,
	// }

	c.String(200, logs)
}

func serviceGetEndpoint(c *napnap.Context) {
	ctx := c.StdContext()

	clusterName := c.Param("cluster_name")
	if len(clusterName) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster_name parameter was invalid"})
	}

	cluster, err := _clusterManager.ClusterByName(ctx, clusterName)
	if err != nil {
		panic(err)
	}

	serviceManager, err := NewServiceManager(cluster, _serviceRepo)
	if err != nil {
		panic(err)
	}

	serviceID := c.Param("service_id")
	if len(serviceID) == 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "service_id parameter was invalid"})
	}

	svc, err := serviceManager.ServiceGetByID(ctx, serviceID)
	if err != nil {
		panic(err)
	}

	if svc == nil {
		panic(app.AppError{ErrorCode: "not_found", Message: "service was not found"})
	}

	c.JSON(200, svc)
}

func serviceRedeployEndpoint(c *napnap.Context) {
	ctx := c.StdContext()

	clusterName := c.Param("cluster_name")
	if len(clusterName) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster_name parameter was invalid"})
	}

	cluster, err := _clusterManager.ClusterByName(ctx, clusterName)
	if err != nil {
		panic(err)
	}

	serviceManager, err := NewServiceManager(cluster, _serviceRepo)
	if err != nil {
		panic(err)
	}

	serviceID := c.Param("service_id")
	if len(serviceID) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "service_id parameter was invalid"})
	}

	svc, err := serviceManager.ServiceGetByID(ctx, serviceID)
	if err != nil {
		panic(err)
	}

	if svc == nil {
		panic(app.AppError{ErrorCode: "not_found", Message: "service was not found"})
	}

	err = serviceManager.Redeploy(ctx, serviceID)
	if err != nil {
		panic(err)
	}

	c.SetStatus(200)
}

func serviceDeleteEndpoint(c *napnap.Context) {
	ctx := c.StdContext()

	clusterName := c.Param("cluster_name")
	if len(clusterName) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster_name parameter was invalid"})
	}

	cluster, err := _clusterManager.ClusterByName(ctx, clusterName)
	if err != nil {
		panic(err)
	}

	if cluster == nil {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster was not found"})
	}

	serviceID := c.Param("service_id")
	if len(serviceID) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "service_id parameter was invalid"})
	}

	serviceManager, err := NewServiceManager(cluster, _serviceRepo)
	if err != nil {
		panic(err)
	}

	svc, err := serviceManager.ServiceGetByID(ctx, serviceID)
	if err != nil {
		panic(err)
	}

	if svc == nil {
		panic(app.AppError{ErrorCode: "not_found", Message: "service was not found"})
	}

	err = serviceManager.ServiceDelete(ctx, serviceID)
	if err != nil {
		panic(err)
	}
	c.SetStatus(204)
}

func serviceCreateEndpoint(c *napnap.Context) {
	ctx := c.StdContext()

	clusterName := c.Param("cluster_name")
	if len(clusterName) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster_name parameter was invalid"})
	}

	cluster, err := _clusterManager.ClusterByName(ctx, clusterName)
	if err != nil {
		panic(err)
	}
	if cluster == nil {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster was not found"})
	}

	serviceManager, err := NewServiceManager(cluster, _serviceRepo)
	if err != nil {
		panic(err)
	}

	var service types.Service
	err = c.BindJSON(&service)
	if err != nil {
		panic(err)
	}

	err = serviceManager.ServiceCreate(ctx, &service)
	if err != nil {
		panic(err)
	}

	// dockerClient := serviceManager.DockerClient()
	// defer dockerClient.Close()

	// var serviceSpec swarm.ServiceSpec
	// c.BindJSON(&serviceSpec)

	// createOptions := dockerTypes.ServiceCreateOptions{}
	// svcResp, err := dockerClient.ServiceCreate(ctx, serviceSpec, createOptions)
	// if err != nil {
	// 	if strings.Contains(err.Error(), "name conflicts") {
	// 		panic(app.AppError{ErrorCode: "service_exists", Message: "name conflicts with an existing service"})
	// 	}
	// 	log.Panicf("abb: create service err: %s", err.Error())
	// }

	// serviceInspectWithRawOpt := dockerTypes.ServiceInspectOptions{}
	// result, _, err := dockerClient.ServiceInspectWithRaw(ctx, svcResp.ID, serviceInspectWithRawOpt)
	// if err != nil {
	// 	log.Panicf("abb: get service err: %s", err.Error())
	// }

	c.JSON(200, service)

}

func serviceUpdateEndpoint(c *napnap.Context) {
	ctx := c.StdContext()

	clusterName := c.Param("cluster_name")
	if len(clusterName) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster_name parameter was invalid"})
	}

	cluster, err := _clusterManager.ClusterByName(ctx, clusterName)
	if err != nil {
		panic(err)
	}
	if cluster == nil {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster was not found"})
	}

	serviceID := c.Param("service_id")
	if len(serviceID) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "service_id parameter was invalid"})
	}

	serviceManager, err := NewServiceManager(cluster, _serviceRepo)
	if err != nil {
		panic(err)
	}

	oldService, err := serviceManager.ServiceGetByID(ctx, serviceID)
	if err != nil {
		panic(err)
	}

	if oldService == nil {
		panic(app.AppError{ErrorCode: "not_found", Message: "the service was not found"})
	}

	var service types.Service
	err = c.BindJSON(&service)
	if err != nil {
		panic(err)
	}

	service.ID = oldService.ID
	service.CreatedAt = oldService.CreatedAt
	err = serviceManager.ServiceUpdate(ctx, &service)
	if err != nil {
		panic(err)
	}

	c.JSON(200, service)
}

func serviceListEndpoint(c *napnap.Context) {
	ctx := c.StdContext()
	pagination := app.GetPaginationFromContext(c)

	clusterName := c.Param("cluster_name")
	if len(clusterName) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster_name parameter was invalid"})
	}

	cluster, err := _clusterManager.ClusterByName(ctx, clusterName)
	if err != nil {
		panic(err)
	}

	serviceManager, err := NewServiceManager(cluster, _serviceRepo)
	if err != nil {
		panic(err)
	}

	opts := types.ServiceListOptions{}
	result, err := serviceManager.List(ctx, opts)
	if err != nil {
		panic(err)
	}

	pagination.SetTotalCount(len(result))
	apiResult := app.ApiPagiationResult{
		Pagination: pagination,
		Data:       result,
	}

	c.JSON(200, apiResult)
}
