// Package config handles application configuration for VoidVPN.
package config

import (
	"os"
	"path/filepath"
	"runtime"
)

func ConfigDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "VoidVPN")
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "voidvpn")
	default:
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return filepath.Join(xdg, "voidvpn")
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "voidvpn")
	}
}

func ConfigFile() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

func ServersDir() string {
	return filepath.Join(ConfigDir(), "servers")
}

func StateDir() string {
	return filepath.Join(ConfigDir(), "state")
}

func StateFile() string {
	return filepath.Join(StateDir(), "connection.json")
}

func EnsureDirs() error {
	dirs := []string{ConfigDir(), ServersDir(), StateDir()}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0700); err != nil {
			return err
		}
	}
	return nil
}
