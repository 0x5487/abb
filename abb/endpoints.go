package abb

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/jasonsoft/abb/app"
	"github.com/jasonsoft/abb/identity"
	"github.com/jasonsoft/abb/types"
	"github.com/jasonsoft/go-audit"
	"github.com/jasonsoft/log"
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

	// config
	router.Get("/v1/clusters/:cluster_name/configs", configListEndpoint)
	router.Get("/v1/clusters/:cluster_name/configs/:config_id", configGetEndpoint)
	router.Post("/v1/clusters/:cluster_name/configs", configCreateEndpoint)
	router.Delete("/v1/clusters/:cluster_name/configs/:config_id", configDeleteEndpoint)

	// health
	router.Get("/v1/clusters/:cluster_name/healthcheck", healthCheckListEndpoint)
	router.Get("/v1/clusters/:cluster_name/healthcheck/:health_id", healthCheckGetEndpoint)
	router.Post("/v1/clusters/:cluster_name/healthcheck", healthCheckCreateEndpoint)
	router.Delete("/v1/clusters/:cluster_name/healthcheck/:health_id", healthCheckDeleteEndpoint)

	return router
}

func healthCheckListEndpoint(c *napnap.Context) {
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
	if cluster == nil {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster was not found"})
	}

	manager, err := NewHealthCheckerManager(_healthCheckRepo)
	if err != nil {
		panic(err)
	}

	opts := types.HealthCheckFilterOptions{
		ClusterID: cluster.ID,
		IsEnabled: -1,
	}
	list, err := manager.List(ctx, opts)
	if err != nil {
		panic(err)
	}

	if len(list) == 0 {
		list = []*types.HealthCheck{}
	}

	pagination.SetTotalCount(len(list))
	apiResult := app.ApiPagiationResult{
		Pagination: pagination,
		Data:       list,
	}

	c.JSON(200, apiResult)
}

func healthCheckGetEndpoint(c *napnap.Context) {
	log.Debug("begin health get")
}

func healthCheckCreateEndpoint(c *napnap.Context) {
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

	manager, err := NewHealthCheckerManager(_healthCheckRepo)
	if err != nil {
		panic(err)
	}

	var healthCheck types.HealthCheck
	err = c.BindJSON(&healthCheck)
	if err != nil {
		panic(err)
	}

	healthCheck.ClusterID = cluster.ID
	err = manager.Create(ctx, &healthCheck)
	if err != nil {
		panic(err)
	}

	// audit the action
	claims, _ := identity.FromContext(ctx)
	actor := claims["sub"].(string)
	namespace := fmt.Sprintf("%s.healthcheck", clusterName)
	event := &audit.Event{
		Namespace: namespace,
		TargetID:  healthCheck.Name,
		Actor:     actor,
		Action:    "create",
		State:     audit.SUCCESS,
	}
	audit.Log(event)
	c.JSON(201, healthCheck)
}

func healthCheckDeleteEndpoint(c *napnap.Context) {

}

func configGetEndpoint(c *napnap.Context) {
	ctx := c.StdContext()

	clusterName := c.Param("cluster_name")
	if len(clusterName) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster_name parameter was invalid"})
	}

	cluster, err := _clusterManager.ClusterByName(ctx, clusterName)
	if err != nil {
		panic(err)
	}

	// check permission
	claims, found := identity.FromContext(ctx)
	if found == false {
		appError := app.AppError{ErrorCode: "invalid_input", Message: "user not found."}
		panic(appError)
	}
	roles := claims["roles"]

	slicB, err := json.Marshal(roles)
	if err != nil {
		panic(err)
	}

	var newRoles []*identity.Role
	err = json.Unmarshal(slicB, &newRoles)
	if err != nil {
		panic(err)
	}

	isValid := false
	for _, role := range newRoles {
		for _, rule := range role.Rules {
			if rule.Namespace != "*" && rule.Namespace != clusterName {
				continue
			}

			for _, res := range rule.Resources {
				if res != "*" && res != "configs" {
					continue
				}

				for _, verb := range rule.Verbs {
					if verb == "*" || verb == "get" {
						log.Debugf("rule: %v", rule)
						isValid = true
					}
				}
			}
		}
	}

	if isValid == false {
		c.SetStatus(403)
		return
	}

	configID := c.Param("config_id")
	if len(configID) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "config_id parameter was invalid"})
	}

	configManager, err := newConfigManager(cluster)
	if err != nil {
		panic(err)
	}
	defer configManager.Close(ctx)

	config, err := configManager.Get(ctx, configID)
	if err != nil {
		panic(err)
	}

	c.JSON(200, config)
}

