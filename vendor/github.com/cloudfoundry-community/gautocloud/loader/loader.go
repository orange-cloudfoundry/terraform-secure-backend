// It has the responsibility to find the *CloudEnv* where your program run, store *Connector*s and retrieve
// services from *CloudEnv* which corresponds to one or many *Connector* and finally it will pass to *Connector* the service
// and store the result from connector.
package loader

import (
	"encoding/json"
	"fmt"
	"github.com/cloudfoundry-community/gautocloud/cloudenv"
	"github.com/cloudfoundry-community/gautocloud/connectors"
	"github.com/cloudfoundry-community/gautocloud/decoder"
	ldlogger "github.com/cloudfoundry-community/gautocloud/logger"
	"log"
	"reflect"
	"strings"
)

type Loader interface {
	ReloadConnectors()
	RegisterConnector(connector connectors.Connector)
	Inject(service interface{}) error
	InjectFromId(id string, service interface{}) error
	GetFirst(id string) (interface{}, error)
	GetAll(id string) ([]interface{}, error)
	CloudEnvs() []cloudenv.CloudEnv
	Connectors() map[string]connectors.Connector
	Store() map[string][]StoredService
	CleanConnectors()
	SetLogger(logger *log.Logger, lvl ldlogger.Level)
	CurrentCloudEnv() cloudenv.CloudEnv
	GetAppInfo() cloudenv.AppInfo
	IsInACloudEnv() bool
}

type GautocloudLoader struct {
	cloudEnvs  []cloudenv.CloudEnv
	connectors map[string]connectors.Connector
	logger     *ldlogger.LoggerLoader
	store      map[string][]StoredService
}
type StoredService struct {
	Data        interface{}
	ReflectType reflect.Type
}

// Create a new loader with cloud environment given
func NewLoader(cloudEnvs []cloudenv.CloudEnv) Loader {
	loader := &GautocloudLoader{
		cloudEnvs: cloudEnvs,
		logger:    ldlogger.NewLoggerLoader(),
	}
	loader.connectors = make(map[string]connectors.Connector)
	loader.store = make(map[string][]StoredService)
	loader.LoadCloudEnvs()
	return loader
}

// Create a new loader with cloud environment given and a logger
func NewLoaderWithLogger(cloudEnvs []cloudenv.CloudEnv, logger *log.Logger, lvl ldlogger.Level) Loader {
	loader := &GautocloudLoader{
		cloudEnvs: cloudEnvs,
		logger:    ldlogger.NewLoggerLoader(),
	}
	loader.connectors = make(map[string]connectors.Connector)
	loader.store = make(map[string][]StoredService)
	loader.SetLogger(logger, lvl)
	loader.LoadCloudEnvs()
	return loader
}

// Return all cloud environments loaded
func (l GautocloudLoader) CloudEnvs() []cloudenv.CloudEnv {
	return l.cloudEnvs
}

// Remove all registered connectors
func (l *GautocloudLoader) CleanConnectors() {
	l.connectors = make(map[string]connectors.Connector)
}

// Return all services loaded
func (l *GautocloudLoader) Store() map[string][]StoredService {
	return l.store
}

// Pass a logger to the loader to let you have the possibility to see logs
// the parameter lvl is the level of verbosity which can be
//  - logger.Lall
//  - logger.Loff
//  - logger.Ldebug
//  - logger.Linfo
//  - logger.Lwarning
//  - logger.Lerror
//  - logger.Lsevere
func (l *GautocloudLoader) SetLogger(logger *log.Logger, lvl ldlogger.Level) {
	l.logger.SetLevel(lvl)
	l.logger.SetLogger(logger)
}

// Register a connector in the loader
// This is mainly use for connectors creators
func (l *GautocloudLoader) RegisterConnector(connector connectors.Connector) {
	if _, ok := l.connectors[connector.Id()]; ok {
		l.logger.Error("During registering connector: A connector with id '%s' already exists.", connector.Id())
		return
	}
	l.logger.Debug("Loading connector '%s' ...", connector.Id())
	l.connectors[connector.Id()] = connector
	storedServices := l.load(connector)
	err := l.checkInCloudEnv()
	if err != nil {
		l.logger.Info("Skipping loading connector '%s': %s", connector.Id(), err.Error())
		return
	}
	if len(storedServices) == 0 {
		return
	}
	l.store[connector.Id()] = storedServices
	l.logger.Debug("Finished loading connector '%s' .", connector.Id())
}

