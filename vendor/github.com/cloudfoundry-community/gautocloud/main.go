package gautocloud

import (
	"github.com/cloudfoundry-community/gautocloud/cloudenv"
	"github.com/cloudfoundry-community/gautocloud/connectors"
	"github.com/cloudfoundry-community/gautocloud/loader"
)

// Return the loader used by the facade
func Loader() loader.Loader {
	return defaultLoader
}

// Reload connectors to find services
func ReloadConnectors() {
	defaultLoader.ReloadConnectors()
}

// Register a connector in the loader
// This is mainly use for connectors creators
func RegisterConnector(connector connectors.Connector) {
	defaultLoader.RegisterConnector(connector)
}

// Inject service(s) found by connectors with given type
// Example:
//  var svc *dbtype.MysqlDB
//  err = gautocloud.Inject(&svc)
//  // svc will have the value of the first service found with type *dbtype.MysqlDB
// If service parameter is not a slice it will give the first service found
// If you pass a slice of a type in service parameter, it will inject in the slice all services found with this type
// It returns an error if parameter is not a pointer or if no service(s) can be found
func Inject(service interface{}) error {
	return defaultLoader.Inject(service)
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
func InjectFromId(id string, service interface{}) error {
	return defaultLoader.InjectFromId(id, service)
}

// Return the first service found by a connector
// id is the id of a connector
// Example:
//  var svc *dbtype.MysqlDB
//  data, err = gautocloud.GetFirst("mysql")
//  svc = data.(*dbtype.MysqlDB)
//
// It returns the first service found or an error if no service can be found or if the connector doesn't exists
func GetFirst(id string) (interface{}, error) {
	return defaultLoader.GetFirst(id)
}

// Return all services found by a connector
// id is the id of a connector
// Example:
//  var svc []interface{}
//  data, err = gautocloud.GetAll("mysql")
//  svc = data[0].(*dbtype.MysqlDB)
//
// warning: a connector may give you different types that's why GetAll return a slice of interface{}
// It returns the first service found or an error if no service can be found or if the connector doesn't exists
func GetAll(id string) ([]interface{}, error) {
	return defaultLoader.GetAll(id)
}

// Return all cloud environments loaded
func CloudEnvs() []cloudenv.CloudEnv {
	return defaultLoader.CloudEnvs()
}

// Return all registered connectors
func Connectors() map[string]connectors.Connector {
	return defaultLoader.Connectors()
}

// Return all services loaded
func Store() map[string][]loader.StoredService {
	return defaultLoader.Store()
}

// Remove all registered connectors
func CleanConnectors() {
	defaultLoader.CleanConnectors()
}

// Return the current cloud env detected
func CurrentCloudEnv() cloudenv.CloudEnv {
	return defaultLoader.CurrentCloudEnv()
}

// Return informations about instance of the running application
func GetAppInfo() cloudenv.AppInfo {
	return defaultLoader.GetAppInfo()
}

// Return true if you are in a cloud environment
func IsInACloudEnv() bool {
	return defaultLoader.IsInACloudEnv()
}

// Show previous logs entries created at initialization
// Prefer set a GAUTOCLOUD_DEBUG env var to true (or `json` to see logs as json) to see debug message at load too
// In some situation, this can be useful.
func ShowPreviousLog() {
	defaultLoader.ShowPreviousLog()
}
