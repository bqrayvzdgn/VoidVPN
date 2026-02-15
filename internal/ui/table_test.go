package ui

import (
	"strings"
	"testing"
)

func TestRenderTable(t *testing.T) {
	columns := []TableColumn{
		{Header: "Name", Width: 10},
		{Header: "Value", Width: 10},
	}
	rows := []TableRow{
		{"foo", "bar"},
		{"hello", "world"},
	}

	result := RenderTable(columns, rows)

	if !strings.Contains(result, "Name") {
		t.Error("table should contain header 'Name'")
	}
	if !strings.Contains(result, "foo") {
		t.Error("table should contain row value 'foo'")
	}
	if !strings.Contains(result, "world") {
		t.Error("table should contain row value 'world'")
	}
}

func TestRenderTableEmpty(t *testing.T) {
	columns := []TableColumn{
		{Header: "Name", Width: 10},
	}

	result := RenderTable(columns, nil)
	if !strings.Contains(result, "No entries found") {
		t.Error("empty table should show 'No entries found'")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1048576, "1.00 MB"},
		{1073741824, "1.00 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatBytes(tt.bytes)
			if got != tt.want {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		input string
		width int
		want  string
	}{
		{"hi", 5, "hi   "},
		{"hello", 3, "hel"},
		{"exact", 5, "exact"},
		{"", 3, "   "},
	}

	for _, tt := range tests {
		got := padRight(tt.input, tt.width)
		if got != tt.want {
			t.Errorf("padRight(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.want)
		}
	}
}
