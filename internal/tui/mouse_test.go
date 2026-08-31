package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/ramackersjp/taxiCheck/internal/routing"
)

func sized(lang string) Model {
	m := sizedModel(lang)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	return updated.(Model)
}

func click(m tea.Model, substr string) (tea.Model, tea.Cmd) {
	x, y := findText(m.(Model), substr)
	if y < 0 {
		return m, nil
	}
	return m.Update(tea.MouseMsg{
		X:      x,
		Y:      y,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
}

func findText(m Model, substr string) (x, y int) {
	lines := strings.Split(ansi.Strip(m.View()), "\n")
	for i, line := range lines {
		idx := strings.Index(line, substr)
		if idx >= 0 {
			return idx, i
		}
	}
	return -1, -1
}

func TestMouseClickMainMenuOpensCalc(t *testing.T) {
	m := sized("en")
	updated, _ := click(m, "Calculate Fare")
	if updated.(Model).screen != screenCalc {
		t.Fatalf("screen=%d, want calc; view:\n%s", updated.(Model).screen, ansi.Strip(m.View()))
	}
}

func TestMouseClickMainMenuQuit(t *testing.T) {
	m := sized("en")
	_, cmd := click(m, "Quit")
	if cmd == nil {
		t.Fatal("clicking Quit should quit")
	}
}

func TestMouseRightClickGoesBack(t *testing.T) {
	m := sized("en")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("3")})
	if updated.(Model).screen != screenHelp {
		t.Fatalf("setup: screen=%d", updated.(Model).screen)
	}
	updated, _ = updated.Update(tea.MouseMsg{
		X:      10,
		Y:      10,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonRight,
	})
	if updated.(Model).screen != screenMain {
		t.Fatalf("right-click should go back, screen=%d", updated.(Model).screen)
	}
}

func TestMouseClickUninstallYes(t *testing.T) {
	m := sized("en")
	updated, _ := click(m, "Uninstall")
	if updated.(Model).screen != screenUninstall {
		t.Fatalf("screen=%d, want uninstall", updated.(Model).screen)
	}
	updated, _ = click(updated.(Model), "Yes, continue")
	if updated.(Model).uninstallStep != 1 {
		t.Fatalf("uninstallStep=%d, want 1", updated.(Model).uninstallStep)
	}
}

func TestMouseClickCalcF2(t *testing.T) {
	m := sized("en")
	updated, _ := click(m, "Calculate Fare")
	mm := updated.(Model)
	if mm.routeMode != "fastest" {
		t.Fatalf("routeMode=%q", mm.routeMode)
	}
	updated, _ = click(mm, "F2")
	if updated.(Model).routeMode != "shortest" {
		t.Fatalf("F2 click: routeMode=%q", updated.(Model).routeMode)
	}
}

func TestMouseClickBranchSelects(t *testing.T) {
	m := sized("en")
	m.screen = screenBranch
	m.branchList = []string{"dev", "v1.1.0"}
	m.currentBranch = "dev"
	m.branchIdx = 0
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	m = updated.(Model)
	updated, cmd := click(m, "v1.1.0")
	mm := updated.(Model)
	if mm.branchIdx != 1 {
		t.Fatalf("branchIdx=%d, want 1; view:\n%s", mm.branchIdx, ansi.Strip(m.View()))
	}
	if cmd != nil {
		t.Fatal("first click should only select")
	}
}

func TestMouseClickSuggestionPicksAddress(t *testing.T) {
	in := textinput.New()
	in.Focus()
	m := sized("en")
	m.screen = screenCalc
	m.inputs = []textinput.Model{in, textinput.New(), textinput.New()}
	m.focusIdx = 0
	m.showSuggest = true
	m.suggestions = []routing.AddressSuggestion{{Display: "Dam, Amsterdam"}}
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	m = updated.(Model)
	updated, _ = click(m, "Dam, Amsterdam")
	if got := updated.(Model).inputs[0].Value(); got != "Dam, Amsterdam" {
		t.Fatalf("value=%q", got)
	}
}

func TestMouseIgnoresMotion(t *testing.T) {
	m := sized("en")
	updated, cmd := m.Update(tea.MouseMsg{
		X:      0,
		Y:      0,
		Action: tea.MouseActionMotion,
		Button: tea.MouseButtonLeft,
	})
	if cmd != nil || updated.(Model).screen != screenMain {
		t.Fatal("motion must be ignored")
	}
}

func TestMouseClickSettings(t *testing.T) {
	m := sized("en")
	updated, _ := click(m, "Settings")
	if updated.(Model).screen != screenSettings {
		t.Fatalf("screen=%d view:\n%s", updated.(Model).screen, ansi.Strip(m.View()))
	}
}
