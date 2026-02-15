package cli

import (
	"testing"
)

func TestProtocolLabel(t *testing.T) {
	tests := []struct {
		protocol string
		want     string
	}{
		{"openvpn", "OVPN"},
		{"wireguard", "WG"},
		{"", "WG"},
		{"other", "WG"},
	}
	for _, tt := range tests {
		t.Run(tt.protocol, func(t *testing.T) {
			got := protocolLabel(tt.protocol)
			if got != tt.want {
				t.Errorf("protocolLabel(%q) = %q, want %q", tt.protocol, got, tt.want)
			}
		})
	}
}

func TestProtocolLabelDefault(t *testing.T) {
	got := protocolLabel("unknown")
	if got != "WG" {
		t.Errorf("protocolLabel(\"unknown\") = %q, want %q", got, "WG")
	}
}
