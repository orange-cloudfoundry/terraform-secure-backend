package loader

import (
	"strings"
	"fmt"
)

type ErrPtrNotGiven struct {
}

func NewErrPtrNotGiven() error {
	return ErrPtrNotGiven{

	}
}
func (e ErrPtrNotGiven) Error() string {
	return "You must pass a pointer."
}

type ErrNoConnectorFound struct {
	id string
}

func NewErrNoConnectorFound(id string) error {
	return ErrNoConnectorFound{
		id: id,
	}
}
func (e ErrNoConnectorFound) Error() string {
	return "Connector with id '" + e.id + "' not found."
}

type ErrNotInCloud struct {
	cloudEnvs []string
}

func NewErrNotInCloud(cloudEnvs []string) error {
	return ErrNotInCloud{
		cloudEnvs: cloudEnvs,
	}
}
func (e ErrNotInCloud) Error() string {
	return fmt.Sprintf(
		"You are not in any cloud environments (available environments are: [ %s ]).",
		strings.Join(e.cloudEnvs, ", "),
	)
}

type ErrGiveService struct {
	content string
}

func NewErrGiveService(content string) error {
	return ErrGiveService{
		content: content,
	}
}
func (e ErrGiveService) Error() string {
	return e.content
}
