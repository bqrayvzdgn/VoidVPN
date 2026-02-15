package keystore

import (
	"github.com/zalando/go-keyring"
)

const serviceName = "VoidVPN"

type keyringStore struct{}

func (k *keyringStore) Store(name string, key string) error {
	return keyring.Set(serviceName, name, key)
}

func (k *keyringStore) Load(name string) (string, error) {
	return keyring.Get(serviceName, name)
}

func (k *keyringStore) Delete(name string) error {
	return keyring.Delete(serviceName, name)
}

func (k *keyringStore) Exists(name string) bool {
	_, err := keyring.Get(serviceName, name)
	return err == nil
}
