package abb

import (
	"context"
	"io/ioutil"
	"strings"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/jasonsoft/abb/app"
	"github.com/jasonsoft/abb/types"
	"github.com/jasonsoft/log"
	uuid "github.com/satori/go.uuid"
)

func newDockerServiceSpec(target *types.Service, networks []dockerTypes.NetworkResource) swarm.ServiceSpec {
	spec := swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			RestartPolicy: &swarm.RestartPolicy{
				Condition:   swarm.RestartPolicyConditionNone,
				Delay:       &target.Spec.Deploy.RestartPolicy.Delay,
				MaxAttempts: &target.Spec.Deploy.RestartPolicy.MaxAttempts,
				Window:      &target.Spec.Deploy.RestartPolicy.Window,
			},
			Placement:     &swarm.Placement{},
			ContainerSpec: &swarm.ContainerSpec{},
		},
		EndpointSpec: &swarm.EndpointSpec{},
	}

	switch strings.ToLower(target.Spec.Deploy.RestartPolicy.Condition) {
	case "any":
		spec.TaskTemplate.RestartPolicy.Condition = swarm.RestartPolicyConditionAny
	}

	spec.Annotations.Name = target.Name
	spec.TaskTemplate.ContainerSpec.Image = target.Spec.Image
	spec.TaskTemplate.ContainerSpec.Env = target.Spec.Environments

	switch strings.ToLower(target.Spec.Deploy.Mode) {
	case "global":
		spec.Mode.Global = &swarm.GlobalService{}
	case "replicated":
		spec.Mode.Replicated = &swarm.ReplicatedService{
			Replicas: &target.Spec.Deploy.Replicas,
		}
	}

	// mounts
	for _, volume := range target.Spec.Volumes {
		switch strings.ToLower(volume.Type) {
		case "bind":
			mount := mount.Mount{
				Type:     mount.TypeBind,
				Source:   volume.Source,
				Target:   volume.Target,
				ReadOnly: volume.ReadOnly,
			}

			spec.TaskTemplate.ContainerSpec.Mounts = append(spec.TaskTemplate.ContainerSpec.Mounts, mount)
		}
	}

	// networks
	for _, network := range target.Spec.Networks {
		for _, dockerNetwork := range networks {
			if len(network) > 0 && network == dockerNetwork.Name {
				network := swarm.NetworkAttachmentConfig{
					Target:  dockerNetwork.ID,
					Aliases: []string{network},
				}
				spec.TaskTemplate.Networks = append(spec.TaskTemplate.Networks, network)
			}
		}
	}

	// ports
	for _, port := range target.Spec.Ports {
		portConfig := swarm.PortConfig{
			TargetPort:    port.Target,
			PublishedPort: port.Published,
		}

		switch strings.ToLower(port.Protocol) {
		case "tcp":
			portConfig.Protocol = swarm.PortConfigProtocolTCP
		case "udp":
			portConfig.Protocol = swarm.PortConfigProtocolUDP
		}

		switch strings.ToLower(port.Mode) {
		case "host":
			portConfig.PublishMode = swarm.PortConfigPublishModeHost
		case "ingress":
			portConfig.PublishMode = swarm.PortConfigPublishModeIngress
		}

		spec.EndpointSpec.Ports = append(spec.EndpointSpec.Ports, portConfig)
	}

	// placement
	for _, placement := range target.Spec.Deploy.Constraints {
		spec.TaskTemplate.Placement.Constraints = append(spec.TaskTemplate.Placement.Constraints, placement)
	}

	return spec
}

// ************************
// Business
// ************************

type ServiceManager struct {
	client  *client.Client
	cluster *types.Cluster
	repo    types.ServiceRepository
}

func NewServiceManager(cluster *types.Cluster, repo types.ServiceRepository) (types.ServiceService, error) {
	client, err := client.NewClient(cluster.Host, "v1.30", nil, nil)
	if err != nil {
		return nil, err
	}

	return &ServiceManager{
		client:  client,
		cluster: cluster,
		repo:    repo,
	}, nil
}

