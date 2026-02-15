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
