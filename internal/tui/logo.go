package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var nowFn = time.Now

var logoBorderStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(黄色).
	Padding(0, 1)

// Same inner width as the previous side-view logo box (do not grow/shrink it).
const logoInnerWidth = 17

// taxiArtColored is a front-facing taxi (hood toward the viewer) with a
// roof sign. Four lines, same footprint as the previous side-view art.
func taxiArtColored() string {
	body := lipgloss.NewStyle().Foreground(黄色).Bold(true)
	sign := lipgloss.NewStyle().Foreground(黑色).Background(黄色).Bold(true)
	lamp := lipgloss.NewStyle().Foreground(黄色).Bold(true)
	light := lipgloss.NewStyle().Foreground(白色).Bold(true)
	wheel := lipgloss.NewStyle().Foreground(灰色).Bold(true)
	l1 := "     " + lamp.Render("▄▄██▄▄")
	l2 := "   " + body.Render("┌┤") + sign.Render("TAXI") + body.Render("├┐")
	l3 := "   " + body.Render("│ ") + light.Render("●") + body.Render("  ") + light.Render("●") + body.Render(" │")
	l4 := "   " + body.Render("└") + wheel.Render("○") + body.Render("────") + wheel.Render("○") + body.Render("┘")
	return strings.Join([]string{l1, l2, l3, l4}, "\n")
}

func GetLogo() string {
	var b strings.Builder
	b.WriteString(taxiArtColored())
	b.WriteString("\n")
	// Inline styles without MarginBottom/MarginTop so the version stays
	// on the same line as the title (titleStyle + helpStyle split it).
	title := lipgloss.NewStyle().Bold(true).Foreground(黄色).Render("TaxiCheck")
	ver := appVersion
	if !strings.HasPrefix(ver, "v") {
		ver = "v" + ver
	}
	b.WriteString(title + " " + lipgloss.NewStyle().Foreground(灰色).Faint(true).Render(ver))
	return logoBorderStyle.Width(logoInnerWidth).Render(b.String())
}

func GetLogoCentered(width int) string {
	logo := GetLogo()
	return lipgloss.PlaceHorizontal(width, lipgloss.Left, logo)
}

func (m Model) frameWidth() int {
	// Content panel is borderStyle.Width(contentWidth()); rounded borders
	// add 2 columns. The header row must match that outer width.
	return m.contentWidth() + 2
}

func (m Model) viewHeader() string {
	logo := GetLogo()
	lw := lipgloss.Width(logo)
	lh := lipgloss.Height(logo)
	gap := 1
	total := m.frameWidth()
	infoW := total - lw - gap
	if infoW < 16 {
		infoW = 16
	}
	info := m.viewInfoPanel(infoW, lh)
	row := lipgloss.JoinHorizontal(lipgloss.Top, logo, strings.Repeat(" ", gap), info)
	if w := lipgloss.Width(row); w < total {
		return lipgloss.PlaceHorizontal(total, lipgloss.Left, row)
	}
	return row
}

func (m Model) viewInfoPanel(totalW, totalH int) string {
	innerW := totalW - 2
	if innerW < 12 {
		innerW = 12
	}
	innerH := totalH - 2
	if innerH < 5 {
		innerH = 5
	}
	t0 := m.clockTime()
	val := lipgloss.NewStyle().Foreground(白色)
	lines := []string{
		fieldLabelOn.Render(t(m.lang, "info_date")) + ": " + val.Render(m.formatDate(t0)),
		fieldLabelOn.Render(t(m.lang, "info_time")) + ": " + val.Render(t0.Format("15:04:05")),
		fieldLabelOn.Render(t(m.lang, "info_source")) + ": " + helpStyle.Render(t(m.lang, "info_source_val")),
		fieldLabelOn.Render(t(m.lang, "info_license")) + ": " + helpStyle.Render(t(m.lang, "info_license_val")),
		helpStyle.Render(t(m.lang, "info_copyright")),
	}
	for len(lines) < innerH {
		lines = append(lines, "")
	}
	if len(lines) > innerH {
		lines = lines[:innerH]
	}
	inner := strings.Join(lines, "\n")
	return logoBorderStyle.Width(innerW).Render(inner)
}

func (m Model) clockTime() time.Time {
	if !m.clock.IsZero() {
		return m.clock
	}
	return nowFn()
}

func (m Model) formatDate(t0 time.Time) string {
	if m.lang == "nl" {
		days := []string{"zondag", "maandag", "dinsdag", "woensdag", "donderdag", "vrijdag", "zaterdag"}
		months := []string{"januari", "februari", "maart", "april", "mei", "juni", "juli", "augustus", "september", "oktober", "november", "december"}
		return fmt.Sprintf("%s %d %s %d", days[t0.Weekday()], t0.Day(), months[t0.Month()-1], t0.Year())
	}
	return t0.Format("Monday, 2 January 2006")
}

type clockMsg time.Time

func tickClock() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return clockMsg(t)
	})
}
