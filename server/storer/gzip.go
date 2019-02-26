package storer

import (
	"compress/gzip"
	"fmt"
	"io"
)

type Gzip struct {
	next Storer
}

func NewGzip(next Storer) *Gzip {
	return &Gzip{next: next}
}

func (s Gzip) Store(path string, reader io.ReadCloser) error {
	defer reader.Close()
	pipeRead, pipeWrite := io.Pipe()
	zw, err := gzip.NewWriterLevel(pipeWrite, gzip.BestCompression)
	if err != nil {
		return err
	}
	go func() {
		defer pipeWrite.Close()
		defer zw.Close()
		_, err := io.Copy(zw, reader)
		if err != nil {
			panic(err)
		}
	}()
	return s.next.Store(path, pipeRead)
}

func (s Gzip) Retrieve(path string) (io.ReadCloser, error) {
	origReader, err := s.next.Retrieve(path)
	if err != nil {
		return nil, fmt.Errorf("storer/gzip: %s", err.Error())
	}
	zr, err := gzip.NewReader(origReader)
	if err != nil {
		return nil, fmt.Errorf("storer/gzip: %s", err.Error())
	}

	return zr, nil
}

func (s Gzip) Delete(path string) error {
	return s.next.Delete(path)
}
