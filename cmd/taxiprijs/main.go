package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ramackersjp/taxiCheck/internal/routing"
	"github.com/ramackersjp/taxiCheck/internal/tui"
)

func main() {
	routing.LoadEnv()

	tui.DetectVersion()

	p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
