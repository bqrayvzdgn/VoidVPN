package wireguard

import (
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"
)

func TestBuildIPCConfig(t *testing.T) {
	cfg := &TunnelConfig{
		PrivateKey:          base64.StdEncoding.EncodeToString(make([]byte, 32)),
		PeerPublicKey:       base64.StdEncoding.EncodeToString(make([]byte, 32)),
		PeerEndpoint:        "1.2.3.4:51820",
		PeerAllowedIPs:      []string{"0.0.0.0/0", "::/0"},
		PersistentKeepalive: 25,
	}

	result, err := BuildIPCConfig(cfg)
	if err != nil {
		t.Fatalf("BuildIPCConfig() error: %v", err)
	}

	if !strings.Contains(result, "private_key=") {
		t.Error("IPC config should contain private_key")
	}
	if !strings.Contains(result, "public_key=") {
		t.Error("IPC config should contain public_key")
	}
	if !strings.Contains(result, "endpoint=1.2.3.4:51820") {
		t.Error("IPC config should contain endpoint")
	}
	if !strings.Contains(result, "allowed_ip=0.0.0.0/0") {
		t.Error("IPC config should contain allowed_ip 0.0.0.0/0")
	}
	if !strings.Contains(result, "allowed_ip=::/0") {
		t.Error("IPC config should contain allowed_ip ::/0")
	}
	if !strings.Contains(result, "persistent_keepalive_interval=25") {
		t.Error("IPC config should contain persistent_keepalive_interval")
	}
}

func TestBuildIPCConfigWithPresharedKey(t *testing.T) {
	cfg := &TunnelConfig{
		PrivateKey:       base64.StdEncoding.EncodeToString(make([]byte, 32)),
		PeerPublicKey:    base64.StdEncoding.EncodeToString(make([]byte, 32)),
		PeerPresharedKey: base64.StdEncoding.EncodeToString(make([]byte, 32)),
		PeerEndpoint:     "1.2.3.4:51820",
		PeerAllowedIPs:   []string{"0.0.0.0/0"},
	}

	result, err := BuildIPCConfig(cfg)
	if err != nil {
		t.Fatalf("BuildIPCConfig() error: %v", err)
	}
	if !strings.Contains(result, "preshared_key=") {
		t.Error("IPC config should contain preshared_key when set")
	}
}

func TestKeyToHex(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	b64Key := base64.StdEncoding.EncodeToString(key)

	result, err := keyToHex(b64Key)
	if err != nil {
		t.Fatalf("keyToHex() error: %v", err)
	}
	expected := hex.EncodeToString(key)

	if result != expected {
		t.Errorf("keyToHex() = %q, want %q", result, expected)
	}
}

func TestKeyToHexInvalid(t *testing.T) {
	_, err := keyToHex("not-valid-base64!!!")
	if err == nil {
		t.Error("keyToHex() should return error for invalid input")
	}
}

func TestBuildIPCConfigEndpointNewline(t *testing.T) {
	cfg := &TunnelConfig{
		PrivateKey:     base64.StdEncoding.EncodeToString(make([]byte, 32)),
		PeerPublicKey:  base64.StdEncoding.EncodeToString(make([]byte, 32)),
		PeerEndpoint:   "1.2.3.4:51820\nhack",
		PeerAllowedIPs: []string{"0.0.0.0/0"},
	}
	_, err := BuildIPCConfig(cfg)
	if err == nil {
		t.Fatal("BuildIPCConfig() should reject endpoint with newline")
	}
	if !strings.Contains(err.Error(), "newline") {
		t.Errorf("error = %q, want mention of newline", err.Error())
	}
}

func TestBuildIPCConfigEndpointCR(t *testing.T) {
	cfg := &TunnelConfig{
		PrivateKey:     base64.StdEncoding.EncodeToString(make([]byte, 32)),
		PeerPublicKey:  base64.StdEncoding.EncodeToString(make([]byte, 32)),
		PeerEndpoint:   "1.2.3.4:51820\rhack",
		PeerAllowedIPs: []string{"0.0.0.0/0"},
	}
	_, err := BuildIPCConfig(cfg)
	if err == nil {
		t.Fatal("BuildIPCConfig() should reject endpoint with carriage return")
	}
	if !strings.Contains(err.Error(), "newline") {
		t.Errorf("error = %q, want mention of newline", err.Error())
	}
}

func TestBuildIPCConfigAllowedIPNewline(t *testing.T) {
	cfg := &TunnelConfig{
		PrivateKey:     base64.StdEncoding.EncodeToString(make([]byte, 32)),
		PeerPublicKey:  base64.StdEncoding.EncodeToString(make([]byte, 32)),
		PeerEndpoint:   "1.2.3.4:51820",
		PeerAllowedIPs: []string{"0.0.0.0/0\nhack"},
	}
	_, err := BuildIPCConfig(cfg)
	if err == nil {
		t.Fatal("BuildIPCConfig() should reject AllowedIP with newline")
	}
	if !strings.Contains(err.Error(), "newline") {
		t.Errorf("error = %q, want mention of newline", err.Error())
	}
}

func TestKeyToHexWrongLength(t *testing.T) {
	shortKey := base64.StdEncoding.EncodeToString(make([]byte, 16))
	_, err := keyToHex(shortKey)
	if err == nil {
		t.Fatal("keyToHex() should reject 16-byte key")
	}
	if !strings.Contains(err.Error(), "expected 32 bytes") {
		t.Errorf("error = %q, want mention of expected 32 bytes", err.Error())
	}
}

func TestBuildIPCConfigNoKeepalive(t *testing.T) {
	cfg := &TunnelConfig{
		PrivateKey:          base64.StdEncoding.EncodeToString(make([]byte, 32)),
		PeerPublicKey:       base64.StdEncoding.EncodeToString(make([]byte, 32)),
		PeerEndpoint:        "1.2.3.4:51820",
		PeerAllowedIPs:      []string{"0.0.0.0/0"},
		PersistentKeepalive: 0,
	}
	result, err := BuildIPCConfig(cfg)
	if err != nil {
		t.Fatalf("BuildIPCConfig() error: %v", err)
	}
	if strings.Contains(result, "persistent_keepalive") {
		t.Error("IPC config should not contain persistent_keepalive when set to 0")
	}
}

func TestBuildIPCConfigMultipleAllowedIPs(t *testing.T) {
	cfg := &TunnelConfig{
		PrivateKey:     base64.StdEncoding.EncodeToString(make([]byte, 32)),
		PeerPublicKey:  base64.StdEncoding.EncodeToString(make([]byte, 32)),
		PeerEndpoint:   "1.2.3.4:51820",
		PeerAllowedIPs: []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
	}
	result, err := BuildIPCConfig(cfg)
	if err != nil {
		t.Fatalf("BuildIPCConfig() error: %v", err)
	}
	count := strings.Count(result, "allowed_ip=")
	if count != 3 {
		t.Errorf("allowed_ip= count = %d, want 3", count)
	}
}
