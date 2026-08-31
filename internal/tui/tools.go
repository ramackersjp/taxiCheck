package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ramackersjp/taxiCheck/internal/tools"
)

const (
	stepLang    = 0
	stepTools   = 1
	stepPricing = 2
)

type toolsStatusMsg struct {
	status tools.Status
}

type toolsResultMsg struct {
	kind string
	err  error
}

var (
	probeTools     = tools.Probe
	installMissing = tools.InstallMissing
	installGitFn   = tools.InstallGit
	installGHFn    = tools.InstallGH
	openSignup     = tools.OpenSignup
	loginCmd       = tools.LoginCommand
)

func (m Model) refreshTools() tea.Cmd {
	return func() tea.Msg {
		return toolsStatusMsg{status: probeTools()}
	}
}

func (m Model) beginPricing() (tea.Model, tea.Cmd) {
	m.inputs = make([]textinput.Model, len(m.config.PassengerGroups)*4)
	pw := m.priceInputWidth()
	j := 0
	for _, g := range m.config.PassengerGroups {
		m.inputs[j] = textinput.New()
		m.inputs[j].CharLimit = 10
		m.inputs[j].Width = pw
		m.inputs[j].SetValue(fmt.Sprintf("%.2f", g.BoardFee))
		j++
		m.inputs[j] = textinput.New()
		m.inputs[j].CharLimit = 10
		m.inputs[j].Width = pw
		m.inputs[j].SetValue(fmt.Sprintf("%.2f", g.PerKm))
		j++
		m.inputs[j] = textinput.New()
		m.inputs[j].CharLimit = 10
		m.inputs[j].Width = pw
		m.inputs[j].SetValue(fmt.Sprintf("%.2f", g.PerMinute))
		j++
		m.inputs[j] = textinput.New()
		m.inputs[j].CharLimit = 10
		m.inputs[j].Width = pw
		m.inputs[j].SetValue(fmt.Sprintf("%.2f", g.WaitMinute))
		j++
	}
	m.focusIdx = 0
	if len(m.inputs) > 0 {
		m.inputs[0].Focus()
	}
	m.err = ""
	return m, textinput.Blink
}

func (m Model) updateTools(msg tea.KeyMsg, setup bool) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "n", "N", "enter":
		if setup {
			m.setupStep = stepPricing
		} else {
			m.settingsStep = stepPricing
		}
		return m.beginPricing()
	case "y", "Y":
		st := m.tools
		return m.runOp(func() tea.Msg {
			err := installMissing(st)
			return toolsResultMsg{kind: "all", err: err}
		})
	case "i", "I":
		return m.runOp(func() tea.Msg {
			err := installGitFn()
			return toolsResultMsg{kind: "git", err: err}
		})
	case "g", "G":
		return m.runOp(func() tea.Msg {
			err := installGHFn()
			return toolsResultMsg{kind: "gh", err: err}
		})
	case "a", "A":
		if err := openSignup(); err != nil {
			m.err = err.Error()
		}
		return m, nil
	case "l", "L":
		if !m.tools.GH {
			m.err = t(m.lang, "tools_need_gh")
			return m, nil
		}
		return m, tea.ExecProcess(loginCmd(), func(err error) tea.Msg {
			return toolsResultMsg{kind: "login", err: err}
		})
	}
	return m, nil
}

func (m Model) viewTools(setup bool) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(t(m.lang, "tools_title")))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render(t(m.lang, "tools_intro")) + "\n\n")

	b.WriteString(t(m.lang, "tools_git") + "  " + toolsState(m.lang, m.tools.Git) + "\n")
	b.WriteString(t(m.lang, "tools_gh") + "  " + toolsState(m.lang, m.tools.GH) + "\n")
	login := t(m.lang, "tools_logged_out")
	if m.tools.LoggedIn {
		login = successStyle.Render(t(m.lang, "tools_logged_in"))
		if m.tools.User != "" {
			login += " " + m.tools.User
		}
	} else {
		login = errorStyle.Render(login)
	}
	b.WriteString(t(m.lang, "tools_github") + "  " + login + "\n\n")

	b.WriteString(keyStyle.Render("I") + t(m.lang, "tools_install_git") + "\n")
	b.WriteString(keyStyle.Render("G") + t(m.lang, "tools_install_gh") + "\n")
	b.WriteString(keyStyle.Render("L") + t(m.lang, "tools_login") + "\n")
	b.WriteString(keyStyle.Render("A") + t(m.lang, "tools_account") + "\n")
	b.WriteString(keyStyle.Render("Y") + t(m.lang, "tools_install_all") + "\n")
	if setup {
		b.WriteString(keyStyle.Render("N") + t(m.lang, "tools_skip") + "\n")
	} else {
		b.WriteString(keyStyle.Render("Enter") + t(m.lang, "tools_continue") + "\n")
	}
	b.WriteString("\n")
	if setup {
		b.WriteString(helpStyle.Render(t(m.lang, "tools_setup_help")))
	} else {
		b.WriteString(helpStyle.Render(t(m.lang, "tools_settings_help")))
	}
	return b.String()
}

func toolsState(lang string, ok bool) string {
	if ok {
		return successStyle.Render(t(lang, "tools_ok"))
	}
	return errorStyle.Render(t(lang, "tools_missing"))
}
