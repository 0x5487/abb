package abb

import (
	"github.com/jasonsoft/abb/app"
	"github.com/jasonsoft/abb/types"
	"github.com/jasonsoft/log"
	"github.com/jasonsoft/napnap"

	"fmt"
	"strings"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
)

func NewAbbRouter() *napnap.Router {
	router := napnap.NewRouter()

	router.Get("/v1/clusters", clusterListEndpoint)

	// nodes
	router.Get("/v1/clusters/:cluster_name/nodes", nodeListEndpoint)
	router.Get("/v1/clusters/:cluster_name/nodes/:node_id", nodeGetEndpoint)
	router.Post("/v1/clusters/:cluster_name/nodes/:node_id", nodeUpdateEndpoint)

	// network
	router.Get("/v1/clusters/:cluster_name/networks", networkListEndpoint)

	// service
	router.Post("/v1/clusters/:cluster_name/services/:service_id/scale/:scale_num", serviceScaleEndpoint)
	router.Post("/v1/clusters/:cluster_name/services/:service_id/redeploy", serviceRedeployEndpoint)
	router.Post("/v1/clusters/:cluster_name/services/:service_id/force-update", serviceForceUpdateEndpoint)
	router.Post("/v1/clusters/:cluster_name/services/:service_id/rollback", serviceRollbackEndpoint)
	router.Get("/v1/clusters/:cluster_name/services/:service_id", serviceGetEndpoint)
	router.Get("/v1/clusters/:cluster_name/services", serviceListEndpoint)
	router.Post("/v1/clusters/:cluster_name/services", serviceCreateEndpoint)
	router.Delete("/v1/clusters/:cluster_name/services/:service_id", serviceDeleteEndpoint)

	return router
}

func clusterListEndpoint(c *napnap.Context) {
	ctx := c.StdContext()
	pagination := app.GetPaginationFromContext(c)

	clusters, err := _clusterManager.ClusterList(ctx)
	if err != nil {
		panic(err)
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

	serviceManager, err := NewServiceManager(cluster)
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

	serviceManager, err := NewServiceManager(cluster)
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

	serviceManager, err := NewServiceManager(cluster)
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

	serviceManager, err := NewServiceManager(cluster)
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

func serviceForceUpdateEndpoint(c *napnap.Context) {
	ctx := c.StdContext()

	clusterName := c.Param("cluster_name")
	if len(clusterName) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster_name parameter was invalid"})
	}

	cluster, err := _clusterManager.ClusterByName(ctx, clusterName)
	if err != nil {
		panic(err)
	}

	serviceManager, err := NewServiceManager(cluster)
	if err != nil {
		panic(err)
	}

	dockerClient := serviceManager.DockerClient()
	defer dockerClient.Close()

	serviceID := c.Param("service_id")
	if len(serviceID) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "service_id parameter was invalid"})
	}

	// get old spec
	getServiceOpt := types.ServiceGetOptions{
		ServiceID: serviceID,
	}
	oldSvc, err := serviceManager.ServiceGet(ctx, getServiceOpt)
	if err != nil {
		panic(err)
	}
	newSpec := oldSvc.Spec

	// create newSpec
	log.Infof("abb: force-update image: %s", newSpec.TaskTemplate.ContainerSpec.Image)
	imagePaths := strings.Split(newSpec.TaskTemplate.ContainerSpec.Image, "@")
	if len(imagePaths) > 0 {
		newSpec.TaskTemplate.ContainerSpec.Image = imagePaths[0]
	}

	newSpec.TaskTemplate.ForceUpdate = uint64(1)
	updateOpt := dockerTypes.ServiceUpdateOptions{}
	_, err = dockerClient.ServiceUpdate(ctx, serviceID, oldSvc.Version, newSpec, updateOpt)
	if err != nil {
		log.Panicf("abb: update service fail: %s", err.Error())
	}

	c.SetStatus(200)
}

