package storer

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
)

type B64 struct {
	next Storer
}

func NewB64(next Storer) *B64 {
	return &B64{next: next}
}

func (s B64) Store(path string, reader io.ReadCloser) error {
	defer reader.Close()
	pipeRead, pipeWrite := io.Pipe()
	w := base64.NewEncoder(base64.StdEncoding, pipeWrite)
	go func() {
		defer pipeWrite.Close()
		defer w.Close()
		_, err := io.Copy(w, reader)
		if err != nil {
			panic(err)
		}
	}()
	return s.next.Store(path, pipeRead)
}

func (s B64) Retrieve(path string) (io.ReadCloser, error) {
	origReader, err := s.next.Retrieve(path)
	if err != nil {
		return nil, fmt.Errorf("storer/b64: %s", err.Error())
	}

	r := base64.NewDecoder(base64.StdEncoding, origReader)
	return ioutil.NopCloser(r), nil
}

func (s B64) Delete(path string) error {
	return s.next.Delete(path)
}
