// Package keystore provides secure storage for WireGuard private keys.
package keystore

import "fmt"

type Keystore interface {
	Store(name string, key string) error
	Load(name string) (string, error)
	Delete(name string) error
	Exists(name string) bool
}

func New() Keystore {
	ks := &keyringStore{}
	// Test if keyring is available
	testKey := "voidvpn-keyring-test"
	if err := ks.Store(testKey, "test"); err != nil {
		// Keyring unavailable, fall back to file store
		fmt.Println("  OS keyring unavailable, using encrypted file storage")
		return &fileStore{}
	}
	_ = ks.Delete(testKey)
	return ks
}
