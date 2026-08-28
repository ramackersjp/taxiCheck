package tui

import (
	"errors"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
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
	m := Model{screen: screenCalc, lang: "en", lastInputVal: "amsterdam", suggestGen: 1}
	updated, cmd := m.Update(suggestMsg{
		inputIdx: 0,
		query:    "amsterdam",
		gen:      1,
		err:      errors.New("rate limited"),
	})
	if cmd != nil {
		t.Fatalf("expected no command, got %v", cmd)
	}
	if mm := updated.(Model); mm.err != "" {
		t.Fatalf("expected the suggestion error to stay silent, got %q", mm.err)
	}
}

func TestStaleSuggestionTickIsIgnored(t *testing.T) {
	in := textinput.New()
	in.SetValue("amsterdam")
	m := Model{
		screen:     screenCalc,
		lang:       "en",
		inputs:     []textinput.Model{in, textinput.New(), textinput.New()},
		focusIdx:   0,
		suggestGen: 4,
	}
	updated, cmd := m.Update(tickMsg{gen: 1})
	if cmd != nil {
		t.Fatalf("stale debounce tick must not fetch, got %v", cmd)
	}
	if mm := updated.(Model); mm.suggestFetching {
		t.Fatal("stale tick must not start a fetch")
	}
}

func TestSuggestSuccessDoesNotSetError(t *testing.T) {
	m := Model{
		screen:       screenCalc,
		lang:         "en",
		lastInputVal: "damrak",
		suggestInput: 0,
		suggestGen:   2,
		err:          "leftover",
	}
	updated, cmd := m.Update(suggestMsg{
		inputIdx: 0,
		query:    "damrak",
		gen:      2,
		suggests: nil,
	})
	if cmd != nil {
		t.Fatalf("expected no command, got %v", cmd)
	}
	mm := updated.(Model)
	if mm.err != "leftover" {
		t.Fatalf("suggestion results must not clear/set the calc error, got %q", mm.err)
	}
	if mm.showSuggest {
		t.Fatal("empty suggestion list should not be shown")
	}
}
