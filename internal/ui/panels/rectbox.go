package panels

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderRectBox(content string, width, height int, borderColor lipgloss.Color, alignV lipgloss.Position) string {
	panelWidth := width
	panelHeight := height
	if panelWidth < 6 || panelHeight < 4 {
		return ""
	}

	innerWidth := panelWidth - 2
	innerHeight := panelHeight - 2

	placed := lipgloss.Place(innerWidth, innerHeight, lipgloss.Left, alignV, content)
	lines := strings.Split(placed, "\n")
	for len(lines) < innerHeight {
		lines = append(lines, "")
	}

	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	horiz := strings.Repeat("─", innerWidth)

	out := make([]string, 0, innerHeight+2)
	out = append(out, borderStyle.Render("┌"+horiz+"┐"))
	for i := 0; i < innerHeight; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		line = lipgloss.NewStyle().Width(innerWidth).MaxWidth(innerWidth).Render(line)
		out = append(out, borderStyle.Render("│")+line+borderStyle.Render("│"))
	}
	out = append(out, borderStyle.Render("└"+horiz+"┘"))
	return strings.Join(out, "\n")
}

func renderRectBoxAligned(content string, width, height int, borderColor lipgloss.Color, alignH, alignV lipgloss.Position) string {
	panelWidth := width
	panelHeight := height
	if panelWidth < 6 || panelHeight < 4 {
		return ""
	}

	innerWidth := panelWidth - 2
	innerHeight := panelHeight - 2

	placed := lipgloss.Place(innerWidth, innerHeight, alignH, alignV, content)
	lines := strings.Split(placed, "\n")
	for len(lines) < innerHeight {
		lines = append(lines, "")
	}

	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	horiz := strings.Repeat("─", innerWidth)

	out := make([]string, 0, innerHeight+2)
	out = append(out, borderStyle.Render("┌"+horiz+"┐"))
	for i := 0; i < innerHeight; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		line = lipgloss.NewStyle().Width(innerWidth).MaxWidth(innerWidth).Render(line)
		out = append(out, borderStyle.Render("│")+line+borderStyle.Render("│"))
	}
	out = append(out, borderStyle.Render("└"+horiz+"┘"))
	return strings.Join(out, "\n")
}
