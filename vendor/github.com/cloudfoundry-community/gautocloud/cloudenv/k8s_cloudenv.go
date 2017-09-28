package cloudenv

import (
	"os"
	"strconv"
	"net/url"
	"strings"
)

type KubernetesCloudEnv struct {
	EnvVarCloudEnv
}

func NewKubernetesCloudEnv() CloudEnv {
	return &KubernetesCloudEnv{NewEnvVarCloudEnv()}
}
func NewKubernetesCloudEnvEnvironment(environ []string) CloudEnv {
	return &KubernetesCloudEnv{NewEnvVarCloudEnvEnvironment(environ)}
}

func (c *KubernetesCloudEnv) Load() error {
	c.InitEnv(c.SanitizeEnvVars(os.Environ()))
	return nil
}

func (c KubernetesCloudEnv) Name() string {
	return "kubernetes"
}

func (c KubernetesCloudEnv) SanitizeEnvVars(envVars []string) []string {
	finalEnvVars := make([]string, 0)
	for _, envVar := range envVars {

		splitEnvVar := strings.Split(envVar, "=")
		if splitEnvVar[0] == "KUBERNETES_PORT" {
			finalEnvVars = append(finalEnvVars, "KUBERNETES_URI=" + strings.Join(splitEnvVar[1:], "="))

		}
		splitEnvVar[0] = strings.Replace(splitEnvVar[0], "_SERVICE", "", -1)
		finalEnvVars = append(finalEnvVars, strings.Join(splitEnvVar, "="))
	}
	return finalEnvVars
}

func (c KubernetesCloudEnv) IsInCloudEnv() bool {
	_, isIn := os.LookupEnv("KUBERNETES_PORT")
	return isIn
}

func (c KubernetesCloudEnv) GetAppInfo() AppInfo {
	properties := make(map[string]interface{})
	var host string
	var port int
	name := c.GetEnvValueName("HOSTNAME")
	k8sPortParsed, _ := url.Parse(c.GetEnvValueName("KUBERNETES_URI"))
	if k8sPortParsed != nil {
		parsedHost := strings.Split(k8sPortParsed.Host, ":")
		host = parsedHost[0]
		if len(parsedHost) > 1 {
			port, _ = strconv.Atoi(parsedHost[1])
		}
	}
	svc := c.GetServicesFromName("KUBERNETES")
	if len(svc) > 0 {
		properties = svc[0].Credentials
	}
	properties["host"] = host
	properties["port"] = port
	return AppInfo{
		Id: name,
		Name: name,
		Properties: properties,
	}
}
