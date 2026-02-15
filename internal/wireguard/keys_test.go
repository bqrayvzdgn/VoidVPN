package wireguard

import (
	"encoding/base64"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error: %v", err)
	}

	if kp.PrivateKey == "" {
		t.Error("PrivateKey should not be empty")
	}
	if kp.PublicKey == "" {
		t.Error("PublicKey should not be empty")
	}

	// Validate key lengths
	privBytes, err := base64.StdEncoding.DecodeString(kp.PrivateKey)
	if err != nil {
		t.Fatalf("failed to decode private key: %v", err)
	}
	if len(privBytes) != 32 {
		t.Errorf("private key length = %d, want 32", len(privBytes))
	}

	pubBytes, err := base64.StdEncoding.DecodeString(kp.PublicKey)
	if err != nil {
		t.Fatalf("failed to decode public key: %v", err)
	}
	if len(pubBytes) != 32 {
		t.Errorf("public key length = %d, want 32", len(pubBytes))
	}
}

func TestGenerateKeyPairUniqueness(t *testing.T) {
	kp1, _ := GenerateKeyPair()
	kp2, _ := GenerateKeyPair()

	if kp1.PrivateKey == kp2.PrivateKey {
		t.Error("two generated keypairs should have different private keys")
	}
	if kp1.PublicKey == kp2.PublicKey {
		t.Error("two generated keypairs should have different public keys")
	}
}

func TestPublicKeyFromPrivate(t *testing.T) {
	kp, _ := GenerateKeyPair()

	pubKey, err := PublicKeyFromPrivate(kp.PrivateKey)
	if err != nil {
		t.Fatalf("PublicKeyFromPrivate() error: %v", err)
	}
	if pubKey != kp.PublicKey {
		t.Errorf("derived public key doesn't match: got %q, want %q", pubKey, kp.PublicKey)
	}
}

func TestPublicKeyFromPrivateInvalid(t *testing.T) {
	_, err := PublicKeyFromPrivate("not-valid-base64!!!")
	if err == nil {
		t.Error("should error on invalid base64")
	}

	_, err = PublicKeyFromPrivate(base64.StdEncoding.EncodeToString([]byte("short")))
	if err == nil {
		t.Error("should error on wrong key length")
	}
}

func TestValidateKey(t *testing.T) {
	kp, _ := GenerateKeyPair()

	if err := ValidateKey(kp.PublicKey); err != nil {
		t.Errorf("ValidateKey() should pass for valid key: %v", err)
	}

	if err := ValidateKey("not-base64!!!"); err == nil {
		t.Error("ValidateKey() should fail for invalid base64")
	}

	if err := ValidateKey(base64.StdEncoding.EncodeToString([]byte("short"))); err == nil {
		t.Error("ValidateKey() should fail for wrong length")
	}
}
