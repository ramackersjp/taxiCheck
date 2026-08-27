package tui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
	screenUpdate
	screenBranch
	screenUninstall
)

var appVersion = "dev"

func SetVersion(v string) {
	appVersion = v
}

func DetectVersion() {
	dir := gitRepoDir()
	args := []string{"symbolic-ref", "--short", "HEAD"}
	if dir != "" {
		args = append([]string{"-C", dir}, args...)
	}
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err == nil {
		branch := strings.TrimSpace(string(output))
		if branch == "v1.0.0" || branch == "dev" {
			appVersion = branch
		}
	}
}

type routeMsg struct {
	result *calc.FareResult
	route  *routing.RouteResult
	start  string
	end    string
}

type routeErrMsg struct {
	err error
}

type tickMsg time.Time

type suggestMsg struct {
	inputIdx int
	query    string
	suggests []routing.AddressSuggestion
}

type updateCheckMsg struct {
	latestTag string
	hasUpdate bool
	err       error
}

type updateResultMsg struct {
	success bool
	err     error
}

type branchResultMsg struct {
	current  string
	branches []string
	err      error
}

type branchSwitchMsg struct {
	success bool
	err     error
	newTag  string
}

type uninstallResultMsg struct {
	success bool
	err     error
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
	routeMode string
	width     int
	height    int

	suggestions    []routing.AddressSuggestion
	suggestionIdx  int
	showSuggest    bool
	suggestInput   int
	lastInputVal   string

	latestTag      string
	hasUpdate      bool
	updateChecked  bool
	updateStatus   string
	currentBranch  string
	branchList     []string
	branchIdx      int
	branchStatus   string
	uninstallStep  int
	uninstallStatus string
	settingsStep   int
	settingsLang   string
	setupDone      bool
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
			setupDone: false,
		}
	}

	lang := cfg.Language
	if lang == "" {
		lang = "en"
	}

	return Model{
		screen:    screenMain,
		config:    cfg,
		lang:      lang,
		routeMode: "fastest",
		setupDone: true,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) inputWidth() int {
	w := m.width - 14
	if w < 20 {
		w = 20
	}
	if w > 50 {
		w = 50
	}
	return w
}

func (m Model) priceInputWidth() int {
	w := m.width - 30
	if w < 10 {
		w = 10
	}
	if w > 15 {
		w = 15
	}
	return w
}

func gitRepoDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	dir := filepath.Dir(exe)
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func (m Model) contentWidth() int {
	w := m.width - 6
	if w < 30 {
		w = 30
	}
	return w
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		inputW := m.inputWidth()
		for i := range m.inputs {
			if i < 2 {
				m.inputs[i].Width = inputW
			}
		}
		return m, nil
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
	case tickMsg:
		return m.fetchSuggestions()
	case suggestMsg:
		if msg.inputIdx == m.suggestInput && m.lastInputVal == msg.query {
			m.suggestions = msg.suggests
			m.suggestionIdx = 0
			m.showSuggest = len(msg.suggests) > 0
		}
		return m, nil
	case updateCheckMsg:
		m.loading = false
		m.updateChecked = true
		if msg.err != nil {
			m.err = msg.err.Error()
		} else {
			m.latestTag = msg.latestTag
			m.hasUpdate = msg.hasUpdate
			if msg.hasUpdate {
				m.updateStatus = fmt.Sprintf("%s -> %s", appVersion, msg.latestTag)
			} else {
				m.updateStatus = t(m.lang, "update_up_to_date")
			}
		}
		return m, nil
	case updateResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
		} else {
			m.updateStatus = t(m.lang, "update_success")
			m.hasUpdate = false
		}
		return m, nil
	case branchResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
		} else {
			m.currentBranch = msg.current
			if msg.current == "v1.0.0" || msg.current == "dev" {
				appVersion = msg.current
			}
			m.branchList = msg.branches
			m.branchIdx = 0
			for i, b := range msg.branches {
				if b == msg.current {
					m.branchIdx = i
					break
				}
			}
		}
		return m, nil
	case branchSwitchMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
		} else {
			appVersion = msg.newTag
			m.branchStatus = t(m.lang, "branch_switch_success")
			m.screen = screenMain
		}
		return m, nil
	case uninstallResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
		} else {
			m.uninstallStatus = t(m.lang, "uninstall_success")
		}
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
	case screenUpdate:
		return m.updateUpdate(msg)
	case screenBranch:
		return m.updateBranch(msg)
	case screenUninstall:
		return m.updateUninstall(msg)
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
		inputW := m.inputWidth()
		for i := range m.inputs {
			m.inputs[i] = textinput.New()
			m.inputs[i].CharLimit = 80
			m.inputs[i].Width = inputW
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
		m.settingsStep = 0
		m.settingsLang = m.lang
		m.inputs = nil
		m.err = ""
		return m, nil
	case "3":
		m.screen = screenHelp
	case "4":
		if !m.setupDone {
			m.screen = screenSetup
			m.setupStep = 0
			m.inputs = nil
			return m, nil
		}
		return m, nil
	case "5":
		m.screen = screenUpdate
		m.updateStatus = ""
		m.err = ""
		if !m.updateChecked {
			m.loading = true
			return m, m.checkUpdate()
		}
		return m, nil
	case "6":
		m.screen = screenBranch
		m.branchStatus = ""
		m.err = ""
		m.loading = true
		return m, m.fetchBranches()
	case "7":
		m.screen = screenUninstall
		m.uninstallStep = 0
		m.uninstallStatus = ""
		m.err = ""
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
	case "esc":
		if m.settingsStep == 0 {
			m.screen = screenMain
			m.err = ""
			return m, nil
		}
		m.settingsStep = 0
		m.inputs = nil
		m.err = ""
		return m, nil
	}

	if m.settingsStep == 0 {
		return m.updateSettingsLang(msg)
	}
	return m.updateSettingsPricing(msg)
}

