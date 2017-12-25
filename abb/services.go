package abb

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/go-sql-driver/mysql"
	"github.com/jasonsoft/abb/app"
	"github.com/jasonsoft/abb/types"
	"github.com/jasonsoft/log"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"
)

func newDockerServiceSpec(target *types.Service, networks []dockerTypes.NetworkResource, swarmConfigs []swarm.Config, swarmSecrets []swarm.Secret) swarm.ServiceSpec {
	if len(target.Spec.Deploy.UpdateConfig.Order) == 0 {
		target.Spec.Deploy.UpdateConfig.Order = "stop-first"
	}

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
		UpdateConfig: &swarm.UpdateConfig{
			Order: target.Spec.Deploy.UpdateConfig.Order,
		},
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

	// endpoint mode
	if strings.EqualFold(target.Spec.Deploy.EndpointMode, "dnsrr") {
		// dnsrr mode
		spec.EndpointSpec.Mode = swarm.ResolutionModeDNSRR
	} else {
		// vip mode
		spec.EndpointSpec.Mode = swarm.ResolutionModeVIP
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

	// secrets
	secretRefs := []*swarm.SecretReference{}
	for _, secret := range target.Spec.Secrets {
		fileTarget := swarm.SecretReferenceFileTarget{
			Name: secret.Target,
			UID:  "0",
			GID:  "0",
			Mode: os.FileMode(0444),
		}

		// secretID and SecretName are mandatory, we have invalid references without them
		secretID := ""
		for _, swarmSecret := range swarmSecrets {
			if swarmSecret.Spec.Name == secret.Source {
				secretID = swarmSecret.ID
				break
			}
		}

		secretRef := swarm.SecretReference{
			File:       &fileTarget,
			SecretName: secret.Source,
			SecretID:   secretID,
		}

		secretRefs = append(secretRefs, &secretRef)
	}
	spec.TaskTemplate.ContainerSpec.Secrets = secretRefs

	// config
	configRefs := []*swarm.ConfigReference{}
	for _, config := range target.Spec.Configs {
		fileTarget := swarm.ConfigReferenceFileTarget{
			Name: config.Target,
			UID:  "0",
			GID:  "0",
			Mode: os.FileMode(0444),
		}

		// ConfigID and ConfigName are mandatory, we have invalid references without them
		configID := ""
		for _, swarmConfig := range swarmConfigs {
			if swarmConfig.Spec.Name == config.Source {
				configID = swarmConfig.ID
				break
			}
		}

		configRef := swarm.ConfigReference{
			File:       &fileTarget,
			ConfigName: config.Source,
			ConfigID:   configID,
		}

		configRefs = append(configRefs, &configRef)
	}
	spec.TaskTemplate.ContainerSpec.Configs = configRefs

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
	client, err := client.NewClient(cluster.Host, "1.30", nil, nil)
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
	target.ID = uuid.NewV4().String()
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
	opts := types.ServiceFilterOptions{
		ClusterID: m.cluster.ID,
		ServiceID: id,
	}

	service, err := m.repo.FindOne(ctx, opts)
	if err != nil {
		return nil, err
	}
	if service == nil {
		opts := types.ServiceFilterOptions{
			ClusterID:   m.cluster.ID,
			ServiceName: id,
		}
		service, err = m.repo.FindOne(ctx, opts)
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

	opts := types.ServiceFilterOptions{
		ClusterID: m.cluster.ID,
		ServiceID: id,
	}
	service, err := m.repo.FindOne(ctx, opts)
	if err != nil {
		return nil, err
	}
	if service == nil {
		opts := types.ServiceFilterOptions{
			ClusterID:   m.cluster.ID,
			ServiceName: id,
		}
		service, err = m.repo.FindOne(ctx, opts)
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
		if client.IsErrNotFound(err) {
			return nil, app.AppError{ErrorCode: "not_found", Message: "service raw was not found"}
		}
		logger.Errorf("abb: get service error: %v", err)
		return nil, err
	}

	return &dockerSvc, nil
}

func (m *ServiceManager) ServiceLogsByID(ctx context.Context, id string) (string, error) {
	logger := log.FromContext(ctx)

	opts := types.ServiceFilterOptions{
		ClusterID: m.cluster.ID,
		ServiceID: id,
	}
	service, err := m.repo.FindOne(ctx, opts)
	if err != nil {
		return "", err
	}
	if service == nil {
		opts := types.ServiceFilterOptions{
			ClusterID:   m.cluster.ID,
			ServiceName: id,
		}
		service, err = m.repo.FindOne(ctx, opts)
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
	opts := types.ServiceFilterOptions{
		ClusterID: m.cluster.ID,
		ServiceID: id,
	}
	service, err := m.repo.FindOne(ctx, opts)
	if err != nil {
		return nil, err
	}
	if service == nil {
		opts := types.ServiceFilterOptions{
			ClusterID:   m.cluster.ID,
			ServiceName: id,
		}
		service, err = m.repo.FindOne(ctx, opts)
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
	opts := types.ServiceFilterOptions{
		ClusterID:   m.cluster.ID,
		ServiceName: name,
	}
	return m.repo.FindOne(ctx, opts)
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

func (m *ServiceManager) List(ctx context.Context, opts types.ServiceFilterOptions) ([]*types.Service, error) {

	// get all services
	serviceListOpts := types.ServiceFilterOptions{
		ClusterID: m.cluster.ID,
	}
	svcList, err := m.repo.Find(ctx, serviceListOpts)
	if err != nil {
		panic(err)
	}

	if len(svcList) > 0 {
		dockerServiceStatus := m.deploymentStatusList(ctx)

		for _, svc := range svcList {
			for _, dockerStatus := range dockerServiceStatus {
				if svc.Name == dockerStatus.ServiceName {
					svc.DeploymentStatus = dockerStatus
				}
			}
		}
	}

	return svcList, nil
}

func (m *ServiceManager) Redeploy(ctx context.Context, id string) error {
	logger := log.FromContext(ctx)

	// get docker networks
	networkOpts := dockerTypes.NetworkListOptions{}
	networkList, err := m.client.NetworkList(ctx, networkOpts)

	// get docker configs
	configOpts := dockerTypes.ConfigListOptions{}
	configList, err := m.client.ConfigList(ctx, configOpts)

	// get docker secrets
	secretOpts := dockerTypes.SecretListOptions{}
	secretList, err := m.client.SecretList(ctx, secretOpts)

	// get service
	service, err := m.ServiceGetByID(ctx, id)
	if err != nil {
		return err
	}

	dockerSvcSpec := newDockerServiceSpec(service, networkList, configList, secretList)

	// get old spec
	serviceInspectOptions := dockerTypes.ServiceInspectOptions{}
	dockerOldSvc, _, err := m.client.ServiceInspectWithRaw(ctx, service.Name, serviceInspectOptions)
	if err != nil {
		if client.IsErrNotFound(err) {
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
		return app.AppError{ErrorCode: "stop_service_first", Message: "It seems the service is still running, you need to stop the service before delete it"}
	}

	dockerClient := m.DockerClient()
	defer dockerClient.Close()

	err = dockerClient.ServiceRemove(ctx, service.Name)
	if err != nil {
		if !client.IsErrNotFound(err) {
			logger.Errorf("abb: delete service error: %v", err)
			return err
		}
	}

	return m.repo.Delete(ctx, service.ID)
}

// ************************
// Database
// ************************

type serviceDAO struct {
	db *sqlx.DB
}

func newServiceDAO(db *sqlx.DB) types.ServiceRepository {
	return &serviceDAO{
		db: db,
	}
}

const insertServiceSQL = "INSERT INTO `services` (`id`, `cluster_id`, `name`, `specJSON`, `created_at`, `updated_at`) VALUES (UNHEX(:id), UNHEX(:cluster_id), :name, :specJSON, :created_at, :updated_at);"

func (repo *serviceDAO) Insert(ctx context.Context, entity *types.Service) error {
	logger := log.FromContext(ctx)

	nowUTC := time.Now().UTC()
	entity.ID = strings.Replace(entity.ID, "-", "", -1)
	entity.CreatedAt = &nowUTC
	entity.UpdatedAt = &nowUTC

	strB, err := json.Marshal(entity.Spec)
	if err != nil {
		return err
	}
	entity.SpecJSON = strB
	//entity.SpecStr = string(strB)

	_, err = repo.db.NamedExec(insertServiceSQL, entity)
	if err != nil {
		mysqlerr, ok := err.(*mysql.MySQLError)
		if ok && mysqlerr.Number == 1062 {
			return app.AppError{ErrorCode: "service_name_exists", Message: "service name already exists"}
		}
		logger.Errorf("abb: insert service fail: %v", err)
		return err
	}

	return nil
}

const updateServiceSQL = "UPDATE `services` SET `cluster_id`=  UNHEX(:cluster_id), `name`= :name, `specJSON`= :specJSON, `updated_at`= :updated_at WHERE id = UNHEX(:id);"

func (repo *serviceDAO) Update(ctx context.Context, entity *types.Service) error {
	logger := log.FromContext(ctx)

	nowUTC := time.Now().UTC()
	entity.UpdatedAt = &nowUTC

	strB, err := json.Marshal(entity.Spec)
	if err != nil {
		return err
	}
	entity.SpecJSON = strB

	_, err = repo.db.NamedExec(updateServiceSQL, entity)
	if err != nil {
		logger.Errorf("service: update service fail: %v", err)
		return err
	}
	return nil
}

const deleteServiceSQL = "DELETE FROM `services` WHERE `id` = UNHEX(:id);"

func (repo *serviceDAO) Delete(ctx context.Context, id string) error {
	logger := log.FromContext(ctx)
	m := map[string]interface{}{
		"id": id,
	}

	_, err := repo.db.NamedExec(deleteServiceSQL, m)
	if err != nil {
		logger.Errorf("service: delete service fail: %v", err)
		return err
	}
	return nil
}

const listServiceListSQL = "SELECT LOWER(HEX(id)) as `id`, LOWER(HEX(cluster_id)) as `cluster_id`, `name`, `specJSON`, created_at, updated_at FROM services WHERE 1=1"

func (repo *serviceDAO) Find(ctx context.Context, opts types.ServiceFilterOptions) ([]*types.Service, error) {
	logger := log.FromContext(ctx)

	findServiceSQL := listServiceListSQL
	param := map[string]interface{}{}

	if len(opts.ClusterID) > 0 {
		findServiceSQL += " AND cluster_id = UNHEX(:cluster_id)"
		logger.Debugf("service: find service: cluster_id: %s", opts.ClusterID)
		param["cluster_id"] = opts.ClusterID
	}

	if len(opts.ServiceID) > 0 {
		findServiceSQL += " AND id = UNHEX(:id)"
		logger.Debugf("service: find service: service_id: %s", opts.ServiceID)
		param["id"] = opts.ServiceID
	}

	if len(opts.ServiceName) > 0 {
		findServiceSQL += " AND name = :name"
		logger.Debugf("service: find service: service_name: %s", opts.ServiceName)
		param["name"] = opts.ServiceName
	}

	var services []*types.Service

	findServiceSQLStmt, err := repo.db.PrepareNamed(findServiceSQL)
	if err != nil {
		log.Errorf("service: prepare sql fail: %v", err)
		return nil, err
	}
	defer findServiceSQLStmt.Close()

	err = findServiceSQLStmt.Select(&services, param)
	if err != nil {
		logger.Errorf("abb: list services fail: %v", err)
	}

	for _, svc := range services {
		if err := json.Unmarshal(svc.SpecJSON, &svc.Spec); err != nil {
			return nil, err
		}
	}

	return services, nil
}

func (repo *serviceDAO) FindOne(ctx context.Context, opts types.ServiceFilterOptions) (*types.Service, error) {
	result, err := repo.Find(ctx, opts)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result[0], nil
}

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
	target.CreatedAt = &nowUTC
	target.UpdatedAt = &nowUTC
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
	target.UpdatedAt = &nowUTC

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

func (repo *ServiceMongo) Find(ctx context.Context, opts types.ServiceFilterOptions) ([]*types.Service, error) {
	logger := log.FromContext(ctx)

	session := _mongoSession.Clone()
	defer session.Close()

	filters := bson.M{}

	if len(opts.ClusterID) > 0 {
		filters["cluster_id"] = opts.ClusterID
	}

	services := []*types.Service{}
	col := session.DB("abb").C("services")
	err := col.Find(filters).Sort("-created_at").All(&services)
	if err != nil {
		if err.Error() == "not found" {
			return nil, nil
		}
		logger.Errorf("abb: findByID error: %v", err)
		return nil, err
	}
	return services, nil
}

func (repo *ServiceMongo) FindOne(ctx context.Context, opts types.ServiceFilterOptions) (*types.Service, error) {
	return nil, nil
}
