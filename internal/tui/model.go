package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jp/taxiprijs/internal/calc"
	"github.com/jp/taxiprijs/internal/config"
	"github.com/jp/taxiprijs/internal/routing"
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

type routeMsg struct {
	result *calc.FareResult
	route  *routing.RouteResult
	start  string
	end    string
}

type routeErrMsg struct {
	err error
}

type Model struct {
	screen    screen
	config    *config.Config
	inputs    []textinput.Model
	focusIdx  int
	passGroup int
	calcInput calc.FareInput
	result    *calc.FareResult
	routeInfo *routing.RouteResult
	startAddr string
	endAddr   string
	err       string
	lang      string
	setupStep int
	loading   bool
}

func NewModel() Model {
	cfg, err := config.Load()
	if err != nil {
		return Model{
			screen: screenMain,
			lang:   "en",
			err:    err.Error(),
		}
	}

	if cfg == nil {
		cfg = config.DefaultConfig()
		return Model{
			screen:    screenSetup,
			config:    cfg,
			lang:      "en",
			setupStep: 0,
		}
	}

	lang := cfg.Language
	if lang == "" {
		lang = "en"
	}

	return Model{
		screen: screenMain,
		config: cfg,
		lang:   lang,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case routeMsg:
		m.loading = false
		m.result = msg.result
		m.routeInfo = msg.route
		m.startAddr = msg.start
		m.endAddr = msg.end
		m.screen = screenResult
		m.err = ""
		return m, nil
	case routeErrMsg:
		m.loading = false
		m.err = msg.err.Error()
		return m, nil
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.loading {
		return m, nil
	}

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
			m.inputs[i].CharLimit = 80
			m.inputs[i].Width = 40
		}
		m.inputs[0].Placeholder = t(m.lang, "calc_placeholder_start")
		m.inputs[0].Focus()
		m.inputs[1].Placeholder = t(m.lang, "calc_placeholder_end")
		m.inputs[2].Placeholder = t(m.lang, "calc_placeholder_passengers")
		m.inputs[2].CharLimit = 2
		m.inputs[2].Width = 5
		m.focusIdx = 0
		return m, textinput.Blink
	case "2":
		m.screen = screenSettings
		m.inputs = make([]textinput.Model, len(m.config.PassengerGroups)*4)
		j := 0
		for i, g := range m.config.PassengerGroups {
			m.inputs[j] = textinput.New()
			m.inputs[j].Placeholder = fmt.Sprintf("Group %d board fee", i+1)
			m.inputs[j].SetValue(fmt.Sprintf("%.2f", g.BoardFee))
			m.inputs[j].CharLimit = 10
			m.inputs[j].Width = 10
			j++
			m.inputs[j] = textinput.New()
			m.inputs[j].Placeholder = fmt.Sprintf("Group %d per km", i+1)
			m.inputs[j].SetValue(fmt.Sprintf("%.2f", g.PerKm))
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
		m.setupStep = 0
		m.inputs = nil
		return m, nil
	}
	return m, nil
}

func (m Model) updateSetup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc":
		if m.setupStep == 0 {
			m.screen = screenMain
			m.err = ""
			return m, nil
		}
		m.setupStep = 0
		m.inputs = nil
		m.err = ""
		return m, nil
	}

	if m.setupStep == 0 {
		return m.updateSetupLang(msg)
	}
	return m.updateSetupPricing(msg)
}

