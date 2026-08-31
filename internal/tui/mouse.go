package tui

import (
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func (m Model) pasteClipboard() (tea.Model, tea.Cmd) {
	s, err := clipboard.ReadAll()
	if err != nil || s == "" {
		return m, nil
	}
	return m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s), Paste: true})
}

func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "f2":
		return tea.KeyMsg{Type: tea.KeyF2}
	case "f3":
		return tea.KeyMsg{Type: tea.KeyF3}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if tea.MouseEvent(msg).IsWheel() {
		if msg.Button == tea.MouseButtonWheelUp {
			return m.handleKey(keyMsg("up"))
		}
		if msg.Button == tea.MouseButtonWheelDown {
			return m.handleKey(keyMsg("down"))
		}
		return m, nil
	}
	if msg.Action != tea.MouseActionPress {
		return m, nil
	}
	switch msg.Button {
	case tea.MouseButtonRight:
		if m.isTyping() {
			return m.pasteClipboard()
		}
		return m.handleKey(keyMsg("esc"))
	case tea.MouseButtonLeft:
		return m.handleClick(msg.X, msg.Y)
	}
	return m, nil
}

func (m Model) handleClick(x, y int) (tea.Model, tea.Cmd) {
	lines := strings.Split(ansi.Strip(m.View()), "\n")
	if y < 0 || y >= len(lines) {
		return m, nil
	}
	line := lines[y]
	if !xOnContent(line, x) {
		return m, nil
	}
	content := contentOf(line)

	switch m.screen {
	case screenHelp:
		if k := leadingKey(content); k != "" {
			switch k {
			case "1", "2", "3", "4", "5", "6", "7", "8", "q":
				return m.updateMain(keyMsg(k))
			}
		}
		if k := leadingKey(content); k != "" {
			return m.handleKey(keyMsg(k))
		}
		if content != "" {
			return m.handleKey(keyMsg("esc"))
		}
	case screenCalc:
		return m.clickCalc(lines, y, content)
	case screenBranch:
		return m.clickBranch(content)
	case screenSetup:
		return m.clickSetup(lines, y, content)
	case screenSettings:
		return m.clickSettings(lines, y, content)
	case screenReport:
		return m.clickReport(lines, y, content)
	case screenResult:
		if content != "" {
			return m.handleKey(keyMsg("enter"))
		}
	default:
		if k := leadingKey(content); k != "" {
			return m.handleKey(keyMsg(k))
		}
	}
	return m, nil
}

func (m Model) clickCalc(lines []string, y int, content string) (tea.Model, tea.Cmd) {
	if k := leadingKey(content); k == "f2" {
		return m.handleKey(keyMsg("f2"))
	}
	if m.showSuggest {
		for i, s := range m.suggestions {
			if s.Display != "" && strings.Contains(content, truncateDisplay(s.Display, m.contentWidth()-8)) {
				m.suggestionIdx = i
				return m.handleKey(keyMsg("enter"))
			}
		}
	}
	labels := []string{
		strings.TrimSpace(t(m.lang, "calc_label_start")),
		strings.TrimSpace(t(m.lang, "calc_label_end")),
		strings.TrimSpace(t(m.lang, "calc_label_passengers")),
	}
	if idx := fieldIndexAt(lines, y, labels); idx >= 0 {
		m.showSuggest = false
		m.suggestions = nil
		m.suggestFetching = false
		m.suggestPending = false
		m.suggestGen++
		return m.focusInputAt(idx)
	}
	return m, nil
}

func (m Model) clickBranch(content string) (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(content)
	name = strings.TrimPrefix(name, "▸")
	name = strings.TrimPrefix(name, "*")
	name = strings.TrimSpace(name)
	for i, b := range m.branchList {
		if name == b {
			if m.branchIdx == i && b != m.currentBranch {
				return m.handleKey(keyMsg(" "))
			}
			m.branchIdx = i
			return m, nil
		}
	}
	if k := leadingKey(content); k != "" {
		return m.handleKey(keyMsg(k))
	}
	return m, nil
}

func (m Model) clickSetup(lines []string, y int, content string) (tea.Model, tea.Cmd) {
	if m.setupStep == 0 {
		return m.clickLang(content)
	}
	labels := []string{
		strings.TrimSpace(t(m.lang, "setup_board_fee")),
		strings.TrimSpace(t(m.lang, "setup_per_km")),
		strings.TrimSpace(t(m.lang, "setup_per_minute")),
		strings.TrimSpace(t(m.lang, "setup_wait_minute")),
	}
	if idx := repeatingFieldIndex(lines, y, labels, len(m.inputs)); idx >= 0 {
		return m.focusInputAt(idx)
	}
	if k := leadingKey(content); k != "" {
		return m.handleKey(keyMsg(k))
	}
	return m, nil
}