// Return all registered connectors
func (l GautocloudLoader) Connectors() map[string]connectors.Connector {
	return l.connectors
}
func (l GautocloudLoader) LoadCloudEnvs() {
	for _, cloudEnv := range l.cloudEnvs {
		if !cloudEnv.IsInCloudEnv() {
			l.logger.Debug("You are not in a '%s' environment", cloudEnv.Name())
			continue
		}
		err := cloudEnv.Load()
		if err != nil {
			l.logger.Error(
				"Error during loading cloud environment %s: %s",
				cloudEnv.Name(),
				err.Error(),
			)
		}
		l.logger.Info("Environment '%s' detected and loaded.", cloudEnv.Name())
	}
}

// Reload connectors to find services
func (l *GautocloudLoader) ReloadConnectors() {
	l.LoadCloudEnvs()
	err := l.checkInCloudEnv()
	if err != nil {
		l.logger.Info("Skipping reloading connectors: " + err.Error())
		return
	}
	l.logger.Info("Reloading connectors ...")
	for _, connector := range l.connectors {
		storedServices := l.load(connector)
		l.store[connector.Id()] = storedServices
	}
	l.logger.Info("Finished reloading connectors ...")
}

// Inject service(s) found by connectors with given type
// Example:
//  var svc *dbtype.MysqlDB
//  err = loader.Inject(&svc)
//  // svc will have the value of the first service found with type *dbtype.MysqlDB
// If service parameter is not a slice it will give the first service found
// If you pass a slice of a type in service parameter, it will inject in the slice all services found with this type
// It returns an error if parameter is not a pointer or if no service(s) can be found
func (l GautocloudLoader) Inject(service interface{}) error {
	err := l.checkInCloudEnv()
	if err != nil {
		return err
	}
	notFound := true
	for id, _ := range l.connectors {
		err = l.InjectFromId(id, service)
		if err == nil && service != nil {
			notFound = false
		}
	}
	if !notFound {
		return nil
	}
	if reflect.TypeOf(service).Kind() != reflect.Ptr {
		return NewErrPtrNotGiven()
	}
	reflectType := reflect.TypeOf(service).Elem()

	if reflectType.Kind() == reflect.Slice {
		reflectType = reflectType.Elem()
	}
	return NewErrGiveService("Service with the type " + reflectType.String() + " cannot be found. (perhaps no services match any connectors)")
}

// Return the current cloud env detected
func (l GautocloudLoader) CurrentCloudEnv() cloudenv.CloudEnv {
	return l.getFirstValidCloudEnv()
}

// Return informations about instance of the running application
func (l GautocloudLoader) GetAppInfo() cloudenv.AppInfo {
	return l.getFirstValidCloudEnv().GetAppInfo()
}
func (l GautocloudLoader) checkInCloudEnv() error {
	if l.IsInACloudEnv() {
		return nil
	}
	return NewErrNotInCloud(l.getCloudEnvNames())
}
func (l GautocloudLoader) getCloudEnvNames() []string {
	names := make([]string, 0)
	for _, cloudEnv := range l.cloudEnvs {
		names = append(names, cloudEnv.Name())
	}
	return names
}

// Return true if you are in a cloud environment
func (l GautocloudLoader) IsInACloudEnv() bool {
	for _, cloudEnv := range l.cloudEnvs {
		if !cloudEnv.IsInCloudEnv() {
			continue
		}
		return true
	}
	return false
}
func (l GautocloudLoader) getFirstValidCloudEnv() cloudenv.CloudEnv {
	var finalCloudEnv cloudenv.CloudEnv
	for _, cloudEnv := range l.cloudEnvs {
		finalCloudEnv = cloudEnv
		if cloudEnv.IsInCloudEnv() {
			break
		}
	}
	return finalCloudEnv
}

// Inject service(s) found by a connector with given type
// id is the id of a connector
// Example:
//  var svc *dbtype.MysqlDB
//  err = gautocloud.InjectFromId("mysql", &svc)
//  // svc will have the value of the first service found with type *dbtype.MysqlDB in this case
// If service parameter is not a slice it will give the first service found
// If you pass a slice of a type in service parameter, it will inject in the slice all services found with this type
// It returns an error if service parameter is not a pointer, if no service(s) can be found and if connector with given id doesn't exist

