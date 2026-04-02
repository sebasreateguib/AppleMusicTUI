package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"SR-Player/internal/ui"
)

func main() {
	// Use alternate screen and mouse support
	p := tea.NewProgram(
		ui.NewApp(0, 0),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running SR-Player: %v\n", err)
		os.Exit(1)
	}
}
