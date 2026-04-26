package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	topBrandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5C6166")).
			Bold(true).
			Padding(0, 1)

	topTabActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF4D67")).
				Bold(true).
				Padding(0, 3).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(lipgloss.Color("#FF4D67"))

	topTabInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#767A80")).
				Padding(0, 3)
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
	brand := topBrandStyle.Render("SR PLAYER")

	innerW := width - 2
	if innerW < 36 {
		innerW = 36
	}

	brandW := lipgloss.Width(brand)
	navW := innerW - brandW - 2
	if navW < lipgloss.Width(tabsRow) {
		navW = lipgloss.Width(tabsRow)
	}

	navContent := lipgloss.JoinHorizontal(
		lipgloss.Center,
		lipgloss.NewStyle().
			Width(brandW+2).
			AlignHorizontal(lipgloss.Left).
			Render(brand),
		lipgloss.NewStyle().
			Width(navW).
			MaxWidth(navW).
			AlignHorizontal(lipgloss.Center).
			Render(tabsRow),
	)

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color("#2D3136")).
		Padding(1, 0, 0, 0).
		Render(navContent)
}
