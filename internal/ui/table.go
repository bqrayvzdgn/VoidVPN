package ui

import (
	"fmt"
	"strings"
)

type TableColumn struct {
	Header string
	Width  int
}

type TableRow []string

func RenderTable(columns []TableColumn, rows []TableRow) string {
	var sb strings.Builder

	// Header
	var headerParts []string
	var separatorParts []string
	for _, col := range columns {
		headerParts = append(headerParts, TitleStyle.Render(padRight(col.Header, col.Width)))
		separatorParts = append(separatorParts, DimStyle.Render(strings.Repeat("─", col.Width)))
	}
	sb.WriteString(strings.Join(headerParts, DimStyle.Render(" │ ")))
	sb.WriteString("\n")
	sb.WriteString(strings.Join(separatorParts, DimStyle.Render("─┼─")))
	sb.WriteString("\n")

	// Rows
	for _, row := range rows {
		var parts []string
		for i, col := range columns {
			val := ""
			if i < len(row) {
				val = row[i]
			}
			parts = append(parts, ValueStyle.Render(padRight(val, col.Width)))
		}
		sb.WriteString(strings.Join(parts, DimStyle.Render(" │ ")))
		sb.WriteString("\n")
	}

	if len(rows) == 0 {
		sb.WriteString(DimStyle.Render("  No entries found."))
		sb.WriteString("\n")
	}

	return sb.String()
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
