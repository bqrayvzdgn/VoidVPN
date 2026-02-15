// Package ui provides terminal UI components and styling for VoidVPN.
package ui

import "github.com/charmbracelet/lipgloss"

var (
	Purple = lipgloss.Color("#7B2FBE")
	Cyan   = lipgloss.Color("#00D4FF")
	White  = lipgloss.Color("#FFFFFF")
	Gray   = lipgloss.Color("#808080")
	Red    = lipgloss.Color("#FF4444")
	Green  = lipgloss.Color("#44FF44")
	Yellow = lipgloss.Color("#FFFF44")

	TitleStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Purple).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Green).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Red).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Yellow)

	DimStyle = lipgloss.NewStyle().
			Foreground(Gray)

	AccentStyle = lipgloss.NewStyle().
			Foreground(Cyan)

	PurpleStyle = lipgloss.NewStyle().
			Foreground(Purple)

	LabelStyle = lipgloss.NewStyle().
			Foreground(Gray).
			Width(18)

	ValueStyle = lipgloss.NewStyle().
			Foreground(White)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Purple).
			Padding(1, 2)
)
