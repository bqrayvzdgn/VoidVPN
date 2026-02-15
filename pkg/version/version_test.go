package version

import (
	"strings"
	"testing"
)

func TestFull(t *testing.T) {
	result := Full()
	if !strings.Contains(result, "VoidVPN") {
		t.Error("Full() should contain 'VoidVPN'")
	}
	if !strings.Contains(result, Version) {
		t.Errorf("Full() should contain version %q", Version)
	}
}

func TestShort(t *testing.T) {
	if Short() != Version {
		t.Errorf("Short() = %q, want %q", Short(), Version)
	}
}
