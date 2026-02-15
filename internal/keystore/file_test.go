package keystore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupKeystoreEnv(t *testing.T) func() {
	t.Helper()
	tmpDir := t.TempDir()
	orig := os.Getenv("APPDATA")
	os.Setenv("APPDATA", tmpDir)
	os.MkdirAll(keysDir(), 0700)
	return func() {
		os.Setenv("APPDATA", orig)
	}
}

func TestValidKeyName(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"mykey"},
		{"a"},
		{"key-with-dash"},
		{"key_under"},
		{strings.Repeat("a", 63)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !validKeyName.MatchString(tt.name) {
				t.Errorf("validKeyName should match %q", tt.name)
			}
		})
	}
}

func TestInvalidKeyName(t *testing.T) {
	tests := []struct {
		name  string
		label string
	}{
		{"", "empty"},
		{"-start", "starts with dash"},
		{"has space", "contains space"},
		{"a/b", "contains slash"},
		{"../x", "path traversal"},
		{strings.Repeat("a", 64), "too long"},
		{"a\nb", "contains newline"},
	}
	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			if validKeyName.MatchString(tt.name) {
				t.Errorf("validKeyName should not match %q (%s)", tt.name, tt.label)
			}
		})
	}
}

func TestKeyFileReturnsPath(t *testing.T) {
	cleanup := setupKeystoreEnv(t)
	defer cleanup()

	path, err := keyFile("mykey")
	if err != nil {
		t.Fatalf("keyFile() error: %v", err)
	}
	if !strings.HasSuffix(path, "mykey.key") {
		t.Errorf("keyFile() = %q, want suffix mykey.key", path)
	}
	dir := keysDir()
	if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(dir)) {
		t.Errorf("keyFile() path %q not under keysDir %q", path, dir)
	}
}

func TestKeyFilePathTraversal(t *testing.T) {
	cleanup := setupKeystoreEnv(t)
	defer cleanup()

	_, err := keyFile("../etc/passwd")
	if err == nil {
		t.Error("keyFile() should reject path traversal attempt")
	}
}

func TestFileStoreRoundTrip(t *testing.T) {
	cleanup := setupKeystoreEnv(t)
	defer cleanup()

	fs := &fileStore{}
	if err := fs.Store("test", "my-secret-key"); err != nil {
		t.Fatalf("Store() error: %v", err)
	}
	got, err := fs.Load("test")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got != "my-secret-key" {
		t.Errorf("Load() = %q, want %q", got, "my-secret-key")
	}
}

func TestFileStoreOverwrite(t *testing.T) {
	cleanup := setupKeystoreEnv(t)
	defer cleanup()

	fs := &fileStore{}
	fs.Store("overwrite", "first-value")
	fs.Store("overwrite", "second-value")

	got, err := fs.Load("overwrite")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got != "second-value" {
		t.Errorf("Load() = %q, want %q", got, "second-value")
	}
}

func TestFileStoreDelete(t *testing.T) {
	cleanup := setupKeystoreEnv(t)
	defer cleanup()

	fs := &fileStore{}
	fs.Store("delme", "value")
	if err := fs.Delete("delme"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	_, err := fs.Load("delme")
	if err == nil {
		t.Error("Load() should error after Delete()")
	}
}

func TestFileStoreExists(t *testing.T) {
	cleanup := setupKeystoreEnv(t)
	defer cleanup()

	fs := &fileStore{}
	fs.Store("existkey", "value")

	if !fs.Exists("existkey") {
		t.Error("Exists() should return true after Store()")
	}
	fs.Delete("existkey")
	if fs.Exists("existkey") {
		t.Error("Exists() should return false after Delete()")
	}
}

func TestFileStoreLoadNonExistent(t *testing.T) {
	cleanup := setupKeystoreEnv(t)
	defer cleanup()

	fs := &fileStore{}
	_, err := fs.Load("nope")
	if err == nil {
		t.Error("Load() should error for nonexistent key")
	}
}

func TestFileStoreCorruptedCiphertext(t *testing.T) {
	cleanup := setupKeystoreEnv(t)
	defer cleanup()

	// Write garbage to a key file
	path, err := keyFile("corrupt")
	if err != nil {
		t.Fatalf("keyFile() error: %v", err)
	}
	os.WriteFile(path, []byte("dGhpcyBpcyBub3QgdmFsaWQgY2lwaGVydGV4dA=="), 0600)

	fs := &fileStore{}
	_, err = fs.Load("corrupt")
	if err == nil {
		t.Error("Load() should error for corrupted ciphertext")
	}
}

func TestDeriveKeyConsistency(t *testing.T) {
	cleanup := setupKeystoreEnv(t)
	defer cleanup()

	// Ensure keysDir exists for salt file
	os.MkdirAll(keysDir(), 0700)

	key1, err := deriveKey()
	if err != nil {
		t.Fatalf("deriveKey() first call error: %v", err)
	}
	key2, err := deriveKey()
	if err != nil {
		t.Fatalf("deriveKey() second call error: %v", err)
	}

	if len(key1) != 32 {
		t.Errorf("deriveKey() returned %d bytes, want 32", len(key1))
	}
	for i := range key1 {
		if key1[i] != key2[i] {
			t.Fatal("deriveKey() returned different keys for same salt")
		}
	}
}
