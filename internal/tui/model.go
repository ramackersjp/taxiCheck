package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jp/taxiprijs/internal/calc"
	"github.com/jp/taxiprijs/internal/config"
)

type screen int

const (
	screenMain screen = iota
	screenSetup
	screenSettings
	screenHelp
	screenCalc
	screenResult
)

type Model struct {
	screen    screen
	config    *config.Config
	inputs    []textinput.Model
	focusIdx  int
	passGroup int
	calcInput calc.FareInput
	result    *calc.FareResult
	err       string
}

func NewModel() Model {
	cfg, err := config.Load()
	if err != nil {
		return Model{
			screen: screenMain,
			err:    err.Error(),
		}
	}

	if cfg == nil {
		return Model{
			screen: screenSetup,
		}
	}

	return Model{
		screen: screenMain,
		config: cfg,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenMain:
		return m.updateMain(msg)
	case screenSetup:
		return m.updateSetup(msg)
	case screenSettings:
		return m.updateSettings(msg)
	case screenHelp:
		return m.updateHelp(msg)
	case screenCalc:
		return m.updateCalc(msg)
	case screenResult:
		return m.updateResult(msg)
	}
	return m, nil
}

func (m Model) updateMain(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "1":
		m.screen = screenCalc
		m.calcInput = calc.FareInput{}
		m.inputs = make([]textinput.Model, 3)
		for i := range m.inputs {
			m.inputs[i] = textinput.New()
			m.inputs[i].CharLimit = 10
			m.inputs[i].Width = 10
		}
		m.inputs[0].Placeholder = "minutes"
		m.inputs[0].Focus()
		m.inputs[1].Placeholder = "wait minutes"
		m.inputs[2].Placeholder = "passengers"
		m.focusIdx = 0
		return m, textinput.Blink
	case "2":
		m.screen = screenSettings
		m.inputs = make([]textinput.Model, len(m.config.PassengerGroups)*3)
		j := 0
		for i, g := range m.config.PassengerGroups {
			m.inputs[j] = textinput.New()
			m.inputs[j].Placeholder = fmt.Sprintf("Group %d board fee", i+1)
			m.inputs[j].SetValue(fmt.Sprintf("%.2f", g.BoardFee))
			m.inputs[j].CharLimit = 10
			m.inputs[j].Width = 10
			j++
			m.inputs[j] = textinput.New()
			m.inputs[j].Placeholder = fmt.Sprintf("Group %d per minute", i+1)
			m.inputs[j].SetValue(fmt.Sprintf("%.2f", g.PerMinute))
			m.inputs[j].CharLimit = 10
			m.inputs[j].Width = 10
			j++
			m.inputs[j] = textinput.New()
			m.inputs[j].Placeholder = fmt.Sprintf("Group %d wait minute", i+1)
			m.inputs[j].SetValue(fmt.Sprintf("%.2f", g.WaitMinute))
			m.inputs[j].CharLimit = 10
			m.inputs[j].Width = 10
			j++
		}
		m.focusIdx = 0
		if len(m.inputs) > 0 {
			m.inputs[0].Focus()
		}
		return m, textinput.Blink
	case "3":
		m.screen = screenHelp
	case "4":
		m.screen = screenSetup
		m.inputs = make([]textinput.Model, len(m.config.PassengerGroups)*3)
		j := 0
		for i := range m.config.PassengerGroups {
			m.inputs[j] = textinput.New()
			m.inputs[j].Placeholder = fmt.Sprintf("Group %d board fee", i+1)
			m.inputs[j].CharLimit = 10
			m.inputs[j].Width = 10
			j++
			m.inputs[j] = textinput.New()
			m.inputs[j].Placeholder = fmt.Sprintf("Group %d per minute", i+1)
			m.inputs[j].CharLimit = 10
			m.inputs[j].Width = 10
			j++
			m.inputs[j] = textinput.New()
			m.inputs[j].Placeholder = fmt.Sprintf("Group %d wait minute", i+1)
			m.inputs[j].CharLimit = 10
			m.inputs[j].Width = 10
			j++
		}
		m.focusIdx = 0
		if len(m.inputs) > 0 {
			m.inputs[0].Focus()
		}
		return m, textinput.Blink
	}
	return m, nil
}