func (m Model) updateSetupLang(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "1", "up", "down":
		if msg.String() == "1" || msg.String() == "up" {
			m.lang = "en"
		} else {
			m.lang = "nl"
		}
		return m, nil
	case "2":
		m.lang = "nl"
		return m, nil
	case "enter":
		m.config.Language = m.lang
		m.setupStep = 1
		m.inputs = make([]textinput.Model, len(m.config.PassengerGroups)*4)
		j := 0
		for _, g := range m.config.PassengerGroups {
			m.inputs[j] = textinput.New()
			m.inputs[j].CharLimit = 10
			m.inputs[j].Width = 10
			m.inputs[j].SetValue(fmt.Sprintf("%.2f", g.BoardFee))
			j++
			m.inputs[j] = textinput.New()
			m.inputs[j].CharLimit = 10
			m.inputs[j].Width = 10
			m.inputs[j].SetValue(fmt.Sprintf("%.2f", g.PerKm))
			j++
			m.inputs[j] = textinput.New()
			m.inputs[j].CharLimit = 10
			m.inputs[j].Width = 10
			m.inputs[j].SetValue(fmt.Sprintf("%.2f", g.PerMinute))
			j++
			m.inputs[j] = textinput.New()
			m.inputs[j].CharLimit = 10
			m.inputs[j].Width = 10
			m.inputs[j].SetValue(fmt.Sprintf("%.2f", g.WaitMinute))
			j++
		}
		if len(m.inputs) > 0 {
			m.inputs[0].Focus()
		}
		return m, textinput.Blink
	}
	return m, nil
}

