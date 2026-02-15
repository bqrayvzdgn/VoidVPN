package platform

import "golang.zx2c4.com/wireguard/tun"

// CreateTUN creates a platform-appropriate TUN device.
func CreateTUN(name string, mtu int) (tun.Device, error) {
	return createTUN(name, mtu)
}
