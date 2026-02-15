package openvpn

import (
	"strings"
	"testing"

	"github.com/voidvpn/voidvpn/internal/config"
)

func TestBuildOVPNConfigProtoTCP(t *testing.T) {
	cfg := &config.ServerConfig{
		Endpoint: "vpn.example.com:443",
		Proto:    "tcp",
	}
	result := BuildOVPNConfig(cfg, 12345)
	if !strings.Contains(result, "proto tcp") {
		t.Error("should contain 'proto tcp'")
	}
}

func TestBuildOVPNConfigDefaultProto(t *testing.T) {
	cfg := &config.ServerConfig{
		Endpoint: "vpn.example.com:1194",
	}
	result := BuildOVPNConfig(cfg, 12345)
	if !strings.Contains(result, "proto udp") {
		t.Error("should default to 'proto udp'")
	}
}

func TestBuildOVPNConfigCipher(t *testing.T) {
	cfg := &config.ServerConfig{
		Endpoint: "1.2.3.4:1194",
		Cipher:   "AES-256-GCM",
	}
	result := BuildOVPNConfig(cfg, 12345)
	if !strings.Contains(result, "cipher AES-256-GCM") {
		t.Error("should contain cipher")
	}
}

func TestBuildOVPNConfigAuth(t *testing.T) {
	cfg := &config.ServerConfig{
		Endpoint: "1.2.3.4:1194",
		Auth:     "SHA256",
	}
	result := BuildOVPNConfig(cfg, 12345)
	if !strings.Contains(result, "auth SHA256") {
		t.Error("should contain auth")
	}
}

func TestBuildOVPNConfigCompLZO(t *testing.T) {
	cfg := &config.ServerConfig{
		Endpoint: "1.2.3.4:1194",
		CompLZO:  true,
	}
	result := BuildOVPNConfig(cfg, 12345)
	if !strings.Contains(result, "comp-lzo") {
		t.Error("should contain comp-lzo")
	}
}

func TestBuildOVPNConfigNoCipher(t *testing.T) {
	cfg := &config.ServerConfig{
		Endpoint: "1.2.3.4:1194",
	}
	result := BuildOVPNConfig(cfg, 12345)
	if strings.Contains(result, "cipher") {
		t.Error("should not contain cipher when empty")
	}
}

func TestBuildOVPNConfigNoAuth(t *testing.T) {
	cfg := &config.ServerConfig{
		Endpoint: "1.2.3.4:1194",
	}
	result := BuildOVPNConfig(cfg, 12345)
	if strings.Contains(result, "auth") {
		t.Error("should not contain auth when empty")
	}
}

func TestBuildOVPNConfigNoCompLZO(t *testing.T) {
	cfg := &config.ServerConfig{
		Endpoint: "1.2.3.4:1194",
	}
	result := BuildOVPNConfig(cfg, 12345)
	if strings.Contains(result, "comp-lzo") {
		t.Error("should not contain comp-lzo when false")
	}
}

func TestBuildOVPNConfigTLSAuth(t *testing.T) {
	cfg := &config.ServerConfig{
		Endpoint: "1.2.3.4:1194",
		TLSAuth:  "fake-tls-key-data",
	}
	result := BuildOVPNConfig(cfg, 12345)
	if !strings.Contains(result, "<tls-auth>") {
		t.Error("should contain tls-auth block")
	}
	if !strings.Contains(result, "key-direction 1") {
		t.Error("should contain key-direction")
	}
}

func TestBuildOVPNConfigManagementPort(t *testing.T) {
	cfg := &config.ServerConfig{
		Endpoint: "1.2.3.4:1194",
	}
	result := BuildOVPNConfig(cfg, 55555)
	if !strings.Contains(result, "management 127.0.0.1 55555") {
		t.Error("should contain management with correct port")
	}
}

func TestParseEndpointNoPort(t *testing.T) {
	host, port := parseEndpoint("vpn.example.com")
	if host != "vpn.example.com" {
		t.Errorf("host = %q, want \"vpn.example.com\"", host)
	}
	if port != "1194" {
		t.Errorf("port = %q, want \"1194\"", port)
	}
}

func TestParseEndpointWithPort(t *testing.T) {
	host, port := parseEndpoint("vpn.example.com:443")
	if host != "vpn.example.com" {
		t.Errorf("host = %q, want \"vpn.example.com\"", host)
	}
	if port != "443" {
		t.Errorf("port = %q, want \"443\"", port)
	}
}

func TestParseEndpointIPv6(t *testing.T) {
	host, port := parseEndpoint("[2001:db8::1]:1194")
	// LastIndex of ":" finds the last colon
	if port != "1194" {
		t.Errorf("port = %q, want \"1194\"", port)
	}
	if host != "[2001:db8::1]" {
		t.Errorf("host = %q, want \"[2001:db8::1]\"", host)
	}
}
