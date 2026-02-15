//go:build !windows

package network

import (
	"strings"
	"testing"
)

func TestUnixDNSSetEmpty(t *testing.T) {
	d := &unixDNS{}
	err := d.Set("", nil)
	if err != nil {
		t.Errorf("Set() with nil servers should return nil, got %v", err)
	}
}

func TestUnixDNSSetInvalidIP(t *testing.T) {
	d := &unixDNS{}
	err := d.Set("", []string{"bad"})
	if err == nil {
		t.Fatal("Set() should error for invalid DNS IP")
	}
	if !strings.Contains(err.Error(), "invalid DNS") {
		t.Errorf("error = %q, want mention of invalid DNS", err.Error())
	}
}

func TestUnixDNSRestoreNilOriginal(t *testing.T) {
	d := &unixDNS{}
	err := d.Restore()
	if err != nil {
		t.Errorf("Restore() with nil origResolvConf should return nil, got %v", err)
	}
}
