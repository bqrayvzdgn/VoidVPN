//go:build windows

package network

import (
	"fmt"
	"net"
	"os/exec"
)

type windowsDNS struct {
	iface string
}

func newDNSManager() DNSManager {
	return &windowsDNS{}
}

func (d *windowsDNS) Set(iface string, servers []string) error {
	d.iface = iface

	if len(servers) == 0 {
		return nil
	}

	// Validate all DNS server addresses
	for _, server := range servers {
		if net.ParseIP(server) == nil {
			return fmt.Errorf("invalid DNS server address: %q", server)
		}
	}

	// Set primary DNS
	cmd := exec.Command("netsh", "interface", "ip", "set", "dns",
		fmt.Sprintf("name=%s", iface), "static", servers[0])
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set primary DNS: %w", err)
	}

	// Set additional DNS servers
	for _, server := range servers[1:] {
		cmd := exec.Command("netsh", "interface", "ip", "add", "dns",
			fmt.Sprintf("name=%s", iface), server, "index=2")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to add DNS server %s: %w", server, err)
		}
	}

	return nil
}

func (d *windowsDNS) Restore() error {
	if d.iface == "" {
		return nil
	}

	// Reset DNS on the TUN interface before it is destroyed.
	cmd := exec.Command("netsh", "interface", "ip", "set", "dns",
		fmt.Sprintf("name=%s", d.iface), "dhcp")
	return cmd.Run()
}
