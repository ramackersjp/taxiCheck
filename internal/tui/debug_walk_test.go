package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/ramackersjp/taxiCheck/internal/config"
)

func sizedModel(lang string) Model {
	m := Model{
		screen:    screenMain,
		lang:      lang,
		config:    config.DefaultConfig(),
		setupDone: true,
		routeMode: "fastest",
		width:     80,
		height:    40,
	}
	m.ensureCalcInputs()
	return m
}

func press(m tea.Model, keys ...string) tea.Model {
	for _, k := range keys {
		var msg tea.KeyMsg
		switch k {
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		case "tab":
			msg = tea.KeyMsg{Type: tea.KeyTab}
		case "up":
			msg = tea.KeyMsg{Type: tea.KeyUp}
		case "down":
			msg = tea.KeyMsg{Type: tea.KeyDown}
		case "f2":
			msg = tea.KeyMsg{Type: tea.KeyF2}
		case "f3":
			msg = tea.KeyMsg{Type: tea.KeyF3}
		case "space":
			msg = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		m, _ = m.Update(msg)
	}
	return m
}

func TestWalkAllScreensDoNotPanic(t *testing.T) {
	m := sizedModel("nl")
	var tm tea.Model = m
	tm, _ = tm.Update(tea.WindowSizeMsg{Width: 80, Height: 40})

	screens := []struct {
		key  string
		want string
	}{
		{"esc", "Menu"},
		{"1", "Hoofdmenu"},
		{"1", "Instellingen"},
		{"esc", "Menu"},
		{"1", "Hoofdmenu"},
		{"2", "Help"},
		{"esc", "Menu"},
		{"1", "Hoofdmenu"},
		{"3", "Updates"},
		{"esc", "Menu"},
		{"1", "Hoofdmenu"},
		{"4", "Wissel Branch"},
		{"esc", "Menu"},
		{"1", "Hoofdmenu"},
		{"5", "Verwijderen"},
		{"esc", "Menu"},
		{"1", "Hoofdmenu"},
		{"6", "Probleem Melden"},
		{"esc", "Menu"},
	}

	for _, s := range screens {
		tm = press(tm, s.key)
		out := tm.View()
		if out == "" {
			t.Fatalf("empty view after %q", s.key)
		}
		if !strings.Contains(out, s.want) {
			t.Fatalf("after %q, view missing %q\n%s", s.key, s.want, out)
		}
		if s.want == "Help" {
			if !strings.Contains(out, "druk U om te pullen") || !strings.Contains(out, "F3") {
				t.Fatalf("help must explain pull then F3, got:\n%s", out)
			}
		}
	}
}

func TestBranchListKeepsMarkerOnSameLine(t *testing.T) {
	m := sizedModel("nl")
	m.screen = screenBranch
	m.currentBranch = "dev"
	m.branchList = []string{"dev", "v1.0.1", "v1.0.0"}
	m.branchIdx = 1
	out := m.View()
	if !strings.Contains(out, "*") {
		t.Fatalf("expected current-branch marker, got:\n%s", out)
	}
	// The marker and the branch name must share a line. helpStyle used to
	// inject MarginTop, which rendered as:
	//   *
	//   dev
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "*") && !strings.Contains(line, "dev") {
			t.Fatalf("branch marker wrapped onto its own line:\n%s", out)
		}
	}
}

func TestSettingsPricingStartsAtFirstField(t *testing.T) {
	m := sizedModel("en")
	m.focusIdx = 2 // leftover from the calc passenger field
	m.inputs = make([]textinput.Model, 3)
	var tm tea.Model = m
	tm = press(tm, "esc")   // blur fare fields so menu keys work
	tm = press(tm, "1")     // menu
	tm = press(tm, "1")     // settings
	tm = press(tm, "enter") // language -> git/github tools
	tm = press(tm, "enter") // tools -> pricing
	mm := tm.(Model)
	if mm.screen != screenSettings || mm.settingsStep != stepPricing {
		t.Fatalf("screen=%d step=%d", mm.screen, mm.settingsStep)
	}
	if mm.focusIdx != 0 {
		t.Fatalf("focusIdx=%d, want 0 so typing edits the first rate", mm.focusIdx)
	}
	if !mm.inputs[0].Focused() {
		t.Fatal("first pricing field must be focused")
	}
	if mm.inputs[2].Focused() {
		t.Fatal("stale calc focus must not remain on field 2")
	}
}

