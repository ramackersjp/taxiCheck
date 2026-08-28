package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Regression test: opening the report screen (menu option 8) used to panic
// because initReportInputs had a value receiver, so the freshly created
// textarea/textarea fields were discarded and the zero textarea's View()
// panicked on its nil cache.
func TestReportScreenDoesNotPanic(t *testing.T) {
	var m tea.Model = NewModel()
	var cmd tea.Cmd
	m, cmd = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	_ = cmd

	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("8")})
	_ = cmd

	if out := m.View(); out == "" {
		t.Fatal("expected a rendered report screen, got empty output")
	}
}
