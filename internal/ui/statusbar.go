package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"SR-Player/internal/models"
)

type StatusBar struct {
	width int
}

var (
	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#111111")).
			Foreground(lipgloss.Color("#CCCCCC")).
			Padding(0, 2)

	progressFilledStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1DB954"))

	progressEmptyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#333333"))

	controlActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1DB954")).
				Bold(true)

	controlInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888"))

	timeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))
)

func NewStatusBar(width int) StatusBar {
	return StatusBar{width: width}
}

func (s *StatusBar) SetWidth(width int) {
	s.width = width
}

func (s *StatusBar) View(np models.NowPlaying) string {
	prevBtn := "⏮"
	playBtn := "▶"
	if np.State == models.StatePlaying {
		playBtn = "⏸"
	}
	nextBtn := "⏭"

	shuffleBtn := controlInactiveStyle.Render("⇄")
	if np.ShuffleEnabled {
		shuffleBtn = controlActiveStyle.Render("⇄")
	}

	repeatIcon := "↺"
	repeatStr := controlInactiveStyle.Render(repeatIcon)
	switch np.RepeatMode {
	case models.RepeatAll:
		repeatStr = controlActiveStyle.Render("↺")
	case models.RepeatOne:
		repeatStr = controlActiveStyle.Render("↺¹")
	}

	controls := fmt.Sprintf("  %s  %s  %s  ",
		controlInactiveStyle.Render(prevBtn),
		controlActiveStyle.Render(playBtn),
		controlInactiveStyle.Render(nextBtn),
	)
	flagControls := fmt.Sprintf("  %s  %s  ", shuffleBtn, repeatStr)

	elapsed := np.Position
	total := np.Track.Duration

	elapsedStr := formatTime(elapsed)
	totalStr := formatTime(total)

	controlsLen := lipgloss.Width(controls)
	flagLen := lipgloss.Width(flagControls)
	timeLen := len(elapsedStr) + len(totalStr) + 3 
	padding := 4

	barWidth := s.width - controlsLen - flagLen - timeLen - padding
	if barWidth < 4 {
		barWidth = 4
	}

	progressBar := renderProgressBar(elapsed, total, barWidth)

	row := lipgloss.JoinHorizontal(
		lipgloss.Center,
		controls,
		timeStyle.Render(elapsedStr),
		" ",
		progressBar,
		" ",
		timeStyle.Render(totalStr),
		flagControls,
	)

	keyhints := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#444444")).
		Render("  [space] play/pause  [n/p] next/prev  [s] shuffle  [r] repeat  [←/→] seek  [q] quit")

	full := lipgloss.JoinVertical(lipgloss.Left,
		row,
		keyhints,
	)

	return statusBarStyle.Width(s.width).MaxWidth(s.width).MaxHeight(2).Render(full)
}

func renderProgressBar(elapsed, total float64, width int) string {
	if width <= 0 {
		return ""
	}
	if total <= 0 {
		return progressEmptyStyle.Render(strings.Repeat("─", width))
	}

	ratio := elapsed / total
	if ratio > 1 {
		ratio = 1
	}
	if ratio < 0 {
		ratio = 0
	}

	filled := int(math.Round(ratio * float64(width)))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	empty := width - filled
	if empty < 0 {
		empty = 0
	}

	if filled > 0 && filled < width {
		filledPart := progressFilledStyle.Render(strings.Repeat("━", filled-1))
		head := controlActiveStyle.Render("●")
		rest := progressEmptyStyle.Render(strings.Repeat("─", empty))
		return filledPart + head + rest
	}

	return progressFilledStyle.Render(strings.Repeat("━", filled)) +
		progressEmptyStyle.Render(strings.Repeat("─", empty))
}

func formatTime(secs float64) string {
	if secs < 0 {
		secs = 0
	}
	total := int(math.Round(secs))
	m := total / 60
	s := total % 60
	return fmt.Sprintf("%d:%02d", m, s)
}
