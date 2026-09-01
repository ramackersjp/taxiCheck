package tui

import "github.com/charmbracelet/lipgloss"

var (
	黄色 = lipgloss.Color("226")
	黑色 = lipgloss.Color("0")
	白色 = lipgloss.Color("15")
	灰色 = lipgloss.Color("245")
	红色 = lipgloss.Color("196")
	绿色 = lipgloss.Color("46")

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(黄色).
			Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(黄色).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(白色).
			MarginBottom(1)

	keyStyle = lipgloss.NewStyle().
			Foreground(黄色).
			Bold(true)

	// No MarginTop: helpStyle is concatenated into list rows (branch
	// names, suggestions). A top margin put "* " and "dev" on two lines.
	helpStyle = lipgloss.NewStyle().
			Foreground(灰色)

	errorStyle = lipgloss.NewStyle().
			Foreground(红色).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(绿色).
			Bold(true)

	arrowStyle = lipgloss.NewStyle().
			Foreground(黄色).
			Bold(true)

	closeStyle = lipgloss.NewStyle().
			Foreground(黑色).
			Background(黄色).
			Bold(true).
			Padding(0, 1)

	fieldLabelOn = lipgloss.NewStyle().
			Foreground(黄色).
			Bold(true)

	fieldBoxOn = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(黄色).
			Padding(0, 1)

	fieldBoxOff = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(灰色).
			Padding(0, 1)
)