func (m *ServiceManager) DockerClient() *client.Client {
	return m.client
}

func (m *ServiceManager) ServiceCreate(ctx context.Context, target *types.Service) error {
	return m.repo.Insert(ctx, target)
}

func (m *ServiceManager) ServiceStop(ctx context.Context, id string) error {
	service, err := m.ServiceGetByID(ctx, id)

	err = m.client.ServiceRemove(ctx, service.Name)
	return err
}

func (m *ServiceManager) ServiceUpdate(ctx context.Context, target *types.Service) error {
	return m.repo.Update(ctx, target)
}

func (m *ServiceManager) ServiceGetByID(ctx context.Context, id string) (*types.Service, error) {
	// logger := log.FromContext(ctx)
	service, err := m.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if service == nil {
		service, err = m.repo.FindByName(ctx, id)
		if err != nil {
			return nil, err
		}
	}

	if service == nil {
		return nil, nil
	}

	dockerServiceStatus := m.deploymentStatusList(ctx)

	for _, dockerStatus := range dockerServiceStatus {
		if service.Name == dockerStatus.ServiceName {
			service.DeploymentStatus = dockerStatus
		}
	}

	return service, nil
}

func (m *ServiceManager) ServiceRawByID(ctx context.Context, id string) (*swarm.Service, error) {
	logger := log.FromContext(ctx)
	service, err := m.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if service == nil {
		service, err = m.repo.FindByName(ctx, id)
		if err != nil {
			return nil, err
		}
	}

	if service == nil {
		return nil, nil
	}

	// get old spec
	serviceInspectOptions := dockerTypes.ServiceInspectOptions{}
	dockerSvc, _, err := m.client.ServiceInspectWithRaw(ctx, service.Name, serviceInspectOptions)
	if err != nil {
		if client.IsErrServiceNotFound(err) {
			return nil, app.AppError{ErrorCode: "not_found", Message: "service raw was not found"}
		}
		logger.Errorf("abb: get service error: %v", err)
		return nil, err
	}

	return &dockerSvc, nil
}

func (m *ServiceManager) ServiceLogsByID(ctx context.Context, id string) (string, error) {
	logger := log.FromContext(ctx)
	service, err := m.repo.FindByID(ctx, id)
	if err != nil {
		return "", err
	}
	if service == nil {
		service, err = m.repo.FindByName(ctx, id)
		if err != nil {
			return "", err
		}
	}

	if service == nil {
		return "", nil
	}

	// get service logs
	logOpts := dockerTypes.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Since:      "0",
		Timestamps: false,
		Tail:       "500",
		Details:    false,
	}
	resp, err := m.client.ServiceLogs(ctx, service.Name, logOpts)
	if err != nil {
		logger.Errorf("abb: get service logs fail: %v", err)
		return "", err
	}
	body, err := ioutil.ReadAll(resp)
	if err != nil {
		logger.Errorf("abb: read service logs fail: %v", err)
		return "", err
	}
	return string(body), nil
}

func (m *ServiceManager) ServiceTaskListByID(ctx context.Context, id string) ([]swarm.Task, error) {
	logger := log.FromContext(ctx)
	service, err := m.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if service == nil {
		service, err = m.repo.FindByName(ctx, id)
		if err != nil {
			return nil, err
		}
	}

	if service == nil {
		return nil, nil
	}

	// get task per service
	filterArgs := filters.NewArgs()
	filterArgs.Add("desired-state", "running")
	filterArgs.Add("service", service.Name)

	taskListOpt := dockerTypes.TaskListOptions{
		Filters: filterArgs,
	}
	taskList, err := m.client.TaskList(ctx, taskListOpt)
	if err != nil {
		logger.Errorf("abb: list service task fail: %v", err)
		return nil, err
	}
	return taskList, nil
}

func (m *ServiceManager) ServiceGetByName(ctx context.Context, name string) (*types.Service, error) {
	return m.repo.FindByName(ctx, name)
}

