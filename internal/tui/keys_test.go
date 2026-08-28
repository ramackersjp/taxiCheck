package tui

import (
	"errors"
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

// Regression test: failures while fetching address suggestions must not show
// an error message (e.g. rate-limit notices during typing).
func TestSuggestErrorIsSilent(t *testing.T) {
	m := Model{screen: screenCalc, lang: "en", lastInputVal: "amsterdam"}
	updated, cmd := m.Update(suggestMsg{
		inputIdx: 0,
		query:    "amsterdam",
		err:      errors.New("rate limited"),
	})
	if cmd != nil {
		t.Fatalf("expected no command, got %v", cmd)
	}
	if mm := updated.(Model); mm.err != "" {
		t.Fatalf("expected the suggestion error to stay silent, got %q", mm.err)
	}
}