func (l GautocloudLoader) InjectFromId(id string, service interface{}) error {
	err := l.checkInCloudEnv()
	if err != nil {
		return err
	}
	err = l.checkConnectorIdExist(id)
	if err != nil {
		return err
	}
	if reflect.TypeOf(service).Kind() != reflect.Ptr {
		return NewErrPtrNotGiven()
	}
	reflectType := reflect.TypeOf(service).Elem()

	vService := reflect.ValueOf(service).Elem()
	isArray := false
	if reflectType.Kind() == reflect.Slice {
		isArray = true
		reflectType = reflectType.Elem()
	}
	dataSlice := make([]interface{}, 0)
	for _, store := range l.store[id] {
		if store.ReflectType == reflectType {
			dataSlice = append(dataSlice, store.Data)
		}
	}

	if len(dataSlice) == 0 {
		return NewErrGiveService(
			fmt.Sprintf(
				"Connector with id '%s' doesn't give a service with the type '%s'. (perhaps no services match the connector)",
				id,
				reflectType.String(),
			),
		)
	}
	if !isArray {
		vService.Set(reflect.ValueOf(dataSlice[0]))
		return nil
	}
	loadSchemas := reflect.MakeSlice(reflect.SliceOf(reflectType), 0, 0)
	for _, data := range dataSlice {
		loadSchemas = reflect.Append(loadSchemas, reflect.ValueOf(data))
	}
	if service == nil {
		vService.Set(loadSchemas)
		return nil
	}
	for i := 0; i < vService.Len(); i++ {
		loadSchemas = reflect.Append(loadSchemas, vService.Index(i))
	}
	vService.Set(loadSchemas)

	return nil
}

// Return the first service found by a connector
// id is the id of a connector
// Example:
//  var svc *dbtype.MysqlDB
//  data, err = gautocloud.GetFirst("mysql")
//  svc = data.(*dbtype.MysqlDB)
// It returns the first service found or an error if no service can be found or if the connector doesn't exists
func (l GautocloudLoader) GetFirst(id string) (interface{}, error) {
	err := l.checkInCloudEnv()
	if err != nil {
		return nil, err
	}
	err = l.checkConnectorIdExist(id)
	if err != nil {
		return nil, err
	}
	if len(l.store[id]) == 0 {
		return nil, NewErrGiveService("No content have been given by connector with id '" + id + "' (no services match the connector).")
	}
	return l.store[id][0].Data, nil
}
func (l GautocloudLoader) checkConnectorIdExist(id string) error {
	if _, ok := l.connectors[id]; !ok {
		return NewErrNoConnectorFound(id)
	}
	return nil
}

// Return all services found by a connector
// id is the id of a connector
// Example:
//  var svc []interface{}
//  data, err = gautocloud.GetAll("mysql")
//  svc = data[0].(*dbtype.MysqlDB)
// warning: a connector may give you different types that's why GetAll return a slice of interface{}
// It returns the first service found or an error if no service can be found or if the connector doesn't exists
func (l GautocloudLoader) GetAll(id string) ([]interface{}, error) {
	err := l.checkInCloudEnv()
	if err != nil {
		return nil, err
	}
	err = l.checkConnectorIdExist(id)
	if err != nil {
		return nil, err
	}

	dataSlice := make([]interface{}, 0)
	for _, store := range l.store[id] {
		dataSlice = append(dataSlice, store.Data)
	}
	return dataSlice, nil
}

func (l *GautocloudLoader) load(connector connectors.Connector) []StoredService {
	services := make([]cloudenv.Service, 0)
	storedServices := make([]StoredService, 0)
	cloudEnv := l.getFirstValidCloudEnv()
	services = append(services, cloudEnv.GetServicesFromTags(connector.Tags())...)
	services = l.addService(services, cloudEnv.GetServicesFromName(connector.Name())...)
	if len(services) == 0 {
		l.logger.Debug(
			"No service found for connector '%s' \n\twith name: '%s' \n\tor tags: [ %s ]",
			connector.Id(),
			connector.Name(),
			strings.Join(connector.Tags(), ", "),
		)
		return storedServices
	}
	serviceType := reflect.TypeOf(connector.Schema())
	for _, service := range services {
		element := reflect.New(serviceType)
		decoder.UnmarshalToValue(service.Credentials, element)
		eltInterface := element.Elem().Interface()
		loadedService, err := connector.Load(eltInterface)
		if err != nil {
			l.logger.Error("Error occured during loading connector '%s': %s\n", connector.Id(), err.Error())
			continue
		}
		reflectType := reflect.TypeOf(loadedService)
		b, _ := json.MarshalIndent(service.Credentials, "\t", "\t")
		l.logger.Debug("Connector '%s' load a service which give type '%s' from credentials:\n\t%s\n",
			connector.Id(),
			reflectType.String(),
			string(b),
		)
		storedServices = append(storedServices, StoredService{
			ReflectType: reflectType,
			Data:        loadedService,
		})
	}
	l.logger.Info("Connector '%s' load %d service(s).", connector.Id(), len(storedServices))
	return storedServices
}
func (l GautocloudLoader) addService(services []cloudenv.Service, toAdd ...cloudenv.Service) []cloudenv.Service {
	for _, service := range toAdd {
		if l.serviceAlreadyExists(services, service) {
			continue
		}
		services = append(services, service)
	}
	return services
}
func (l GautocloudLoader) serviceAlreadyExists(services []cloudenv.Service, toFind cloudenv.Service) bool {
	for _, service := range services {
		if reflect.DeepEqual(service, toFind) {
			return true
		}
	}
	return false
}
