// Package network provides platform-specific network configuration for VoidVPN.
package network

// DNSManager handles DNS configuration for the VPN tunnel.
type DNSManager interface {
	Set(iface string, servers []string) error
	Restore() error
}

// NewDNSManager returns a platform-appropriate DNS manager.
func NewDNSManager() DNSManager {
	return newDNSManager()
}
