package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/voidvpn/voidvpn/internal/config"
	"github.com/voidvpn/voidvpn/internal/network"
	"github.com/voidvpn/voidvpn/internal/tunnel"
)

type Daemon struct {
	tunnel    tunnel.Tunnel
	server    *config.ServerConfig
	dns       network.DNSManager
	routes    network.RouteManager
	ipc       *IPCServer
	cancel    context.CancelFunc
	Connected chan struct{} // closed when tunnel is connected
}

func New(tun tunnel.Tunnel, server *config.ServerConfig) *Daemon {
	return &Daemon{
		tunnel:    tun,
		server:    server,
		dns:       network.NewDNSManager(),
		routes:    network.NewRouteManager(),
		Connected: make(chan struct{}),
	}
}

func (d *Daemon) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	d.cancel = cancel
	defer cancel()

	slog.Info("connecting tunnel", "server", d.server.Name, "protocol", d.server.Protocol, "endpoint", d.server.Endpoint)

	// Connect tunnel
	if err := d.tunnel.Connect(ctx); err != nil {
		return err
	}
	defer d.cleanup()

	// Get tunnel status for interface name
	status, err := d.tunnel.Status()
	if err != nil {
		return err
	}

	slog.Debug("tunnel device ready", "interface", status.InterfaceName)

	// OpenVPN handles IP/DNS/routing via its own process.
	// For WireGuard, we must configure the network stack ourselves.
	if d.server.Protocol != "openvpn" {
		// Assign IP address to the tunnel interface
		slog.Debug("assigning address", "interface", status.InterfaceName, "address", d.server.Address)
		if err := network.AssignAddress(status.InterfaceName, d.server.Address); err != nil {
			return fmt.Errorf("failed to assign address to tunnel interface: %w", err)
		}

		// Add VPN routes (0.0.0.0/1 + 128.0.0.0/1 via TUN interface, endpoint via default gw)
		endpointHost := network.ExtractEndpointHost(d.server.Endpoint)
		hasIPv6 := false
		for _, aip := range d.server.AllowedIPs {
			if strings.Contains(aip, ":") {
				hasIPv6 = true
				break
			}
		}
		slog.Debug("adding VPN routes", "endpoint", endpointHost, "ipv6", hasIPv6)
		if err := d.routes.AddVPNRoutes(status.InterfaceName, endpointHost, hasIPv6); err != nil {
			return fmt.Errorf("failed to add VPN routes: %w", err)
		}
		slog.Info("routes configured", "interface", status.InterfaceName, "ipv6", hasIPv6)

		// Configure DNS
		if len(d.server.DNS) > 0 {
			slog.Debug("setting DNS", "interface", status.InterfaceName, "servers", d.server.DNS)
			if err := d.dns.Set(status.InterfaceName, d.server.DNS); err != nil {
				slog.Warn("failed to set DNS", "error", err)
			} else {
				slog.Info("DNS configured", "servers", d.server.DNS)
			}
		}
	}

	// Signal connected only AFTER network is fully configured (IP, routes, DNS).
	// This ensures the spinner shows success only when traffic can actually flow.
	close(d.Connected)

	// Save connection state
	state := &ConnectionState{
		Server:        d.server.Name,
		ConnectedAt:   time.Now(),
		InterfaceName: status.InterfaceName,
		TunnelIP:      d.server.Address,
		Endpoint:      d.server.Endpoint,
		PID:           os.Getpid(),
		Protocol:      d.server.Protocol,
	}
	if err := SaveState(state); err != nil {
		slog.Warn("failed to save state", "error", err)
	}

	// Start IPC server
	ipc, err := NewIPCServer(d.handleIPC)
	if err != nil {
		slog.Warn("failed to start IPC server", "error", err)
	} else {
		d.ipc = ipc
		go ipc.Serve()
		slog.Debug("IPC server started")
	}

	slog.Info("connected", "server", d.server.Name, "tunnel_ip", d.server.Address)

	// Wait for signal or IPC disconnect
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		slog.Info("signal received, disconnecting", "signal", sig)
	case <-ctx.Done():
		slog.Info("context cancelled, disconnecting")
	}

	return nil
}

func (d *Daemon) handleIPC(req *IPCRequest) *IPCResponse {
	slog.Debug("IPC request", "command", req.Command)
	switch req.Command {
	case "status":
		state, err := LoadState()
		if err != nil {
			return &IPCResponse{Success: false, Error: err.Error()}
		}
		// Update traffic stats
		if d.tunnel != nil {
			if status, err := d.tunnel.Status(); err == nil {
				state.TxBytes = status.TxBytes
				state.RxBytes = status.RxBytes
			}
		}
		return &IPCResponse{Success: true, State: state}
	case "disconnect":
		if d.cancel != nil {
			d.cancel()
		}
		return &IPCResponse{Success: true}
	default:
		return &IPCResponse{Success: false, Error: "unknown command"}
	}
}

func (d *Daemon) cleanup() {
	slog.Info("cleaning up")

	if d.ipc != nil {
		d.ipc.Close()
	}

	// Remove routes before disconnecting tunnel (routes reference the tunnel gateway)
	if err := d.routes.RemoveVPNRoutes(); err != nil {
		slog.Warn("failed to remove VPN routes", "error", err)
	} else {
		slog.Debug("VPN routes removed")
	}

	d.dns.Restore()
	slog.Debug("DNS restored")

	d.tunnel.Disconnect()
	ClearState()

	slog.Info("cleanup complete")
}
