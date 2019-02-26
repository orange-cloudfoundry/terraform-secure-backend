package credhub

import (
	"code.cloudfoundry.org/credhub-cli/credhub/credentials"
	"code.cloudfoundry.org/credhub-cli/credhub/credentials/values"
	"github.com/sirupsen/logrus"
)

type CredhubClient interface {
	GetLatestJSON(name string) (credentials.JSON, error)
	Delete(name string) error
	SetJSON(name string, value values.JSON) (credentials.JSON, error)
	FindByPath(path string) (credentials.FindResults, error)
	SetValue(name string, value values.Value) (credentials.Value, error)
	GetLatestValue(name string) (credentials.Value, error)
}

type NullCredhubClient struct {
}

func (NullCredhubClient) GetLatestJSON(name string) (credentials.JSON, error) {
	return credentials.JSON{}, nil
}

func (NullCredhubClient) Delete(name string) error {
	return nil
}

func (NullCredhubClient) SetJSON(name string, value values.JSON) (credentials.JSON, error) {
	logrus.WithField("path", name).WithField("type", "set-json").Info(value)
	return credentials.JSON{}, nil
}

func (NullCredhubClient) FindByPath(path string) (credentials.FindResults, error) {
	return credentials.FindResults{}, nil
}

func (NullCredhubClient) SetValue(name string, value values.Value) (credentials.Value, error) {
	logrus.WithField("path", name).WithField("type", "set-value").Info(value)
	return credentials.Value{}, nil
}

func (NullCredhubClient) GetLatestValue(name string) (credentials.Value, error) {
	return credentials.Value{}, nil
}
