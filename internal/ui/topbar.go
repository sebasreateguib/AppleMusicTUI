package ui

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderVisualizerBar(tick int, width int, height int, isPlaying bool, bpm int) string {
	if height <= 0 || width <= 0 {
		return ""
	}
	blocks := []string{" ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

	effectiveBPM := bpm
	if effectiveBPM <= 0 {
		effectiveBPM = 120
	}

	tSec := float64(tick) * 0.10
	bps := float64(effectiveBPM) / 60.0
	pulse := math.Pow(math.Abs(math.Sin(math.Pi*bps*tSec)), 4)

	numCols := width / 2
	if numCols < 1 {
		numCols = 1
	}

	var out strings.Builder
	for y := 0; y < height; y++ {
		for x := 0; x < numCols; x++ {
			if x > 0 {
				out.WriteString(" ")
			}

			h := 0
			if isPlaying {
				v1 := math.Sin(float64(x)*0.8 + tSec*bps*2.0)
				v2 := math.Cos(float64(x)*1.3 - tSec*bps*1.5)
				baseWave := (v1 + v2 + 2.0) / 4.0
				noise := float64((tick*x)%5) * 0.05
				bandMultiplier := 0.5 + 0.5*math.Sin(float64(x))
				val := baseWave*0.4 + (pulse*bandMultiplier)*0.6 + noise
				if val > 1.0 {
					val = 1.0
				}
				if val < 0 {
					val = 0
				}
				h = int(val * float64(height*8))
			}

			baseVal := (height - 1 - y) * 8
			cellH := h - baseVal
			if cellH < 0 {
				cellH = 0
			}
			if cellH > 7 {
				cellH = 7
			}

			if cellH == 0 {
				out.WriteString(" ")
			} else {
				color := lipgloss.Color("#FA243C")
				out.WriteString(lipgloss.NewStyle().Foreground(color).Render(blocks[cellH]))
			}
		}
		if y < height-1 {
			out.WriteString("\n")
		}
	}

	return lipgloss.NewStyle().Width(width).Render(out.String())
}

func renderTopBar(width int, logoLines []string, searchView string, isSearchFocused bool) string {
	var logoStr string
	if len(logoLines) > 0 {
		logoStr = strings.Join(logoLines, "\n")
	}

	logoW := 0
	if logoStr != "" {
		logoW = lipgloss.Width(logoStr)
	}

	searchW := width - logoW - 4
	if searchW < 20 {
		searchW = 20
	}

	searchStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#333333"))
	if isSearchFocused {
		searchStyle = searchStyle.BorderForeground(lipgloss.Color("#FA243C"))
	}

	searchRendered := searchStyle.Width(searchW).Render(searchView)

	topRow := lipgloss.JoinHorizontal(lipgloss.Bottom,
		lipgloss.NewStyle().Width(logoW).Render(logoStr),
		"  ",
		searchRendered,
	)

	return lipgloss.NewStyle().Width(width).Render(topRow)
}

func renderASCIILogo(rawLines []string, maxLines int) []string {
	if len(rawLines) == 0 {
		return nil
	}

	var result []string
	count := 0
	for _, line := range rawLines {
		if count >= maxLines {
			break
		}
		if line == "" && count == 0 {
			continue
		}
		result = append(result, line)
		count++
	}

	if len(result) > 0 {
		var colored []string
		colors := []string{"#4ade80", "#4ade80", "#facc15", "#f97316", "#FA243C", "#a855f7", "#3b82f6"}
		for i, line := range result {
			c := colors[i%len(colors)]
			colored = append(colored, lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Render(line))
		}
		return colored
	}
	return nil
}
