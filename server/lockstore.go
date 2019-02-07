package server

import (
	"code.cloudfoundry.org/credhub-cli/credhub/credentials/values"
	"github.com/hashicorp/terraform/state"
	"strings"
)

type LockStore struct {
	credhubClient CredhubClient
}

func NewLockStore(credhubClient CredhubClient) *LockStore {
	return &LockStore{credhubClient}
}

func (s LockStore) Lock(name string, info *state.LockInfo) error {
	return s.toggleLock(name, info, true)
}

func (s LockStore) toggleLock(name string, info *state.LockInfo, lockState bool) error {
	if !lockState {
		return s.DeleteLock(name)
	}
	_, err := s.credhubClient.SetValue(name+LOCK_SUFFIX, values.Value(info.ID))
	return err
}

func (s LockStore) UnLock(name string, info *state.LockInfo) error {
	return s.toggleLock(name, info, false)
}

func (s LockStore) IsLocked(name string) (*state.LockInfo, bool) {
	cred, err := s.credhubClient.GetLatestValue(name + LOCK_SUFFIX)
	if err != nil {
		return nil, false
	}
	return &state.LockInfo{
		ID: string(cred.Value),
	}, true
}

func (s LockStore) DeleteLock(name string) error {
	err := s.credhubClient.Delete(name + LOCK_SUFFIX)
	if err != nil && strings.Contains(err.Error(), "does not exist") {
		return nil
	}
	return err
}
