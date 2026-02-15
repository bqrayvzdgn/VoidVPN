package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportWireGuardConfig(t *testing.T) {
	tmpDir := t.TempDir()

	confContent := `[Interface]
PrivateKey = yNGmpMvlEWbSI1iqKVlHPBXMRTf5pPjAi0CE5vIVp0I=
Address = 10.0.0.2/24
DNS = 1.1.1.1, 8.8.8.8
MTU = 1380

[Peer]
PublicKey = xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=
Endpoint = vpn.example.com:51820
AllowedIPs = 0.0.0.0/0, ::/0
PersistentKeepalive = 25
`
	confPath := filepath.Join(tmpDir, "myserver.conf")
	if err := os.WriteFile(confPath, []byte(confContent), 0600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	server, privateKey, err := ImportWireGuardConfig(confPath)
	if err != nil {
		t.Fatalf("ImportWireGuardConfig() error: %v", err)
	}

	if server.Name != "myserver" {
		t.Errorf("Name = %q, want %q", server.Name, "myserver")
	}
	if server.Endpoint != "vpn.example.com:51820" {
		t.Errorf("Endpoint = %q, want %q", server.Endpoint, "vpn.example.com:51820")
	}
	if server.PublicKey != "xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=" {
		t.Errorf("PublicKey doesn't match")
	}
	if server.Address != "10.0.0.2/24" {
		t.Errorf("Address = %q, want %q", server.Address, "10.0.0.2/24")
	}
	if server.MTU != 1380 {
		t.Errorf("MTU = %d, want %d", server.MTU, 1380)
	}
	if len(server.DNS) != 2 || server.DNS[0] != "1.1.1.1" {
		t.Errorf("DNS = %v, want [1.1.1.1 8.8.8.8]", server.DNS)
	}
	if privateKey != "yNGmpMvlEWbSI1iqKVlHPBXMRTf5pPjAi0CE5vIVp0I=" {
		t.Errorf("PrivateKey doesn't match")
	}
	if server.PersistentKeepalive != 25 {
		t.Errorf("PersistentKeepalive = %d, want 25", server.PersistentKeepalive)
	}
}

func TestImportMissingPeer(t *testing.T) {
	tmpDir := t.TempDir()
	confContent := `[Interface]
PrivateKey = yNGmpMvlEWbSI1iqKVlHPBXMRTf5pPjAi0CE5vIVp0I=
Address = 10.0.0.2/24
`
	confPath := filepath.Join(tmpDir, "noendpoint.conf")
	os.WriteFile(confPath, []byte(confContent), 0600)

	_, _, err := ImportWireGuardConfig(confPath)
	if err == nil {
		t.Error("should error when Peer Endpoint is missing")
	}
}

func TestImportInvalidFile(t *testing.T) {
	_, _, err := ImportWireGuardConfig("/nonexistent/path.conf")
	if err == nil {
		t.Error("should error for nonexistent file")
	}
}

func TestImportOpenVPNConfig(t *testing.T) {
	tmpDir := t.TempDir()

	ovpnContent := `client
dev tun
proto udp
remote vpn.example.com 1194
cipher AES-256-GCM
auth SHA256
comp-lzo
resolv-retry infinite
nobind
persist-key
persist-tun
verb 3

<ca>
-----BEGIN CERTIFICATE-----
MIIBfakecacert
-----END CERTIFICATE-----
</ca>

<cert>
-----BEGIN CERTIFICATE-----
MIIBfakeclientcert
-----END CERTIFICATE-----
</cert>

<key>
-----BEGIN PRIVATE KEY-----
MIIBfakeclientkey
-----END PRIVATE KEY-----
</key>

<tls-auth>
-----BEGIN OpenVPN Static key V1-----
faketlsauthkey
-----END OpenVPN Static key V1-----
</tls-auth>
`
	ovpnPath := filepath.Join(tmpDir, "myovpn.ovpn")
	if err := os.WriteFile(ovpnPath, []byte(ovpnContent), 0600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	server, err := ImportOpenVPNConfig(ovpnPath)
	if err != nil {
		t.Fatalf("ImportOpenVPNConfig() error: %v", err)
	}

	if server.Name != "myovpn" {
		t.Errorf("Name = %q, want %q", server.Name, "myovpn")
	}
	if server.Protocol != "openvpn" {
		t.Errorf("Protocol = %q, want %q", server.Protocol, "openvpn")
	}
	if server.Endpoint != "vpn.example.com:1194" {
		t.Errorf("Endpoint = %q, want %q", server.Endpoint, "vpn.example.com:1194")
	}
	if server.Proto != "udp" {
		t.Errorf("Proto = %q, want %q", server.Proto, "udp")
	}
	if server.Cipher != "AES-256-GCM" {
		t.Errorf("Cipher = %q, want %q", server.Cipher, "AES-256-GCM")
	}
	if server.Auth != "SHA256" {
		t.Errorf("Auth = %q, want %q", server.Auth, "SHA256")
	}
	if !server.CompLZO {
		t.Error("CompLZO should be true")
	}
	if server.CACert == "" {
		t.Error("CACert should not be empty")
	}
	if server.ClientCert == "" {
		t.Error("ClientCert should not be empty")
	}
	if server.ClientKey == "" {
		t.Error("ClientKey should not be empty")
	}
	if server.TLSAuth == "" {
		t.Error("TLSAuth should not be empty")
	}
	if server.RemotePort != 1194 {
		t.Errorf("RemotePort = %d, want 1194", server.RemotePort)
	}
}

func TestImportOpenVPNNoRemote(t *testing.T) {
	tmpDir := t.TempDir()
	ovpnContent := `client
dev tun
proto udp
`
	ovpnPath := filepath.Join(tmpDir, "noremote.ovpn")
	os.WriteFile(ovpnPath, []byte(ovpnContent), 0600)

	_, err := ImportOpenVPNConfig(ovpnPath)
	if err == nil {
		t.Error("should error when remote directive is missing")
	}
}

func TestImportOpenVPNInvalidFile(t *testing.T) {
	_, err := ImportOpenVPNConfig("/nonexistent/path.ovpn")
	if err == nil {
		t.Error("should error for nonexistent file")
	}
}

func TestImportWireGuardMissingAddress(t *testing.T) {
	tmpDir := t.TempDir()
	confContent := `[Interface]
PrivateKey = yNGmpMvlEWbSI1iqKVlHPBXMRTf5pPjAi0CE5vIVp0I=

[Peer]
PublicKey = xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=
Endpoint = vpn.example.com:51820
`
	confPath := filepath.Join(tmpDir, "noaddr.conf")
	os.WriteFile(confPath, []byte(confContent), 0600)

	_, _, err := ImportWireGuardConfig(confPath)
	if err == nil {
		t.Error("should error when Address is missing")
	}
	if !strings.Contains(err.Error(), "Address") {
		t.Errorf("error = %q, want mention of Address", err.Error())
	}
}

func TestImportWireGuardMissingPublicKey(t *testing.T) {
	tmpDir := t.TempDir()
	confContent := `[Interface]
PrivateKey = yNGmpMvlEWbSI1iqKVlHPBXMRTf5pPjAi0CE5vIVp0I=
Address = 10.0.0.2/24

[Peer]
Endpoint = vpn.example.com:51820
`
	confPath := filepath.Join(tmpDir, "nopub.conf")
	os.WriteFile(confPath, []byte(confContent), 0600)

	_, _, err := ImportWireGuardConfig(confPath)
	if err == nil {
		t.Error("should error when PublicKey is missing")
	}
	if !strings.Contains(err.Error(), "PublicKey") {
		t.Errorf("error = %q, want mention of PublicKey", err.Error())
	}
}

func TestImportWireGuardInvalidMTU(t *testing.T) {
	tmpDir := t.TempDir()
	confContent := `[Interface]
PrivateKey = yNGmpMvlEWbSI1iqKVlHPBXMRTf5pPjAi0CE5vIVp0I=
Address = 10.0.0.2/24
MTU = abc

[Peer]
PublicKey = xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=
Endpoint = vpn.example.com:51820
`
	confPath := filepath.Join(tmpDir, "badmtu.conf")
	os.WriteFile(confPath, []byte(confContent), 0600)

	server, _, err := ImportWireGuardConfig(confPath)
	if err != nil {
		t.Fatalf("ImportWireGuardConfig() error: %v", err)
	}
	if server.MTU != 1420 {
		t.Errorf("MTU = %d, want 1420 (default) when MTU is invalid", server.MTU)
	}
}

func TestImportOpenVPNProtoTCP(t *testing.T) {
	tmpDir := t.TempDir()
	ovpnContent := `remote vpn.example.com 443
proto tcp
`
	ovpnPath := filepath.Join(tmpDir, "tcp.ovpn")
	os.WriteFile(ovpnPath, []byte(ovpnContent), 0600)

	server, err := ImportOpenVPNConfig(ovpnPath)
	if err != nil {
		t.Fatalf("ImportOpenVPNConfig() error: %v", err)
	}
	if server.Proto != "tcp" {
		t.Errorf("Proto = %q, want %q", server.Proto, "tcp")
	}
}

func TestImportWireGuardEmptyAllowedIPs(t *testing.T) {
	tmpDir := t.TempDir()
	confContent := `[Interface]
PrivateKey = yNGmpMvlEWbSI1iqKVlHPBXMRTf5pPjAi0CE5vIVp0I=
Address = 10.0.0.2/24

[Peer]
PublicKey = xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=
Endpoint = vpn.example.com:51820
AllowedIPs =
`
	confPath := filepath.Join(tmpDir, "emptyips.conf")
	os.WriteFile(confPath, []byte(confContent), 0600)

	server, _, err := ImportWireGuardConfig(confPath)
	if err != nil {
		t.Fatalf("ImportWireGuardConfig() error: %v", err)
	}
	// Empty AllowedIPs should keep defaults from DefaultServerConfig
	if len(server.AllowedIPs) != 2 || server.AllowedIPs[0] != "0.0.0.0/0" {
		t.Errorf("AllowedIPs = %v, want default [0.0.0.0/0 ::/0]", server.AllowedIPs)
	}
}

func TestImportOpenVPNMinimal(t *testing.T) {
	tmpDir := t.TempDir()
	ovpnContent := `remote 10.0.0.1 443
proto tcp
`
	ovpnPath := filepath.Join(tmpDir, "minimal.ovpn")
	os.WriteFile(ovpnPath, []byte(ovpnContent), 0600)

	server, err := ImportOpenVPNConfig(ovpnPath)
	if err != nil {
		t.Fatalf("ImportOpenVPNConfig() error: %v", err)
	}
	if server.Endpoint != "10.0.0.1:443" {
		t.Errorf("Endpoint = %q, want %q", server.Endpoint, "10.0.0.1:443")
	}
	if server.Proto != "tcp" {
		t.Errorf("Proto = %q, want %q", server.Proto, "tcp")
	}
	if server.CACert != "" {
		t.Error("CACert should be empty for minimal config")
	}
}
