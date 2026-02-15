package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigDirUsesAppdata(t *testing.T) {
	origAppdata := os.Getenv("APPDATA")
	tmp := t.TempDir()
	os.Setenv("APPDATA", tmp)
	defer os.Setenv("APPDATA", origAppdata)

	dir := ConfigDir()
	if !strings.HasPrefix(dir, tmp) {
		t.Errorf("ConfigDir() = %q, should start with %q", dir, tmp)
	}
	if !strings.HasSuffix(dir, "VoidVPN") {
		t.Errorf("ConfigDir() = %q, should end with 'VoidVPN'", dir)
	}
}

func TestConfigFile(t *testing.T) {
	f := ConfigFile()
	if !strings.HasSuffix(f, "config.yaml") {
		t.Errorf("ConfigFile() = %q, should end with 'config.yaml'", f)
	}
}

func TestServersDir(t *testing.T) {
	d := ServersDir()
	if !strings.HasSuffix(d, "servers") {
		t.Errorf("ServersDir() = %q, should end with 'servers'", d)
	}
}

func TestStateDir(t *testing.T) {
	d := StateDir()
	if !strings.HasSuffix(d, "state") {
		t.Errorf("StateDir() = %q, should end with 'state'", d)
	}
}

func TestStateFile(t *testing.T) {
	f := StateFile()
	if !strings.HasSuffix(f, "connection.json") {
		t.Errorf("StateFile() = %q, should end with 'connection.json'", f)
	}
}

func TestEnsureDirsCreatesDirectories(t *testing.T) {
	origAppdata := os.Getenv("APPDATA")
	tmp := t.TempDir()
	os.Setenv("APPDATA", tmp)
	defer os.Setenv("APPDATA", origAppdata)

	err := EnsureDirs()
	if err != nil {
		t.Fatalf("EnsureDirs() error: %v", err)
	}

	// Verify all dirs were created
	for _, dir := range []string{ConfigDir(), ServersDir(), StateDir()} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("directory %q not created: %v", dir, err)
		} else if !info.IsDir() {
			t.Errorf("%q is not a directory", dir)
		}
	}
}

func TestEnsureDirsIdempotent(t *testing.T) {
	origAppdata := os.Getenv("APPDATA")
	tmp := t.TempDir()
	os.Setenv("APPDATA", tmp)
	defer os.Setenv("APPDATA", origAppdata)

	// Call twice -- should succeed both times
	if err := EnsureDirs(); err != nil {
		t.Fatalf("first EnsureDirs() error: %v", err)
	}
	if err := EnsureDirs(); err != nil {
		t.Fatalf("second EnsureDirs() error: %v", err)
	}
}

func TestConfigDirContainsVoidVPN(t *testing.T) {
	dir := ConfigDir()
	if !strings.Contains(filepath.Base(dir), "oidVPN") && !strings.Contains(filepath.Base(dir), "voidvpn") {
		t.Errorf("ConfigDir() base = %q, should contain 'VoidVPN' or 'voidvpn'", filepath.Base(dir))
	}
}
