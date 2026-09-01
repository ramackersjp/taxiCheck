package tui

import (
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	// Linux SGR often delivers a reliable Release; Windows usually delivers
	// Press. Handle both, but ignore the duplicate Release after a Press at
	// the same cell so a click does not fire twice.
	isLeft := msg.Button == tea.MouseButtonLeft || msg.Type == tea.MouseLeft
	isRight := msg.Button == tea.MouseButtonRight || msg.Type == tea.MouseRight
	switch msg.Action {
	case tea.MouseActionPress, tea.MouseActionRelease:
	default:
		return m, nil
	}
	if msg.Action == tea.MouseActionRelease && m.lastMousePress &&
		msg.X == m.lastMouseX && msg.Y == m.lastMouseY {
		m.lastMousePress = false
		return m, nil
	}
	if msg.Action == tea.MouseActionPress {
		m.lastMouseX, m.lastMouseY = msg.X, msg.Y
		m.lastMousePress = true
	} else {
		m.lastMousePress = false
	}
	switch {
	case isRight:
		if m.isTyping() {
			return m.pasteClipboard()
		}
		return m.handleKey(keyMsg("esc"))
	case isLeft:
		return m.handleClick(msg.X, msg.Y)
	}
	return m, nil
}

func (m Model) handleClick(x, y int) (tea.Model, tea.Cmd) {
	lines := strings.Split(ansi.Strip(m.View()), "\n")
	if m.screen != screenMain {
		for _, yy := range []int{y, y - 1, y + 1} {
			if yy >= 0 && yy < len(lines) && clickIsClose(lines[yy], x) {
				return m.handleKey(keyMsg("esc"))
			}
		}
	}
	content, y := hitContent(lines, x, y)
	if content == "" {
		return m, nil
	}

	switch m.screen {
	case screenMain:
		k := leadingKey(content)
		switch k {
		case "2", "3", "4", "5", "6", "7", "8", "q":
			return m.openMenu(k)
		}
		if m.setupDone {
			return m.clickCalc(lines, x, y, content)
		}
		if k != "" {
			return m.openMenu(k)
		}
	case screenHelp:
		if k := leadingKey(content); k != "" {
			switch k {
			case "1", "2", "3", "4", "5", "6", "7", "8", "q":
				return m.openMenu(k)
			}
		}
		if k := leadingKey(content); k != "" {
			return m.handleKey(keyMsg(k))
		}
		if content != "" {
			return m.handleKey(keyMsg("esc"))
		}
	case screenCalc:
		return m.clickCalc(lines, x, y, content)
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

func (m Model) clickCalc(lines []string, x, y int, content string) (tea.Model, tea.Cmd) {
	if m.showSuggest {
		for i, s := range m.suggestions {
			if s.Display != "" && strings.Contains(content, truncateDisplay(s.Display, m.contentWidth()-8)) {
				m.suggestionIdx = i
				return m.handleKey(keyMsg("enter"))
			}
		}
	}
	// Route/F2 sits on the same rows as passengers. Only treat the click as
	// F2 when X is in that column (or on the F2 token in the help line).
	if m.clickIsMode(lines, x, y) {
		return m.handleKey(keyMsg("f2"))
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
		return m.focusCalcAt(idx)
	}
	return m, nil
}

func (m Model) clickIsMode(lines []string, x, y int) bool {
	modeLab := strings.TrimSpace(t(m.lang, "calc_mode"))
	modeTop, modeLeft := -1, -1
	for i, line := range lines {
		c := contentOf(line)
		if modeLab == "" || !strings.Contains(c, modeLab) {
			continue
		}
		if strings.Contains(c, "Tab:") {
			continue
		}
		modeTop = i
		modeLeft = colOf(line, modeLab)
		if modeLeft > 1 {
			modeLeft -= 2 // "● "
		}
		break
	}
	if modeTop >= 0 && y >= modeTop-1 && y <= modeTop+4 && modeLeft >= 0 && x >= modeLeft {
		return true
	}
	if y >= 0 && y < len(lines) {
		if c := colOf(lines[y], "F2"); c >= 0 {
			return x >= c-1 && x <= c+3
		}
		if c := colOf(strings.ToLower(lines[y]), "f2"); c >= 0 {
			return x >= c-1 && x <= c+3
		}
	}
	return false
}

func colOf(line, substr string) int {
	if substr == "" {
		return -1
	}
	idx := strings.Index(line, substr)
	if idx < 0 {
		return -1
	}
	return lipgloss.Width(line[:idx])
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
	if m.setupStep == stepLang {
		return m.clickLang(content)
	}
	if m.setupStep == stepTools {
		if k := leadingKey(content); k != "" {
			return m.handleKey(keyMsg(k))
		}
		return m, nil
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
	if m.settingsStep == stepLang {
		return m.clickLang(content)
	}
	if m.settingsStep == stepTools {
		if k := leadingKey(content); k != "" {
			return m.handleKey(keyMsg(k))
		}
		return m, nil
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
	type hit struct{ fi, y int }
	var found []hit
	seen := map[int]bool{}
	for i, line := range lines {
		c := contentOf(line)
		for fi, lab := range labels {
			if lab == "" || seen[fi] || !strings.Contains(c, lab) {
				continue
			}
			found = append(found, hit{fi, i})
			seen[fi] = true
			break
		}
	}
	for i, h := range found {
		// Label plus the rounded box (top, value, bottom). Allow one row
		// above for terminals that report Y off by one.
		y0 := h.y - 1
		y1 := h.y + 4
		if i+1 < len(found) && found[i+1].y-1 < y1 {
			y1 = found[i+1].y - 1
		}
		if y >= y0 && y <= y1 {
			return h.fi
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
	case "1", "2", "3", "4", "5", "6", "7", "8", "q", "y", "n", "Y", "N", "U", "u", "R", "r", "I", "i", "G", "g", "L", "l", "A", "a", "X", "x":
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

// hitContent maps a click to a useful row. Prefer the exact cell, then an
// adjacent line (Linux/Windows can report Y off by one vs lipgloss.Place).
// Do not snap several rows away: that stole clicks from field cards onto
// nearby menu keys.
func hitContent(lines []string, x, y int) (string, int) {
	if c := contentAt(lines, x, y); c != "" {
		return c, y
	}
	for _, d := range []int{-1, 1} {
		if c := contentAt(lines, x, y+d); c != "" {
			return c, y + d
		}
	}
	if y >= 0 && y < len(lines) {
		return contentOf(lines[y]), y
	}
	return "", y
}

func contentAt(lines []string, x, y int) string {
	if y < 0 || y >= len(lines) {
		return ""
	}
	c := contentOf(lines[y])
	if c == "" {
		return ""
	}
	if leadingKey(c) != "" {
		return c
	}
	if xOnContent(lines[y], x) {
		return c
	}
	return ""
}

func xOnContent(line string, x int) bool {
	start, end := -1, -1
	col := 0
	for _, r := range line {
		w := 1
		if r == '\t' {
			w = 4
		}
		if r != ' ' {
			if start < 0 {
				start = col
			}
			end = col + w - 1
		}
		col += w
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
