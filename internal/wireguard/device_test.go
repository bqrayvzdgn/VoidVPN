package wireguard

import (
	"testing"
)

func TestParseAddress(t *testing.T) {
	tests := []struct {
		addr    string
		bits    int
		wantErr bool
	}{
		{"10.0.0.1/24", 24, false},
		{"10.0.0.1", 32, false},
		{"invalid", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			prefix, err := ParseAddress(tt.addr)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseAddress(%q) should return error", tt.addr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseAddress(%q) error: %v", tt.addr, err)
			}
			if prefix.Bits() != tt.bits {
				t.Errorf("ParseAddress(%q).Bits() = %d, want %d", tt.addr, prefix.Bits(), tt.bits)
			}
		})
	}
}

func TestParseAddressIPv6(t *testing.T) {
	tests := []struct {
		addr string
		bits int
	}{
		{"fd00::1/64", 64},
		{"fd00::1", 32}, // appends /32 even for IPv6 (documents current behavior)
	}
	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			prefix, err := ParseAddress(tt.addr)
			if err != nil {
				t.Fatalf("ParseAddress(%q) error: %v", tt.addr, err)
			}
			if prefix.Bits() != tt.bits {
				t.Errorf("ParseAddress(%q).Bits() = %d, want %d", tt.addr, prefix.Bits(), tt.bits)
			}
		})
	}
}
