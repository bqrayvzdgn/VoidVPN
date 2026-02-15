package wireguard

import (
	"fmt"
	"net/netip"
	"strings"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
)

type Device struct {
	dev    *device.Device
	tunDev tun.Device
	name   string
}

func NewDevice(tunDev tun.Device, logger *device.Logger) (*Device, error) {
	bind := conn.NewDefaultBind()
	dev := device.NewDevice(tunDev, bind, logger)

	name, err := tunDev.Name()
	if err != nil {
		name = "unknown"
	}

	return &Device{
		dev:    dev,
		tunDev: tunDev,
		name:   name,
	}, nil
}

func (d *Device) Configure(cfg *TunnelConfig) error {
	ipcConfig, err := BuildIPCConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to build IPC config: %w", err)
	}
	return d.dev.IpcSet(ipcConfig)
}

func (d *Device) Up() error {
	return d.dev.Up()
}

func (d *Device) Down() error {
	d.dev.Down()
	return nil
}

func (d *Device) Close() {
	d.dev.Close()
}

func (d *Device) Name() string {
	return d.name
}

func (d *Device) TUN() tun.Device {
	return d.tunDev
}

type DeviceStats struct {
	TxBytes       int64
	RxBytes       int64
	LastHandshake int64 // Unix timestamp
}

func (d *Device) Stats() (*DeviceStats, error) {
	ipcGet, err := d.dev.IpcGet()
	if err != nil {
		return nil, fmt.Errorf("failed to get device stats: %w", err)
	}

	stats := &DeviceStats{}
	for _, line := range strings.Split(ipcGet, "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]
		switch key {
		case "tx_bytes":
			fmt.Sscanf(value, "%d", &stats.TxBytes)
		case "rx_bytes":
			fmt.Sscanf(value, "%d", &stats.RxBytes)
		case "last_handshake_time_sec":
			fmt.Sscanf(value, "%d", &stats.LastHandshake)
		}
	}

	return stats, nil
}

// ParseAddress parses a CIDR address string into a netip.Prefix.
func ParseAddress(addr string) (netip.Prefix, error) {
	if !strings.Contains(addr, "/") {
		addr += "/32"
	}
	prefix, err := netip.ParsePrefix(addr)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("invalid address %q: %w", addr, err)
	}
	return prefix, nil
}