func (m *ServiceManager) deploymentStatusList(ctx context.Context) map[string]types.DeploymentStatus {
	// get all nodes
	nodeListOpt := dockerTypes.NodeListOptions{}
	nodeList, err := m.client.NodeList(ctx, nodeListOpt)
	if err != nil {
		panic(err)
	}

	// get all docker services
	dockerServiceListOpts := dockerTypes.ServiceListOptions{}
	dockerSvcList, err := m.client.ServiceList(ctx, dockerServiceListOpts)
	if err != nil {
		panic(err)
	}

	// get all tasks
	taskListOpt := dockerTypes.TaskListOptions{}
	taskList, err := m.client.TaskList(ctx, taskListOpt)
	if err != nil {
		panic(err)
	}

	serviceStatus := getServicesStatus(dockerSvcList, nodeList, taskList)
	return serviceStatus
}

func (m *ServiceManager) List(ctx context.Context, opts types.ServiceListOptions) ([]*types.Service, error) {

	// get all services
	serviceListOpts := types.ServiceListOptions{}
	svcList, err := m.repo.Find(ctx, serviceListOpts)
	if err != nil {
		panic(err)
	}

	dockerServiceStatus := m.deploymentStatusList(ctx)

	for _, svc := range svcList {
		for _, dockerStatus := range dockerServiceStatus {
			if svc.Name == dockerStatus.ServiceName {
				svc.DeploymentStatus = dockerStatus
			}
		}
	}

	return svcList, nil
}

func (m *ServiceManager) Redeploy(ctx context.Context, id string) error {
	logger := log.FromContext(ctx)

	// get docker network
	networkOpts := dockerTypes.NetworkListOptions{}
	networkList, err := m.client.NetworkList(ctx, networkOpts)

	// get service
	service, err := m.ServiceGetByID(ctx, id)
	if err != nil {
		return err
	}

	dockerSvcSpec := newDockerServiceSpec(service, networkList)

	// get old spec
	serviceInspectOptions := dockerTypes.ServiceInspectOptions{}
	dockerOldSvc, _, err := m.client.ServiceInspectWithRaw(ctx, service.Name, serviceInspectOptions)
	if err != nil {
		if client.IsErrServiceNotFound(err) {
			// create new docker service
			createOptions := dockerTypes.ServiceCreateOptions{}
			_, err := m.client.ServiceCreate(ctx, dockerSvcSpec, createOptions)
			if err != nil {
				return err
			}
			return nil
		}
		logger.Errorf("abb: get service error: %v", err)
		return err
	}

	// new spec with force update
	dockerSvcSpec.TaskTemplate.ForceUpdate = dockerOldSvc.Spec.TaskTemplate.ForceUpdate + 1
	updateOpt := dockerTypes.ServiceUpdateOptions{}
	_, err = m.client.ServiceUpdate(ctx, dockerOldSvc.ID, dockerOldSvc.Version, dockerSvcSpec, updateOpt)
	if err != nil {
		logger.Panicf("abb: update service fail: %s", err.Error())
	}

	return nil
}

func (m *ServiceManager) ServiceDelete(ctx context.Context, id string) error {
	logger := log.FromContext(ctx)

	// get service
	service, err := m.ServiceGetByID(ctx, id)
	if err != nil {
		return err
	}

	// ensure the service is stop
	if service.DeploymentStatus.AvailableReplicas > 0 {
		return app.AppError{ErrorCode: "service_must_stop", Message: "It seems the service is still running, you need to stop the service before delete it"}
	}

	dockerClient := m.DockerClient()
	defer dockerClient.Close()

	err = dockerClient.ServiceRemove(ctx, service.Name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {

		} else {
			logger.Errorf("abb: delete service error: %v", err)
			return err
		}
	}

	return m.repo.Delete(ctx, service.ID)
}

// ************************
// Database
// ************************

// type serviceDAO struct {
// 	db *sqlx.DB
// }

