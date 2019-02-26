package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hashicorp/terraform/state"
	"github.com/orange-cloudfoundry/terraform-secure-backend/server/credhub"
	"github.com/orange-cloudfoundry/terraform-secure-backend/server/storer"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type ApiController struct {
	basePath      string
	storer        storer.Storer
	store         *LockStore
	credhubClient credhub.CredhubClient
}

func NewApiController(basePath string, credhubClient credhub.CredhubClient, storer storer.Storer, store *LockStore) *ApiController {
	return &ApiController{basePath, storer, store, credhubClient}
}

type CredModel struct {
	CredhubName      string `json:"credhub_name"`
	Name             string `json:"name"`
	VersionCreatedAt string `json:"version_created_at" yaml:"version_created_at"`
	IsLocked         bool   `json:"is_locked"`
	CurrentLockId    string `json:"current_lock_id,omitempty"`
}

func (c ApiController) Store(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	entry := logrus.WithField("action", "store").WithField("name", c.RequestName(req))
	entry.Debug("Storing tfstate")
	err := c.storer.Store(c.CredhubName(req), req.Body)
	if err != nil {
		entry.Error(err)
		panic(err)
	}
}

func (c ApiController) Retrieve(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	entry := logrus.WithField("action", "retrieve").WithField("name", c.RequestName(req))
	entry.Debug("Retrieving tfstate")
	r, err := c.storer.Retrieve(c.CredhubName(req))
	if err != nil && strings.Contains(err.Error(), "does not exist") {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		entry.Error(err)
		panic(err)
	}
	defer r.Close()
	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, r)
}

func (c ApiController) Delete(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	path := c.CredhubName(req)
	entry := logrus.WithField("action", "delete").WithField("name", c.RequestName(req))
	entry.Debug("Deleting tfstate")
	err := c.storer.Delete(path)
	if err != nil {
		entry.Error(err)
		panic(err)
	}
	err = c.store.DeleteLock(path)
	if err != nil {
		entry.Error(err)
		panic(err)
	}
}

func (c ApiController) Lock(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	var info *state.LockInfo
	name := c.CredhubName(req)
	entry := logrus.WithField("action", "lock").WithField("name", c.RequestName(req))
	entry.Debug("Locking tfstate")
	info, locked := c.store.IsLocked(name)
	if locked {
		entry.Debug("Already locked")
		w.WriteHeader(http.StatusLocked)
		w.Header().Set("Content-Type", "application/json")
		w.Write(info.Marshal())
		return
	}
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		entry.Error(err)
		panic(err)
	}
	info = &state.LockInfo{}
	err = json.Unmarshal(b, info)
	if err != nil {
		entry.Error(err)
		panic(err)
	}
	err = c.store.Lock(name, info)
	if err != nil {
		entry.Error(err)
		panic(err)
	}
}

func (c ApiController) CredhubName(req *http.Request) string {
	return fmt.Sprintf("%s/%s", c.basePath, c.RequestName(req))
}

func (c ApiController) RequestName(req *http.Request) string {
	vars := mux.Vars(req)
	return vars["name"]
}

func (c ApiController) UnLock(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	var info *state.LockInfo
	name := c.CredhubName(req)
	entry := logrus.WithField("action", "unlock").WithField("name", c.RequestName(req))
	entry.Debug("Unlocking tfstate")
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		entry.Error(err)
		panic(err)
	}
	info = &state.LockInfo{}
	err = json.Unmarshal(b, info)
	if err != nil {
		entry.Error(err)
		panic(err)
	}
	currentInfo, locked := c.store.IsLocked(name)
	if locked && currentInfo.ID != info.ID {
		w.WriteHeader(http.StatusConflict)
		w.Header().Set("Content-Type", "application/json")
		w.Write(currentInfo.Marshal())
		return
	}
	err = c.store.UnLock(name, info)
	if err != nil {
		entry.Error(err)
		panic(err)
	}
}

func (c ApiController) List(w http.ResponseWriter, req *http.Request) {
	entry := logrus.WithField("action", "list")
	result, err := c.credhubClient.FindByPath(c.basePath)
	if err != nil {
		entry.Error(err)
		panic(err)
	}
	creds := result.Credentials
	backendCreds := make([]CredModel, 0)
	for _, cred := range creds {
		name := cred.Name
		if strings.HasSuffix(name, LOCK_SUFFIX) {
			continue
		}
		info, locked := c.store.IsLocked(name)
		lockId := ""
		if info != nil {
			lockId = info.ID
		}
		backendCreds = append(backendCreds, CredModel{
			Name:             ParseTfName(name),
			CredhubName:      cred.Name,
			VersionCreatedAt: cred.VersionCreatedAt,
			IsLocked:         locked,
			CurrentLockId:    lockId,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.MarshalIndent(backendCreds, "", "\t")
	w.Write(b)
}

func ParseTfName(credhubName string) string {
	splited := strings.Split(credhubName, "/")
	return splited[len(splited)-1]
}
