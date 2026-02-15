package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	LogLevel      string   `yaml:"log_level"`
	DefaultServer string   `yaml:"default_server"`
	AutoConnect   bool     `yaml:"auto_connect"`
	DNSFallback   []string `yaml:"dns_fallback"`
	KillSwitch    bool     `yaml:"kill_switch"`
}

func DefaultConfig() *AppConfig {
	return &AppConfig{
		LogLevel:    "info",
		DNSFallback: []string{"1.1.1.1", "8.8.8.8"},
	}
}

func Load() (*AppConfig, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(ConfigFile())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *AppConfig) Save() error {
	if err := EnsureDirs(); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigFile(), data, 0600)
}

func (c *AppConfig) Get(key string) string {
	switch key {
	case "log_level":
		return c.LogLevel
	case "default_server":
		return c.DefaultServer
	case "kill_switch":
		if c.KillSwitch {
			return "true"
		}
		return "false"
	case "auto_connect":
		if c.AutoConnect {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

func (c *AppConfig) Set(key, value string) bool {
	switch key {
	case "log_level":
		c.LogLevel = value
	case "default_server":
		c.DefaultServer = value
	case "kill_switch":
		c.KillSwitch = value == "true"
	case "auto_connect":
		c.AutoConnect = value == "true"
	default:
		return false
	}
	return true
}
