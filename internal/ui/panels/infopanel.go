package panels

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	infoPanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#333333")).
		Padding(0, 1)

	labelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	valStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Bold(true)
	accentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FA243C"))
)

func ViewInfoPanel(width, height int, device string, vol int) string {
	if width <= 4 || height <= 2 {
		return ""
	}

	// 1. Output Device
	if device == "" {
		device = "Unknown"
	}
	devLabel := "Output: "
	maxDevLen := width - 4 - len(devLabel)
	if maxDevLen < 0 {
		maxDevLen = 0
	}
	devStr := device
	if len([]rune(devStr)) > maxDevLen {
		if maxDevLen > 1 {
			devStr = string([]rune(devStr)[:maxDevLen-1]) + "…"
		} else {
			devStr = ""
		}
	}
	line1 := labelStyle.Render(devLabel) + valStyle.Render(devStr)

	// 2. Volume Bar
	volLabel := "Vol: "
	volSuffix := fmt.Sprintf(" %d%%", vol)
	volW := width - 4 - len(volLabel) - len(volSuffix) - 2 // 2 for brackets []
	if volW < 1 {
		volW = 1
	}

	filled := int(float64(vol) / 100.0 * float64(volW))
	if filled > volW {
		filled = volW
	}
	if filled < 0 {
		filled = 0
	}
	empty := volW - filled

	bar := accentStyle.Render(strings.Repeat("|", filled)) +
		labelStyle.Render(strings.Repeat("-", empty))

	line2 := labelStyle.Render(volLabel) + "[" + bar + "]" + valStyle.Render(volSuffix)

	// 3. Format (Fixed Lossless indicator to match reference)
	line3 := labelStyle.Render("Format: ") + accentStyle.Render("ALAC / Lossless")

	content := lipgloss.JoinVertical(lipgloss.Left, line1, "", line2, line3)

	return infoPanelStyle.Width(width - 2).Height(height - 2).MaxWidth(width - 2).MaxHeight(height - 2).Render(content)
}
