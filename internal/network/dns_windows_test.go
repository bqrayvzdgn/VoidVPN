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

func TestDNSSetStoresIface(t *testing.T) {
	d := &windowsDNS{}
	// Set with valid IPs but will fail on netsh (expected - no admin)
	_ = d.Set("test-iface", []string{"8.8.8.8"})
	if d.iface != "test-iface" {
		t.Errorf("iface = %q, want \"test-iface\"", d.iface)
	}
}

func TestDNSSetMultipleInvalidIP(t *testing.T) {
	d := &windowsDNS{}
	err := d.Set("iface", []string{"8.8.8.8", "not-valid"})
	if err == nil {
		t.Error("expected error for invalid IP in list")
	}
	if !strings.Contains(err.Error(), "invalid DNS") {
		t.Errorf("error should mention invalid DNS: %v", err)
	}
}

func TestDNSRestoreWithIface(t *testing.T) {
	d := &windowsDNS{iface: "test-iface"}
	// This will fail because netsh isn't available in test, but it exercises the code path
	_ = d.Restore()
	// We just verify it doesn't panic
}
