package storer

import (
	"bytes"
	"code.cloudfoundry.org/credhub-cli/credhub/credentials/values"
	"encoding/json"
	"fmt"
	"github.com/orange-cloudfoundry/terraform-secure-backend/server/credhub"
	"io"
	"io/ioutil"
)

type Credhub struct {
	basePath string
	cclient  credhub.CredhubClient
}

func NewCredhub(cclient credhub.CredhubClient) *Credhub {
	return &Credhub{
		cclient: cclient,
	}
}

func (s Credhub) Store(path string, reader io.ReadCloser) error {
	defer reader.Close()
	jDec := json.NewDecoder(reader)
	var dataJson map[string]interface{}
	err := jDec.Decode(&dataJson)
	if err != nil {
		return err
	}
	_, err = s.cclient.SetJSON(path, values.JSON(dataJson))
	return err
}

func (s Credhub) Retrieve(path string) (io.ReadCloser, error) {
	cred, err := s.cclient.GetLatestJSON(path)
	if err != nil {
		return nil, fmt.Errorf("storer/credhub: %s", err.Error())
	}
	buf := &bytes.Buffer{}
	jEnc := json.NewEncoder(buf)
	err = jEnc.Encode(cred.Value)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(buf), nil

}

func (s Credhub) Delete(path string) error {
	return s.cclient.Delete(path)
}
