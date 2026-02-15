package wireguard

import (
	"testing"

	"github.com/voidvpn/voidvpn/internal/config"
)

func TestNewTunnelFieldMapping(t *testing.T) {
	serverCfg := &config.ServerConfig{
		Name:                "test-server",
		Address:             "10.0.0.2/24",
		DNS:                 []string{"8.8.8.8", "1.1.1.1"},
		MTU:                 1420,
		PublicKey:            "testpubkey",
		Endpoint:            "vpn.example.com:51820",
		AllowedIPs:          []string{"0.0.0.0/0", "::/0"},
		PresharedKey:        "testpsk",
		PersistentKeepalive: 25,
	}

	tun := NewTunnel(serverCfg, "testprivkey")

	if tun == nil {
		t.Fatal("NewTunnel returned nil")
	}
	if tun.config.PrivateKey != "testprivkey" {
		t.Errorf("PrivateKey = %q, want \"testprivkey\"", tun.config.PrivateKey)
	}
	if tun.config.Address != "10.0.0.2/24" {
		t.Errorf("Address = %q, want \"10.0.0.2/24\"", tun.config.Address)
	}
	if len(tun.config.DNS) != 2 {
		t.Errorf("DNS len = %d, want 2", len(tun.config.DNS))
	}
	if tun.config.MTU != 1420 {
		t.Errorf("MTU = %d, want 1420", tun.config.MTU)
	}
	if tun.config.PeerPublicKey != "testpubkey" {
		t.Errorf("PeerPublicKey = %q, want \"testpubkey\"", tun.config.PeerPublicKey)
	}
	if tun.config.PeerEndpoint != "vpn.example.com:51820" {
		t.Errorf("PeerEndpoint = %q, want \"vpn.example.com:51820\"", tun.config.PeerEndpoint)
	}
	if len(tun.config.PeerAllowedIPs) != 2 {
		t.Errorf("PeerAllowedIPs len = %d, want 2", len(tun.config.PeerAllowedIPs))
	}
	if tun.config.PeerPresharedKey != "testpsk" {
		t.Errorf("PeerPresharedKey = %q, want \"testpsk\"", tun.config.PeerPresharedKey)
	}
	if tun.config.PersistentKeepalive != 25 {
		t.Errorf("PersistentKeepalive = %d, want 25", tun.config.PersistentKeepalive)
	}
	if tun.server != serverCfg {
		t.Error("server reference mismatch")
	}
}

func TestNewTunnelIsNotActive(t *testing.T) {
	tun := NewTunnel(&config.ServerConfig{Name: "test"}, "key")
	if tun.IsActive() {
		t.Error("new tunnel should not be active")
	}
}

func TestTunnelStatusNotConnected(t *testing.T) {
	tun := NewTunnel(&config.ServerConfig{
		Name:     "test",
		Endpoint: "1.2.3.4:51820",
		Address:  "10.0.0.2/24",
	}, "key")

	status, err := tun.Status()
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}
	if status.Connected {
		t.Error("status.Connected should be false")
	}
	if status.Protocol != "wireguard" {
		t.Errorf("Protocol = %q, want \"wireguard\"", status.Protocol)
	}
	if status.ServerName != "test" {
		t.Errorf("ServerName = %q, want \"test\"", status.ServerName)
	}
}

func TestTunnelDisconnectNilDevice(t *testing.T) {
	tun := NewTunnel(&config.ServerConfig{Name: "test"}, "key")
	err := tun.Disconnect()
	if err != nil {
		t.Errorf("Disconnect() with nil device should not error: %v", err)
	}
}
