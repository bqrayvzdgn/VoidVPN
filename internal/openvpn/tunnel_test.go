package openvpn

import (
	"strings"
	"testing"

	"github.com/voidvpn/voidvpn/internal/config"
)

func TestBuildOVPNConfig(t *testing.T) {
	cfg := &config.ServerConfig{
		Name:     "test-server",
		Protocol: "openvpn",
		Endpoint: "vpn.example.com:1194",
		Proto:    "udp",
		Cipher:   "AES-256-GCM",
		Auth:     "SHA256",
		CompLZO:  true,
		CACert:   "-----BEGIN CERTIFICATE-----\nFAKECA\n-----END CERTIFICATE-----",
		ClientCert: "-----BEGIN CERTIFICATE-----\nFAKECERT\n-----END CERTIFICATE-----",
		ClientKey:  "-----BEGIN PRIVATE KEY-----\nFAKEKEY\n-----END PRIVATE KEY-----",
		TLSAuth:   "-----BEGIN OpenVPN Static key V1-----\nFAKETLS\n-----END OpenVPN Static key V1-----",
	}

	result := BuildOVPNConfig(cfg, 12345)

	checks := []string{
		"client",
		"dev tun",
		"proto udp",
		"remote vpn.example.com 1194",
		"cipher AES-256-GCM",
		"auth SHA256",
		"comp-lzo",
		"management 127.0.0.1 12345",
		"<ca>",
		"FAKECA",
		"</ca>",
		"<cert>",
		"FAKECERT",
		"</cert>",
		"<key>",
		"FAKEKEY",
		"</key>",
		"<tls-auth>",
		"FAKETLS",
		"</tls-auth>",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("BuildOVPNConfig output missing %q", check)
		}
	}
}

func TestBuildOVPNConfigMinimal(t *testing.T) {
	cfg := &config.ServerConfig{
		Name:     "minimal",
		Protocol: "openvpn",
		Endpoint: "1.2.3.4:443",
		Proto:    "tcp",
	}

	result := BuildOVPNConfig(cfg, 9999)

	if !strings.Contains(result, "proto tcp") {
		t.Error("expected 'proto tcp'")
	}
	if !strings.Contains(result, "remote 1.2.3.4 443") {
		t.Error("expected 'remote 1.2.3.4 443'")
	}
	if strings.Contains(result, "comp-lzo") {
		t.Error("comp-lzo should not be present when CompLZO is false")
	}
	if strings.Contains(result, "<ca>") {
		t.Error("<ca> block should not be present when CACert is empty")
	}
}

func TestDetectOpenVPN(t *testing.T) {
	// This test just verifies DetectOpenVPN returns a path or an error
	// without crashing. It doesn't require openvpn to be installed.
	path, err := DetectOpenVPN()
	if err != nil {
		// Expected on systems without openvpn
		if !strings.Contains(err.Error(), "openvpn binary not found") {
			t.Errorf("unexpected error: %v", err)
		}
		return
	}
	if path == "" {
		t.Error("DetectOpenVPN returned empty path with no error")
	}
}

func TestParseEndpoint(t *testing.T) {
	tests := []struct {
		input    string
		wantHost string
		wantPort string
	}{
		{"vpn.example.com:1194", "vpn.example.com", "1194"},
		{"1.2.3.4:443", "1.2.3.4", "443"},
		{"noport.example.com", "noport.example.com", "1194"},
	}

	for _, tt := range tests {
		host, port := parseEndpoint(tt.input)
		if host != tt.wantHost || port != tt.wantPort {
			t.Errorf("parseEndpoint(%q) = (%q, %q), want (%q, %q)",
				tt.input, host, port, tt.wantHost, tt.wantPort)
		}
	}
}

func TestNewTunnel(t *testing.T) {
	cfg := &config.ServerConfig{
		Name:     "test",
		Protocol: "openvpn",
		Endpoint: "vpn.example.com:1194",
	}

	tun := NewTunnel(cfg)
	if tun == nil {
		t.Fatal("NewTunnel returned nil")
	}
	if tun.server != cfg {
		t.Error("tunnel server config mismatch")
	}
	if tun.mgmtPort < 10000 || tun.mgmtPort >= 60000 {
		t.Errorf("unexpected management port: %d", tun.mgmtPort)
	}
	if !tun.IsActive() == true {
		// Should not be active before connecting
	}
	if tun.IsActive() {
		t.Error("tunnel should not be active before Connect()")
	}
}
