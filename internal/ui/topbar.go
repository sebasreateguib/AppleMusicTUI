package ui

import (
	"math/rand"

	"github.com/charmbracelet/lipgloss"
)

var (
	topBrandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF2D55")).
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

// glyphPool is the character set used for the encrypted scramble effect.
var glyphPool = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*<>?")

const brandTarget = "MusicPlayer"

// encryptedBrand returns the brand name with a "decrypting" animation effect.
// Characters are revealed from left to right as tick advances; the rest
// show random glyphs, giving a hacker/encrypted aesthetic.
func encryptedBrand(tick int) string {
	runes := []rune(brandTarget)
	totalChars := len(runes)

	// Full cycle: each character takes ~2 ticks to stabilise, then hold
	// for a full cycle before scrambling again.
	cycleLen := totalChars*2 + 10
	phase := tick % cycleLen

	revealed := phase / 2
	if revealed > totalChars {
		revealed = totalChars
	}

	src := rand.New(rand.NewSource(int64(tick)))

	out := make([]rune, totalChars)
	for i, ch := range runes {
		if i < revealed {
			out[i] = ch
		} else {
			out[i] = glyphPool[src.Intn(len(glyphPool))]
		}
	}
	return string(out)
}

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

func renderTopBar(width int, activeTab TopTab, tick int) string {
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
	brand := topBrandStyle.Render("MusicPlayer")

	innerW := width - 2
	if innerW < 36 {
		innerW = 36
	}

	brandW := lipgloss.Width(brand)
	navW := innerW - brandW - 2
	if navW < lipgloss.Width(tabsRow) {
		navW = lipgloss.Width(tabsRow)
	}

	// Top-align so the brand text sits on the same row as the tab labels.
	// (Tabs are 2 lines tall because of the active-tab underline border;
	// using anything other than Top would push the 1-line brand downward.)
	navContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
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
