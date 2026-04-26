package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	topTabActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FA243C")).
				Bold(true).
				Padding(0, 2).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(lipgloss.Color("#FA243C"))

	topTabInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Padding(0, 2)
)

func navTabTitle(tab TopTab) string {
	switch tab {
	case TabAlbums:
		return "Albums"
	case TabArtists:
		return "Artists"
	case TabSongs:
		return "Songs"
	case TabPlaylists:
		return "Playlists"
	case TabSearch:
		return "Search"
	default:
		return "Songs"
	}
}

func renderTopBar(width int, activeTab TopTab) string {
	tabs := []TopTab{TabAlbums, TabArtists, TabSongs, TabPlaylists, TabSearch}
	var renderedTabs []string
	for _, tab := range tabs {
		label := navTabTitle(tab)
		if tab == activeTab {
			renderedTabs = append(renderedTabs, topTabActiveStyle.Render(label))
		} else {
			renderedTabs = append(renderedTabs, topTabInactiveStyle.Render(label))
		}
	}
	tabsRow := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	innerW := width - 2
	if innerW < 24 {
		innerW = 24
	}

	navContent := lipgloss.NewStyle().
		Width(innerW).
		MaxWidth(innerW).
		AlignHorizontal(lipgloss.Center).
		Render(tabsRow)

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color("#333333")).
		Render(navContent)
}
