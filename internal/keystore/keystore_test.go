package keystore

import (
	"os"
	"testing"
)

func TestNewReturnsKeystore(t *testing.T) {
	tmpDir := t.TempDir()
	orig := os.Getenv("APPDATA")
	os.Setenv("APPDATA", tmpDir)
	defer os.Setenv("APPDATA", orig)

	ks := New()
	if ks == nil {
		t.Fatal("New() returned nil")
	}
}

func TestNewKeystoreRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	orig := os.Getenv("APPDATA")
	os.Setenv("APPDATA", tmpDir)
	defer os.Setenv("APPDATA", orig)

	ks := New()
	if err := ks.Store("roundtrip", "test-secret"); err != nil {
		t.Fatalf("Store() error: %v", err)
	}
	got, err := ks.Load("roundtrip")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got != "test-secret" {
		t.Errorf("Load() = %q, want %q", got, "test-secret")
	}
}
