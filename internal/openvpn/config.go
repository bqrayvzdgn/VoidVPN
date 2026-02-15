package openvpn

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/voidvpn/voidvpn/internal/config"
)

// DetectOpenVPN searches for the openvpn binary on the system.
func DetectOpenVPN() (string, error) {
	// Check PATH first
	if path, err := exec.LookPath("openvpn"); err == nil {
		return path, nil
	}

	// Check common install locations
	var candidates []string
	if runtime.GOOS == "windows" {
		candidates = []string{
			`C:\Program Files\OpenVPN\bin\openvpn.exe`,
			`C:\Program Files (x86)\OpenVPN\bin\openvpn.exe`,
		}
	} else {
		candidates = []string{
			"/usr/sbin/openvpn",
			"/usr/local/sbin/openvpn",
			"/usr/bin/openvpn",
			"/usr/local/bin/openvpn",
		}
	}

	for _, c := range candidates {
		if path, err := exec.LookPath(c); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("openvpn binary not found. Install OpenVPN and ensure it is in your PATH")
}

// BuildOVPNConfig generates .ovpn file content from a ServerConfig.
func BuildOVPNConfig(cfg *config.ServerConfig, mgmtPort int) string {
	var sb strings.Builder

	sb.WriteString("client\n")
	sb.WriteString("dev tun\n")
	sb.WriteString("nobind\n")
	sb.WriteString("persist-key\n")
	sb.WriteString("persist-tun\n")

	// OpenVPN 2.7+ uses ovpn-dco driver by default on Windows.
	// Do NOT disable DCO — it avoids the need for TAP adapters.

	proto := cfg.Proto
	if proto == "" {
		proto = "udp"
	}
	sb.WriteString(fmt.Sprintf("proto %s\n", proto))

	// Parse host:port from endpoint
	host, port := parseEndpoint(cfg.Endpoint)
	sb.WriteString(fmt.Sprintf("remote %s %s\n", host, port))
	sb.WriteString("resolv-retry infinite\n")

	if cfg.Cipher != "" {
		sb.WriteString(fmt.Sprintf("cipher %s\n", cfg.Cipher))
	}
	if cfg.Auth != "" {
		sb.WriteString(fmt.Sprintf("auth %s\n", cfg.Auth))
	}
	if cfg.CompLZO {
		sb.WriteString("comp-lzo\n")
	}

	sb.WriteString("verb 3\n")
	sb.WriteString(fmt.Sprintf("management 127.0.0.1 %d\n", mgmtPort))

	// Inline certificate blocks
	if cfg.CACert != "" {
		sb.WriteString("<ca>\n")
		sb.WriteString(cfg.CACert)
		sb.WriteString("\n</ca>\n")
	}
	if cfg.ClientCert != "" {
		sb.WriteString("<cert>\n")
		sb.WriteString(cfg.ClientCert)
		sb.WriteString("\n</cert>\n")
	}
	if cfg.ClientKey != "" {
		sb.WriteString("<key>\n")
		sb.WriteString(cfg.ClientKey)
		sb.WriteString("\n</key>\n")
	}
	if cfg.TLSAuth != "" {
		sb.WriteString("key-direction 1\n")
		sb.WriteString("<tls-auth>\n")
		sb.WriteString(cfg.TLSAuth)
		sb.WriteString("\n</tls-auth>\n")
	}

	return sb.String()
}

func parseEndpoint(endpoint string) (host, port string) {
	idx := strings.LastIndex(endpoint, ":")
	if idx == -1 {
		return endpoint, "1194"
	}
	return endpoint[:idx], endpoint[idx+1:]
}
