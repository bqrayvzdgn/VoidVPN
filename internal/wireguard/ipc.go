package wireguard

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

// BuildIPCConfig constructs the IPC configuration string for wireguard-go device.IpcSet().
func BuildIPCConfig(cfg *TunnelConfig) (string, error) {
	var sb strings.Builder

	privHex, err := keyToHex(cfg.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}
	sb.WriteString(fmt.Sprintf("private_key=%s\n", privHex))

	// Peer configuration
	pubHex, err := keyToHex(cfg.PeerPublicKey)
	if err != nil {
		return "", fmt.Errorf("invalid peer public key: %w", err)
	}
	sb.WriteString(fmt.Sprintf("public_key=%s\n", pubHex))

	if cfg.PeerPresharedKey != "" {
		pskHex, err := keyToHex(cfg.PeerPresharedKey)
		if err != nil {
			return "", fmt.Errorf("invalid preshared key: %w", err)
		}
		sb.WriteString(fmt.Sprintf("preshared_key=%s\n", pskHex))
	}

	if cfg.PeerEndpoint == "" {
		return "", fmt.Errorf("peer endpoint is required")
	}
	// Validate endpoint contains no newlines (IPC injection prevention)
	if strings.ContainsAny(cfg.PeerEndpoint, "\n\r") {
		return "", fmt.Errorf("invalid peer endpoint: contains newline characters")
	}
	sb.WriteString(fmt.Sprintf("endpoint=%s\n", cfg.PeerEndpoint))

	if len(cfg.PeerAllowedIPs) == 0 {
		return "", fmt.Errorf("peer AllowedIPs is required (at least one entry needed)")
	}
	for _, allowedIP := range cfg.PeerAllowedIPs {
		if strings.ContainsAny(allowedIP, "\n\r") {
			return "", fmt.Errorf("invalid allowed IP: contains newline characters")
		}
		sb.WriteString(fmt.Sprintf("allowed_ip=%s\n", allowedIP))
	}

	if cfg.PersistentKeepalive > 0 {
		sb.WriteString(fmt.Sprintf("persistent_keepalive_interval=%d\n", cfg.PersistentKeepalive))
	}

	return sb.String(), nil
}

// keyToHex converts a base64-encoded WireGuard key to hex encoding for IPC.
func keyToHex(base64Key string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return "", fmt.Errorf("invalid base64 key: %w", err)
	}
	if len(decoded) != 32 {
		return "", fmt.Errorf("invalid key length: expected 32 bytes, got %d", len(decoded))
	}
	return hex.EncodeToString(decoded), nil
}
