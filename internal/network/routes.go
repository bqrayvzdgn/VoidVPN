package network

// RouteManager handles routing table configuration for the VPN tunnel.
type RouteManager interface {
	AddVPNRoutes(iface string, endpoint string, ipv6 bool) error
	RemoveVPNRoutes() error
}

// NewRouteManager returns a platform-appropriate route manager.
func NewRouteManager() RouteManager {
	return newRouteManager()
}
