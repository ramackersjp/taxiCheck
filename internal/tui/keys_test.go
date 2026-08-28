package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Regression test: Ctrl+C must not close the app, so it can be used to copy
// text (CLI text selection, clipboard inside the report textarea) instead.
func TestCtrlCDoesNotQuit(t *testing.T) {
	m := Model{screen: screenMain, lang: "en"}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd != nil {
		t.Fatalf("expected no command (Ctrl+C must not quit), got %v", cmd)
	}
	if updated.(Model).screen != screenMain {
		t.Fatalf("expected to stay on the main screen, got %d", updated.(Model).screen)
	}
}