func (m Model) clickSettings(lines []string, y int, content string) (tea.Model, tea.Cmd) {
	if m.settingsStep == 0 {
		return m.clickLang(content)
	}
	labels := []string{
		strings.TrimSpace(t(m.lang, "settings_board_fee")),
		strings.TrimSpace(t(m.lang, "settings_per_km")),
		strings.TrimSpace(t(m.lang, "settings_per_minute")),
		strings.TrimSpace(t(m.lang, "settings_wait_minute")),
	}
	if idx := repeatingFieldIndex(lines, y, labels, len(m.inputs)); idx >= 0 {
		return m.focusInputAt(idx)
	}
	if k := leadingKey(content); k != "" {
		return m.handleKey(keyMsg(k))
	}
	return m, nil
}

func (m Model) clickLang(content string) (tea.Model, tea.Cmd) {
	if strings.HasPrefix(content, ">") {
		return m.handleKey(keyMsg("enter"))
	}
	if k := leadingKey(content); k != "" {
		return m.handleKey(keyMsg(k))
	}
	return m, nil
}

func (m Model) clickReport(lines []string, y int, content string) (tea.Model, tea.Cmd) {
	if m.reportSubmitted {
		if content != "" {
			return m.handleKey(keyMsg("esc"))
		}
		return m, nil
	}
	descY, errY := -1, -1
	desc := strings.TrimSpace(t(m.lang, "report_desc_label"))
	errl := strings.TrimSpace(t(m.lang, "report_err_label"))
	for i, line := range lines {
		c := contentOf(line)
		if descY < 0 && strings.Contains(c, desc) {
			descY = i
		}
		if errY < 0 && strings.Contains(c, errl) {
			errY = i
		}
	}
	switch {
	case descY >= 0 && y >= descY && (errY < 0 || y < errY):
		m.reportFocus = 0
		m.reportDesc.Focus()
		m.reportErr.Blur()
		return m, textinput.Blink
	case errY >= 0 && y >= errY:
		m.reportFocus = 1
		m.reportErr.Focus()
		m.reportDesc.Blur()
		return m, textinput.Blink
	}
	if k := leadingKey(content); k != "" {
		return m.handleKey(keyMsg(k))
	}
	return m, nil
}

func (m Model) focusInputAt(idx int) (tea.Model, tea.Cmd) {
	if idx < 0 || idx >= len(m.inputs) {
		return m, nil
	}
	if m.focusIdx >= 0 && m.focusIdx < len(m.inputs) {
		m.inputs[m.focusIdx].Blur()
	}
	m.focusIdx = idx
	m.inputs[idx].Focus()
	return m, textinput.Blink
}

func fieldIndexAt(lines []string, y int, labels []string) int {
	for i, line := range lines {
		c := contentOf(line)
		for fi, lab := range labels {
			if lab != "" && strings.Contains(c, lab) {
				if y == i || y == i-1 {
					return fi
				}
			}
		}
	}
	return -1
}

func repeatingFieldIndex(lines []string, y int, labels []string, n int) int {
	idx := 0
	for i, line := range lines {
		c := contentOf(line)
		for _, lab := range labels {
			if lab != "" && strings.Contains(c, lab) {
				if y == i || y == i-1 {
					if idx < n {
						return idx
					}
				}
				idx++
			}
		}
	}
	return -1
}

func leadingKey(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}
	if strings.HasPrefix(content, "F3") || strings.HasPrefix(content, "f3") {
		return "f3"
	}
	if strings.HasPrefix(content, "F2") || strings.HasPrefix(content, "f2") {
		return "f2"
	}
	tok := content
	if i := strings.IndexAny(content, " \t-"); i >= 0 {
		tok = content[:i]
	}
	switch tok {
	case "1", "2", "3", "4", "5", "6", "7", "8", "q", "y", "n", "Y", "N", "U", "u", "R", "r":
		return strings.ToLower(tok)
	}
	return ""
}

func contentOf(line string) string {
	s := strings.TrimSpace(line)
	s = strings.TrimPrefix(s, "│")
	s = strings.TrimSuffix(s, "│")
	return strings.TrimSpace(s)
}

func xOnContent(line string, x int) bool {
	runes := []rune(line)
	start, end := -1, -1
	for i, r := range runes {
		if r != ' ' {
			if start < 0 {
				start = i
			}
			end = i
		}
	}
	if start < 0 {
		return false
	}
	return x >= start && x <= end
}

func truncateDisplay(display string, maxLen int) string {
	if maxLen < 20 {
		maxLen = 20
	}
	if len(display) > maxLen {
		return display[:maxLen-3] + "..."
	}
	return display
}
