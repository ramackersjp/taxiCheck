package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ramackersjp/taxiCheck/internal/routing"
	"github.com/ramackersjp/taxiCheck/internal/tui"
)

// Set at build time by the Makefile: -X main.version=$(VERSION)
var version = "dev"

func main() {
	routing.LoadEnv()

	if version != "" && version != "dev" {
		v := version
		if !strings.HasPrefix(v, "v") {
			v = "v" + v
		}
		tui.SetVersion(v)
	}
	tui.DetectVersion()

	p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