func serviceScaleEndpoint(c *napnap.Context) {
	ctx := c.StdContext()

	clusterName := c.Param("cluster_name")
	if len(clusterName) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster_name parameter was invalid"})
	}

	cluster, err := _clusterManager.ClusterByName(ctx, clusterName)
	if err != nil {
		panic(err)
	}

	serviceManager, err := NewServiceManager(cluster)
	if err != nil {
		panic(err)
	}

	dockerClient := serviceManager.DockerClient()

	serviceID := c.Param("service_id")
	if len(serviceID) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "service_id parameter was invalid"})
	}

	scaleNum, err := c.ParamInt("scale_num")
	if err != nil {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "scale_num parameter was invalid"})
	}

	// get service
	getServiceOpt := types.ServiceGetOptions{
		ServiceID: serviceID,
	}
	svc, err := serviceManager.ServiceGet(ctx, getServiceOpt)
	if err != nil {
		panic(err)
	}

	if svc.Spec.Mode.Replicated == nil {
		panic(app.AppError{ErrorCode: "invalid_action", Message: "scale can only be used with replicated mode"})
	}

	num := uint64(scaleNum)
	svc.Spec.Mode.Replicated.Replicas = &num

	updateOpt := dockerTypes.ServiceUpdateOptions{}
	_, err = dockerClient.ServiceUpdate(ctx, serviceID, svc.Version, svc.Spec, updateOpt)
	if err != nil {
		log.Panicf("abb: update service fail: %s", err.Error())
	}

	c.SetStatus(200)
}

