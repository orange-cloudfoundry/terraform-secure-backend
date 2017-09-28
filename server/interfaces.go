package server

import (
	"github.com/cloudfoundry-incubator/credhub-cli/credhub/credentials"
	"github.com/cloudfoundry-incubator/credhub-cli/credhub/credentials/values"
)

type CredhubClient interface {
	GetLatestJSON(name string) (credentials.JSON, error)
	Delete(name string) error
	SetJSON(name string, value values.JSON, overwrite bool) (credentials.JSON, error)
	FindByPath(path string) ([]credentials.Base, error)
	SetValue(name string, value values.Value, overwrite bool) (credentials.Value, error)
	GetLatestValue(name string) (credentials.Value, error)
}
