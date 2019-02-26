package storer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

type Cutter struct {
	next      Storer
	chunkSize int64
}

type Index struct {
	NumParts int `json:"num-parts"`
}

type Part struct {
	Part string `json:"part"`
}

func NewCutter(next Storer, chunkSize int64) *Cutter {
	return &Cutter{
		next:      next,
		chunkSize: chunkSize,
	}
}

func (s Cutter) Store(path string, reader io.ReadCloser) error {
	defer reader.Close()
	i := 0
	stop := false
	for {
		buf := &bytes.Buffer{}
		buf.WriteString(`{ "part": "`)
		written, err := io.CopyN(buf, reader, s.chunkSize)
		if err != nil && err != io.EOF {
			return err
		}
		if written == 0 {
			i--
			break
		}
		if err != nil && err == io.EOF {
			stop = true
		}

		buf.WriteString(`"}`)
		err = s.next.Store(s.partPath(path, i), ioutil.NopCloser(buf))
		if err != nil {
			return err
		}
		if stop {
			break
		}
		i++
	}
	buf := &bytes.Buffer{}
	b, _ := json.Marshal(Index{i + 1})
	buf.Write(b)
	return s.next.Store(s.indexPath(path), ioutil.NopCloser(buf))
}

func (s Cutter) Retrieve(path string) (io.ReadCloser, error) {
	rIndex, err := s.next.Retrieve(s.indexPath(path))
	if err != nil {
		return nil, fmt.Errorf("storer/cutter: %s", err.Error())
	}
	jDec := json.NewDecoder(rIndex)
	var index Index
	err = jDec.Decode(&index)
	if err != nil {
		return nil, fmt.Errorf("storer/cutter: %s", err.Error())
	}
	piper, pipew := io.Pipe()
	go func() {
		defer pipew.Close()
		for i := 0; i < index.NumParts; i++ {
			r, err := s.next.Retrieve(s.partPath(path, i))
			if err != nil {
				r.Close()
				panic(err)
			}
			jDec := json.NewDecoder(r)
			var part Part
			err = jDec.Decode(&part)
			if err != nil {
				r.Close()
				panic(err)
			}
			_, err = io.WriteString(pipew, part.Part)
			if err != nil {
				r.Close()
				panic(err)
			}
			r.Close()
		}
	}()

	return piper, nil
}

func (s Cutter) Delete(path string) error {
	rIndex, err := s.next.Retrieve(s.indexPath(path))
	if err != nil {
		return err
	}
	jDec := json.NewDecoder(rIndex)
	var index Index
	err = jDec.Decode(&index)
	if err != nil && !strings.Contains(err.Error(), "does not exist") {
		return err
	}

	err = s.next.Delete(s.indexPath(path))
	if err != nil && !strings.Contains(err.Error(), "does not exist") {
		return err
	}

	for i := 0; i < index.NumParts; i++ {
		err = s.next.Delete(s.partPath(path, i))
		if err != nil && !strings.Contains(err.Error(), "does not exist") {
			return err
		}

	}
	return nil
}

func (s Cutter) partPath(path string, index int) string {
	return fmt.Sprintf("%s/%d", path, index)
}

func (s Cutter) indexPath(path string) string {
	return fmt.Sprintf("%s/index", path)
}
