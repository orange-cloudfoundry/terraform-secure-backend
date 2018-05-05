package generic

import (
	"github.com/cloudfoundry-community/gautocloud/connectors"
	"github.com/cloudfoundry-community/gautocloud/interceptor"
	"github.com/satori/go.uuid"
)

type ConfigGenericConnector struct {
	SchemaBasedGenericConnector
}

func NewConfigGenericConnector(config interface{}, interceptors ...interceptor.Intercepter) connectors.Connector {
	return &ConfigGenericConnector{
		SchemaBasedGenericConnector{
			schema:       config,
			id:           uuid.NewV4().String() + ":config",
			name:         ".*config.*",
			tags:         []string{"config.*"},
			interceptors: interceptors,
		},
	}
}

func (c ConfigGenericConnector) Intercepter() interceptor.Intercepter {
	interceptFunc := c.SchemaBasedGenericConnector.Intercepter()
	if interceptFunc != nil {
		return interceptFunc
	}
	return interceptor.NewOverwrite()
}
