package network

import (
	"fmt"
	"net/netip"
	"os/exec"
	"runtime"
	"strings"
)

// AssignAddress assigns an IP address to a network interface.
func AssignAddress(iface string, address string) error {
	if address == "" {
		return fmt.Errorf("tunnel address is empty — check server config")
	}

	prefix, err := netip.ParsePrefix(address)
	if err != nil {
		if addr, err2 := netip.ParseAddr(address); err2 == nil {
			// Use /32 for IPv4, /128 for IPv6
			bits := 32
			if addr.Is6() {
				bits = 128
			}
			prefix = netip.PrefixFrom(addr, bits)
		} else {
			return fmt.Errorf("invalid address %q: %w", address, err)
		}
	}

	switch runtime.GOOS {
	case "windows":
		return assignAddressWindows(iface, prefix)
	default:
		return assignAddressLinux(iface, prefix)
	}
}

func assignAddressWindows(iface string, prefix netip.Prefix) error {
	addr := prefix.Addr().String()
	mask := prefixToMask(prefix.Bits())
	cmd := exec.Command("netsh", "interface", "ip", "set", "address",
		fmt.Sprintf("name=%s", iface), "static", addr, mask)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("netsh set address failed: %s: %w", string(out), err)
	}
	return nil
}

func assignAddressLinux(iface string, prefix netip.Prefix) error {
	cmd := exec.Command("ip", "addr", "add", prefix.String(), "dev", iface)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ip addr add failed: %s: %w", string(out), err)
	}

	// Bring interface up
	linkCmd := exec.Command("ip", "link", "set", iface, "up")
	if out, err := linkCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ip link set up failed: %s: %w", string(out), err)
	}
	return nil
}

func prefixToMask(bits int) string {
	mask := uint32(0xFFFFFFFF) << (32 - bits)
	return fmt.Sprintf("%d.%d.%d.%d",
		(mask>>24)&0xFF, (mask>>16)&0xFF, (mask>>8)&0xFF, mask&0xFF)
}

// ExtractGateway extracts a gateway IP from a tunnel address (e.g., 10.0.0.2/24 -> 10.0.0.1).
func ExtractGateway(address string) string {
	prefix, err := netip.ParsePrefix(address)
	if err != nil {
		// Try as plain IP
		addr, err := netip.ParseAddr(address)
		if err != nil {
			return ""
		}
		prefix = netip.PrefixFrom(addr, 24)
	}

	// Set last octet to 1 for the gateway
	addr := prefix.Addr()
	bytes := addr.As4()
	bytes[3] = 1
	return netip.AddrFrom4(bytes).String()
}

// ExtractEndpointHost extracts the host part from an endpoint (host:port).
func ExtractEndpointHost(endpoint string) string {
	// Handle IPv6 brackets [host]:port
	if strings.HasPrefix(endpoint, "[") {
		idx := strings.Index(endpoint, "]")
		if idx > 0 {
			return endpoint[1:idx]
		}
	}
	// IPv4 or hostname:port
	idx := strings.LastIndex(endpoint, ":")
	if idx > 0 {
		return endpoint[:idx]
	}
	return endpoint
}