func (m Model) updateSettingsLang(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "1", "up", "down":
		if msg.String() == "1" || msg.String() == "up" {
			m.settingsLang = "en"
		} else {
			m.settingsLang = "nl"
		}
		return m, nil
	case "2":
		m.settingsLang = "nl"
		return m, nil
	case "enter":
		m.lang = m.settingsLang
		m.config.Language = m.lang
		m.settingsStep = 1
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
		if len(m.inputs) > 0 {
			m.inputs[0].Focus()
		}
		return m, textinput.Blink
	}
	return m, nil
}

func (m Model) updateSettingsPricing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		return m.saveSettings()
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

func (m Model) updateUpdate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "enter":
		m.screen = screenMain
		m.err = ""
		return m, nil
	case "u":
		if m.hasUpdate {
			m.loading = true
			m.err = ""
			return m, m.pullUpdate()
		}
		return m, nil
	case "r":
		m.loading = true
		m.err = ""
		m.updateChecked = false
		return m, m.checkUpdate()
	}
	return m, nil
}

func (m Model) updateBranch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "enter":
		m.screen = screenMain
		m.err = ""
		return m, nil
	case "up", "k":
		if len(m.branchList) > 0 {
			m.branchIdx = (m.branchIdx - 1 + len(m.branchList)) % len(m.branchList)
		}
		return m, nil
	case "down", "j":
		if len(m.branchList) > 0 {
			m.branchIdx = (m.branchIdx + 1) % len(m.branchList)
		}
		return m, nil
	case " ":
		if len(m.branchList) > 0 {
			target := m.branchList[m.branchIdx]
			if target != m.currentBranch {
				m.loading = true
				m.err = ""
				return m, m.switchBranch(target)
			}
		}
		return m, nil
	}
	return m, nil
}

func (m Model) updateUninstall(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.screen = screenMain
		m.err = ""
		return m, nil
	case "n", "N":
		m.screen = screenMain
		m.err = ""
		return m, nil
	case "y", "Y":
		if m.uninstallStep == 0 {
			m.uninstallStep = 1
			return m, nil
		}
		if m.uninstallStep == 1 {
			m.loading = true
			m.err = ""
			return m, m.runUninstall()
		}
		return m, nil
	}
	return m, nil
}

func (m Model) fetchSuggestions() (tea.Model, tea.Cmd) {
	if m.focusIdx > 1 || len(m.inputs) < 2 {
		return m, nil
	}
	query := m.inputs[m.focusIdx].Value()
	inputIdx := m.focusIdx
	q := query

	return m, func() tea.Msg {
		suggests, _ := routing.SuggestAddresses(q)
		return suggestMsg{inputIdx: inputIdx, query: q, suggests: suggests}
	}
}

