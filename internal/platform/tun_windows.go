//go:build windows

package platform

import (
	"fmt"

	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"
)

func createTUN(name string, mtu int) (tun.Device, error) {
	return tun.CreateTUN(name, mtu)
}

// GetInterfaceLUID returns the LUID for a WireGuard TUN device on Windows.
func GetInterfaceLUID(dev tun.Device) (winipcfg.LUID, error) {
	nativeDev, ok := dev.(*tun.NativeTun)
	if !ok {
		return 0, fmt.Errorf("unexpected TUN device type: %T", dev)
	}
	return winipcfg.LUIDFromIndex(uint32(nativeDev.File().Fd()))
}
