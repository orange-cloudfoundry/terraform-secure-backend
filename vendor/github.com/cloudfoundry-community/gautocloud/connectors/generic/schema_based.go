package generic

import (
	"github.com/cloudfoundry-community/gautocloud/connectors"
	"github.com/cloudfoundry-community/gautocloud/interceptor"
	"reflect"
)

type SchemaBasedGenericConnector struct {
	schema       interface{}
	id           string
	name         string
	tags         []string
	interceptors []interceptor.Intercepter
}

func NewSchemaBasedGenericConnector(id, name string, tags []string, schema interface{}, interceptors ...interceptor.Intercepter) connectors.Connector {
	return &SchemaBasedGenericConnector{
		schema:       schema,
		id:           id,
		name:         name,
		tags:         tags,
		interceptors: interceptors,
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

func (c SchemaBasedGenericConnector) Intercepter() interceptor.Intercepter {
	if _, ok := c.schema.(interceptor.SchemaIntercepter); ok {
		return interceptor.NewSchema()
	}
	schemaPtr := reflect.New(reflect.TypeOf(c.schema))
	if _, ok := schemaPtr.Interface().(interceptor.SchemaIntercepter); ok {
		return interceptor.NewSchema()
	}
	if len(c.interceptors) == 0 {
		return nil
	}
	return interceptor.IntercepterFunc(func(current, found interface{}) (interface{}, error) {
		var err error
		data := found
		for _, i := range c.interceptors {
			data, err = i.Intercept(current, data)
			if err != nil {
				return nil, err
			}
		}
		return data, nil
	})
}
