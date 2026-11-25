package panels

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderVisualizer draws one column per bar level (0-255), growing from the
// bottom, with a blank column between bars.
func renderVisualizer(levels []byte, width int, height int) string {
	if height <= 0 || width <= 0 {
		return ""
	}
	blocks := []rune(" ▁▂▃▄▅▆▇█")
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#FA243C"))
	bars := (width + 1) / 2

	rows := make([]string, height)
	for y := range rows {
		row := []rune(strings.Repeat(" ", width))
		base := (height - 1 - y) * 8
		for i := 0; i < bars && len(levels) > 0; i++ {
			// Spread levels over the columns: the counts only differ while
			// cava restarts after a resize, or on extremely wide terminals.
			h := int(levels[i*len(levels)/bars]) * height * 8 / 255
			row[i*2] = blocks[max(0, min(8, h-base))]
		}
		rows[y] = style.Render(string(row))
	}
	return strings.Join(rows, "\n")
}

// visualizerInnerSize returns the drawing area of a visualizer panel.
func visualizerInnerSize(width, height int) (int, int) {
	return width - 4, height - 2
}

// VisualizerBars returns how many bars fit in a visualizer panel of that width.
func VisualizerBars(width int) int {
	inner, _ := visualizerInnerSize(width, 0)
	return (inner + 1) / 2
}

// RenderVisualizerPanel draws the bars, or notice in their place when non-empty.
func RenderVisualizerPanel(width, height int, levels []byte, notice string) string {
	frame := Frame{Title: "Visualizer"}
	innerWidth, innerHeight := visualizerInnerSize(width, height)
	if innerWidth < 1 || innerHeight < 1 || notice != "" {
		return frame.Render(width, height, message(notice, width-2, innerHeight))
	}
	lines := strings.Split(renderVisualizer(levels, innerWidth, innerHeight), "\n")
	for i, l := range lines {
		lines[i] = " " + l
	}
	return frame.Render(width, height, lines)
}
