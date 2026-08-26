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

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(黄色).
			Padding(0, 1)

	focusedInputStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("226")).
				Padding(0, 1).
				Bold(true)

	buttonStyle = lipgloss.NewStyle().
			Background(黄色).
			Foreground(黑色).
			Padding(0, 2).
			Bold(true)

	activeButtonStyle = lipgloss.NewStyle().
				Background(白色).
				Foreground(黑色).
				Padding(0, 2).
				Bold(true)

	resultStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(黄色).
			Padding(1, 2).
			MarginTop(1)

	keyStyle = lipgloss.NewStyle().
			Foreground(黄色).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(灰色).
			MarginTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(红色).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(绿色).
			Bold(true)

	logoStyle = lipgloss.NewStyle().
			Foreground(黄色).
			Bold(true).
			MarginRight(2)

	arrowStyle = lipgloss.NewStyle().
			Foreground(黄色).
			Bold(true)
)

func ApplyYellowBorder(s lipgloss.Style) lipgloss.Style {
	return s.Border(lipgloss.RoundedBorder()).BorderForeground(黄色)
}