func (m Model) updateSetup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "tab", "shift+tab":
		if len(m.inputs) == 0 {
			return m, nil
		}
		m.inputs[m.focusIdx].Blur()
		if msg.String() == "tab" {
			m.focusIdx = (m.focusIdx + 1) % len(m.inputs)
		} else {
			m.focusIdx = (m.focusIdx - 1 + len(m.inputs)) % len(m.inputs)
		}
		m.inputs[m.focusIdx].Focus()
		return m, textinput.Blink
	case "enter":
		if m.screen == screenSetup {
			return m.saveSetup()
		}
	case "esc":
		m.screen = screenMain
		m.err = ""
		return m, nil
	}

	for i := range m.inputs {
		if i == m.focusIdx {
			m.inputs[i].Update(msg)
		}
	}
	return m, nil
}

func (m Model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "tab", "shift+tab":
		if len(m.inputs) == 0 {
			return m, nil
		}
		m.inputs[m.focusIdx].Blur()
		if msg.String() == "tab" {
			m.focusIdx = (m.focusIdx + 1) % len(m.inputs)
		} else {
			m.focusIdx = (m.focusIdx - 1 + len(m.inputs)) % len(m.inputs)
		}
		m.inputs[m.focusIdx].Focus()
		return m, textinput.Blink
	case "enter":
		return m.saveSettings()
	case "esc":
		m.screen = screenMain
		m.err = ""
		return m, nil
	}

	for i := range m.inputs {
		if i == m.focusIdx {
			m.inputs[i].Update(msg)
		}
	}
	return m, nil
}

func (m Model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "enter":
		m.screen = screenMain
		return m, nil
	}
	return m, nil
}

func (m Model) updateCalc(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "tab", "shift+tab":
		if len(m.inputs) == 0 {
			return m, nil
		}
		m.inputs[m.focusIdx].Blur()
		if msg.String() == "tab" {
			m.focusIdx = (m.focusIdx + 1) % len(m.inputs)
		} else {
			m.focusIdx = (m.focusIdx - 1 + len(m.inputs)) % len(m.inputs)
		}
		m.inputs[m.focusIdx].Focus()
		return m, textinput.Blink
	case "enter":
		return m.calculateFare()
	case "esc":
		m.screen = screenMain
		m.err = ""
		return m, nil
	}

	for i := range m.inputs {
		if i == m.focusIdx {
			m.inputs[i].Update(msg)
		}
	}
	return m, nil
}

func (m Model) updateResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "enter":
		m.screen = screenCalc
		m.result = nil
		m.calcInput = calc.FareInput{}
		m.inputs = make([]textinput.Model, 3)
		for i := range m.inputs {
			m.inputs[i] = textinput.New()
			m.inputs[i].CharLimit = 10
			m.inputs[i].Width = 10
		}
		m.inputs[0].Placeholder = "minutes"
		m.inputs[0].Focus()
		m.inputs[1].Placeholder = "wait minutes"
		m.inputs[2].Placeholder = "passengers"
		m.focusIdx = 0
		return m, textinput.Blink
	}
	return m, nil
}

func (m Model) calculateFare() (tea.Model, tea.Cmd) {
	if len(m.inputs) < 3 {
		m.err = "Invalid input fields"
		return m, nil
	}

	minutes, err := strconv.ParseFloat(m.inputs[0].Value(), 64)
	if err != nil {
		m.err = "Invalid minutes value"
		return m, nil
	}

	waitTime, err := strconv.ParseFloat(m.inputs[1].Value(), 64)
	if err != nil {
		m.err = "Invalid wait time value"
		return m, nil
	}

	passengers, err := strconv.Atoi(m.inputs[2].Value())
	if err != nil || passengers < 1 || passengers > 5 {
		m.err = "Passengers must be 1-5"
		return m, nil
	}

	groups := make([]calc.PassengerGroup, len(m.config.PassengerGroups))
	for i, g := range m.config.PassengerGroups {
		groups[i] = calc.PassengerGroup{
			Name:       g.Name,
			BoardFee:   g.BoardFee,
			PerMinute:  g.PerMinute,
			WaitMinute: g.WaitMinute,
		}
	}

	result := calc.Calculate(calc.FareInput{
		Minutes:    minutes,
		WaitTime:   waitTime,
		Passengers: passengers,
	}, groups)

	m.result = &result
	m.screen = screenResult
	m.err = ""
	return m, nil
}