func (m Model) updateSetupPricing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
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
		return m.saveSetup()
	}

	for i := range m.inputs {
		if i == m.focusIdx {
			m.inputs[i], _ = m.inputs[i].Update(msg)
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
			m.inputs[i], _ = m.inputs[i].Update(msg)
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
		return m.startCalculation()
	case "esc":
		m.screen = screenMain
		m.err = ""
		return m, nil
	}

	for i := range m.inputs {
		if i == m.focusIdx {
			m.inputs[i], _ = m.inputs[i].Update(msg)
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
		m.routeInfo = nil
		m.calcInput = calc.FareInput{}
		m.inputs = make([]textinput.Model, 3)
		for i := range m.inputs {
			m.inputs[i] = textinput.New()
			m.inputs[i].CharLimit = 80
			m.inputs[i].Width = 40
		}
		m.inputs[0].Placeholder = t(m.lang, "calc_placeholder_start")
		m.inputs[0].Focus()
		m.inputs[1].Placeholder = t(m.lang, "calc_placeholder_end")
		m.inputs[2].Placeholder = t(m.lang, "calc_placeholder_passengers")
		m.inputs[2].CharLimit = 2
		m.inputs[2].Width = 5
		m.focusIdx = 0
		return m, textinput.Blink
	}
	return m, nil
}

func (m Model) startCalculation() (tea.Model, tea.Cmd) {
	if len(m.inputs) < 3 {
		m.err = t(m.lang, "err_invalid_input")
		return m, nil
	}

	startAddr := strings.TrimSpace(m.inputs[0].Value())
	endAddr := strings.TrimSpace(m.inputs[1].Value())
	passengersStr := strings.TrimSpace(m.inputs[2].Value())

	if startAddr == "" {
		m.err = t(m.lang, "err_empty_start")
		return m, nil
	}
	if endAddr == "" {
		m.err = t(m.lang, "err_empty_end")
		return m, nil
	}
	if passengersStr == "" {
		m.err = t(m.lang, "err_invalid_pass")
		return m, nil
	}

	passengers, err := strconv.Atoi(passengersStr)
	if err != nil || passengers < 1 || passengers > 8 {
		m.err = t(m.lang, "err_invalid_pass")
		return m, nil
	}

	m.loading = true
	m.err = ""

	groups := make([]calc.PassengerGroup, len(m.config.PassengerGroups))
	for i, g := range m.config.PassengerGroups {
		groups[i] = calc.PassengerGroup{
			Name:       g.Name,
			BoardFee:   g.BoardFee,
			PerKm:      g.PerKm,
			PerMinute:  g.PerMinute,
			WaitMinute: g.WaitMinute,
		}
	}

	start := startAddr
	end := endAddr
	p := passengers
	grps := groups

	return m, func() tea.Msg {
		route, err := routing.CalculateRoute(start, end)
		if err != nil {
			return routeErrMsg{err: err}
		}

		result := calc.Calculate(calc.FareInput{
			DistanceKm:  route.DistanceKm,
			DurationMin: route.DurationMin,
			Passengers:  p,
		}, grps)

		return routeMsg{
			result: &result,
			route:  route,
			start:  start,
			end:    end,
		}
	}
}

func (m Model) saveSetup() (tea.Model, tea.Cmd) {
	if len(m.inputs) < len(m.config.PassengerGroups)*4 {
		m.err = t(m.lang, "err_not_all_filled")
		return m, nil
	}

	for i := range m.config.PassengerGroups {
		boardFee, err := strconv.ParseFloat(m.inputs[i*4].Value(), 64)
		if err != nil {
			m.err = t(m.lang, "err_invalid_board") + fmt.Sprintf("%d", i+1)
			return m, nil
		}
		perKm, err := strconv.ParseFloat(m.inputs[i*4+1].Value(), 64)
		if err != nil {
			m.err = t(m.lang, "err_invalid_per_km") + fmt.Sprintf("%d", i+1)
			return m, nil
		}
		perMinute, err := strconv.ParseFloat(m.inputs[i*4+2].Value(), 64)
		if err != nil {
			m.err = t(m.lang, "err_invalid_per_min") + fmt.Sprintf("%d", i+1)
			return m, nil
		}
		waitMinute, err := strconv.ParseFloat(m.inputs[i*4+3].Value(), 64)
		if err != nil {
			m.err = t(m.lang, "err_invalid_wait_min") + fmt.Sprintf("%d", i+1)
			return m, nil
		}

		m.config.PassengerGroups[i].BoardFee = boardFee
		m.config.PassengerGroups[i].PerKm = perKm
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
	if len(m.inputs) < len(m.config.PassengerGroups)*4 {
		m.err = t(m.lang, "err_not_all_filled")
		return m, nil
	}

	for i := range m.config.PassengerGroups {
		boardFee, err := strconv.ParseFloat(m.inputs[i*4].Value(), 64)
		if err != nil {
			m.err = t(m.lang, "err_invalid_board") + fmt.Sprintf("%d", i+1)
			return m, nil
		}
		perKm, err := strconv.ParseFloat(m.inputs[i*4+1].Value(), 64)
		if err != nil {
			m.err = t(m.lang, "err_invalid_per_km") + fmt.Sprintf("%d", i+1)
			return m, nil
		}
		perMinute, err := strconv.ParseFloat(m.inputs[i*4+2].Value(), 64)
		if err != nil {
			m.err = t(m.lang, "err_invalid_per_min") + fmt.Sprintf("%d", i+1)
			return m, nil
		}
		waitMinute, err := strconv.ParseFloat(m.inputs[i*4+3].Value(), 64)
		if err != nil {
			m.err = t(m.lang, "err_invalid_wait_min") + fmt.Sprintf("%d", i+1)
			return m, nil
		}

		m.config.PassengerGroups[i].BoardFee = boardFee
		m.config.PassengerGroups[i].PerKm = perKm
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
	b.WriteString(titleStyle.Render(t(m.lang, "title")))
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

	if m.loading {
		b.WriteString("\n")
		b.WriteString(helpStyle.Render(t(m.lang, "loading")))
	}

	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(t(m.lang, "err_error") + m.err))
	}

	return borderStyle.Render(b.String())
}

func (m Model) viewMain() string {
	var b strings.Builder
	b.WriteString(subtitleStyle.Render(t(m.lang, "main_menu")))
	b.WriteString("\n\n")
	b.WriteString(keyStyle.Render("1") + t(m.lang, "main_calc") + "\n")
	b.WriteString(keyStyle.Render("2") + t(m.lang, "main_settings") + "\n")
	b.WriteString(keyStyle.Render("3") + t(m.lang, "main_help") + "\n")
	b.WriteString(keyStyle.Render("4") + t(m.lang, "main_setup") + "\n")
	b.WriteString(keyStyle.Render("q") + t(m.lang, "main_quit") + "\n")
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(t(m.lang, "main_select")))
	return b.String()
}