// func newServiceDAO(db *sqlx.DB) types.ServiceRepository {
// 	return &serviceDAO{
// 		db: db,
// 	}
// }

// const insertServiceSQL = "INSERT INTO `services` (`cluster_id`, `name`, `spec`, `created_at`, `updated_at`) VALUES (:cluster_id, :name, :spec, :created_at, :updated_at);"

// func (dao *serviceDAO) Insert(ctx context.Context, entity *types.Service) error {
// 	logger := log.FromContext(ctx)

// 	nowUTC := time.Now().UTC()
// 	entity.CreatedAt = &nowUTC
// 	entity.UpdatedAt = &nowUTC

// 	strB, err := json.Marshal(entity.Spec)
// 	if err != nil {
// 		return err
// 	}
// 	entity.SpecStr = string(strB)

// 	sqlResult, err := dao.db.NamedExec(insertServiceSQL, entity)
// 	if err != nil {
// 		mysqlerr, ok := err.(*mysql.MySQLError)
// 		if ok && mysqlerr.Number == 1062 {
// 			return app.AppError{ErrorCode: "service_name_exists", Message: "service name already exists"}
// 		}
// 		logger.Errorf("abb: insert service fail: %v", err)
// 		return err
// 	}
// 	lastID, err := sqlResult.LastInsertId()
// 	if err != nil {
// 		return err
// 	}
// 	entity.ID = int(lastID)
// 	return nil
// }

// const selectServiceByIDSQL = "SELECT id, cluster_id, `name`, `spec`, created_at, updated_at FROM services where 1=1"

// func (dao *serviceDAO) SelectOne(ctx context.Context, opts types.ServiceGetOptions) (*types.Service, error) {
// 	logger := log.FromContext(ctx)

// 	m := map[string]interface{}{}

// 	var sqlStmt string
// 	if opts.ID > 0 {
// 		sqlStmt = selectServiceByIDSQL + " And (id = :id)"
// 		m["id"] = opts.ID
// 	}

// 	if len(opts.Name) > 0 {
// 		sqlStmt = selectServiceByIDSQL + " And (`name` = :name)"
// 		m["name"] = opts.Name
// 	}

// 	selectServiceByIDStmt, err := dao.db.PrepareNamed(sqlStmt)
// 	if err != nil {
// 		logger.Errorf("abb: prepare sql fail: %v", err)
// 		return nil, err
// 	}
// 	defer selectServiceByIDStmt.Close()

// 	service := types.Service{}
// 	for i := 0; i < 10; i++ {
// 		err = selectServiceByIDStmt.Get(&service, m)
// 		if err == driver.ErrBadConn {
// 			continue
// 		}
// 		if err != nil {
// 			if err == sql.ErrNoRows {
// 				return nil, nil
// 			}
// 			logger.Errorf("abb: get request fail: %v", err)
// 			return nil, err
// 		}
// 		break
// 	}

// 	if err := json.Unmarshal([]byte(service.SpecStr), &service.Spec); err != nil {
// 		return nil, err
// 	}

// 	return &service, nil
// }

// const listServiceList = "SELECT id, cluster_id, `name`, `spec`, created_at, updated_at FROM services"

// func (dao *serviceDAO) List(ctx context.Context, opts types.ServiceListOptions) ([]*types.Service, error) {
// 	logger := log.FromContext(ctx)

// 	var services []*types.Service
// 	err := dao.db.Select(&services, listServiceList)
// 	if err != nil {
// 		logger.Errorf("abb: list services fail: %v", err)
// 	}

// 	for _, svc := range services {
// 		if err := json.Unmarshal([]byte(svc.SpecStr), &svc.Spec); err != nil {
// 			return nil, err
// 		}
// 	}

// 	return services, nil
// }

// func (dao *serviceDAO) Update(ctx context.Context, target *types.Service) error {
// 	return nil
// }

// ************************
// MongoDB
// ************************

type ServiceMongo struct {
}

