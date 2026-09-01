package tui

import (
	"strconv"
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
	return x >= col-2 && x <= col+lipgloss.Width("X")+2
}

func (m Model) fieldCard(label, inner string, focused bool) string {
	w := m.cardInnerWidth()
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

// cardInnerWidth is the lipgloss Width() of a full-width field box. Rounded
// borders add 2 columns on top of that; the result matches the outer frame.
func (m Model) cardInnerWidth() int {
	w := m.contentWidth() - 8
	if w < 20 {
		w = 20
	}
	return w
}

// rowOuterWidth is the on-screen width of a full-width field box (inner + borders).
func (m Model) rowOuterWidth() int {
	return m.cardInnerWidth() + 2
}

func clipLabel(s string, max int) string {
	if max < 4 {
		max = 4
	}
	if lipgloss.Width(s) <= max {
		return s
	}
	runes := []rune(s)
	w := 0
	for i := range runes {
		if w+1 > max-1 {
			return string(runes[:i]) + "…"
		}
		w++
	}
	return s
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

	b.WriteString(m.viewPassModeRow())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(t(m.lang, "calc_help_home")))
	return b.String()
}

func (m Model) viewPassModeRow() string {
	passInner := m.calcInputs[2].View()
	passLab := t(m.lang, "calc_label_passengers")
	passFocused := m.calcFocus == 2 && m.calcInputs[2].Focused()
	modeName := t(m.lang, "calc_mode_fastest")
	if m.routeMode != "fastest" {
		modeName = t(m.lang, "calc_mode_shortest")
	}
	modeInner := keyStyle.Render("F2") + "  " + successStyle.Render(modeName)
	modeLab := t(m.lang, "calc_mode")

	passCard := func(boxW int, title string) string {
		box := fieldBoxOff.Width(boxW)
		lab := helpStyle.Render(title)
		if passFocused {
			box = fieldBoxOn.Width(boxW)
			lab = fieldLabelOn.Render(title)
		}
		return lab + "\n" + box.Render(passInner)
	}
	modeCard := func(boxW int, title string) string {
		return fieldLabelOn.Render(title) + "\n" + fieldBoxOn.Width(boxW).Render(modeInner)
	}

	const gap = 2
	const minModeOuter = 22
	rowOuter := m.rowOuterWidth()
	passDot := "○ "
	if passFocused {
		passDot = "● "
	}
	passTitleRaw := passDot + passLab
	modeTitleRaw := "● " + modeLab
	passNeed := lipgloss.Width(passTitleRaw)
	if passNeed < 16 {
		passNeed = 16
	}
	if passNeed+gap+minModeOuter > rowOuter {
		// Not enough room: stack full-width cards so nothing wraps.
		return passCard(m.cardInnerWidth(), passTitleRaw) + "\n" + modeCard(m.cardInnerWidth(), modeTitleRaw)
	}
	passOuter := passNeed
	modeOuter := rowOuter - passOuter - gap
	passTitle := clipLabel(passTitleRaw, passOuter)
	modeTitle := clipLabel(modeTitleRaw, modeOuter)
	// Width() is content+padding; rounded borders add 2 columns.
	passBoxW := passOuter - 2
	modeBoxW := modeOuter - 2
	if passBoxW < 8 {
		passBoxW = 8
	}
	if modeBoxW < 10 {
		modeBoxW = 10
	}
	passStyled := helpStyle.Render(passTitle)
	if passFocused {
		passStyled = fieldLabelOn.Render(passTitle)
	}
	// Pin the title to the column width so JoinHorizontal does not expand
	// past the box (a longer unstyled label previously wrapped the row).
	passCol := lipgloss.NewStyle().Width(passOuter).MaxWidth(passOuter).Render(passStyled)
	modeCol := lipgloss.NewStyle().Width(modeOuter).MaxWidth(modeOuter).Render(fieldLabelOn.Render(modeTitle))
	passBox := fieldBoxOff.Width(passBoxW)
	if passFocused {
		passBox = fieldBoxOn.Width(passBoxW)
	}
	left := passCol + "\n" + passBox.Render(passInner)
	right := modeCol + "\n" + fieldBoxOn.Width(modeBoxW).Render(modeInner)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", gap), right)
}

func (m Model) viewMenuLink() string {
	return m.viewHomeActions()
}

func (m Model) viewHomeActions() string {
	left := keyStyle.Render("1") + "  " + arrowStyle.Render("▸") + "  " + fieldLabelOn.Render(t(m.lang, "main_more"))
	right := keyStyle.Render("2") + "  " + arrowStyle.Render("⌫") + "  " + fieldLabelOn.Render(t(m.lang, "main_clear"))
	w := m.contentWidth() - 4
	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 2 {
		gap = 2
	}
	line := left + strings.Repeat(" ", gap) + right
	if m.branchStatus != "" {
		return line + "\n\n" + successStyle.Render(m.branchStatus)
	}
	return line
}

func (m Model) clearCalcFields() (tea.Model, tea.Cmd) {
	m.ensureCalcInputs()
	for i := range m.calcInputs {
		m.calcInputs[i].SetValue("")
	}
	m.showSuggest = false
	m.suggestions = nil
	m.suggestFetching = false
	m.suggestPending = false
	m.suggestGen++
	m.err = ""
	m.result = nil
	m.routeInfo = nil
	return m.focusCalcAt(0)
}

func (m Model) menuItemIDs() []string {
	ids := []string{"settings", "help"}
	if !m.setupDone {
		ids = append(ids, "setup")
	}
	return append(ids, "update", "branch", "uninstall", "report")
}

func (m Model) menuItemLabel(id string) string {
	switch id {
	case "settings":
		return t(m.lang, "main_settings")
	case "help":
		return t(m.lang, "main_help")
	case "setup":
		return t(m.lang, "main_setup")
	case "update":
		return t(m.lang, "main_update")
	case "branch":
		return t(m.lang, "main_branch")
	case "uninstall":
		return t(m.lang, "main_uninstall")
	case "report":
		return t(m.lang, "main_report")
	default:
		return ""
	}
}

func (m Model) viewMenuItems(start int) string {
	var b strings.Builder
	n := start
	for _, id := range m.menuItemIDs() {
		b.WriteString(keyStyle.Render(strconv.Itoa(n)) + m.menuItemLabel(id) + "\n")
		n++
	}
	b.WriteString(keyStyle.Render("q") + t(m.lang, "main_quit") + "\n")
	if m.branchStatus != "" {
		b.WriteString("\n")
		b.WriteString(successStyle.Render(m.branchStatus))
	}
	return b.String()
}

func (m Model) viewMenu() string {
	var b strings.Builder
	b.WriteString(m.pageHeader(t(m.lang, "main_menu")))
	b.WriteString("\n")
	b.WriteString(m.viewMenuItems(1))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(t(m.lang, "main_select")))
	return b.String()
}
