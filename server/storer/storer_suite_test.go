package storer_test

import (
	"bytes"
	"encoding/json"
	"github.com/orange-cloudfoundry/terraform-secure-backend/server/storer"
	"io"
	"io/ioutil"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestStorer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storer Suite")
}

var storerRec *StorerRecorder = &StorerRecorder{
	buf:        make(map[string][]byte),
	deleteCall: make(map[string]bool),
}

type StorerRecorder struct {
	buf        map[string][]byte
	deleteCall map[string]bool
}

func (s *StorerRecorder) Store(path string, reader io.ReadCloser) error {
	b, _ := ioutil.ReadAll(reader)
	s.buf[path] = b
	return nil
}

func (s *StorerRecorder) Retrieve(path string) (io.ReadCloser, error) {
	return ioutil.NopCloser(bytes.NewBuffer(s.buf[path])), nil
}

func (s *StorerRecorder) RetrieveString(path string) string {
	return string(s.buf[path])
}

func (s *StorerRecorder) RetrieveIndex(path string) storer.Index {
	var index storer.Index
	json.Unmarshal(s.buf[path+"/index"], &index)
	return index
}

func (s *StorerRecorder) RetrievePart(subPath string) string {
	var part storer.Part
	json.Unmarshal(s.buf[subPath], &part)
	return part.Part
}

func (s *StorerRecorder) RetrieveBytes(path string) []byte {
	return s.buf[path]
}

func (s *StorerRecorder) Reset() {
	s.buf = make(map[string][]byte)
	s.deleteCall = make(map[string]bool)
}

func (s *StorerRecorder) Delete(path string) error {
	delete(s.buf, path)
	s.deleteCall[path] = true
	return nil
}

func (s *StorerRecorder) IsDeletedCall(path string) bool {
	if called, ok := s.deleteCall[path]; ok && called {
		return true
	}
	return false
}

func Str2ReadCloser(s string) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewBufferString(s))
}

func ReadCloserToBytes(r io.ReadCloser) []byte {
	b, _ := ioutil.ReadAll(r)
	return b
}
