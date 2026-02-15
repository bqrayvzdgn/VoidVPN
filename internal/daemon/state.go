// Package daemon manages background VPN connections and IPC communication.
package daemon

import (
	"encoding/json"
	"os"
	"time"

	"github.com/voidvpn/voidvpn/internal/config"
)

type ConnectionState struct {
	Server        string    `json:"server"`
	ConnectedAt   time.Time `json:"connected_at"`
	InterfaceName string    `json:"interface_name"`
	TunnelIP      string    `json:"tunnel_ip"`
	Endpoint      string    `json:"endpoint"`
	PID           int       `json:"pid"`
	TxBytes       int64     `json:"tx_bytes"`
	RxBytes       int64     `json:"rx_bytes"`
	Protocol      string    `json:"protocol"`
}

func SaveState(state *ConnectionState) error {
	if err := config.EnsureDirs(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(config.StateFile(), data, 0600)
}

func LoadState() (*ConnectionState, error) {
	data, err := os.ReadFile(config.StateFile())
	if err != nil {
		return nil, err
	}

	var state ConnectionState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func ClearState() error {
	if err := os.Remove(config.StateFile()); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func IsConnected() bool {
	state, err := LoadState()
	if err != nil {
		return false
	}
	// Verify the daemon process is still alive.
	// If it crashed (SIGKILL), the state file persists but the PID is dead.
	if !isProcessRunning(state.PID) {
		ClearState()
		return false
	}
	return true
}
