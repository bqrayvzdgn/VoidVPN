//go:build windows

package network

import (
	"strings"
	"testing"
)

func TestDNSSetEmptyServers(t *testing.T) {
	d := &windowsDNS{}
	err := d.Set("TestIface", nil)
	if err != nil {
		t.Errorf("Set() with nil servers should return nil, got %v", err)
	}
}

func TestDNSSetInvalidIP(t *testing.T) {
	d := &windowsDNS{}
	err := d.Set("TestIface", []string{"not-an-ip"})
	if err == nil {
		t.Fatal("Set() should error for invalid DNS IP")
	}
	if !strings.Contains(err.Error(), "invalid DNS") {
		t.Errorf("error = %q, want mention of invalid DNS", err.Error())
	}
}

func TestDNSRestoreEmptyIface(t *testing.T) {
	d := &windowsDNS{}
	err := d.Restore()
	if err != nil {
		t.Errorf("Restore() with empty iface should return nil, got %v", err)
	}
}
