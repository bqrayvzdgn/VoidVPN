package ui

import (
	"strings"
	"testing"
)

func TestBanner(t *testing.T) {
	result := Banner()
	if result == "" {
		t.Error("Banner() should return non-empty string")
	}
}

func TestBannerContent(t *testing.T) {
	result := Banner()
	// The banner art uses Unicode block characters, check for those
	if !strings.Contains(result, "██") {
		t.Error("Banner() should contain Unicode block art")
	}
	if !strings.Contains(result, "Secure. Private. Void.") {
		t.Error("Banner() should contain tagline 'Secure. Private. Void.'")
	}
}
