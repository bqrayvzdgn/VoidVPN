package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var validNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9 _-]{0,62}$`)

type ServerConfig struct {
	Name                string   `yaml:"name"`
	Protocol            string   `yaml:"protocol"`
	Endpoint            string   `yaml:"endpoint"`
	PublicKey           string   `yaml:"public_key"`
	AllowedIPs          []string `yaml:"allowed_ips"`
	DNS                 []string `yaml:"dns"`
	Address             string   `yaml:"address"`
	PresharedKey        string   `yaml:"preshared_key,omitempty"`
	PersistentKeepalive int      `yaml:"persistent_keepalive"`
	MTU                 int      `yaml:"mtu"`

	// OpenVPN-specific fields
	CACert     string `yaml:"ca_cert,omitempty"`
	ClientCert string `yaml:"client_cert,omitempty"`
	ClientKey  string `yaml:"client_key,omitempty"`
	TLSAuth    string `yaml:"tls_auth,omitempty"`
	Cipher     string `yaml:"cipher,omitempty"`
	Auth       string `yaml:"auth,omitempty"`
	Proto      string `yaml:"proto,omitempty"`
	CompLZO    bool   `yaml:"comp_lzo,omitempty"`
	RemotePort int    `yaml:"remote_port,omitempty"`
	Username   string `yaml:"username,omitempty"`
	Password   string `yaml:"password,omitempty"`
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Protocol:            "wireguard",
		AllowedIPs:          []string{"0.0.0.0/0", "::/0"},
		DNS:                 []string{"1.1.1.1", "1.0.0.1"},
		PersistentKeepalive: 25,
		MTU:                 1420,
	}
}

// ValidateName checks that a server name is safe for use in file paths.
func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if !validNamePattern.MatchString(name) {
		return fmt.Errorf("invalid name %q: must be alphanumeric with hyphens, underscores, or spaces (max 63 chars)", name)
	}
	return nil
}

func serverFile(name string) (string, error) {
	if err := ValidateName(name); err != nil {
		return "", err
	}
	safe := strings.ReplaceAll(strings.ToLower(name), " ", "-")
	full := filepath.Join(ServersDir(), safe+".yaml")
	// Belt-and-suspenders: verify the resolved path stays under ServersDir
	if !strings.HasPrefix(filepath.Clean(full), filepath.Clean(ServersDir())) {
		return "", fmt.Errorf("invalid name: path traversal detected")
	}
	return full, nil
}

func LoadServer(name string) (*ServerConfig, error) {
	path, err := serverFile(name)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("server '%s' not found", name)
		}
		return nil, err
	}

	var cfg ServerConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse server config: %w", err)
	}
	return &cfg, nil
}

func SaveServer(cfg *ServerConfig) error {
	if err := EnsureDirs(); err != nil {
		return err
	}

	path, err := serverFile(cfg.Name)
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func RemoveServer(name string) error {
	path, err := serverFile(name)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("server '%s' not found", name)
	}
	return os.Remove(path)
}

func ListServers() ([]*ServerConfig, error) {
	if err := EnsureDirs(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(ServersDir())
	if err != nil {
		return nil, err
	}

	var servers []*ServerConfig
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(ServersDir(), entry.Name()))
		if err != nil {
			continue
		}

		var cfg ServerConfig
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			continue
		}
		servers = append(servers, &cfg)
	}
	return servers, nil
}

func ServerExists(name string) bool {
	path, err := serverFile(name)
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}
