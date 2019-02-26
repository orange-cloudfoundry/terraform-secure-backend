package storer

import (
	"io"
)

type Storer interface {
	Store(path string, reader io.ReadCloser) error
	Retrieve(path string) (io.ReadCloser, error)
	Delete(path string) error
}
