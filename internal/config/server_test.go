package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestEnv(t *testing.T) func() {
	t.Helper()
	tmpDir := t.TempDir()
	origConfigDir := os.Getenv("APPDATA")
	os.Setenv("APPDATA", tmpDir)
	EnsureDirs()
	return func() {
		os.Setenv("APPDATA", origConfigDir)
	}
}

func TestServerCRUD(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	server := DefaultServerConfig()
	server.Name = "test-server"
	server.Endpoint = "vpn.example.com:51820"
	server.PublicKey = "dGVzdHB1YmxpY2tleQ=="
	server.Address = "10.0.0.2/24"

	// Create
	if err := SaveServer(server); err != nil {
		t.Fatalf("SaveServer() error: %v", err)
	}

	// Read
	loaded, err := LoadServer("test-server")
	if err != nil {
		t.Fatalf("LoadServer() error: %v", err)
	}
	if loaded.Endpoint != "vpn.example.com:51820" {
		t.Errorf("loaded Endpoint = %q, want %q", loaded.Endpoint, "vpn.example.com:51820")
	}
	if loaded.PublicKey != "dGVzdHB1YmxpY2tleQ==" {
		t.Errorf("loaded PublicKey doesn't match")
	}

	// Exists
	if !ServerExists("test-server") {
		t.Error("ServerExists should return true")
	}
	if ServerExists("nonexistent") {
		t.Error("ServerExists should return false for nonexistent server")
	}

	// List
	servers, err := ListServers()
	if err != nil {
		t.Fatalf("ListServers() error: %v", err)
	}
	if len(servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(servers))
	}

	// Delete
	if err := RemoveServer("test-server"); err != nil {
		t.Fatalf("RemoveServer() error: %v", err)
	}
	if ServerExists("test-server") {
		t.Error("server should be removed")
	}
}

func TestLoadServerNotFound(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	_, err := LoadServer("nonexistent")
	if err == nil {
		t.Error("LoadServer should return error for nonexistent server")
	}
}

func TestRemoveServerNotFound(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	err := RemoveServer("nonexistent")
	if err == nil {
		t.Error("RemoveServer should return error for nonexistent server")
	}
}

func TestValidateName(t *testing.T) {
	valid := []string{"server1", "My Server", strings.Repeat("a", 63)}
	for _, name := range valid {
		if err := ValidateName(name); err != nil {
			t.Errorf("ValidateName(%q) unexpected error: %v", name, err)
		}
	}

	invalid := []string{"", "../x", strings.Repeat("a", 64), "a/b"}
	for _, name := range invalid {
		if err := ValidateName(name); err == nil {
			t.Errorf("ValidateName(%q) should return error", name)
		}
	}
}

func TestServerFileNormalization(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	path, err := serverFile("My Server")
	if err != nil {
		t.Fatalf("serverFile() error: %v", err)
	}
	if !strings.Contains(path, "my-server.yaml") {
		t.Errorf("serverFile(\"My Server\") = %q, want path containing my-server.yaml", path)
	}
}

func TestServerFileTraversal(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	path, err := serverFile("test-server")
	if err != nil {
		t.Fatalf("serverFile() error: %v", err)
	}
	sdir := filepath.Clean(ServersDir())
	if !strings.HasPrefix(filepath.Clean(path), sdir) {
		t.Errorf("serverFile() path %q not under ServersDir %q", path, sdir)
	}
}

func TestDefaultServerConfig(t *testing.T) {
	cfg := DefaultServerConfig()
	if cfg.MTU != 1420 {
		t.Errorf("default MTU = %d, want 1420", cfg.MTU)
	}
	if cfg.PersistentKeepalive != 25 {
		t.Errorf("default PersistentKeepalive = %d, want 25", cfg.PersistentKeepalive)
	}
	if len(cfg.AllowedIPs) != 2 {
		t.Errorf("default AllowedIPs count = %d, want 2", len(cfg.AllowedIPs))
	}
}

func TestLoadServerInvalidYAML(t *testing.T) {
	origAppdata := os.Getenv("APPDATA")
	tmp := t.TempDir()
	os.Setenv("APPDATA", tmp)
	defer os.Setenv("APPDATA", origAppdata)

	EnsureDirs()

	// Write invalid YAML
	serverPath := filepath.Join(ServersDir(), "bad-yaml.yaml")
	os.WriteFile(serverPath, []byte("{{{{invalid yaml!!!!"), 0644)

	_, err := LoadServer("bad-yaml")
	if err == nil {
		t.Error("LoadServer() should fail on invalid YAML")
	}
}

func TestListServersEmpty(t *testing.T) {
	origAppdata := os.Getenv("APPDATA")
	tmp := t.TempDir()
	os.Setenv("APPDATA", tmp)
	defer os.Setenv("APPDATA", origAppdata)

	EnsureDirs()

	servers, err := ListServers()
	if err != nil {
		t.Fatalf("ListServers() error: %v", err)
	}
	if len(servers) != 0 {
		t.Errorf("ListServers() = %d servers, want 0", len(servers))
	}
}

func TestListServersSkipsInvalidYAML(t *testing.T) {
	origAppdata := os.Getenv("APPDATA")
	tmp := t.TempDir()
	os.Setenv("APPDATA", tmp)
	defer os.Setenv("APPDATA", origAppdata)

	EnsureDirs()

	// Save one valid server
	SaveServer(&ServerConfig{Name: "valid-server", Endpoint: "1.2.3.4:51820"})

	// Write one invalid YAML file
	invalidPath := filepath.Join(ServersDir(), "invalid.yaml")
	os.WriteFile(invalidPath, []byte("{{{{not yaml"), 0644)

	servers, err := ListServers()
	if err != nil {
		t.Fatalf("ListServers() error: %v", err)
	}
	// Should have at least the valid server
	if len(servers) < 1 {
		t.Errorf("ListServers() = %d, want >= 1", len(servers))
	}
}

func TestServerExistsTrue(t *testing.T) {
	origAppdata := os.Getenv("APPDATA")
	tmp := t.TempDir()
	os.Setenv("APPDATA", tmp)
	defer os.Setenv("APPDATA", origAppdata)

	EnsureDirs()
	SaveServer(&ServerConfig{Name: "existing"})

	if !ServerExists("existing") {
		t.Error("ServerExists() should return true for saved server")
	}
}

func TestServerExistsFalse(t *testing.T) {
	origAppdata := os.Getenv("APPDATA")
	tmp := t.TempDir()
	os.Setenv("APPDATA", tmp)
	defer os.Setenv("APPDATA", origAppdata)

	EnsureDirs()

	if ServerExists("nonexistent") {
		t.Error("ServerExists() should return false for missing server")
	}
}