func NewServiceMongo() (types.ServiceRepository, error) {
	session := _mongoSession.Clone()
	defer session.Close()
	col := session.DB("abb").C("services")

	// create index
	nameIdx := mgo.Index{
		Name:       "idx_service_name",
		Key:        []string{"name"},
		Background: true,
		Sparse:     true,
		Unique:     true,
	}
	err := col.EnsureIndex(nameIdx)
	if err != nil {
		return nil, err
	}

	return &ServiceMongo{}, nil
}

func (repo *ServiceMongo) Insert(ctx context.Context, target *types.Service) error {
	logger := log.FromContext(ctx)

	session := _mongoSession.Clone()
	defer session.Close()

	col := session.DB("abb").C("services")
	target.ID = uuid.NewV4().String()
	nowUTC := time.Now().UTC()
	target.CreatedAt = nowUTC
	target.UpdatedAt = nowUTC
	err := col.Insert(target)

	if err != nil {
		if strings.HasPrefix(err.Error(), "E11000") {
			return app.AppError{ErrorCode: "duplicate", Message: "the service id or name already exits"}
		}
		logger.Errorf("abb: insert service error: %v", err)
		return err
	}
	return nil
}

func (repo *ServiceMongo) Update(ctx context.Context, target *types.Service) error {
	logger := log.FromContext(ctx)

	if len(target.ID) == 0 {
		return app.AppError{ErrorCode: "invalid_input", Message: "id can't be empty or null."}
	}
	nowUTC := time.Now().UTC()
	target.UpdatedAt = nowUTC

	session := _mongoSession.Clone()
	defer session.Close()

	col := session.DB("abb").C("services")
	colQuerier := bson.M{"_id": target.ID}
	err := col.Update(colQuerier, target)
	if err != nil {
		if strings.HasPrefix(err.Error(), "E11000") {
			return app.AppError{ErrorCode: "duplicate", Message: "the service id or name already exits"}
		}
		logger.Errorf("abb: service update error: %v", err)
		return err
	}
	return nil
}

func (repo *ServiceMongo) Delete(ctx context.Context, id string) error {
	logger := log.FromContext(ctx)

	if len(id) == 0 {
		return app.AppError{ErrorCode: "invalid_input", Message: "id can't be empty or null."}
	}

	session := _mongoSession.Clone()
	defer session.Close()

	col := session.DB("abb").C("services")
	err := col.RemoveId(id)
	if err != nil {
		logger.Errorf("abb: service delete error: %v", err)
		return err
	}
	return nil
}

func (repo *ServiceMongo) FindByName(ctx context.Context, name string) (*types.Service, error) {
	logger := log.FromContext(ctx)

	session := _mongoSession.Clone()
	defer session.Close()

	service := types.Service{}
	col := session.DB("abb").C("services")
	err := col.Find(bson.M{"name": name}).One(&service)
	if err != nil {
		if err.Error() == "not found" {
			return nil, nil
		}
		logger.Errorf("abb: findbyName error: %v", err)
		return nil, err
	}
	return &service, nil
}

func (repo *ServiceMongo) FindByID(ctx context.Context, id string) (*types.Service, error) {
	logger := log.FromContext(ctx)

	session := _mongoSession.Clone()
	defer session.Close()

	service := types.Service{}
	col := session.DB("abb").C("services")
	err := col.FindId(id).One(&service)
	if err != nil {
		if err.Error() == "not found" {
			return nil, nil
		}
		logger.Errorf("abb: findByID error: %v", err)
		return nil, err
	}
	return &service, nil
}

func (repo *ServiceMongo) Find(ctx context.Context, opts types.ServiceListOptions) ([]*types.Service, error) {
	logger := log.FromContext(ctx)

	session := _mongoSession.Clone()
	defer session.Close()

	services := []*types.Service{}
	col := session.DB("abb").C("services")
	err := col.Find(bson.M{}).Sort("-created_at").All(&services)
	if err != nil {
		if err.Error() == "not found" {
			return nil, nil
		}
		logger.Errorf("abb: findByID error: %v", err)
		return nil, err
	}
	return services, nil
}
