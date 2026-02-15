package daemon

import (
	"os"
	"testing"
	"time"

	"github.com/voidvpn/voidvpn/internal/config"
)

func setupStateTest(t *testing.T) func() {
	t.Helper()
	tmpDir := t.TempDir()
	orig := os.Getenv("APPDATA")
	os.Setenv("APPDATA", tmpDir)
	config.EnsureDirs()
	return func() {
		os.Setenv("APPDATA", orig)
	}
}

func TestSaveLoadState(t *testing.T) {
	cleanup := setupStateTest(t)
	defer cleanup()

	state := &ConnectionState{
		Server:        "test-server",
		ConnectedAt:   time.Now().Truncate(time.Second),
		InterfaceName: "voidvpn0",
		TunnelIP:      "10.0.0.2/24",
		Endpoint:      "1.2.3.4:51820",
		PID:           12345,
		TxBytes:       1024,
		RxBytes:       2048,
	}

	if err := SaveState(state); err != nil {
		t.Fatalf("SaveState() error: %v", err)
	}

	loaded, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState() error: %v", err)
	}

	if loaded.Server != "test-server" {
		t.Errorf("Server = %q, want %q", loaded.Server, "test-server")
	}
	if loaded.PID != 12345 {
		t.Errorf("PID = %d, want %d", loaded.PID, 12345)
	}
	if loaded.TxBytes != 1024 {
		t.Errorf("TxBytes = %d, want %d", loaded.TxBytes, 1024)
	}
}

func TestIsConnected(t *testing.T) {
	cleanup := setupStateTest(t)
	defer cleanup()

	if IsConnected() {
		t.Error("should not be connected initially")
	}

	// Use current process PID so the alive check passes
	state := &ConnectionState{Server: "test", PID: os.Getpid()}
	SaveState(state)

	if !IsConnected() {
		t.Error("should be connected after saving state")
	}

	ClearState()
	if IsConnected() {
		t.Error("should not be connected after clearing state")
	}
}

func TestIsConnectedStalePID(t *testing.T) {
	cleanup := setupStateTest(t)
	defer cleanup()

	// Save state with a PID that is not running (0 is never a valid process)
	state := &ConnectionState{Server: "test", PID: 0}
	SaveState(state)

	// IsConnected should detect the dead PID and clean up
	if IsConnected() {
		t.Error("should not be connected when PID is dead")
	}

	// State file should have been cleaned up
	if _, err := LoadState(); err == nil {
		t.Error("state file should have been removed for dead PID")
	}
}

func TestClearStateNonExistent(t *testing.T) {
	cleanup := setupStateTest(t)
	defer cleanup()

	if err := ClearState(); err != nil {
		t.Errorf("ClearState() should not error on non-existent file: %v", err)
	}
}

func TestLoadStateNonExistent(t *testing.T) {
	cleanup := setupStateTest(t)
	defer cleanup()

	_, err := LoadState()
	if err == nil {
		t.Error("LoadState() should error on non-existent file")
	}
}
