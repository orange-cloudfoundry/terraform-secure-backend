// +build gautocloud_mock

package gautocloud

import (
	"github.com/cloudfoundry-community/gautocloud/loader"
	"github.com/cloudfoundry-community/gautocloud/loader/fake"
)

var defaultLoader loader.Loader = fake.NewMockLoader()