func (m Model) saveSetup() (tea.Model, tea.Cmd) {
	if len(m.inputs) < len(m.config.PassengerGroups)*3 {
		m.err = "Not all fields filled"
		return m, nil
	}

	for i := range m.config.PassengerGroups {
		boardFee, err := strconv.ParseFloat(m.inputs[i*3].Value(), 64)
		if err != nil {
			m.err = fmt.Sprintf("Invalid board fee for group %d", i+1)
			return m, nil
		}
		perMinute, err := strconv.ParseFloat(m.inputs[i*3+1].Value(), 64)
		if err != nil {
			m.err = fmt.Sprintf("Invalid per minute rate for group %d", i+1)
			return m, nil
		}
		waitMinute, err := strconv.ParseFloat(m.inputs[i*3+2].Value(), 64)
		if err != nil {
			m.err = fmt.Sprintf("Invalid wait minute rate for group %d", i+1)
			return m, nil
		}

		m.config.PassengerGroups[i].BoardFee = boardFee
		m.config.PassengerGroups[i].PerMinute = perMinute
		m.config.PassengerGroups[i].WaitMinute = waitMinute
	}

	if err := config.Save(m.config); err != nil {
		m.err = err.Error()
		return m, nil
	}

	m.screen = screenMain
	m.err = ""
	return m, nil
}

func (m Model) saveSettings() (tea.Model, tea.Cmd) {
	if len(m.inputs) < len(m.config.PassengerGroups)*3 {
		m.err = "Not all fields filled"
		return m, nil
	}

	for i := range m.config.PassengerGroups {
		boardFee, err := strconv.ParseFloat(m.inputs[i*3].Value(), 64)
		if err != nil {
			m.err = fmt.Sprintf("Invalid board fee for group %d", i+1)
			return m, nil
		}
		perMinute, err := strconv.ParseFloat(m.inputs[i*3+1].Value(), 64)
		if err != nil {
			m.err = fmt.Sprintf("Invalid per minute rate for group %d", i+1)
			return m, nil
		}
		waitMinute, err := strconv.ParseFloat(m.inputs[i*3+2].Value(), 64)
		if err != nil {
			m.err = fmt.Sprintf("Invalid wait minute rate for group %d", i+1)
			return m, nil
		}

		m.config.PassengerGroups[i].BoardFee = boardFee
		m.config.PassengerGroups[i].PerMinute = perMinute
		m.config.PassengerGroups[i].WaitMinute = waitMinute
	}

	if err := config.Save(m.config); err != nil {
		m.err = err.Error()
		return m, nil
	}

	m.screen = screenMain
	m.err = ""
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(GetLogo())
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("Dutch Taxi Fare Calculator"))
	b.WriteString("\n\n")

	switch m.screen {
	case screenMain:
		b.WriteString(m.viewMain())
	case screenSetup:
		b.WriteString(m.viewSetup())
	case screenSettings:
		b.WriteString(m.viewSettings())
	case screenHelp:
		b.WriteString(m.viewHelp())
	case screenCalc:
		b.WriteString(m.viewCalc())
	case screenResult:
		b.WriteString(m.viewResult())
	}

	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("Error: " + m.err))
	}

	return borderStyle.Render(b.String())
}

func (m Model) viewMain() string {
	var b strings.Builder
	b.WriteString(subtitleStyle.Render("Main Menu"))
	b.WriteString("\n\n")
	b.WriteString(keyStyle.Render("1") + " Calculate Fare\n")
	b.WriteString(keyStyle.Render("2") + " Settings\n")
	b.WriteString(keyStyle.Render("3") + " Help/Manual\n")
	b.WriteString(keyStyle.Render("4") + " Initial Setup\n")
	b.WriteString(keyStyle.Render("q") + " Quit\n")
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Select an option"))
	return b.String()
}

func (m Model) viewSetup() string {
	var b strings.Builder
	b.WriteString(subtitleStyle.Render("Initial Setup"))
	b.WriteString("\n\n")
	for i, g := range m.config.PassengerGroups {
		b.WriteString(titleStyle.Render(g.Name))
		b.WriteString("\n")
		idx := i * 3
		if idx < len(m.inputs) {
			b.WriteString(m.inputs[idx].View() + " Board Fee\n")
		}
		idx++
		if idx < len(m.inputs) {
			b.WriteString(m.inputs[idx].View() + " Per Minute\n")
		}
		idx++
		if idx < len(m.inputs) {
			b.WriteString(m.inputs[idx].View() + " Wait Minute\n")
		}
		b.WriteString("\n")
	}
	b.WriteString(helpStyle.Render("Tab: next field | Enter: save | Esc: cancel"))
	return b.String()
}

