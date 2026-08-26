package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const taxiArt = `    ______
   /|_||_\` + "`" + `.__
  (   _    _ _\
  =` + "`" + `-(_)--(_)-'`

var logoBorderStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(黄色).
	Padding(0, 1)

func GetLogo() string {
	var b strings.Builder
	b.WriteString(taxiArt)
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("TaxiCheck"))
	b.WriteString(" " + helpStyle.Render("v"+appVersion))
	return logoBorderStyle.Render(b.String())
}

func GetLogoCentered(width int) string {
	logo := GetLogo()
	return lipgloss.PlaceHorizontal(width, lipgloss.Left, logo)
}
