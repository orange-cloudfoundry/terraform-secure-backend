package interceptor

import (
	"fmt"
	"reflect"
)

// User must implement this interface on his struct/schema to be able to intercept injected value from gautocloud.
// User could now modified his own schema, gautocloud will use user schema as injection.
type SchemaIntercepter interface {
	// Found is the interface found by gautocloud. It has the exactly same type than the struct/schema which implements
	// this interface.
	Intercept(found interface{}) error
}

// Interceptor to let user implements himself write an interception on his struct/schema.
// This interceptor will call function Intercept if schema/struct implements SchemaIntercepter from current.
// It will return an error from intercept function if there is or the current interface after modification.
// Tips: If current not found (user doesn't use inject functions from gautocloud) this is schema found which will be used
func NewSchema() Intercepter {
	return IntercepterFunc(schema)
}

func schema(current, found interface{}) (interface{}, error) {
	schema := current
	if schema == nil {
		schema = found
	}
	if currentIntercept, ok := schema.(SchemaIntercepter); ok {
		err := currentIntercept.(SchemaIntercepter).Intercept(found)
		return schema, err
	}
	currentType := reflect.TypeOf(schema)
	currentValue := reflect.ValueOf(schema)

	if currentType.Kind() == reflect.Ptr {
		currentCopyElem := currentValue.Elem()
		currentCopy := currentCopyElem.Interface()
		if _, ok := currentCopy.(SchemaIntercepter); !ok {
			return nil, fmt.Errorf("schema does not implement SchemaIntercepter.")
		}
		err := currentCopy.(SchemaIntercepter).Intercept(found)
		if err != nil {
			return nil, err
		}
		currentCopyElem.Set(reflect.ValueOf(currentCopy))
		return currentValue.Interface(), nil
	}

	currentCopyPtr := reflect.New(currentType)
	currentCopyPtr.Elem().Set(currentValue)
	currentCopy := currentCopyPtr.Interface()
	if _, ok := currentCopy.(SchemaIntercepter); !ok {
		return nil, fmt.Errorf("schema does not implement SchemaIntercepter.")
	}
	err := currentCopy.(SchemaIntercepter).Intercept(found)
	if err != nil {
		return nil, err
	}
	return currentCopyPtr.Elem().Interface(), nil
}
