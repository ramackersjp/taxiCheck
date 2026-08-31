package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Regression test: Ctrl+C must not close the app, so it can be used to copy
// text (CLI text selection, clipboard inside the report textarea) instead.
func quits(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

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

func TestReportTypingQDoesNotQuit(t *testing.T) {
	m := Model{screen: screenReport, lang: "en"}
	m.initReportInputs()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if quits(cmd) {
		t.Fatal("typing q in the issue form must not quit")
	}
	mm := updated.(Model)
	if mm.screen != screenReport {
		t.Fatalf("screen=%d, want report", mm.screen)
	}
	if !strings.Contains(mm.reportDesc.Value(), "q") {
		t.Fatalf("expected q to be typed, got %q", mm.reportDesc.Value())
	}
}

func TestReportPasteDoesNotQuit(t *testing.T) {
	m := Model{screen: screenReport, lang: "en"}
	m.initReportInputs()
	m.reportFocus = 1
	m.reportDesc.Blur()
	m.reportErr.Focus()
	pasted := "fatal: not a git repository\nrequest failed"
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(pasted), Paste: true})
	if quits(cmd) {
		t.Fatal("pasting into the issue form must not quit")
	}
	mm := updated.(Model)
	if mm.screen != screenReport {
		t.Fatalf("screen=%d, want report", mm.screen)
	}
	if !strings.Contains(mm.reportErr.Value(), "not a git repository") {
		t.Fatalf("paste not inserted, got %q", mm.reportErr.Value())
	}
}

func TestCalcTypingQDoesNotQuit(t *testing.T) {
	in := textinput.New()
	in.Focus()
	m := Model{
		screen:   screenCalc,
		lang:     "en",
		inputs:   []textinput.Model{in, textinput.New(), textinput.New()},
		focusIdx: 0,
	}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if quits(cmd) {
		t.Fatal("typing q in an address field must not quit")
	}
	if updated.(Model).screen != screenCalc {
		t.Fatalf("screen=%d, want calc", updated.(Model).screen)
	}
}

func TestReportRightClickDoesNotLeave(t *testing.T) {
	m := Model{screen: screenReport, lang: "en", width: 80, height: 40}
	m.initReportInputs()
	updated, _ := m.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonRight,
		X:      10,
		Y:      10,
	})
	if updated.(Model).screen != screenReport {
		t.Fatalf("right-click while typing should paste, not leave; screen=%d", updated.(Model).screen)
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
