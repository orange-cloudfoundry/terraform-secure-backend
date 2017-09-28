package cloudenv

import (
	"os"
	"strconv"
	"net"
)

type HerokuCloudEnv struct {
	EnvVarCloudEnv
}

const ENV_KEY_APP_NAME string = "GAUTOCLOUD_APP_NAME"

func NewHerokuCloudEnv() CloudEnv {
	return &HerokuCloudEnv{NewEnvVarCloudEnv()}
}
func NewHerokuCloudEnvEnvironment(environ []string) CloudEnv {
	return &HerokuCloudEnv{NewEnvVarCloudEnvEnvironment(environ)}
}
func (c HerokuCloudEnv) Name() string {
	return "heroku"
}

func (c HerokuCloudEnv) IsInCloudEnv() bool {
	_, isIn := os.LookupEnv("DYNO")
	return isIn
}
func (c HerokuCloudEnv) externalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags & net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags & net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return ""
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String()
		}
	}
	return ""
}

func (c HerokuCloudEnv) GetAppInfo() AppInfo {
	name := c.GetEnvValueName(ENV_KEY_APP_NAME)
	if name == "" {
		name = "<unknown>"
	}
	port, _ := strconv.Atoi(c.GetEnvValueName("PORT"))
	return AppInfo{
		Id: c.GetEnvValueName("DYNO"),
		Name: name,
		Properties: map[string]interface{}{
			"port": port,
			"host": c.externalIP(),
		},
	}
}