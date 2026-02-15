package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.LogLevel != "info" {
		t.Errorf("expected default log level 'info', got '%s'", cfg.LogLevel)
	}
	if len(cfg.DNSFallback) != 2 {
		t.Errorf("expected 2 default DNS fallback servers, got %d", len(cfg.DNSFallback))
	}
	if cfg.KillSwitch {
		t.Error("expected kill switch to be false by default")
	}
}

func TestConfigGetSet(t *testing.T) {
	cfg := DefaultConfig()

	tests := []struct {
		key   string
		value string
		want  string
	}{
		{"log_level", "debug", "debug"},
		{"default_server", "myserver", "myserver"},
		{"kill_switch", "true", "true"},
		{"auto_connect", "true", "true"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if !cfg.Set(tt.key, tt.value) {
				t.Errorf("Set(%q, %q) returned false", tt.key, tt.value)
			}
			got := cfg.Get(tt.key)
			if got != tt.want {
				t.Errorf("Get(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestConfigSetUnknownKey(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Set("nonexistent", "value") {
		t.Error("Set should return false for unknown key")
	}
	if cfg.Get("nonexistent") != "" {
		t.Error("Get should return empty string for unknown key")
	}
}

func TestConfigSaveLoad(t *testing.T) {
	// Use temp directory
	tmpDir := t.TempDir()
	origConfigDir := os.Getenv("APPDATA")
	os.Setenv("APPDATA", tmpDir)
	defer os.Setenv("APPDATA", origConfigDir)

	cfg := DefaultConfig()
	cfg.LogLevel = "debug"
	cfg.DefaultServer = "test-server"
	cfg.KillSwitch = true

	// Ensure dirs exist
	os.MkdirAll(filepath.Join(tmpDir, "VoidVPN"), 0700)

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.LogLevel != "debug" {
		t.Errorf("loaded LogLevel = %q, want %q", loaded.LogLevel, "debug")
	}
	if loaded.DefaultServer != "test-server" {
		t.Errorf("loaded DefaultServer = %q, want %q", loaded.DefaultServer, "test-server")
	}
	if !loaded.KillSwitch {
		t.Error("loaded KillSwitch should be true")
	}
}

func TestLoadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	origConfigDir := os.Getenv("APPDATA")
	os.Setenv("APPDATA", tmpDir)
	defer os.Setenv("APPDATA", origConfigDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() should not error for non-existent file: %v", err)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("should return default config, got LogLevel=%q", cfg.LogLevel)
	}
}
