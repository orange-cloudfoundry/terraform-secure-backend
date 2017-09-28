package generic

import (
	"github.com/cloudfoundry-community/gautocloud/connectors"
	"github.com/satori/go.uuid"
)

type ConfigGenericConnector struct {
	SchemaBasedGenericConnector
}

func NewConfigGenericConnector(config interface{}) connectors.Connector {
	return &ConfigGenericConnector{
		SchemaBasedGenericConnector{
			schema: config,
			id: uuid.NewV4().String() + ":config",
			name: ".*config.*",
			tags: []string{"config.*"},
		},
	}
}