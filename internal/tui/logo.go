package tui

import "github.com/charmbracelet/lipgloss"

var taxiASCII = `      ______
     /|_||_\` + "`" + `.__
    (   _    _ _\
    =` + "`" + `-(_)--(_)-'`

func GetLogo() string {
	return logoStyle.Render(taxiASCII)
}

func GetLogoCentered(width int) string {
	logo := logoStyle.Render(taxiASCII)
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, logo)
}
