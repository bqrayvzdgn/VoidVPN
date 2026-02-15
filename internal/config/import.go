package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
)

func ImportWireGuardConfig(path string) (*ServerConfig, string, error) {
	cfg, err := ini.LoadSources(ini.LoadOptions{
		AllowNonUniqueSections: true,
	}, path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse config file: %w", err)
	}

	iface := cfg.Section("Interface")
	peer := cfg.Section("Peer")

	if peer == nil {
		return nil, "", fmt.Errorf("no [Peer] section found in config")
	}

	// Extract server name from filename
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	server := DefaultServerConfig()
	server.Name = name

	// Interface section
	if key := iface.Key("Address"); key.String() != "" {
		server.Address = key.String()
	} else {
		return nil, "", fmt.Errorf("interface Address is required")
	}
	if key := iface.Key("DNS"); key.String() != "" {
		server.DNS = splitAndTrim(key.String())
	}
	if key := iface.Key("MTU"); key.String() != "" {
		mtu, err := key.Int()
		if err == nil && mtu > 0 {
			server.MTU = mtu
		}
	}

	// Peer section
	if key := peer.Key("PublicKey"); key.String() != "" {
		server.PublicKey = key.String()
	} else {
		return nil, "", fmt.Errorf("peer PublicKey is required")
	}

	if key := peer.Key("Endpoint"); key.String() != "" {
		server.Endpoint = key.String()
	} else {
		return nil, "", fmt.Errorf("peer Endpoint is required")
	}

	if key := peer.Key("AllowedIPs"); key.String() != "" {
		ips := splitAndTrim(key.String())
		if len(ips) > 0 {
			server.AllowedIPs = ips
		}
	}
	// AllowedIPs keeps DefaultServerConfig default (0.0.0.0/0, ::/0) if not specified

	if key := peer.Key("PresharedKey"); key.String() != "" {
		server.PresharedKey = key.String()
	}

	if key := peer.Key("PersistentKeepalive"); key.String() != "" {
		ka, err := key.Int()
		if err == nil {
			server.PersistentKeepalive = ka
		}
	}

	// Extract private key from Interface section
	privateKey := ""
	if key := iface.Key("PrivateKey"); key.String() != "" {
		privateKey = key.String()
	}

	return server, privateKey, nil
}

// ImportOpenVPNConfig parses an .ovpn config file and returns a ServerConfig.
func ImportOpenVPNConfig(path string) (*ServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	base := filepath.Base(path)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	server := &ServerConfig{
		Name:     name,
		Protocol: "openvpn",
		Proto:    "udp",
		DNS:      []string{"1.1.1.1", "1.0.0.1"},
	}

	// Parse directives line by line
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "remote":
			if len(fields) >= 2 {
				host := fields[1]
				port := "1194"
				if len(fields) >= 3 {
					port = fields[2]
				}
				server.Endpoint = host + ":" + port
				if p, err := strconv.Atoi(port); err == nil {
					server.RemotePort = p
				}
			}
		case "proto":
			if len(fields) >= 2 {
				server.Proto = fields[1]
			}
		case "cipher":
			if len(fields) >= 2 {
				server.Cipher = fields[1]
			}
		case "auth":
			if len(fields) >= 2 {
				server.Auth = fields[1]
			}
		case "comp-lzo":
			server.CompLZO = true
		case "auth-user-pass":
			// Credentials will be prompted at connect time
		}
	}

	// Extract inline blocks
	server.CACert = extractInlineBlock(content, "ca")
	server.ClientCert = extractInlineBlock(content, "cert")
	server.ClientKey = extractInlineBlock(content, "key")
	server.TLSAuth = extractInlineBlock(content, "tls-auth")

	if server.Endpoint == "" {
		return nil, fmt.Errorf("no 'remote' directive found in config")
	}

	return server, nil
}

// extractInlineBlock extracts content between <tag> and </tag> markers.
func extractInlineBlock(content, tag string) string {
	start := "<" + tag + ">"
	end := "</" + tag + ">"

	startIdx := strings.Index(content, start)
	if startIdx == -1 {
		return ""
	}
	startIdx += len(start)

	endIdx := strings.Index(content[startIdx:], end)
	if endIdx == -1 {
		return ""
	}

	return strings.TrimSpace(content[startIdx : startIdx+endIdx])
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
