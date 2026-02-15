package daemon

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/voidvpn/voidvpn/internal/config"
	"github.com/voidvpn/voidvpn/internal/tunnel"
)

// Mock implementations
type mockTunnel struct {
	connectErr    error
	disconnectErr error
	statusResp    *tunnel.TunnelStatus
	statusErr     error
	active        bool
	connected     bool
	disconnected  bool
}

func (m *mockTunnel) Connect(ctx context.Context) error {
	m.connected = true
	if m.connectErr != nil {
		return m.connectErr
	}
	return nil
}

func (m *mockTunnel) Disconnect() error {
	m.disconnected = true
	return m.disconnectErr
}

func (m *mockTunnel) Status() (*tunnel.TunnelStatus, error) {
	if m.statusErr != nil {
		return nil, m.statusErr
	}
	if m.statusResp != nil {
		return m.statusResp, nil
	}
	return &tunnel.TunnelStatus{InterfaceName: "test0"}, nil
}

func (m *mockTunnel) IsActive() bool {
	return m.active
}

type mockDNS struct {
	setErr     error
	restoreErr error
	setCalled  bool
	restored   bool
}

func (m *mockDNS) Set(iface string, servers []string) error {
	m.setCalled = true
	return m.setErr
}

func (m *mockDNS) Restore() error {
	m.restored = true
	return m.restoreErr
}

type mockRoutes struct {
	addErr    error
	removeErr error
	added     bool
	removed   bool
}

func (m *mockRoutes) AddVPNRoutes(iface string, endpoint string, ipv6 bool) error {
	m.added = true
	return m.addErr
}

func (m *mockRoutes) RemoveVPNRoutes() error {
	m.removed = true
	return m.removeErr
}

func setupDaemonTest(t *testing.T) func() {
	t.Helper()
	origAppdata := os.Getenv("APPDATA")
	tmp := t.TempDir()
	os.Setenv("APPDATA", tmp)
	return func() {
		os.Setenv("APPDATA", origAppdata)
	}
}

func TestNewDaemon(t *testing.T) {
	tun := &mockTunnel{}
	server := &config.ServerConfig{Name: "test-server"}
	d := New(tun, server)

	if d == nil {
		t.Fatal("New() returned nil")
	}
	if d.Connected == nil {
		t.Error("Connected channel not initialized")
	}
	if d.tunnel == nil {
		t.Error("tunnel not set")
	}
	if d.server != server {
		t.Error("server not set correctly")
	}
}

func TestHandleIPCStatus(t *testing.T) {
	cleanup := setupDaemonTest(t)
	defer cleanup()

	tun := &mockTunnel{
		statusResp: &tunnel.TunnelStatus{
			TxBytes: 1000,
			RxBytes: 2000,
		},
	}
	server := &config.ServerConfig{Name: "test"}
	d := &Daemon{tunnel: tun, server: server}

	// Save state first so LoadState succeeds
	state := &ConnectionState{Server: "test", PID: os.Getpid()}
	SaveState(state)

	req := &IPCRequest{Command: "status"}
	resp := d.handleIPC(req)

	if !resp.Success {
		t.Errorf("status command failed: %s", resp.Error)
	}
	if resp.State == nil {
		t.Error("response state is nil")
	}
}

func TestHandleIPCStatusNoState(t *testing.T) {
	cleanup := setupDaemonTest(t)
	defer cleanup()

	tun := &mockTunnel{}
	server := &config.ServerConfig{Name: "test"}
	d := &Daemon{tunnel: tun, server: server}

	req := &IPCRequest{Command: "status"}
	resp := d.handleIPC(req)

	if resp.Success {
		t.Error("status should fail when no state file exists")
	}
}

func TestHandleIPCDisconnect(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d := &Daemon{cancel: cancel}

	req := &IPCRequest{Command: "disconnect"}
	resp := d.handleIPC(req)

	if !resp.Success {
		t.Error("disconnect should succeed")
	}

	select {
	case <-ctx.Done():
		// Expected — cancel was called
	default:
		t.Error("context should have been cancelled by disconnect")
	}
}

func TestHandleIPCDisconnectNilCancel(t *testing.T) {
	d := &Daemon{}

	req := &IPCRequest{Command: "disconnect"}
	resp := d.handleIPC(req)

	if !resp.Success {
		t.Error("disconnect with nil cancel should still succeed")
	}
}

func TestHandleIPCUnknownCommand(t *testing.T) {
	d := &Daemon{}

	req := &IPCRequest{Command: "invalid"}
	resp := d.handleIPC(req)

	if resp.Success {
		t.Error("unknown command should fail")
	}
	if !strings.Contains(resp.Error, "unknown") {
		t.Errorf("error should contain 'unknown': %s", resp.Error)
	}
}

func TestHandleIPCStatusUpdatesTraffic(t *testing.T) {
	cleanup := setupDaemonTest(t)
	defer cleanup()

	tun := &mockTunnel{
		statusResp: &tunnel.TunnelStatus{
			TxBytes: 5000,
			RxBytes: 10000,
		},
	}
	server := &config.ServerConfig{Name: "test"}
	d := &Daemon{tunnel: tun, server: server}

	state := &ConnectionState{Server: "test", PID: os.Getpid()}
	SaveState(state)

	resp := d.handleIPC(&IPCRequest{Command: "status"})
	if !resp.Success {
		t.Fatalf("status failed: %s", resp.Error)
	}
	if resp.State.TxBytes != 5000 {
		t.Errorf("TxBytes = %d, want 5000", resp.State.TxBytes)
	}
	if resp.State.RxBytes != 10000 {
		t.Errorf("RxBytes = %d, want 10000", resp.State.RxBytes)
	}
}

func TestCleanupCallsAll(t *testing.T) {
	cleanup := setupDaemonTest(t)
	defer cleanup()

	tun := &mockTunnel{}
	dns := &mockDNS{}
	routes := &mockRoutes{}

	d := &Daemon{
		tunnel: tun,
		server: &config.ServerConfig{Name: "test"},
		dns:    dns,
		routes: routes,
	}

	d.cleanup()

	if !tun.disconnected {
		t.Error("tunnel.Disconnect() not called")
	}
	if !dns.restored {
		t.Error("dns.Restore() not called")
	}
	if !routes.removed {
		t.Error("routes.RemoveVPNRoutes() not called")
	}
}

func TestLoadStateInvalidJSON(t *testing.T) {
	cleanup := setupDaemonTest(t)
	defer cleanup()

	// Write invalid JSON
	config.EnsureDirs()
	os.WriteFile(config.StateFile(), []byte("{invalid json!!!}"), 0644)

	_, err := LoadState()
	if err == nil {
		t.Error("LoadState() should fail on invalid JSON")
	}
}

func TestRunTunnelConnectError(t *testing.T) {
	cleanup := setupDaemonTest(t)
	defer cleanup()

	tun := &mockTunnel{
		connectErr: fmt.Errorf("connection refused"),
	}
	server := &config.ServerConfig{Name: "test", Protocol: "openvpn"}

	d := &Daemon{
		tunnel:    tun,
		server:    server,
		dns:       &mockDNS{},
		routes:    &mockRoutes{},
		Connected: make(chan struct{}),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := d.Run(ctx)
	if err == nil {
		t.Error("Run() should return error when tunnel.Connect fails")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("error should contain 'connection refused': %v", err)
	}
}
