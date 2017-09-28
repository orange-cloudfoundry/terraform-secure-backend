package generic

import (
	"github.com/cloudfoundry-community/gautocloud/connectors"
)

type SchemaBasedGenericConnector struct {
	schema interface{}
	id     string
	name   string
	tags   []string
}

func NewSchemaBasedGenericConnector(id, name string, tags []string, schema interface{}) connectors.Connector {
	return &SchemaBasedGenericConnector{
		schema: schema,
		id: id,
		name: name,
		tags: tags,
	}
}
func (c SchemaBasedGenericConnector) Id() string {
	return c.id
}
func (c SchemaBasedGenericConnector) Name() string {
	return c.name
}
func (c SchemaBasedGenericConnector) Tags() []string {
	return c.tags
}
func (c SchemaBasedGenericConnector) Load(schema interface{}) (interface{}, error) {
	return schema, nil
}
func (c SchemaBasedGenericConnector) Schema() interface{} {
	return c.schema
}

