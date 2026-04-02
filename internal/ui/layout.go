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

var (
	tabActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1DB954")).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color("#1DB954"))

	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#555555")).
				Padding(0, 1)
)

func NewLayout(width, height int) Layout {
	return Layout{width: width, height: height}
}

func (l *Layout) SetSize(width, height int) {
	l.width = width
	l.height = height
}

func (l *Layout) PanelSizes() (leftW, centerW, rightW, mainH int) {
	statusH := 2
	topH := 3 // Box length of search input
	mainH = l.height - statusH - topH
	if mainH < 5 {
		mainH = 5
	}

	leftW = 28

	rightW = 38
	if rightW > l.width/3 {
		rightW = l.width / 3
	}
	if rightW < 36 {
		rightW = 36
	}

	centerW = l.width - leftW - rightW
	if centerW < 20 {
		centerW = 20
	}
	return
}

func (l *Layout) Render(
	topView string,
	playlistView string,
	centerContentView string,
	rightContentView string,
	statusView string,
	activeRightTab RightView,
) string {
	leftW, centerW, rightW, mainH := l.PanelSizes()

	tabBar := renderTabBar(activeRightTab, rightW)

	rightPanel := lipgloss.JoinVertical(lipgloss.Left, tabBar, rightContentView)

	mainRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(leftW).MaxHeight(mainH).Render(playlistView),
		lipgloss.NewStyle().Width(centerW).MaxHeight(mainH).Render(centerContentView),
		lipgloss.NewStyle().Width(rightW).MaxHeight(mainH).Render(rightPanel),
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		topView,
		mainRow,
		statusView,
	)
}

func renderTabBar(active RightView, width int) string {
	tabs := []struct {
		label string
		view  RightView
		key   string
	}{
		{"Now Playing", ViewNowPlaying, "1"},
		{"Queue", ViewQueue, "2"},
	}

	var rendered []string
	for _, tab := range tabs {
		label := tab.label + " [" + tab.key + "]"
		if tab.view == active {
			rendered = append(rendered, tabActiveStyle.Render(label))
		} else {
			rendered = append(rendered, tabInactiveStyle.Render(label))
		}
	}

	bar := lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
	return lipgloss.NewStyle().Width(width).MaxWidth(width).
		Background(lipgloss.Color("#0a0a0a")).
		Render(bar)
}
