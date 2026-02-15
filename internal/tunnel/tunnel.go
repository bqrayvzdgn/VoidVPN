// Package tunnel defines a protocol-agnostic interface for VPN tunnels.
package tunnel

import (
	"context"
	"time"
)

// Tunnel is the interface that all VPN protocol implementations must satisfy.
type Tunnel interface {
	Connect(ctx context.Context) error
	Disconnect() error
	Status() (*TunnelStatus, error)
	IsActive() bool
}

// TunnelStatus holds the current state of a VPN tunnel, regardless of protocol.
type TunnelStatus struct {
	Connected     bool
	Protocol      string
	ServerName    string
	Endpoint      string
	TunnelIP      string
	ConnectedAt   time.Time
	TxBytes       int64
	RxBytes       int64
	LastHandshake time.Time
	InterfaceName string
}
