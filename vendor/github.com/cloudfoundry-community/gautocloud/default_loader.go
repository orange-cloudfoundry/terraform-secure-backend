// +build !gautocloud_mock

package gautocloud

import (
	"github.com/cloudfoundry-community/gautocloud/cloudenv"
	"github.com/cloudfoundry-community/gautocloud/loader"
)

var defaultLoader loader.Loader = loader.NewFacadeLoader(
	[]cloudenv.CloudEnv{
		cloudenv.NewCfCloudEnv(),
		cloudenv.NewHerokuCloudEnv(),
		cloudenv.NewKubernetesCloudEnv(),
		cloudenv.NewLocalCloudEnv(),
	},
)
