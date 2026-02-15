//go:build windows

package network

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
)

type windowsRoutes struct {
	// endpoint route goes via the physical default gateway
	endpointRoute *gatewayRoute
	// VPN split routes go via the TUN interface directly (on-link)
	ifaceRoutes []ifaceRoute
}

type gatewayRoute struct {
	network string
	mask    string
	gateway string
}

type ifaceRoute struct {
	prefix string
	iface  string
}

func newRouteManager() RouteManager {
	return &windowsRoutes{}
}

func (r *windowsRoutes) AddVPNRoutes(iface string, endpoint string, ipv6 bool) error {
	// Get current default gateway for endpoint-specific route
	defaultGW, err := getDefaultGateway()
	if err != nil {
		return fmt.Errorf("failed to get default gateway: %w", err)
	}

	// Validate endpoint IP
	if net.ParseIP(endpoint) == nil {
		// Resolve hostname to IP
		ips, err := net.LookupIP(endpoint)
		if err != nil || len(ips) == 0 {
			return fmt.Errorf("failed to resolve endpoint %q: %w", endpoint, err)
		}
		endpoint = ips[0].String()
	}

	// Add route to VPN endpoint via current default gateway (keep it reachable)
	if err := addGatewayRoute(endpoint, "255.255.255.255", defaultGW); err != nil {
		return fmt.Errorf("failed to add endpoint route: %w", err)
	}
	r.endpointRoute = &gatewayRoute{endpoint, "255.255.255.255", defaultGW}

	// Add IPv4 split routes (0.0.0.0/1 + 128.0.0.0/1) via TUN interface directly.
	// Uses on-link routing — no gateway needed, works with /32 tunnel addresses.
	for _, prefix := range []string{"0.0.0.0/1", "128.0.0.0/1"} {
		if err := addInterfaceRoute(prefix, iface); err != nil {
			return fmt.Errorf("failed to add route %s: %w", prefix, err)
		}
		r.ifaceRoutes = append(r.ifaceRoutes, ifaceRoute{prefix, iface})
	}

	// Add IPv6 split routes (::/1 + 8000::/1) to prevent IPv6 traffic leaking
	// outside the tunnel when AllowedIPs includes ::/0.
	if ipv6 {
		for _, prefix := range []string{"::/1", "8000::/1"} {
			if err := addInterfaceRouteV6(prefix, iface); err != nil {
				return fmt.Errorf("failed to add IPv6 route %s: %w", prefix, err)
			}
			r.ifaceRoutes = append(r.ifaceRoutes, ifaceRoute{prefix, iface})
		}
	}

	return nil
}

func (r *windowsRoutes) RemoveVPNRoutes() error {
	var lastErr error

	// Remove interface routes first
	for i := len(r.ifaceRoutes) - 1; i >= 0; i-- {
		rt := r.ifaceRoutes[i]
		var err error
		if strings.Contains(rt.prefix, ":") {
			err = deleteInterfaceRouteV6(rt.prefix, rt.iface)
		} else {
			err = deleteInterfaceRoute(rt.prefix, rt.iface)
		}
		if err != nil {
			lastErr = err
		}
	}
	r.ifaceRoutes = nil

	// Remove endpoint route
	if r.endpointRoute != nil {
		rt := r.endpointRoute
		if err := deleteGatewayRoute(rt.network, rt.mask, rt.gateway); err != nil {
			lastErr = err
		}
		r.endpointRoute = nil
	}

	return lastErr
}

func addGatewayRoute(network, mask, gateway string) error {
	if net.ParseIP(network) == nil {
		return fmt.Errorf("invalid network address: %q", network)
	}
	if net.ParseIP(mask) == nil {
		return fmt.Errorf("invalid subnet mask: %q", mask)
	}
	if net.ParseIP(gateway) == nil {
		return fmt.Errorf("invalid gateway address: %q", gateway)
	}
	cmd := exec.Command("route", "add", network, "mask", mask, gateway, "metric", "5")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("route add failed: %s: %w", string(out), err)
	}
	return nil
}

func deleteGatewayRoute(network, mask, gateway string) error {
	cmd := exec.Command("route", "delete", network, "mask", mask, gateway)
	return cmd.Run()
}

func addInterfaceRoute(prefix, iface string) error {
	cmd := exec.Command("netsh", "interface", "ip", "add", "route",
		prefix, iface, "metric=5")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("netsh add route %s via %s failed: %s: %w", prefix, iface, string(out), err)
	}
	return nil
}

func deleteInterfaceRoute(prefix, iface string) error {
	cmd := exec.Command("netsh", "interface", "ip", "delete", "route",
		prefix, iface)
	return cmd.Run()
}

func addInterfaceRouteV6(prefix, iface string) error {
	cmd := exec.Command("netsh", "interface", "ipv6", "add", "route",
		prefix, iface, "metric=5")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("netsh add ipv6 route %s via %s failed: %s: %w", prefix, iface, string(out), err)
	}
	return nil
}

func deleteInterfaceRouteV6(prefix, iface string) error {
	cmd := exec.Command("netsh", "interface", "ipv6", "delete", "route",
		prefix, iface)
	return cmd.Run()
}

func getDefaultGateway() (string, error) {
	out, err := exec.Command("route", "print", "0.0.0.0").Output()
	if err != nil {
		return "", err
	}

	// Parse output to find default gateway
	lines := splitLines(string(out))
	for _, line := range lines {
		fields := splitFields(line)
		if len(fields) >= 3 && fields[0] == "0.0.0.0" && fields[1] == "0.0.0.0" {
			return fields[2], nil
		}
	}
	return "", fmt.Errorf("default gateway not found")
}

func splitLines(s string) []string {
	var lines []string
	for _, line := range append([]string{}, split(s, '\n')...) {
		line = trimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func split(s string, sep byte) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func splitFields(s string) []string {
	var fields []string
	field := ""
	for _, c := range s {
		if c == ' ' || c == '\t' {
			if field != "" {
				fields = append(fields, field)
				field = ""
			}
		} else {
			field += string(c)
		}
	}
	if field != "" {
		fields = append(fields, field)
	}
	return fields
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r' || s[start] == '\n') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r' || s[end-1] == '\n') {
		end--
	}
	return s[start:end]
}
