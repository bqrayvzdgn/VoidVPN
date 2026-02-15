//go:build windows

package network

import (
	"strings"
	"testing"
)

func TestSplitLines(t *testing.T) {
	input := "line1\r\nline2\r\nline3\r\n"
	got := splitLines(input)
	if len(got) != 3 {
		t.Errorf("splitLines() returned %d lines, want 3", len(got))
	}
	for _, line := range got {
		if strings.TrimSpace(line) == "" {
			t.Error("splitLines() should not return empty lines")
		}
	}
}

func TestSplitLinesEmpty(t *testing.T) {
	got := splitLines("")
	if len(got) != 0 {
		t.Errorf("splitLines(\"\") returned %d lines, want 0", len(got))
	}
}

func TestSplitFields(t *testing.T) {
	got := splitFields("  foo  bar  baz  ")
	want := []string{"foo", "bar", "baz"}
	if len(got) != len(want) {
		t.Fatalf("splitFields() returned %d fields, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("splitFields()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSplitFieldsEmpty(t *testing.T) {
	got := splitFields("")
	if len(got) != 0 {
		t.Errorf("splitFields(\"\") returned %d fields, want 0", len(got))
	}
}

func TestTrimSpace(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"  hello  ", "hello"},
		{"\thello\t", "hello"},
		{"\r\nhello\r\n", "hello"},
		{"hello", "hello"},
		{"  \t  ", ""},
	}
	for _, tt := range tests {
		got := trimSpace(tt.input)
		if got != tt.want {
			t.Errorf("trimSpace(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestAddGatewayRouteInvalidNetwork(t *testing.T) {
	err := addGatewayRoute("not-an-ip", "255.255.255.255", "192.168.1.1")
	if err == nil {
		t.Fatal("addGatewayRoute should error for invalid network")
	}
	if !strings.Contains(err.Error(), "invalid network") {
		t.Errorf("error = %q, want mention of invalid network", err.Error())
	}
}

func TestAddGatewayRouteInvalidMask(t *testing.T) {
	err := addGatewayRoute("192.168.1.0", "bad", "192.168.1.1")
	if err == nil {
		t.Fatal("addGatewayRoute should error for invalid subnet mask")
	}
	if !strings.Contains(err.Error(), "invalid subnet mask") {
		t.Errorf("error = %q, want mention of invalid subnet mask", err.Error())
	}
}
