package cloudenv

import (
	"os"
	"strings"
	"encoding/json"
	"bytes"
)

type EnvVarCloudEnv struct {
	envVars []EnvVar
}
type EnvVar struct {
	Key   string
	Value string
}

func NewEnvVarCloudEnv() EnvVarCloudEnv {
	return EnvVarCloudEnv{}

}
func NewEnvVarCloudEnvEnvironment(environ []string) EnvVarCloudEnv {
	cloudEnv := EnvVarCloudEnv{}
	cloudEnv.InitEnv(environ)
	return cloudEnv
}
func (c *EnvVarCloudEnv) Load() error {
	c.InitEnv(os.Environ())
	return nil
}
func (c *EnvVarCloudEnv) InitEnv(environ []string) {
	envVars := make([]EnvVar, 0)
	for _, envVar := range environ {
		splitEnv := strings.Split(envVar, "=")
		envVars = append(envVars, EnvVar{
			Key: strings.ToLower(splitEnv[0]),
			Value: strings.TrimSpace(strings.Join(splitEnv[1:], "=")),
		})
	}
	c.envVars = envVars
}
func (c EnvVarCloudEnv) GetServicesFromTags(tags []string) ([]Service) {
	services := make([]Service, 0)
	for _, tag := range tags {
		services = append(services, c.getServicesFromPrefix(tag)...)
	}
	return services
}
func (c EnvVarCloudEnv) GetServicesFromName(name string) ([]Service) {
	return c.getServicesFromPrefix(name)
}
func (c EnvVarCloudEnv) getServicesFromPrefix(prefix string) []Service {
	services := make(map[string]Service)
	for _, envVar := range c.envVars {
		splitKey := c.splitKey(envVar.Key)
		posKey := c.findPosInKey(strings.ToLower(prefix), splitKey)
		if posKey == -1 {
			continue
		}
		toSplitPos := posKey + 1
		name := splitKey[0]
		if len(splitKey) > 1 {
			name = strings.Join(splitKey[0:toSplitPos], "_")
		}
		if _, ok := services[name]; !ok {
			services[name] = Service{
				Credentials: make(map[string]interface{}),
			}
		}
		if len(splitKey) == 1 {
			services[name].Credentials[splitKey[0]] = c.extractCredValue(splitKey, envVar.Value)
			services[name].Credentials["uri"] = c.extractCredValue(splitKey, envVar.Value)
			jsonCreds := c.decodeJson(envVar.Value)
			for key, value := range jsonCreds {
				services[name].Credentials[key] = value
			}
			continue
		}
		keyName := strings.Join(splitKey[toSplitPos:], "_")
		if keyName != "" {
			services[name].Credentials[keyName] = c.extractCredValue(splitKey[toSplitPos:], envVar.Value)
		}
		jsonCreds := c.decodeJson(envVar.Value)
		for key, value := range jsonCreds {
			services[name].Credentials[key] = value
		}
	}
	sliceServices := make([]Service, 0)
	for _, service := range services {
		sliceServices = append(sliceServices, service)
	}
	return sliceServices
}
func (c EnvVarCloudEnv) extractCredValue(splitKey []string, value string) interface{} {
	if len(splitKey) > 1 {
		return c.extractCredValue(splitKey[1:], value)
	}
	return value
}
func (c EnvVarCloudEnv) isJson(value string) bool {
	return strings.HasPrefix(value, "{") || strings.HasPrefix(value, "[")
}
func (c EnvVarCloudEnv) decodeJson(value string) map[string]interface{} {
	creds := make(map[string]interface{})
	if !c.isJson(value) {
		return creds
	}
	decoder := json.NewDecoder(bytes.NewReader([]byte(value)))
	decoder.UseNumber()
	decoder.Decode(&creds)
	return creds
}
func (c EnvVarCloudEnv) EnvVars() []EnvVar {
	return c.envVars
}

func (c EnvVarCloudEnv) splitKey(key string) []string {
	return strings.Split(key, "_")
}
func (c EnvVarCloudEnv) findPosInKey(matcher string, splitKey []string) int {
	splitMatcher := c.splitKey(matcher)
	for index, key := range splitKey {
		if len(splitMatcher) == 1 && match(splitMatcher[0], key) {
			return index
		}
		nextIndex := index + 1
		if len(splitKey) > nextIndex && match(splitMatcher[0], key) {
			pos := c.findPosInKey(strings.Join(splitMatcher[1:], "_"), splitKey[nextIndex:])
			if pos == -1 {
				return -1
			}
			return pos + 1
		}
	}
	return -1
}

func (c EnvVarCloudEnv) GetEnvValueName(key string) string {
	for _, envVar := range c.EnvVars() {
		if envVar.Key == strings.ToLower(key) {
			return envVar.Value
		}
	}
	return ""
}
