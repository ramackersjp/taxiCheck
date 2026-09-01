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
	if m.screen != screenMain {
		t.Fatal("home should be the fare form")
	}
	if !strings.Contains(ansi.Strip(m.View()), "Start address") {
		t.Fatalf("home missing start field:\n%s", ansi.Strip(m.View()))
	}
	updated, _ := click(m, "Start address")
	if updated.(Model).screen != screenMain {
		t.Fatalf("screen=%d, want main", updated.(Model).screen)
	}
	if !updated.(Model).calcFocused() {
		t.Fatal("start field should be focused")
	}
}

func TestMouseClickMainMenuQuit(t *testing.T) {
	m := sized("en")
	updated, _ := click(m, "Menu")
	_, cmd := click(updated.(Model), "Quit")
	if cmd == nil {
		t.Fatal("clicking Quit should quit")
	}
}

func TestMouseRightClickGoesBack(t *testing.T) {
	m := sized("en")
	updated, _ := click(m, "Menu")
	updated, _ = click(updated.(Model), "Help")
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
	updated, _ := click(m, "Menu")
	updated, _ = click(updated.(Model), "Uninstall")
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
	if m.routeMode != "fastest" {
		t.Fatalf("routeMode=%q", m.routeMode)
	}
	updated, _ := click(m, "F2")
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
	m.screen = screenMain
	m.calcInputs = []textinput.Model{in, textinput.New(), textinput.New()}
	m.calcFocus = 0
	m.showSuggest = true
	m.suggestions = []routing.AddressSuggestion{{Display: "Dam, Amsterdam"}}
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	m = updated.(Model)
	updated, _ = click(m, "Dam, Amsterdam")
	if got := updated.(Model).calcInputs[0].Value(); got != "Dam, Amsterdam" {
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

func TestMouseReleaseOpensCalc(t *testing.T) {
	m := sized("en")
	x, y := findText(m, "Start address")
	updated, _ := m.Update(tea.MouseMsg{
		X:      x,
		Y:      y,
		Action: tea.MouseActionRelease,
		Button: tea.MouseButtonLeft,
	})
	if !updated.(Model).calcFocused() {
		t.Fatal("release click should focus the start field")
	}
}

func TestMouseClickOffByOneStillHits(t *testing.T) {
	m := sized("en")
	x, y := findText(m, "Start address")
	lines := strings.Split(ansi.Strip(m.View()), "\n")
	ty := y - 1
	if ty < 0 || strings.TrimSpace(contentOf(lines[ty])) != "" {
		ty = y + 2
		if ty >= len(lines) || strings.TrimSpace(contentOf(lines[ty])) != "" {
			t.Skip("no blank line next to the menu item")
		}
	}
	updated, _ := m.Update(tea.MouseMsg{
		X:      x,
		Y:      ty,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	if !updated.(Model).calcFocused() && updated.(Model).screen != screenMain {
		t.Fatalf("blank-line click at Y=%d (item %d) screen=%d", ty, y, updated.(Model).screen)
	}
}

func TestMouseClickRowMarginHitsKey(t *testing.T) {
	m := sized("en")
	_, y := findText(m, "Menu")
	updated, _ := m.Update(tea.MouseMsg{
		X:      0,
		Y:      y,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	if updated.(Model).screen != screenMenu {
		t.Fatalf("margin click screen=%d, want menu", updated.(Model).screen)
	}
}

func TestMousePressThenReleaseDoesNotDouble(t *testing.T) {
	m := sized("en")
	x, y := findText(m, "Menu")
	updated, _ := m.Update(tea.MouseMsg{
		X: x, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft,
	})
	if updated.(Model).screen != screenMenu {
		t.Fatal("press should open the menu")
	}
	updated, _ = updated.Update(tea.MouseMsg{
		X: x, Y: y, Action: tea.MouseActionRelease, Button: tea.MouseButtonLeft,
	})
	if updated.(Model).screen != screenMenu {
		t.Fatalf("release after press should not leave the menu, screen=%d", updated.(Model).screen)
	}
}

func TestCloseXReturnsToMain(t *testing.T) {
	m := sized("en")
	updated, _ := click(m, "Menu")
	updated, _ = click(updated.(Model), "Settings")
	if updated.(Model).screen != screenSettings {
		t.Fatal("need settings")
	}
	x, y := findText(updated.(Model), "X")
	if y < 0 {
		t.Fatalf("missing close X:\n%s", ansi.Strip(updated.(Model).View()))
	}
	updated, _ = updated.Update(tea.MouseMsg{
		X: x, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft,
	})
	if updated.(Model).screen != screenMain {
		t.Fatalf("X should return home, screen=%d", updated.(Model).screen)
	}
}

func TestMouseClickSettings(t *testing.T) {
	m := sized("en")
	updated, _ := click(m, "Menu")
	updated, _ = click(updated.(Model), "Settings")
	if updated.(Model).screen != screenSettings {
		t.Fatalf("screen=%d view:\n%s", updated.(Model).screen, ansi.Strip(m.View()))
	}
}

func TestHomeMenuLinkAndMenuNumbers(t *testing.T) {
	m := sized("en")
	home := ansi.Strip(m.View())
	if !strings.Contains(home, "▸") || !strings.Contains(home, "Menu") {
		t.Fatalf("home should show a ▸ Menu link:\n%s", home)
	}
	if strings.Contains(home, "2 Settings") || strings.Contains(home, "3 Help") {
		t.Fatalf("home must not list the numbered menu:\n%s", home)
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	mm := updated.(Model)
	if mm.screen != screenMenu {
		t.Fatalf("1 should open the menu, screen=%d", mm.screen)
	}
	menu := ansi.Strip(mm.View())
	if !strings.Contains(menu, "1 Settings") {
		t.Fatalf("menu should start at 1 Settings:\n%s", menu)
	}
	if strings.Contains(menu, "2 Settings") {
		t.Fatalf("settings must not be 2:\n%s", menu)
	}
	if !strings.Contains(menu, "2 Help") || !strings.Contains(menu, "3 Check for Updates") {
		t.Fatalf("menu numbering off:\n%s", menu)
	}
}

func TestMouseClickInsideStartBoxFocusesField(t *testing.T) {
	m := sized("en")
	_, y := findText(m, "Start address")
	if y < 0 {
		t.Fatal("missing start label")
	}
	lines := strings.Split(ansi.Strip(m.View()), "\n")
	boxY := y + 2
	if boxY >= len(lines) || !strings.Contains(lines[boxY], "Dam Square") {
		boxY = y + 1
	}
	x := strings.Index(lines[boxY], "Dam")
	if x < 0 {
		x = strings.Index(lines[boxY], "│")
		if x < 0 {
			t.Fatalf("start box not at y=%d:\n%s", boxY, ansi.Strip(m.View()))
		}
		x++
	}
	updated, _ := m.Update(tea.MouseMsg{
		X: x, Y: boxY, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft,
	})
	if !updated.(Model).calcFocused() || updated.(Model).calcFocus != 0 {
		t.Fatalf("click inside start box: focus=%d focused=%v\n%s",
			updated.(Model).calcFocus, updated.(Model).calcFocused(), ansi.Strip(m.View()))
	}
}

func TestMouseClickPassengersDoesNotToggleF2(t *testing.T) {
	m := sized("en")
	if m.routeMode != "fastest" {
		t.Fatal("expected fastest")
	}
	x, y := findText(m, "Number of passengers")
	if y < 0 {
		t.Fatal("missing passengers label")
	}
	updated, _ := m.Update(tea.MouseMsg{
		X: x, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft,
	})
	mm := updated.(Model)
	if mm.routeMode != "fastest" {
		t.Fatal("clicking passengers must not toggle F2")
	}
	if mm.calcFocus != 2 {
		t.Fatalf("calcFocus=%d, want passengers", mm.calcFocus)
	}
}

func TestMouseClickF2ColumnTogglesMode(t *testing.T) {
	m := sized("en")
	x, y := findText(m, "F2")
	if y < 0 {
		t.Fatal("missing F2")
	}
	updated, _ := m.Update(tea.MouseMsg{
		X: x, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft,
	})
	if updated.(Model).routeMode != "shortest" {
		t.Fatalf("F2 column should toggle route, got %q", updated.(Model).routeMode)
	}
}

func TestPassModeRowDoesNotWrap(t *testing.T) {
	m := sized("en")
	lines := strings.Split(ansi.Strip(m.View()), "\n")
	_, y := findText(m, "Number of passengers")
	if y < 0 || y+1 >= len(lines) {
		t.Fatal("missing passengers row")
	}
	box := contentOf(lines[y+1])
	if strings.Count(box, "╭") != 2 {
		t.Fatalf("passenger and route boxes should sit on one row, got %q", box)
	}
	if strings.Contains(contentOf(lines[y+2]), "╭") {
		t.Fatalf("route box wrapped onto the next line:\n%s", strings.Join(lines[y:y+5], "\n"))
	}
}
