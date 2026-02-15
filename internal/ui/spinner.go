package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ConnectMsg struct{ Err error }

type SpinnerModel struct {
	spinner  spinner.Model
	message  string
	quitting bool
	err      error
}

func NewSpinner(message string) SpinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(Cyan)
	return SpinnerModel{spinner: s, message: message}
}

func (m SpinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			m.quitting = true
			return m, tea.Quit
		}
	case ConnectMsg:
		m.quitting = true
		m.err = msg.Err
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m SpinnerModel) View() string {
	if m.quitting {
		if m.err != nil {
			return ErrorStyle.Render(fmt.Sprintf("✗ %s\n", m.err))
		}
		return SuccessStyle.Render("✓ Connected!\n")
	}
	return fmt.Sprintf("%s %s\n", m.spinner.View(), AccentStyle.Render(m.message))
}
