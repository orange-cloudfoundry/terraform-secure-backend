package cli

import (
	"fmt"
	"github.com/cloudfoundry-community/gautocloud"
	"github.com/orange-cloudfoundry/terraform-secure-backend/server"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type ServerApp struct {
	*cli.App
}

func NewApp(version string) *ServerApp {
	app := &ServerApp{cli.NewApp()}
	app.Name = "terraform-secure-backend"
	app.Version = version
	app.Usage = "An http server to store terraform state file securely"
	app.ErrWriter = os.Stderr
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config-path, c",
			Value: "backend-config.yml",
			Usage: "Path to the config file",
		},
	}
	app.Action = app.RunServer
	return app
}

func (a *ServerApp) Run(arguments []string) (err error) {
	a.Action = a.RunServer
	return a.App.Run(arguments)
}
func (a *ServerApp) RunServer(c *cli.Context) error {
	if gautocloud.IsInACloudEnv() && gautocloud.CurrentCloudEnv().Name() != "localcloud" {
		gobisServer, err := server.NewCloudServer(a.Version)
		if err != nil {
			return err
		}
		return gobisServer.Run()
	}
	config, err := a.loadServerConfig(c)
	if err != nil {
		return err
	}
	gobisServer, err := server.NewServer(a.Version, config)
	if err != nil {
		return err
	}
	return gobisServer.Run()
}
func (a ServerApp) loadServerConfig(c *cli.Context) (*server.ServerConfig, error) {
	confPath := c.GlobalString("config-path")
	if confPath == "" {
		return &server.ServerConfig{}, nil
	}
	return a.loadConfigFromFile(confPath)
}
func (a ServerApp) loadConfigFromFile(confPath string) (*server.ServerConfig, error) {
	dat, err := ioutil.ReadFile(confPath)
	if err != nil {
		return &server.ServerConfig{}, nil
	}
	confFile := &server.ServerConfig{}
	err = yaml.Unmarshal(dat, &confFile)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal config file found: %s", err.Error())
	}
	return confFile, nil
}
