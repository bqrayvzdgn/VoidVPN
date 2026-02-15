package ui

import (
	"strings"
	"testing"
	"time"
)

func TestRenderStatusDisconnected(t *testing.T) {
	s := StatusInfo{Connected: false}
	result := RenderStatus(s)
	if !strings.Contains(result, "Disconnected") {
		t.Error("RenderStatus should contain 'Disconnected' when not connected")
	}
}

func TestRenderStatusConnected(t *testing.T) {
	s := StatusInfo{
		Connected:   true,
		Protocol:    "wireguard",
		ServerName:  "my-server",
		Endpoint:    "1.2.3.4:51820",
		TunnelIP:    "10.0.0.2",
		ConnectedAt: time.Now().Add(-5 * time.Minute),
		TxBytes:     1024,
		RxBytes:     2048,
	}
	result := RenderStatus(s)
	if !strings.Contains(result, "my-server") {
		t.Error("RenderStatus should contain server name")
	}
	if !strings.Contains(result, "1.2.3.4:51820") {
		t.Error("RenderStatus should contain endpoint")
	}
	if !strings.Contains(result, "10.0.0.2") {
		t.Error("RenderStatus should contain tunnel IP")
	}
}

func TestFormatHandshakeNever(t *testing.T) {
	got := formatHandshake(time.Time{})
	if got != "never" {
		t.Errorf("formatHandshake(zero) = %q, want %q", got, "never")
	}
}

func TestFormatHandshakeAgo(t *testing.T) {
	got := formatHandshake(time.Now().Add(-5 * time.Minute))
	if !strings.Contains(got, "ago") {
		t.Errorf("formatHandshake(5min ago) = %q, want to contain 'ago'", got)
	}
}

func TestRenderStatusConnectedOpenVPN(t *testing.T) {
	info := StatusInfo{
		Connected:   true,
		Protocol:    "openvpn",
		ServerName:  "my-ovpn",
		Endpoint:    "vpn.example.com:443",
		TunnelIP:    "10.8.0.2",
		ConnectedAt: time.Now().Add(-10 * time.Minute),
		TxBytes:     1024 * 1024,
		RxBytes:     2048 * 1024,
	}
	result := RenderStatus(info)
	if !strings.Contains(result, "my-ovpn") {
		t.Error("should contain server name")
	}
	if !strings.Contains(result, "OpenVPN") {
		t.Error("should contain 'OpenVPN' for openvpn protocol")
	}
	if !strings.Contains(result, "10.8.0.2") {
		t.Error("should contain tunnel IP")
	}
}

func TestRenderStatusWireGuard(t *testing.T) {
	info := StatusInfo{
		Connected:   true,
		Protocol:    "wireguard",
		ServerName:  "wg-server",
		Endpoint:    "1.2.3.4:51820",
		TunnelIP:    "10.0.0.2",
		ConnectedAt: time.Now(),
	}
	result := RenderStatus(info)
	if !strings.Contains(result, "WireGuard") {
		t.Error("should contain 'WireGuard' for wireguard protocol")
	}
}

func TestFormatHandshakeRecent(t *testing.T) {
	recent := time.Now().Add(-30 * time.Second)
	result := formatHandshake(recent)
	if !strings.Contains(result, "ago") {
		t.Errorf("formatHandshake(30s ago) = %q, should contain 'ago'", result)
	}
}

func TestRenderStatusDisconnectedContent(t *testing.T) {
	info := StatusInfo{Connected: false}
	result := RenderStatus(info)
	if !strings.Contains(result, "Disconnected") {
		t.Error("should show 'Disconnected'")
	}
}