func configDeleteEndpoint(c *napnap.Context) {
	ctx := c.StdContext()

	var config types.Config
	err := c.BindJSON(&config)
	if err != nil {
		panic(err)
	}

	clusterName := c.Param("cluster_name")
	if len(clusterName) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster_name parameter was invalid"})
	}

	cluster, err := _clusterManager.ClusterByName(ctx, clusterName)
	if err != nil {
		panic(err)
	}

	configID := c.Param("config_id")
	if len(configID) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "config_id parameter was invalid"})
	}

	configManager, err := newConfigManager(cluster)
	if err != nil {
		panic(err)
	}
	defer configManager.Close(ctx)

	err = configManager.Delete(ctx, configID)
	if err != nil {
		panic(err)
	}

	c.SetStatus(204)
}

func configCreateEndpoint(c *napnap.Context) {
	ctx := c.StdContext()

	var config types.Config
	err := c.BindJSON(&config)
	if err != nil {
		panic(err)
	}

	clusterName := c.Param("cluster_name")
	if len(clusterName) <= 0 {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "cluster_name parameter was invalid"})
	}

	cluster, err := _clusterManager.ClusterByName(ctx, clusterName)
	if err != nil {
		panic(err)
	}

	configManager, err := newConfigManager(cluster)
	if err != nil {
		panic(err)
	}
	defer configManager.Close(ctx)

	err = configManager.Create(ctx, &config)
	if err != nil {
		panic(err)
	}

	c.JSON(201, config)
}