func TestCalcValidationAndF2(t *testing.T) {
	m := sizedModel("en")
	var tm tea.Model = m
	tm = press(tm, "enter")
	mm := tm.(Model)
	if mm.err == "" {
		t.Fatal("empty start address should error")
	}
	tm = press(tm, "f2")
	mm = tm.(Model)
	if mm.routeMode != "shortest" {
		t.Fatalf("routeMode=%q, want shortest", mm.routeMode)
	}
}

func TestHeaderShowsDateTimeSourceLicense(t *testing.T) {
	orig := nowFn
	nowFn = func() time.Time { return time.Date(2026, 9, 1, 14, 32, 8, 0, time.UTC) }
	defer func() { nowFn = orig }()
	m := sized("en")
	out := ansi.Strip(m.View())
	for _, want := range []string{"TAXI", "TaxiCheck", "Date", "Time", "14:32:08", "github.com/ramackersjp/taxiCheck", "MIT", "J.P. Ramackers"} {
		if !strings.Contains(out, want) {
			t.Fatalf("header missing %q:\n%s", want, out)
		}
	}
	h := m.viewHeader()
	content := borderStyle.Width(m.contentWidth()).Render("x")
	if lipgloss.Width(h) != lipgloss.Width(content) {
		t.Fatalf("header width %d, content %d", lipgloss.Width(h), lipgloss.Width(content))
	}
	if lipgloss.Height(h) != lipgloss.Height(GetLogo()) {
		t.Fatalf("info column height %d, logo %d", lipgloss.Height(h), lipgloss.Height(GetLogo()))
	}
	if lipgloss.Width(GetLogo()) != 19 {
		t.Fatalf("logo box width %d, want 19", lipgloss.Width(GetLogo()))
	}
}

func TestLogoPutsVersionBesideTitle(t *testing.T) {
	orig := appVersion
	appVersion = "dev"
	defer func() { appVersion = orig }()
	out := GetLogo()
	if !strings.Contains(out, "TaxiCheck") {
		t.Fatalf("logo missing title:\n%s", out)
	}
	if !strings.Contains(out, "vdev") && !strings.Contains(out, "dev") {
		t.Fatalf("logo missing version:\n%s", out)
	}
	// titleStyle MarginBottom + helpStyle MarginTop used to push the
	// version onto a later line than "TaxiCheck".
	lines := strings.Split(out, "\n")
	found := false
	for _, line := range lines {
		if strings.Contains(line, "TaxiCheck") && strings.Contains(line, "dev") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("version should sit on the same line as TaxiCheck:\n%s", out)
	}
}

func TestUninstallTwoStepStaysOnScreenUntilConfirm(t *testing.T) {
	m := sizedModel("en")
	var tm tea.Model = m
	tm = press(tm, "esc")
	tm = press(tm, "1")
	tm = press(tm, "5")
	tm = press(tm, "y")
	mm := tm.(Model)
	if mm.uninstallStep != 1 {
		t.Fatalf("step=%d, want 1", mm.uninstallStep)
	}
	if mm.loading {
		t.Fatal("must not uninstall after the first yes")
	}
	out := mm.View()
	if !strings.Contains(out, "REALLY") && !strings.Contains(out, "HELEMAAL") {
		t.Fatalf("expected second confirmation, got:\n%s", out)
	}
}

func TestEscCancelsLoading(t *testing.T) {
	m := sizedModel("nl")
	m.screen = screenUpdate
	m.loading = true
	m.opGen = 3
	var tm tea.Model = m
	tm = press(tm, "esc")
	mm := tm.(Model)
	if mm.loading {
		t.Fatal("Esc must cancel loading")
	}
	if mm.screen != screenMain {
		t.Fatalf("screen=%d, want main", mm.screen)
	}
	if mm.opGen != 4 {
		t.Fatalf("opGen=%d, want 4 so in-flight results are ignored", mm.opGen)
	}

	stale := asyncMsg{gen: 3, inner: updateCheckMsg{hasUpdate: true, latestTag: "v9.9.9"}}
	tm, _ = tm.Update(stale)
	mm = tm.(Model)
	if mm.hasUpdate {
		t.Fatal("cancelled update check must not apply")
	}
}

func TestReportEmptySubmitIsRejected(t *testing.T) {
	m := sizedModel("en")
	var tm tea.Model = m
	tm, _ = tm.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	tm = press(tm, "esc")
	tm = press(tm, "1")
	tm = press(tm, "6")
	tm = press(tm, "enter")
	mm := tm.(Model)
	if mm.err == "" {
		t.Fatal("empty report should be rejected")
	}
	if mm.loading {
		t.Fatal("empty report must not submit")
	}
}