func (m Model) viewSetup() string {
	var b strings.Builder
	b.WriteString(subtitleStyle.Render(t(m.lang, "setup_title")))
	b.WriteString("\n\n")

	if m.setupStep == 0 {
		b.WriteString(titleStyle.Render(t(m.lang, "setup_lang_title")))
		b.WriteString("\n\n")
		b.WriteString(keyStyle.Render("1") + " " + t(m.lang, "setup_lang_en") + "\n")
		b.WriteString(keyStyle.Render("2") + " " + t(m.lang, "setup_lang_nl") + "\n")
		b.WriteString("\n")
		if m.lang == "en" {
			b.WriteString("  > " + t(m.lang, "setup_lang_en") + " <\n")
		} else {
			b.WriteString("  > " + t(m.lang, "setup_lang_nl") + " <\n")
		}
		b.WriteString("\n")
		b.WriteString(helpStyle.Render(t(m.lang, "setup_lang_help")))
		return b.String()
	}

	for i, g := range m.config.PassengerGroups {
		b.WriteString(titleStyle.Render(g.Name))
		b.WriteString("\n")
		idx := i * 4
		if idx < len(m.inputs) {
			b.WriteString(m.inputs[idx].View() + t(m.lang, "setup_board_fee") + "\n")
		}
		idx++
		if idx < len(m.inputs) {
			b.WriteString(m.inputs[idx].View() + t(m.lang, "setup_per_km") + "\n")
		}
		idx++
		if idx < len(m.inputs) {
			b.WriteString(m.inputs[idx].View() + t(m.lang, "setup_per_minute") + "\n")
		}
		idx++
		if idx < len(m.inputs) {
			b.WriteString(m.inputs[idx].View() + t(m.lang, "setup_wait_minute") + "\n")
		}
		b.WriteString("\n")
	}
	b.WriteString(helpStyle.Render(t(m.lang, "setup_help")))
	return b.String()
}

func (m Model) viewSettings() string {
	var b strings.Builder
	b.WriteString(subtitleStyle.Render(t(m.lang, "settings_title")))
	b.WriteString("\n\n")
	for i, g := range m.config.PassengerGroups {
		b.WriteString(titleStyle.Render(g.Name))
		b.WriteString("\n")
		idx := i * 4
		if idx < len(m.inputs) {
			b.WriteString(m.inputs[idx].View() + t(m.lang, "settings_board_fee") + "\n")
		}
		idx++
		if idx < len(m.inputs) {
			b.WriteString(m.inputs[idx].View() + t(m.lang, "settings_per_km") + "\n")
		}
		idx++
		if idx < len(m.inputs) {
			b.WriteString(m.inputs[idx].View() + t(m.lang, "settings_per_minute") + "\n")
		}
		idx++
		if idx < len(m.inputs) {
			b.WriteString(m.inputs[idx].View() + t(m.lang, "settings_wait_minute") + "\n")
		}
		b.WriteString("\n")
	}
	b.WriteString(helpStyle.Render(t(m.lang, "settings_help")))
	return b.String()
}

