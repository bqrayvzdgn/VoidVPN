//go:build !windows

package network

import (
	"testing"
)

func TestValidIfaceName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"eth0", true},
		{"wg0", true},
		{"has space", false},
		{"", false},
		{"a/b", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validIfaceName.MatchString(tt.name)
			if got != tt.want {
				t.Errorf("validIfaceName.MatchString(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestRemoveVPNRoutesEmptyUnix(t *testing.T) {
	r := &unixRoutes{}
	err := r.RemoveVPNRoutes()
	if err != nil {
		t.Errorf("RemoveVPNRoutes() on empty should return nil, got %v", err)
	}
}

func TestAddVPNRoutesInvalidInterface(t *testing.T) {
	r := &unixRoutes{}
	err := r.AddVPNRoutes("bad iface!", "1.2.3.4", false)
	if err == nil {
		t.Error("AddVPNRoutes should error for invalid interface name")
	}
}
