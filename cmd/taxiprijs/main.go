package main

import (
	"fmt"
	"os"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jp/taxiprijs/internal/routing"
	"github.com/jp/taxiprijs/internal/tui"
)

var version = "dev"

var currentVersion atomic.Value

func init() {
	currentVersion.Store(version)
}

func main() {
	routing.LoadEnv()

	tui.SetVersion(currentVersion.Load().(string))

	p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