func (m Model) viewSettings() string {
	var b strings.Builder
	b.WriteString(subtitleStyle.Render("Settings"))
	b.WriteString("\n\n")
	for i, g := range m.config.PassengerGroups {
		b.WriteString(titleStyle.Render(g.Name))
		b.WriteString("\n")
		idx := i * 3
		if idx < len(m.inputs) {
			b.WriteString(m.inputs[idx].View() + " Board Fee\n")
		}
		idx++
		if idx < len(m.inputs) {
			b.WriteString(m.inputs[idx].View() + " Per Minute\n")
		}
		idx++
		if idx < len(m.inputs) {
			b.WriteString(m.inputs[idx].View() + " Wait Minute\n")
		}
		b.WriteString("\n")
	}
	b.WriteString(helpStyle.Render("Tab: next field | Enter: save | Esc: cancel"))
	return b.String()
}

func (m Model) viewHelp() string {
	var b strings.Builder
	b.WriteString(subtitleStyle.Render("Help & Manual"))
	b.WriteString("\n\n")
	b.WriteString("Keyboard Controls:\n")
	b.WriteString("  " + keyStyle.Render("1") + " - Calculate Fare\n")
	b.WriteString("  " + keyStyle.Render("2") + " - Settings\n")
	b.WriteString("  " + keyStyle.Render("3") + " - Help/Manual\n")
	b.WriteString("  " + keyStyle.Render("4") + " - Initial Setup\n")
	b.WriteString("  " + keyStyle.Render("q") + " - Quit\n")
	b.WriteString("  " + keyStyle.Render("Tab") + " - Next field\n")
	b.WriteString("  " + keyStyle.Render("Enter") + " - Submit/Save\n")
	b.WriteString("  " + keyStyle.Render("Esc") + " - Back\n")
	b.WriteString("\n")
	b.WriteString("Configuration is stored in:\n")
	b.WriteString("  ~/.taxiprijs/config.toml\n")
	b.WriteString("\n")
	b.WriteString("Passenger Categories:\n")
	b.WriteString("  1-4 passengers: Standard group\n")
	b.WriteString("  1-5 passengers: Larger group\n")
	b.WriteString("\n")
	b.WriteString("Pricing:\n")
	b.WriteString("  Board Fee: Initial charge\n")
	b.WriteString("  Per Minute: Cost per minute of ride\n")
	b.WriteString("  Wait Minute: Cost per minute of waiting\n")
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Press Esc or Enter to return to main menu"))
	return b.String()
}

func (m Model) viewCalc() string {
	var b strings.Builder
	b.WriteString(subtitleStyle.Render("Calculate Fare"))
	b.WriteString("\n\n")
	if len(m.inputs) >= 3 {
		b.WriteString(m.inputs[0].View() + " Minutes\n")
		b.WriteString(m.inputs[1].View() + " Wait Time (minutes)\n")
		b.WriteString(m.inputs[2].View() + " Passengers (1-5)\n")
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Tab: next field | Enter: calculate | Esc: cancel"))
	return b.String()
}

func (m Model) viewResult() string {
	var b strings.Builder
	b.WriteString(subtitleStyle.Render("Fare Result"))
	b.WriteString("\n\n")
	if m.result != nil {
		b.WriteString("Passenger Group: " + m.result.Group + "\n\n")
		b.WriteString("Board Fee:      €" + fmt.Sprintf("%.2f", m.result.BaseFee) + "\n")
		b.WriteString("Time Fee:       €" + fmt.Sprintf("%.2f", m.result.TimeFee) + "\n")
		b.WriteString("Wait Fee:       €" + fmt.Sprintf("%.2f", m.result.WaitFee) + "\n")
		b.WriteString("───────────────────\n")
		b.WriteString("Total:          " + successStyle.Render("€"+fmt.Sprintf("%.2f", m.result.Total)) + "\n")
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Press Enter or Esc to calculate again | q: quit"))
	return b.String()
}
