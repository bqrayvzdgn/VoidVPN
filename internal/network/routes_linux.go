//go:build !windows

package network

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
)

var validIfaceName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type unixRoutes struct {
	addedRoutes []string
	endpoint    string
	defaultGW   string
}

func newRouteManager() RouteManager {
	return &unixRoutes{}
}

func (r *unixRoutes) AddVPNRoutes(iface string, endpoint string, ipv6 bool) error {
	// Validate inputs
	if !validIfaceName.MatchString(iface) {
		return fmt.Errorf("invalid interface name: %q", iface)
	}
	if net.ParseIP(endpoint) == nil {
		// Resolve hostname to IP
		ips, err := net.LookupIP(endpoint)
		if err != nil || len(ips) == 0 {
			return fmt.Errorf("failed to resolve endpoint %q: %w", endpoint, err)
		}
		endpoint = ips[0].String()
	}

	r.endpoint = endpoint

	// Get current default gateway
	out, err := exec.Command("ip", "route", "show", "default").Output()
	if err != nil {
		return fmt.Errorf("failed to get default route: %w", err)
	}

	fields := strings.Fields(string(out))
	if len(fields) >= 3 {
		r.defaultGW = fields[2]
	}

	// Route VPN endpoint via current default gateway
	if r.defaultGW != "" {
		cmd := exec.Command("ip", "route", "add", endpoint+"/32", "via", r.defaultGW)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to add endpoint route: %w", err)
		}
		r.addedRoutes = append(r.addedRoutes, endpoint+"/32")
	}

	// Add IPv4 split routes
	for _, cidr := range []string{"0.0.0.0/1", "128.0.0.0/1"} {
		cmd := exec.Command("ip", "route", "add", cidr, "dev", iface)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to add route %s: %w", cidr, err)
		}
		r.addedRoutes = append(r.addedRoutes, cidr)
	}

	// Add IPv6 split routes to prevent IPv6 traffic leaking outside the tunnel
	if ipv6 {
		for _, cidr := range []string{"::/1", "8000::/1"} {
			cmd := exec.Command("ip", "-6", "route", "add", cidr, "dev", iface)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to add IPv6 route %s: %w", cidr, err)
			}
			r.addedRoutes = append(r.addedRoutes, "v6:"+cidr)
		}
	}

	return nil
}

func (r *unixRoutes) RemoveVPNRoutes() error {
	var lastErr error
	for i := len(r.addedRoutes) - 1; i >= 0; i-- {
		route := r.addedRoutes[i]
		var cmd *exec.Cmd
		if strings.HasPrefix(route, "v6:") {
			cmd = exec.Command("ip", "-6", "route", "delete", strings.TrimPrefix(route, "v6:"))
		} else {
			cmd = exec.Command("ip", "route", "delete", route)
		}
		if err := cmd.Run(); err != nil {
			lastErr = err
		}
	}
	r.addedRoutes = nil
	return lastErr
}
