package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) ensureCalcInputs() {
	if len(m.calcInputs) == 3 {
		m.styleCalcInputs()
		return
	}
	inputW := m.inputWidth()
	m.calcInputs = make([]textinput.Model, 3)
	for i := range m.calcInputs {
		m.calcInputs[i] = textinput.New()
		m.calcInputs[i].CharLimit = 80
		m.calcInputs[i].Width = inputW
		m.calcInputs[i].Prompt = ""
		m.calcInputs[i].PlaceholderStyle = helpStyle
	}
	m.calcInputs[0].Placeholder = t(m.lang, "calc_placeholder_start")
	m.calcInputs[1].Placeholder = t(m.lang, "calc_placeholder_end")
	m.calcInputs[2].Placeholder = t(m.lang, "calc_placeholder_passengers")
	m.calcInputs[2].CharLimit = 2
	m.calcInputs[2].Width = 4
	m.calcFocus = 0
	m.calcInputs[0].Focus()
	m.styleCalcInputs()
}

func (m *Model) styleCalcInputs() {
	if len(m.calcInputs) < 3 {
		return
	}
	on := lipgloss.NewStyle().Foreground(白色)
	off := lipgloss.NewStyle().Foreground(灰色)
	for i := range m.calcInputs {
		m.calcInputs[i].Prompt = ""
		if i == m.calcFocus && m.calcInputs[i].Focused() {
			m.calcInputs[i].TextStyle = on
			m.calcInputs[i].Cursor.Style = keyStyle
		} else {
			m.calcInputs[i].TextStyle = off
		}
	}
}

func (m Model) calcFocused() bool {
	if len(m.calcInputs) < 3 || m.calcFocus < 0 || m.calcFocus > 2 {
		return false
	}
	return m.calcInputs[m.calcFocus].Focused()
}

func (m Model) blurCalc() Model {
	for i := range m.calcInputs {
		m.calcInputs[i].Blur()
	}
	m.showSuggest = false
	m.suggestions = nil
	return m
}

func (m Model) focusCalcAt(idx int) (tea.Model, tea.Cmd) {
	if idx < 0 || idx >= len(m.calcInputs) {
		return m, nil
	}
	for i := range m.calcInputs {
		m.calcInputs[i].Blur()
	}
	m.calcFocus = idx
	m.calcInputs[idx].Focus()
	m.styleCalcInputs()
	m.showSuggest = false
	m.suggestFetching = false
	m.suggestPending = false
	m.suggestGen++
	return m, textinput.Blink
}

func (m Model) pageHeader(title string) string {
	btn := closeStyle.Render("X")
	left := lipgloss.NewStyle().Foreground(白色).Render(title)
	w := m.contentWidth() - 4
	if w < 10 {
		w = 10
	}
	gap := w - lipgloss.Width(left) - lipgloss.Width(btn)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + btn + "\n"
}

func clickIsClose(line string, x int) bool {
	idx := strings.LastIndex(line, "X")
	if idx < 0 {
		return false
	}
	col := lipgloss.Width(line[:idx])
	trimmed := strings.TrimSpace(contentOf(line))
	if !strings.HasSuffix(trimmed, "X") {
		return false
	}
	return x >= col-1 && x <= col+lipgloss.Width("X")+1
}

func (m Model) fieldCard(label, inner string, focused bool) string {
	w := m.contentWidth() - 8
	if w < 20 {
		w = 20
	}
	dot := "○ "
	lab := helpStyle.Render(dot + label)
	box := fieldBoxOff.Width(w)
	if focused {
		dot = "● "
		lab = fieldLabelOn.Render(dot + label)
		box = fieldBoxOn.Width(w)
	}
	return lab + "\n" + box.Render(inner)
}

func (m Model) viewFareCard() string {
	var b strings.Builder
	b.WriteString(fieldLabelOn.Render(t(m.lang, "calc_title")) + "\n")
	if len(m.calcInputs) < 3 {
		return b.String()
	}
	m.styleCalcInputs()
	b.WriteString(m.fieldCard(t(m.lang, "calc_label_start"), m.calcInputs[0].View(), m.calcFocus == 0 && m.calcInputs[0].Focused()))
	b.WriteString("\n")
	if m.showSuggest && m.suggestInput == 0 {
		b.WriteString(m.renderSuggestions())
	}
	b.WriteString(m.fieldCard(t(m.lang, "calc_label_end"), m.calcInputs[1].View(), m.calcFocus == 1 && m.calcInputs[1].Focused()))
	b.WriteString("\n")
	if m.showSuggest && m.suggestInput == 1 {
		b.WriteString(m.renderSuggestions())
	}

	passW := 14
	if passW > m.contentWidth()/3 {
		passW = 8
	}
	passInner := m.calcInputs[2].View()
	passLab := t(m.lang, "calc_label_passengers")
	passFocused := m.calcFocus == 2 && m.calcInputs[2].Focused()
	passDot := "○ "
	passTitle := helpStyle.Render(passDot + passLab)
	passBox := fieldBoxOff.Width(passW)
	if passFocused {
		passTitle = fieldLabelOn.Render("● " + passLab)
		passBox = fieldBoxOn.Width(passW)
	}
	passCard := passTitle + "\n" + passBox.Render(passInner)

	mode := t(m.lang, "calc_mode_fastest")
	if m.routeMode != "fastest" {
		mode = t(m.lang, "calc_mode_shortest")
	}
	modeInner := keyStyle.Render("F2") + "  " + successStyle.Render(mode)
	modeW := m.contentWidth() - 8 - passW - 2
	if modeW < 16 {
		modeW = 16
	}
	modeCard := fieldLabelOn.Render("● "+t(m.lang, "calc_mode")) + "\n" + fieldBoxOn.Width(modeW).Render(modeInner)
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, passCard, "  ", modeCard))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(t(m.lang, "calc_help_home")))
	return b.String()
}

func (m Model) viewMainMenuKeys() string {
	var b strings.Builder
	b.WriteString(helpStyle.Render("─ "+t(m.lang, "main_menu")+" ─") + "\n")
	b.WriteString(keyStyle.Render("2") + t(m.lang, "main_settings") + "\n")
	b.WriteString(keyStyle.Render("3") + t(m.lang, "main_help") + "\n")
	if !m.setupDone {
		b.WriteString(keyStyle.Render("4") + t(m.lang, "main_setup") + "\n")
	}
	b.WriteString(keyStyle.Render("5") + t(m.lang, "main_update") + "\n")
	b.WriteString(keyStyle.Render("6") + t(m.lang, "main_branch") + "\n")
	b.WriteString(keyStyle.Render("7") + t(m.lang, "main_uninstall") + "\n")
	b.WriteString(keyStyle.Render("8") + t(m.lang, "main_report") + "\n")
	b.WriteString(keyStyle.Render("q") + t(m.lang, "main_quit") + "\n")
	if m.branchStatus != "" {
		b.WriteString("\n")
		b.WriteString(successStyle.Render(m.branchStatus))
	}
	return b.String()
}
