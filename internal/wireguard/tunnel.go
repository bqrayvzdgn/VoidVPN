package wireguard

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"golang.zx2c4.com/wireguard/device"

	"github.com/voidvpn/voidvpn/internal/config"
	"github.com/voidvpn/voidvpn/internal/platform"
	"github.com/voidvpn/voidvpn/internal/tunnel"
)

type Tunnel struct {
	config      *TunnelConfig
	server      *config.ServerConfig
	device      *Device
	cancelFunc  context.CancelFunc
	connectedAt time.Time
}

func NewTunnel(serverCfg *config.ServerConfig, privateKey string) *Tunnel {
	tunnelCfg := &TunnelConfig{
		PrivateKey:          privateKey,
		Address:             serverCfg.Address,
		DNS:                 serverCfg.DNS,
		MTU:                 serverCfg.MTU,
		PeerPublicKey:       serverCfg.PublicKey,
		PeerEndpoint:        serverCfg.Endpoint,
		PeerAllowedIPs:      serverCfg.AllowedIPs,
		PeerPresharedKey:    serverCfg.PresharedKey,
		PersistentKeepalive: serverCfg.PersistentKeepalive,
	}

	return &Tunnel{
		config: tunnelCfg,
		server: serverCfg,
	}
}

func (t *Tunnel) Connect(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	t.cancelFunc = cancel

	// Create TUN device
	slog.Debug("creating TUN device", "name", "voidvpn0", "mtu", t.config.MTU)
	tunDev, err := platform.CreateTUN("voidvpn0", t.config.MTU)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create TUN device: %w", err)
	}

	// Create WireGuard device
	logger := device.NewLogger(device.LogLevelSilent, "")
	dev, err := NewDevice(tunDev, logger)
	if err != nil {
		tunDev.Close()
		cancel()
		return fmt.Errorf("failed to create WireGuard device: %w", err)
	}
	t.device = dev

	// Configure device
	slog.Debug("configuring WireGuard device", "endpoint", t.config.PeerEndpoint, "allowed_ips", t.config.PeerAllowedIPs)
	if err := dev.Configure(t.config); err != nil {
		dev.Close()
		cancel()
		return fmt.Errorf("failed to configure device: %w", err)
	}

	// Bring device up
	if err := dev.Up(); err != nil {
		dev.Close()
		cancel()
		return fmt.Errorf("failed to bring device up: %w", err)
	}

	t.connectedAt = time.Now()
	slog.Info("WireGuard tunnel connected", "server", t.server.Name, "endpoint", t.server.Endpoint)
	return nil
}

func (t *Tunnel) Disconnect() error {
	if t.cancelFunc != nil {
		t.cancelFunc()
	}
	if t.device != nil {
		t.device.Close()
	}
	slog.Info("WireGuard tunnel disconnected", "server", t.server.Name)
	return nil
}

func (t *Tunnel) Status() (*tunnel.TunnelStatus, error) {
	status := &tunnel.TunnelStatus{
		Protocol:   "wireguard",
		ServerName: t.server.Name,
		Endpoint:   t.server.Endpoint,
		TunnelIP:   t.config.Address,
		Connected:  t.device != nil,
	}

	if t.device != nil {
		status.ConnectedAt = t.connectedAt
		stats, err := t.device.Stats()
		if err == nil {
			status.TxBytes = stats.TxBytes
			status.RxBytes = stats.RxBytes
			if stats.LastHandshake > 0 {
				status.LastHandshake = time.Unix(stats.LastHandshake, 0)
			}
		}
		status.InterfaceName = t.device.Name()
	}

	return status, nil
}

func (t *Tunnel) IsActive() bool {
	return t.device != nil
}