func (m Model) viewHelp() string {
	var b strings.Builder
	b.WriteString(subtitleStyle.Render(t(m.lang, "help_title")))
	b.WriteString("\n\n")
	b.WriteString(t(m.lang, "help_controls") + "\n")
	b.WriteString("  " + keyStyle.Render("1") + t(m.lang, "help_calc") + "\n")
	b.WriteString("  " + keyStyle.Render("2") + t(m.lang, "help_settings") + "\n")
	b.WriteString("  " + keyStyle.Render("3") + t(m.lang, "help_help") + "\n")
	b.WriteString("  " + keyStyle.Render("4") + t(m.lang, "help_setup") + "\n")
	b.WriteString("  " + keyStyle.Render("q") + t(m.lang, "help_quit") + "\n")
	b.WriteString("  " + keyStyle.Render("Tab") + t(m.lang, "help_tab") + "\n")
	b.WriteString("  " + keyStyle.Render("Enter") + t(m.lang, "help_enter") + "\n")
	b.WriteString("  " + keyStyle.Render("Esc") + t(m.lang, "help_esc") + "\n")
	b.WriteString("\n")
	b.WriteString(t(m.lang, "help_config_title") + "\n")
	b.WriteString(t(m.lang, "help_config_path") + "\n")
	b.WriteString("\n")
	b.WriteString(t(m.lang, "help_api_title") + "\n")
	b.WriteString(t(m.lang, "help_api_desc") + "\n")
	b.WriteString("\n")
	b.WriteString(t(m.lang, "help_pass_title") + "\n")
	b.WriteString(t(m.lang, "help_pass_1") + "\n")
	b.WriteString(t(m.lang, "help_pass_2") + "\n")
	b.WriteString("\n")
	b.WriteString(t(m.lang, "help_pricing_title") + "\n")
	b.WriteString(t(m.lang, "help_pricing_board") + "\n")
	b.WriteString(t(m.lang, "help_pricing_km") + "\n")
	b.WriteString(t(m.lang, "help_pricing_time") + "\n")
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(t(m.lang, "help_return")))
	return b.String()
}

func (m Model) viewCalc() string {
	var b strings.Builder
	b.WriteString(subtitleStyle.Render(t(m.lang, "calc_title")))
	b.WriteString("\n\n")
	if len(m.inputs) >= 3 {
		b.WriteString(m.inputs[0].View() + "\n")
		b.WriteString(helpStyle.Render("  "+t(m.lang, "calc_label_start")) + "\n\n")
		b.WriteString(m.inputs[1].View() + "\n")
		b.WriteString(helpStyle.Render("  "+t(m.lang, "calc_label_end")) + "\n\n")
		b.WriteString(m.inputs[2].View() + "\n")
		b.WriteString(helpStyle.Render("  "+t(m.lang, "calc_label_passengers")) + "\n")
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(t(m.lang, "calc_help")))
	return b.String()
}

func (m Model) viewResult() string {
	var b strings.Builder
	b.WriteString(subtitleStyle.Render(t(m.lang, "result_title")))
	b.WriteString("\n\n")

	b.WriteString(t(m.lang, "result_route") + "\n")
	b.WriteString("  " + m.startAddr + "\n")
	b.WriteString("  " + arrowStyle.Render(" ↓ ") + "\n")
	b.WriteString("  " + m.endAddr + "\n\n")

	if m.routeInfo != nil {
		b.WriteString(t(m.lang, "result_distance") + fmt.Sprintf("%.1f km", m.routeInfo.DistanceKm) + "\n")
		b.WriteString(t(m.lang, "result_duration") + fmt.Sprintf("%.0f min", m.routeInfo.DurationMin) + "\n\n")
	}

	if m.result != nil {
		b.WriteString(t(m.lang, "result_group") + m.result.Group + "\n\n")
		b.WriteString(t(m.lang, "result_board") + fmt.Sprintf("€%.2f", m.result.BaseFee) + "\n")
		b.WriteString(t(m.lang, "result_km") + fmt.Sprintf("€%.2f", m.result.KmFee) + "\n")
		b.WriteString(t(m.lang, "result_time") + fmt.Sprintf("€%.2f", m.result.TimeFee) + "\n")
		b.WriteString("───────────────────\n")
		b.WriteString(t(m.lang, "result_total") + successStyle.Render("€"+fmt.Sprintf("%.2f", m.result.Total)) + "\n")
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(t(m.lang, "result_help")))
	return b.String()
}