func serviceRollbackEndpoint(c *napnap.Context) {
	ctx := c.StdContext()

	clusterName := c.Param("cluster_name")
	if len(clusterName) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster_name parameter was invalid"})
	}

	cluster, err := _clusterManager.ClusterByName(ctx, clusterName)
	if err != nil {
		panic(err)
	}

	serviceManager, err := NewServiceManager(cluster)
	if err != nil {
		panic(err)
	}

	dockerClient := serviceManager.DockerClient()
	defer dockerClient.Close()

	serviceID := c.Param("service_id")
	if len(serviceID) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "service_id parameter was invalid"})
	}

	// get service and rollback
	getServiceOpt := types.ServiceGetOptions{
		ServiceID: serviceID,
	}
	svc, err := serviceManager.ServiceGet(ctx, getServiceOpt)
	if err != nil {
		panic(err)
	}

	updateOpt := dockerTypes.ServiceUpdateOptions{
		Rollback: "previous",
	}
	_, err = dockerClient.ServiceUpdate(ctx, serviceID, svc.Version, svc.Spec, updateOpt)
	if err != nil {
		log.Panicf("abb: rollback service fail: %s", err.Error())
	}

	c.SetStatus(200)
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

	serviceManager, err := NewServiceManager(cluster)
	if err != nil {
		panic(err)
	}

	dockerClient := serviceManager.DockerClient()
	defer dockerClient.Close()

	serviceID := c.Param("service_id")
	if len(serviceID) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "service_id parameter was invalid"})
	}

	getServiceOpt := types.ServiceGetOptions{
		ServiceID: serviceID,
	}
	svc, err := serviceManager.ServiceGet(ctx, getServiceOpt)
	if err != nil {
		panic(err)
	}
	svcList := []swarm.Service{}
	svcList = append(svcList, *svc)

	// get all nodes
	nodeListOpt := dockerTypes.NodeListOptions{}
	nodeList, err := dockerClient.NodeList(ctx, nodeListOpt)
	if err != nil {
		panic(err)
	}

	// get all tasks
	taskListOpt := dockerTypes.TaskListOptions{}
	taskList, err := dockerClient.TaskList(ctx, taskListOpt)
	if err != nil {
		panic(err)
	}

	serviceStatus := getServicesStatus(svcList, nodeList, taskList)

	for _, svc := range svcList {
		newService := types.Service{}
		newService.Service = svc
		newService.Status = serviceStatus[svc.ID]
		c.JSON(200, newService)
	}
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

	serviceManager, err := NewServiceManager(cluster)
	if err != nil {
		panic(err)
	}

	dockerClient := serviceManager.DockerClient()
	defer dockerClient.Close()

	serviceID := c.Param("service_id")
	if len(serviceID) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "service_id parameter was invalid"})
	}

	// get old spec
	getServiceOpt := types.ServiceGetOptions{
		ServiceID: serviceID,
	}
	oldSvc, err := serviceManager.ServiceGet(ctx, getServiceOpt)
	if err != nil {
		panic(err)
	}
	newSpec := oldSvc.Spec

	// create newSpec
	zero := uint64(0)
	replicated := swarm.ReplicatedService{
		Replicas: &zero,
	}
	newSpec.Mode.Replicated = &replicated
	updateOpt := dockerTypes.ServiceUpdateOptions{}
	_, err = dockerClient.ServiceUpdate(ctx, serviceID, oldSvc.Version, newSpec, updateOpt)
	if err != nil {
		log.Panicf("abb: update service fail: %s", err.Error())
	}

	// get new service and rollback
	newSvc, err := serviceManager.ServiceGet(ctx, getServiceOpt)
	if err != nil {
		panic(err)
	}

	updateOpt = dockerTypes.ServiceUpdateOptions{
		Rollback: "previous",
	}
	_, err = dockerClient.ServiceUpdate(ctx, serviceID, newSvc.Version, oldSvc.Spec, updateOpt)
	if err != nil {
		log.Panicf("abb: rollback service fail: %s", err.Error())
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

	serviceManager, err := NewServiceManager(cluster)
	if err != nil {
		panic(err)
	}

	dockerClient := serviceManager.DockerClient()
	defer dockerClient.Close()

	serviceID := c.Param("service_id")
	if len(serviceID) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "service_id parameter was invalid"})
	}

	err = dockerClient.ServiceRemove(ctx, serviceID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			notFoundMsg := fmt.Sprintf("%s was not found", serviceID)
			panic(app.AppError{ErrorCode: "not_found", Message: notFoundMsg})
		}
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

	serviceManager, err := NewServiceManager(cluster)
	if err != nil {
		panic(err)
	}

	dockerClient := serviceManager.DockerClient()
	defer dockerClient.Close()

	var serviceSpec swarm.ServiceSpec
	c.BindJSON(&serviceSpec)

	createOptions := dockerTypes.ServiceCreateOptions{}
	svcResp, err := dockerClient.ServiceCreate(ctx, serviceSpec, createOptions)
	if err != nil {
		if strings.Contains(err.Error(), "name conflicts") {
			panic(app.AppError{ErrorCode: "service_exists", Message: "name conflicts with an existing service"})
		}
		log.Panicf("abb: create service err: %s", err.Error())
	}

	serviceInspectWithRawOpt := dockerTypes.ServiceInspectOptions{}
	result, _, err := dockerClient.ServiceInspectWithRaw(ctx, svcResp.ID, serviceInspectWithRawOpt)
	if err != nil {
		log.Panicf("abb: get service err: %s", err.Error())
	}

	c.JSON(200, result)

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

	serviceManager, err := NewServiceManager(cluster)
	if err != nil {
		panic(err)
	}

	dockerClient := serviceManager.DockerClient()
	defer dockerClient.Close()

	// get all nodes
	nodeListOpt := dockerTypes.NodeListOptions{}
	nodeList, err := dockerClient.NodeList(ctx, nodeListOpt)
	if err != nil {
		panic(err)
	}

	// get all services
	serviceListOpt := dockerTypes.ServiceListOptions{}
	svcList, err := dockerClient.ServiceList(ctx, serviceListOpt)
	if err != nil {
		panic(err)
	}

	// get all tasks
	taskListOpt := dockerTypes.TaskListOptions{}
	taskList, err := dockerClient.TaskList(ctx, taskListOpt)
	if err != nil {
		panic(err)
	}

	serviceStatus := getServicesStatus(svcList, nodeList, taskList)

	result := []*types.Service{}
	for _, svc := range svcList {
		newService := types.Service{}
		newService.Service = svc
		newService.Status = serviceStatus[svc.ID]
		result = append(result, &newService)
	}

	pagination.SetTotalCount(len(result))
	apiResult := app.ApiPagiationResult{
		Pagination: pagination,
		Data:       result,
	}

	c.JSON(200, apiResult)
}
