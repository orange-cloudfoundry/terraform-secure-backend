package server

import (
	"code.cloudfoundry.org/credhub-cli/credhub/credentials/values"
	"github.com/hashicorp/terraform/state"
	"github.com/orange-cloudfoundry/terraform-secure-backend/server/credhub"
	"strings"
)

type LockStore struct {
	credhubClient credhub.CredhubClient
}

func NewLockStore(credhubClient credhub.CredhubClient) *LockStore {
	return &LockStore{credhubClient}
}

func (s LockStore) Lock(path string, info *state.LockInfo) error {
	return s.toggleLock(path, info, true)
}

func (s LockStore) toggleLock(path string, info *state.LockInfo, lockState bool) error {
	if !lockState {
		return s.DeleteLock(path)
	}
	_, err := s.credhubClient.SetValue(path+LOCK_SUFFIX, values.Value(info.ID))
	return err
}

func (s LockStore) UnLock(path string, info *state.LockInfo) error {
	return s.toggleLock(path, info, false)
}

func (s LockStore) IsLocked(path string) (*state.LockInfo, bool) {
	cred, err := s.credhubClient.GetLatestValue(path + LOCK_SUFFIX)
	if err != nil {
		return nil, false
	}
	return &state.LockInfo{
		ID: string(cred.Value),
	}, true
}

func (s LockStore) DeleteLock(path string) error {
	err := s.credhubClient.Delete(path + LOCK_SUFFIX)
	if err != nil && strings.Contains(err.Error(), "does not exist") {
		return nil
	}
	return err
}
