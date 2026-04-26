package ui

import (
	"github.com/charmbracelet/lipgloss"
)

type RightView int

const (
	ViewNowPlaying RightView = iota
	ViewQueue
)

type Layout struct {
	width  int
	height int
}

func NewLayout(width, height int) Layout {
	return Layout{width: width, height: height}
}

func (l *Layout) SetSize(width, height int) {
	l.width = width
	l.height = height
}

func (l *Layout) PanelSizes() (centerW, rightW, mainH int) {
	statusH := 2
	topH := 4
	mainH = l.height - statusH - topH
	if mainH < 5 {
		mainH = 5
	}

	rightW = 58
	maxRight := (l.width * 46) / 100
	if rightW > maxRight {
		rightW = maxRight
	}
	if rightW < 36 {
		rightW = 36
	}

	centerW = l.width - rightW
	if centerW < 24 {
		centerW = 24
	}
	return
}

func (l *Layout) Render(
	topView string,
	centerContentView string,
	rightContentView string,
	statusView string,
) string {
	centerW, rightW, mainH := l.PanelSizes()

	centerCol := lipgloss.NewStyle().
		Width(centerW).
		Height(mainH).
		MaxWidth(centerW).
		MaxHeight(mainH).
		Render(centerContentView)

	rightCol := lipgloss.NewStyle().
		Width(rightW).
		Height(mainH).
		MaxWidth(rightW).
		MaxHeight(mainH).
		Render(rightContentView)

	mainRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		centerCol,
		rightCol,
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		topView,
		mainRow,
		statusView,
	)
}
