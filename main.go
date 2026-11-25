package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sebasreateguib/applemusictui/internal/artwork"
	"github.com/sebasreateguib/applemusictui/internal/config"
	"github.com/sebasreateguib/applemusictui/internal/ui"
)

// version is set by the release build.
var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println("applemusictui", version)
		return
	}

	if err := ensureMusic(); err != nil {
		if errors.Is(err, context.Canceled) {
			os.Exit(130) // Ctrl+C
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Use alternate screen and mouse support (clicks and drags)
	app := ui.NewApp(0, 0, config.Load(), artwork.Supported())
	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		// Twice the visualizer rate: at 60 the renderer's clock would drift
		// against cava's 60 fps and drop frames.
		tea.WithFPS(120),
	)

	_, err := p.Run()
	app.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running applemusictui: %v\n", err)
		os.Exit(1)
	}
}