func (m Model) updateCalc(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "f2":
		if m.routeMode == "fastest" {
			m.routeMode = "shortest"
		} else {
			m.routeMode = "fastest"
		}
		return m, nil
	case "down", "up":
		if m.showSuggest {
			if msg.String() == "down" {
				m.suggestionIdx = (m.suggestionIdx + 1) % len(m.suggestions)
			} else {
				m.suggestionIdx = (m.suggestionIdx - 1 + len(m.suggestions)) % len(m.suggestions)
			}
			return m, nil
		}
		if len(m.inputs) == 0 {
			return m, nil
		}
		m.inputs[m.focusIdx].Blur()
		if msg.String() == "down" {
			m.focusIdx = (m.focusIdx + 1) % len(m.inputs)
		} else {
			m.focusIdx = (m.focusIdx - 1 + len(m.inputs)) % len(m.inputs)
		}
		m.inputs[m.focusIdx].Focus()
		m.showSuggest = false
		return m, textinput.Blink
	case "enter":
		if m.showSuggest && m.suggestionIdx < len(m.suggestions) {
			sel := m.suggestions[m.suggestionIdx]
			m.inputs[m.focusIdx].SetValue(sel.Display)
			m.showSuggest = false
			m.suggestions = nil
			return m, nil
		}
		return m.startCalculation()
	case "esc":
		if m.showSuggest {
			m.showSuggest = false
			m.suggestions = nil
			return m, nil
		}
		m.screen = screenMain
		m.err = ""
		return m, nil
	case "tab", "shift+tab":
		m.showSuggest = false
		m.suggestions = nil
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
	}

	for i := range m.inputs {
		if i == m.focusIdx {
			oldVal := m.inputs[i].Value()
			m.inputs[i], _ = m.inputs[i].Update(msg)
			newVal := m.inputs[i].Value()
			if i < 2 && newVal != oldVal && len(newVal) >= 2 {
				m.suggestInput = i
				m.lastInputVal = newVal
				return m, tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
					return tickMsg(t)
				})
			}
			if i < 2 && newVal != oldVal {
				m.showSuggest = false
				m.suggestions = nil
			}
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
		inputW := m.inputWidth()
		for i := range m.inputs {
			m.inputs[i] = textinput.New()
			m.inputs[i].CharLimit = 80
			m.inputs[i].Width = inputW
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

func (m Model) checkUpdate() tea.Cmd {
	return func() tea.Msg {
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get("https://api.github.com/repos/jp/taxiprijs/releases/latest")
		if err != nil {
			return updateCheckMsg{err: err}
		}
		defer resp.Body.Close()

		if resp.StatusCode == 404 {
			return updateCheckMsg{hasUpdate: false}
		}

		var release struct {
			TagName string `json:"tag_name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return updateCheckMsg{err: err}
		}

		hasUpdate := release.TagName != appVersion
		return updateCheckMsg{
			latestTag: release.TagName,
			hasUpdate: hasUpdate,
		}
	}
}

func (m Model) pullUpdate() tea.Cmd {
	return func() tea.Msg {
		dir := gitRepoDir()
		args := []string{"pull"}
		if dir != "" {
			args = append([]string{"-C", dir}, args...)
		}
		cmd := exec.Command("git", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return updateResultMsg{err: fmt.Errorf("%s: %s", err, string(output))}
		}
		return updateResultMsg{success: true}
	}
}

func (m Model) fetchBranches() tea.Cmd {
	return func() tea.Msg {
		dir := gitRepoDir()
		args := []string{"symbolic-ref", "--short", "HEAD"}
		if dir != "" {
			args = append([]string{"-C", dir}, args...)
		}
		cmd := exec.Command("git", args...)
		output, err := cmd.Output()
		current := "HEAD"
		if err == nil {
			current = strings.TrimSpace(string(output))
		}

		branches := []string{"dev", "v1.0.0"}

		return branchResultMsg{
			current:  current,
			branches: branches,
		}
	}
}

func (m Model) switchBranch(branch string) tea.Cmd {
	return func() tea.Msg {
		dir := gitRepoDir()
		args := []string{"checkout", branch}
		if dir != "" {
			args = append([]string{"-C", dir}, args...)
		}
		cmd := exec.Command("git", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return branchSwitchMsg{err: fmt.Errorf("%s: %s", err, string(output))}
		}
		return branchSwitchMsg{success: true, newTag: branch}
	}
}

func (m Model) runUninstall() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("make", "uninstall")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return uninstallResultMsg{err: fmt.Errorf("%s: %s", err, string(output))}
		}
		return uninstallResultMsg{success: true}
	}
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
	mode := m.routeMode

	return m, func() tea.Msg {
		route, err := routing.CalculateRoute(start, end, mode)
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
	m.setupDone = true
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
	case screenUpdate:
		b.WriteString(m.viewUpdate())
	case screenBranch:
		b.WriteString(m.viewBranch())
	case screenUninstall:
		b.WriteString(m.viewUninstall())
	}

	if m.loading {
		b.WriteString("\n")
		b.WriteString(helpStyle.Render(t(m.lang, "loading")))
	}

	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(t(m.lang, "err_error") + m.err))
	}

	content := borderStyle.
		Width(m.contentWidth()).
		Render(b.String())

	var full strings.Builder
	full.WriteString("\n")
	full.WriteString(GetLogoCentered(m.width))
	full.WriteString("\n\n")
	full.WriteString(content)

	if m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, full.String())
	}
	return full.String()
}

func (m Model) viewMain() string {
	var b strings.Builder
	b.WriteString(subtitleStyle.Render(t(m.lang, "main_menu")))
	b.WriteString("\n\n")
	b.WriteString(keyStyle.Render("1") + t(m.lang, "main_calc") + "\n")
	b.WriteString(keyStyle.Render("2") + t(m.lang, "main_settings") + "\n")
	b.WriteString(keyStyle.Render("3") + t(m.lang, "main_help") + "\n")
	if !m.setupDone {
		b.WriteString(keyStyle.Render("4") + t(m.lang, "main_setup") + "\n")
	}
	b.WriteString(keyStyle.Render("5") + t(m.lang, "main_update") + "\n")
	b.WriteString(keyStyle.Render("6") + t(m.lang, "main_branch") + "\n")
	b.WriteString(keyStyle.Render("7") + t(m.lang, "main_uninstall") + "\n")
	b.WriteString(keyStyle.Render("q") + t(m.lang, "main_quit") + "\n")

	if m.branchStatus != "" {
		b.WriteString("\n")
		b.WriteString(successStyle.Render(m.branchStatus))
		m.branchStatus = ""
	}

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

	if m.settingsStep == 0 {
		b.WriteString(titleStyle.Render(t(m.lang, "settings_lang_title")))
		b.WriteString("\n\n")
		b.WriteString(keyStyle.Render("1") + " " + t(m.lang, "settings_lang_en") + "\n")
		b.WriteString(keyStyle.Render("2") + " " + t(m.lang, "settings_lang_nl") + "\n")
		b.WriteString("\n")
		if m.settingsLang == "en" {
			b.WriteString("  > " + t(m.lang, "settings_lang_en") + " <\n")
		} else {
			b.WriteString("  > " + t(m.lang, "settings_lang_nl") + " <\n")
		}
		b.WriteString("\n")
		b.WriteString(helpStyle.Render(t(m.lang, "settings_lang_help")))
		return b.String()
	}

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
	if !m.setupDone {
		b.WriteString("  " + keyStyle.Render("4") + t(m.lang, "help_setup") + "\n")
	}
	b.WriteString("  " + keyStyle.Render("5") + t(m.lang, "help_update") + "\n")
	b.WriteString("  " + keyStyle.Render("6") + t(m.lang, "help_branch") + "\n")
	b.WriteString("  " + keyStyle.Render("7") + t(m.lang, "help_uninstall") + "\n")
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
		b.WriteString(helpStyle.Render("  "+t(m.lang, "calc_label_start")) + "\n")
		if m.showSuggest && m.suggestInput == 0 {
			b.WriteString(m.renderSuggestions())
		}
		b.WriteString("\n")
		b.WriteString(m.inputs[1].View() + "\n")
		b.WriteString(helpStyle.Render("  "+t(m.lang, "calc_label_end")) + "\n")
		if m.showSuggest && m.suggestInput == 1 {
			b.WriteString(m.renderSuggestions())
		}
		b.WriteString("\n")
		b.WriteString(m.inputs[2].View() + "\n")
		b.WriteString(helpStyle.Render("  "+t(m.lang, "calc_label_passengers")) + "\n")
	}
	b.WriteString("\n")
	if m.routeMode == "fastest" {
		b.WriteString(keyStyle.Render("F2") + " " + t(m.lang, "calc_mode") + ": " + successStyle.Render(t(m.lang, "calc_mode_fastest")) + "\n")
	} else {
		b.WriteString(keyStyle.Render("F2") + " " + t(m.lang, "calc_mode") + ": " + successStyle.Render(t(m.lang, "calc_mode_shortest")) + "\n")
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(t(m.lang, "calc_help")))
	return b.String()
}

func (m Model) renderSuggestions() string {
	maxLen := m.contentWidth() - 8
	if maxLen < 20 {
		maxLen = 20
	}
	var b strings.Builder
	for i, s := range m.suggestions {
		display := s.Display
		if len(display) > maxLen {
			display = display[:maxLen-3] + "..."
		}
		if i == m.suggestionIdx {
			b.WriteString("  " + successStyle.Render("▸ "+display) + "\n")
		} else {
			b.WriteString("  " + helpStyle.Render("  "+display) + "\n")
		}
	}
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

	if m.routeMode == "shortest" {
		b.WriteString(t(m.lang, "result_mode") + t(m.lang, "calc_mode_shortest") + "\n")
	} else {
		b.WriteString(t(m.lang, "result_mode") + t(m.lang, "calc_mode_fastest") + "\n")
	}

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

func (m Model) viewUpdate() string {
	var b strings.Builder
	b.WriteString(subtitleStyle.Render(t(m.lang, "update_title")))
	b.WriteString("\n\n")

	ver := appVersion
	if !strings.HasPrefix(ver, "v") {
		ver = "v" + ver
	}
	b.WriteString(t(m.lang, "update_current") + successStyle.Render(ver) + "\n")
	b.WriteString("\n")

	if m.updateStatus != "" {
		if m.hasUpdate {
			b.WriteString(t(m.lang, "update_available") + successStyle.Render(m.updateStatus) + "\n")
			b.WriteString("\n")
			b.WriteString(keyStyle.Render("u") + " " + t(m.lang, "update_pull") + "\n")
			b.WriteString(keyStyle.Render("r") + " " + t(m.lang, "update_recheck") + "\n")
		} else {
			b.WriteString(successStyle.Render(m.updateStatus) + "\n")
			b.WriteString("\n")
			b.WriteString(keyStyle.Render("r") + " " + t(m.lang, "update_recheck") + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render(t(m.lang, "update_help")))
	return b.String()
}

func (m Model) viewBranch() string {
	var b strings.Builder
	b.WriteString(subtitleStyle.Render(t(m.lang, "branch_title")))
	b.WriteString("\n\n")

	b.WriteString(t(m.lang, "branch_current") + successStyle.Render(m.currentBranch) + "\n\n")

	if len(m.branchList) > 0 {
		b.WriteString(t(m.lang, "branch_list") + "\n\n")
		for i, branch := range m.branchList {
			prefix := "  "
			if branch == m.currentBranch {
				prefix = "  " + successStyle.Render("* ")
			}
			if i == m.branchIdx && branch != m.currentBranch {
				b.WriteString(prefix + successStyle.Render("▸ "+branch) + "\n")
			} else {
				b.WriteString(prefix + helpStyle.Render(branch) + "\n")
			}
		}
		b.WriteString("\n")
		b.WriteString(helpStyle.Render(t(m.lang, "branch_help")))
	} else {
		b.WriteString(t(m.lang, "branch_none") + "\n")
		b.WriteString("\n")
		b.WriteString(helpStyle.Render(t(m.lang, "branch_help")))
	}

	return b.String()
}

func (m Model) viewUninstall() string {
	var b strings.Builder
	b.WriteString(subtitleStyle.Render(t(m.lang, "uninstall_title")))
	b.WriteString("\n\n")

	if m.uninstallStatus != "" {
		b.WriteString(successStyle.Render(m.uninstallStatus) + "\n")
	} else if m.uninstallStep == 0 {
		b.WriteString(t(m.lang, "uninstall_confirm") + "\n\n")
		b.WriteString(keyStyle.Render("y") + " " + t(m.lang, "uninstall_yes") + "\n")
		b.WriteString(keyStyle.Render("n") + " " + t(m.lang, "uninstall_no") + "\n")
	} else {
		b.WriteString(errorStyle.Render(t(m.lang, "uninstall_final")) + "\n\n")
		b.WriteString(keyStyle.Render("y") + " " + t(m.lang, "uninstall_final_yes") + "\n")
		b.WriteString(keyStyle.Render("n") + " " + t(m.lang, "uninstall_no") + "\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render(t(m.lang, "uninstall_help")))
	return b.String()
}
