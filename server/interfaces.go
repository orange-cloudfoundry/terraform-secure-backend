package server

import (
	"github.com/cloudfoundry-incubator/credhub-cli/credhub/credentials"
	"github.com/cloudfoundry-incubator/credhub-cli/credhub/credentials/values"
	"github.com/cloudfoundry-incubator/credhub-cli/credhub"
)

type CredhubClient interface {
	GetLatestJSON(name string) (credentials.JSON, error)
	Delete(name string) error
	SetJSON(name string, value values.JSON, overwrite credhub.Mode) (credentials.JSON, error)
	FindByPath(path string) (credentials.FindResults, error)
	SetValue(name string, value values.Value, overwrite credhub.Mode) (credentials.Value, error)
	GetLatestValue(name string) (credentials.Value, error)
}
