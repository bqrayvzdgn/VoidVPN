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

func TestAddGatewayRouteInvalidGateway(t *testing.T) {
	err := addGatewayRoute("192.168.1.0", "255.255.255.0", "not-an-ip")
	if err == nil {
		t.Error("expected error for invalid gateway")
	}
	if !strings.Contains(err.Error(), "invalid gateway") {
		t.Errorf("error should contain 'invalid gateway': %v", err)
	}
}

func TestSplitNoSeparator(t *testing.T) {
	result := split("nosep", '\n')
	if len(result) != 1 || result[0] != "nosep" {
		t.Errorf("split() = %v, want [\"nosep\"]", result)
	}
}

func TestSplitEmptyString(t *testing.T) {
	result := split("", '\n')
	if len(result) != 1 || result[0] != "" {
		t.Errorf("split(\"\") = %v, want [\"\"]", result)
	}
}

func TestSplitMultipleSeparators(t *testing.T) {
	result := split("a\nb\nc", '\n')
	if len(result) != 3 {
		t.Errorf("split() len = %d, want 3", len(result))
	}
	if result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Errorf("split() = %v, want [a b c]", result)
	}
}

func TestSplitConsecutiveSeparators(t *testing.T) {
	result := split("a\n\nb", '\n')
	if len(result) != 3 {
		t.Errorf("split() len = %d, want 3", len(result))
	}
	if result[1] != "" {
		t.Errorf("split()[1] = %q, want \"\"", result[1])
	}
}

func TestRemoveVPNRoutesEmpty(t *testing.T) {
	r := &windowsRoutes{}
	err := r.RemoveVPNRoutes()
	if err != nil {
		t.Errorf("RemoveVPNRoutes() on empty should return nil: %v", err)
	}
}

func TestSplitFieldsMultipleSpaces(t *testing.T) {
	result := splitFields("  one   two   three  ")
	if len(result) != 3 {
		t.Errorf("splitFields() len = %d, want 3", len(result))
	}
	want := []string{"one", "two", "three"}
	for i, w := range want {
		if result[i] != w {
			t.Errorf("splitFields()[%d] = %q, want %q", i, result[i], w)
		}
	}
}

func TestSplitFieldsTabs(t *testing.T) {
	result := splitFields("foo\tbar\tbaz")
	if len(result) != 3 {
		t.Errorf("splitFields() len = %d, want 3", len(result))
	}
}

func TestTrimSpaceNoWhitespace(t *testing.T) {
	result := trimSpace("hello")
	if result != "hello" {
		t.Errorf("trimSpace(\"hello\") = %q, want \"hello\"", result)
	}
}

func TestTrimSpaceAllWhitespace(t *testing.T) {
	result := trimSpace("   \t\r\n  ")
	if result != "" {
		t.Errorf("trimSpace(whitespace) = %q, want \"\"", result)
	}
}
