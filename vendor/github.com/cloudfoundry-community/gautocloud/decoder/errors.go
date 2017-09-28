package decoder

import (
	"reflect"
	"fmt"
)

type ErrTypeNotSupported struct {
	field reflect.Value
}

func NewErrTypeNotSupported(field reflect.Value) error {
	return ErrTypeNotSupported{
		field: field,
	}
}
func (e ErrTypeNotSupported) Error() string {
	return fmt.Sprintf("Type %s is not supported", e.field.Type().String())
}

type ErrDecode struct {
	content string
}

func NewErrDecode(content string) error {
	return ErrDecode{
		content: content,
	}
}
func (e ErrDecode) Error() string {
	return e.content
}
