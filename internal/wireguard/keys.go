// Package wireguard provides WireGuard tunnel management for VoidVPN.
package wireguard

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

type KeyPair struct {
	PrivateKey string
	PublicKey  string
}

func GenerateKeyPair() (*KeyPair, error) {
	var privateKey [32]byte
	if _, err := rand.Read(privateKey[:]); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Clamp the private key per WireGuard spec
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	var publicKey [32]byte
	curve25519.ScalarBaseMult(&publicKey, &privateKey)

	return &KeyPair{
		PrivateKey: base64.StdEncoding.EncodeToString(privateKey[:]),
		PublicKey:  base64.StdEncoding.EncodeToString(publicKey[:]),
	}, nil
}

func PublicKeyFromPrivate(privateKeyBase64 string) (string, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return "", fmt.Errorf("invalid private key encoding: %w", err)
	}
	if len(privateKeyBytes) != 32 {
		return "", fmt.Errorf("invalid private key length: expected 32 bytes, got %d", len(privateKeyBytes))
	}

	var privateKey, publicKey [32]byte
	copy(privateKey[:], privateKeyBytes)
	curve25519.ScalarBaseMult(&publicKey, &privateKey)

	return base64.StdEncoding.EncodeToString(publicKey[:]), nil
}

func ValidateKey(keyBase64 string) error {
	decoded, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return fmt.Errorf("invalid base64 encoding: %w", err)
	}
	if len(decoded) != 32 {
		return fmt.Errorf("invalid key length: expected 32 bytes, got %d", len(decoded))
	}
	return nil
}
