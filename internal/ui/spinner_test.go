package ui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewSpinnerMessage(t *testing.T) {
	m := NewSpinner("Connecting...")
	if m.message != "Connecting..." {
		t.Errorf("message = %q, want \"Connecting...\"", m.message)
	}
	if m.quitting {
		t.Error("should not be quitting initially")
	}
	if m.err != nil {
		t.Error("err should be nil initially")
	}
}

func TestSpinnerViewConnecting(t *testing.T) {
	m := NewSpinner("Connecting to server...")
	view := m.View()
	if !strings.Contains(view, "Connecting to server...") {
		t.Errorf("View() should contain message: %q", view)
	}
}

func TestSpinnerViewSuccess(t *testing.T) {
	m := NewSpinner("test")
	m.quitting = true
	m.err = nil
	view := m.View()
	if !strings.Contains(view, "Connected") {
		t.Errorf("View() should show success: %q", view)
	}
}

func TestSpinnerViewError(t *testing.T) {
	m := NewSpinner("test")
	m.quitting = true
	m.err = errors.New("auth failed")
	view := m.View()
	if !strings.Contains(view, "auth failed") {
		t.Errorf("View() should show error: %q", view)
	}
}

func TestSpinnerUpdateConnectMsgSuccess(t *testing.T) {
	m := NewSpinner("test")
	updated, cmd := m.Update(ConnectMsg{Err: nil})
	model := updated.(SpinnerModel)
	if !model.quitting {
		t.Error("should be quitting after ConnectMsg")
	}
	if model.err != nil {
		t.Error("err should be nil on success")
	}
	if cmd == nil {
		t.Error("should return tea.Quit cmd")
	}
}

func TestSpinnerUpdateConnectMsgError(t *testing.T) {
	m := NewSpinner("test")
	updated, _ := m.Update(ConnectMsg{Err: errors.New("failed")})
	model := updated.(SpinnerModel)
	if !model.quitting {
		t.Error("should be quitting after error")
	}
	if model.err == nil {
		t.Error("err should be set")
	}
}

func TestSpinnerUpdateCtrlC(t *testing.T) {
	m := NewSpinner("test")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	model := updated.(SpinnerModel)
	if !model.quitting {
		t.Error("should be quitting after ctrl+c")
	}
	if cmd == nil {
		t.Error("should return tea.Quit cmd")
	}
}

func TestSpinnerUpdateQKey(t *testing.T) {
	m := NewSpinner("test")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model := updated.(SpinnerModel)
	if !model.quitting {
		t.Error("should be quitting after 'q' key")
	}
	if cmd == nil {
		t.Error("should return tea.Quit cmd")
	}
}

func TestSpinnerUpdateUnknownMsg(t *testing.T) {
	m := NewSpinner("test")
	_, cmd := m.Update("unknown message")
	if cmd != nil {
		t.Error("unknown message should return nil cmd")
	}
}

func TestSpinnerInit(t *testing.T) {
	m := NewSpinner("test")
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return spinner tick command")
	}
}
