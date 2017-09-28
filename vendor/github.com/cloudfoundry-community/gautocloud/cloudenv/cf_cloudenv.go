package cloudenv

import (
	"github.com/cloudfoundry-community/go-cfenv"
	"os"
)

type CfCloudEnv struct {
	appEnv *cfenv.App
}

func NewCfCloudEnv() CloudEnv {
	return &CfCloudEnv{}

}
func NewCfCloudEnvWithAppEnv(appEnv *cfenv.App) CloudEnv {
	return &CfCloudEnv{
		appEnv: appEnv,
	}
}
func (c CfCloudEnv) Name() string {
	return "cloudfoundry"
}
func (c CfCloudEnv) GetServicesFromTags(tags []string) ([]Service) {
	if len(tags) == 0 {
		return make([]Service, 0)
	}
	services := make([]cfenv.Service, 0)
	for _, tag := range tags {
		servicesFound, err := c.appEnv.Services.WithTagUsingPattern(tag)
		if err != nil {
			continue
		}
		services = append(services, servicesFound...)
	}
	return c.convertCfServices(services)
}
func (c CfCloudEnv) convertCfServices(cfServices []cfenv.Service) []Service {
	services := make([]Service, 0)
	for _, cfService := range cfServices {
		services = append(services, Service{
			Credentials: cfService.Credentials,
		})
	}
	return services
}
func (c CfCloudEnv) initAppEnv() (error) {
	if !c.IsInCloudEnv() {
		return nil
	}
	if c.appEnv != nil {
		return nil
	}
	var err error
	c.appEnv, err = cfenv.Current()
	if err != nil {
		c.appEnv = nil
	}
	return err
}
func (c *CfCloudEnv) Load() error {
	if !c.IsInCloudEnv() {
		return nil
	}
	var err error
	c.appEnv, err = cfenv.Current()
	if err != nil {
		c.appEnv = nil
	}
	return err
}
func (c CfCloudEnv) GetServicesFromName(name string) ([]Service) {
	servicesFound, err := c.appEnv.Services.WithNameUsingPattern(name)
	if err != nil {
		return make([]Service, 0)
	}
	return c.convertCfServices(servicesFound)
}
func (c CfCloudEnv) IsInCloudEnv() bool {
	if os.Getenv("VCAP_APPLICATION") != "" {
		return true
	}
	return false
}
func (c CfCloudEnv) GetAppInfo() AppInfo {
	return AppInfo{
		Id: c.appEnv.ID,
		Name: c.appEnv.Name,
		Properties: map[string]interface{}{
			"uris": c.appEnv.ApplicationURIs,
			"host": c.appEnv.Host,
			"home": c.appEnv.Home,
			"index": c.appEnv.Index,
			"memory_limit": c.appEnv.MemoryLimit,
			"port": c.appEnv.Port,
			"space_id": c.appEnv.SpaceID,
			"space_name": c.appEnv.SpaceName,
			"temp_dir": c.appEnv.TempDir,
			"user": c.appEnv.User,
			"version": c.appEnv.Version,
			"working_dir": c.appEnv.WorkingDir,
		},
	}
}