func configListEndpoint(c *napnap.Context) {
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

	// check permission
	claims, found := identity.FromContext(ctx)
	if found == false {
		appError := app.AppError{ErrorCode: "invalid_input", Message: "user not found."}
		panic(appError)
	}
	roles := claims["roles"]

	slicB, err := json.Marshal(roles)
	if err != nil {
		panic(err)
	}

	var newRoles []*identity.Role
	err = json.Unmarshal(slicB, &newRoles)
	if err != nil {
		panic(err)
	}

	isValid := false
	for _, role := range newRoles {
		for _, rule := range role.Rules {
			if rule.Namespace != "*" && rule.Namespace != clusterName {
				continue
			}

			for _, res := range rule.Resources {
				if res != "*" && res != "configs" {
					continue
				}

				for _, verb := range rule.Verbs {
					if verb == "*" || verb == "list" {
						log.Debugf("rule: %v", rule)
						isValid = true
					}
				}
			}
		}
	}

	if isValid == false {
		c.SetStatus(403)
		return
	}

	configManager, err := newConfigManager(cluster)
	if err != nil {
		panic(err)
	}
	defer configManager.Close(ctx)

	opts := types.ConfigListOption{}
	configList, err := configManager.List(ctx, opts)
	if err != nil {
		panic(err)
	}

	pagination.SetTotalCount(len(configList))
	apiResult := app.ApiPagiationResult{
		Pagination: pagination,
		Data:       configList,
	}

	c.JSON(200, apiResult)
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

	opts := types.TaskListOption{}

	serviceID := c.Query("service_id")
	if len(serviceID) > 0 {
		opts.ServiceID = serviceID
	}

	desiredState := c.Query("desired-state")
	if len(desiredState) > 0 {
		opts.DesiredState = desiredState
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

	// audit the action
	claims, _ := identity.FromContext(ctx)
	actor := claims["sub"].(string)
	namespace := "cluster"
	event := &audit.Event{
		Namespace: namespace,
		TargetID:  cluster.Name,
		Actor:     actor,
		Action:    "create",
		State:     audit.SUCCESS,
	}
	audit.Log(event)

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

	// check permission
	claims, found := identity.FromContext(ctx)
	if found == false {
		appError := app.AppError{ErrorCode: "invalid_input", Message: "user not found."}
		panic(appError)
	}
	roles := claims["roles"]

	slicB, err := json.Marshal(roles)
	if err != nil {
		panic(err)
	}

	var newRoles []*identity.Role
	err = json.Unmarshal(slicB, &newRoles)
	if err != nil {
		panic(err)
	}

	var clustersFound map[string]bool
	clustersFound = make(map[string]bool)

	isValid := false
	resultClusters := []*types.Cluster{}
	for _, role := range newRoles {
		for _, rule := range role.Rules {
			if len(rule.Namespace) > 0 && rule.Namespace != "*" {
				continue
			}

			for _, res := range rule.Resources {
				if res != "*" && res != "clusters" {
					continue
				}

				for _, resName := range rule.ResourceNames {
					if resName == "*" {
						resultClusters = clusters
						isValid = true
						break
					}

					for _, verb := range rule.Verbs {
						if verb == "*" || verb == "list" {
							log.Debugf("rule: %v", rule)
							isValid = true
							for _, cluster := range clusters {
								if _, ok := clustersFound[resName]; !ok && cluster.Name == resName {
									resultClusters = append(resultClusters, cluster)
									clustersFound[resName] = true
								}
							}
						}
					}
				}
			}
		}
	}

	if isValid == false {
		c.SetStatus(403)
		return
	}

	//Sort number from small to larger
	sort.Slice(resultClusters, func(i, j int) bool { return resultClusters[i].Sort < resultClusters[j].Sort })

	pagination.SetTotalCount(len(resultClusters))
	apiResult := app.ApiPagiationResult{
		Pagination: pagination,
		Data:       resultClusters,
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

	// audit the action
	claims, _ := identity.FromContext(ctx)
	actor := claims["sub"].(string)
	namespace := fmt.Sprintf("%s.services", clusterName)
	event := &audit.Event{
		Namespace: namespace,
		TargetID:  serviceID,
		Actor:     actor,
		Action:    "stop",
	}

	if err != nil {
		event.State = audit.FAILED
		event.Message = err.Error()
		audit.Log(event)
		panic(err)
	}

	event.State = audit.SUCCESS
	audit.Log(event)

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

	// audit the action
	claims, _ := identity.FromContext(ctx)
	actor := claims["sub"].(string)
	namespace := fmt.Sprintf("%s.services", clusterName)
	event := &audit.Event{
		Namespace: namespace,
		TargetID:  serviceID,
		Actor:     actor,
		Action:    "redeploy",
		State:     audit.SUCCESS,
	}
	audit.Log(event)

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

	// audit the action
	claims, _ := identity.FromContext(ctx)
	actor := claims["sub"].(string)
	namespace := fmt.Sprintf("%s.services", clusterName)
	event := &audit.Event{
		Namespace: namespace,
		TargetID:  serviceID,
		Actor:     actor,
		Action:    "delete",
		State:     audit.SUCCESS,
	}
	audit.Log(event)
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

	// audit the action
	claims, _ := identity.FromContext(ctx)
	actor := claims["sub"].(string)
	namespace := fmt.Sprintf("%s.services", clusterName)
	event := &audit.Event{
		Namespace: namespace,
		TargetID:  service.Name,
		Actor:     actor,
		Action:    "create",
		State:     audit.SUCCESS,
	}
	audit.Log(event)
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

	// audit the action
	claims, _ := identity.FromContext(ctx)
	actor := claims["sub"].(string)
	namespace := fmt.Sprintf("%s.services", clusterName)
	event := &audit.Event{
		Namespace: namespace,
		TargetID:  serviceID,
		Actor:     actor,
		Action:    "save",
		State:     audit.SUCCESS,
	}
	audit.Log(event)
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
	if cluster == nil {
		panic(app.AppError{ErrorCode: "not_found", Message: "cluster doesn't exist"})
	}

	serviceManager, err := NewServiceManager(cluster, _serviceRepo)
	if err != nil {
		panic(err)
	}

	opts := types.ServiceFilterOptions{}
	result, err := serviceManager.List(ctx, opts)
	if err != nil {
		panic(err)
	}

	if len(result) == 0 {
		result = []*types.Service{}
	}

	// check permission
	claims, found := identity.FromContext(ctx)
	if found == false {
		appError := app.AppError{ErrorCode: "invalid_input", Message: "user not found."}
		panic(appError)
	}
	roles := claims["roles"]

	slicB, err := json.Marshal(roles)
	if err != nil {
		panic(err)
	}

	var newRoles []*identity.Role
	err = json.Unmarshal(slicB, &newRoles)
	if err != nil {
		panic(err)
	}

	var serviceFound map[string]bool
	serviceFound = make(map[string]bool)

	resultService := []*types.Service{}
	isValid := false
	for _, role := range newRoles {
		for _, rule := range role.Rules {
			if rule.Namespace != "*" && rule.Namespace != clusterName {
				continue
			}

			for _, res := range rule.Resources {
				if res != "*" && res != "services" {
					continue
				}

				for _, resName := range rule.ResourceNames {
					if resName == "*" {
						resultService = result
						isValid = true
						break
					}

					for _, verb := range rule.Verbs {
						if verb == "*" || verb == "list" {
							log.Debugf("rule: %v", rule)
							isValid = true
							for _, service := range result {
								if _, ok := serviceFound[resName]; !ok && service.Name == resName {
									resultService = append(resultService, service)
									serviceFound[resName] = true
								}
							}
						}
					}
				}
			}
		}
	}

	if isValid == false {
		c.SetStatus(403)
		return
	}

	pagination.SetTotalCount(len(resultService))
	apiResult := app.ApiPagiationResult{
		Pagination: pagination,
		Data:       resultService,
	}

	c.JSON(200, apiResult)
}
