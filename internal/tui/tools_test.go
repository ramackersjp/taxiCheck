package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ramackersjp/taxiCheck/internal/config"
	"github.com/ramackersjp/taxiCheck/internal/tools"
)

func setupOnLang(lang string) Model {
	m := sized(lang)
	m.screen = screenSetup
	m.setupStep = stepLang
	m.setupDone = false
	m.config = config.DefaultConfig()
	return m
}

func TestSetupEnterOpensToolsThenNSkipsToPricing(t *testing.T) {
	orig := probeTools
	probeTools = func() tools.Status { return tools.Status{} }
	defer func() { probeTools = orig }()

	m := setupOnLang("en")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		updated, _ = updated.Update(cmd())
	}
	mm := updated.(Model)
	if mm.setupStep != stepTools {
		t.Fatalf("setupStep=%d, want tools", mm.setupStep)
	}
	out := mm.View()
	if !strings.Contains(out, "Git") || !strings.Contains(out, "Skip") {
		t.Fatalf("tools view missing Git/Skip:\n%s", out)
	}

	updated, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	mm = updated.(Model)
	if mm.setupStep != stepPricing {
		t.Fatalf("N should skip to pricing, step=%d", mm.setupStep)
	}
	if !strings.Contains(mm.View(), "Board Fee") && !strings.Contains(mm.View(), "Instap") {
		t.Fatalf("expected pricing fields, got:\n%s", mm.View())
	}
}

func TestSettingsEnterOpensTools(t *testing.T) {
	orig := probeTools
	probeTools = func() tools.Status {
		return tools.Status{Git: true, GH: true, LoggedIn: true, User: "tester"}
	}
	defer func() { probeTools = orig }()

	m := sized("en")
	m.screen = screenSettings
	m.settingsStep = stepLang
	m.settingsLang = "en"
	m.config = config.DefaultConfig()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		updated, _ = updated.Update(cmd())
	}
	mm := updated.(Model)
	if mm.settingsStep != stepTools {
		t.Fatalf("settingsStep=%d, want tools", mm.settingsStep)
	}
	if !strings.Contains(mm.View(), "tester") {
		t.Fatalf("expected logged-in user in settings tools:\n%s", mm.View())
	}
}

func TestToolsYInstallsMissing(t *testing.T) {
	origP, origI := probeTools, installMissing
	probeTools = func() tools.Status { return tools.Status{} }
	var called bool
	installMissing = func(s tools.Status) error {
		called = true
		if s.Git || s.GH {
			t.Fatal("expected missing tools")
		}
		return nil
	}
	defer func() {
		probeTools = origP
		installMissing = origI
	}()

	m := sized("en")
	m.screen = screenSetup
	m.setupStep = stepTools
	m.config = config.DefaultConfig()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if cmd == nil {
		t.Fatal("expected install command")
	}
	updated, cmd = updated.Update(cmd())
	if !called {
		t.Fatal("InstallMissing not called")
	}
	if cmd != nil {
		updated, _ = updated.Update(cmd())
	}
	if updated.(Model).setupStep != stepTools {
		t.Fatal("should stay on tools after install")
	}
}

func TestToolsNeedGHBeforeLogin(t *testing.T) {
	m := sized("en")
	m.screen = screenSettings
	m.settingsStep = stepTools
	m.tools = tools.Status{GH: false}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	if cmd != nil {
		t.Fatal("login without gh must not exec")
	}
	if updated.(Model).err == "" {
		t.Fatal("expected an error asking to install gh")
	}
}
