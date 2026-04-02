package panels

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"SR-Player/internal/models"
)

type playlistItem struct {
	playlist models.Playlist
}

func (p playlistItem) Title() string       { return p.playlist.Name }
func (p playlistItem) Description() string { return fmt.Sprintf("%d tracks", p.playlist.Count) }
func (p playlistItem) FilterValue() string { return p.playlist.Name }

type PlaylistsPanel struct {
	list    list.Model
	width   int
	height  int
	focused bool
}

var (
	playlistFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#1DB954")).
				Padding(0, 1)

	playlistBlurredStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#333333")).
				Padding(0, 1)
)

func NewPlaylistsPanel(width, height int) PlaylistsPanel {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#1DB954")).
		BorderLeftForeground(lipgloss.Color("#1DB954"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#1DB954")).
		BorderLeftForeground(lipgloss.Color("#1DB954"))

	l := list.New([]list.Item{}, delegate, width-4, height-4)
	l.Title = "Playlists"
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return PlaylistsPanel{
		list:   l,
		width:  width,
		height: height,
	}
}

func (p *PlaylistsPanel) SetPlaylists(playlists []models.Playlist) {
	items := make([]list.Item, len(playlists))
	for i, pl := range playlists {
		items[i] = playlistItem{playlist: pl}
	}
	p.list.SetItems(items)
}

func (p *PlaylistsPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.list.SetSize(width-4, height-2)
}

func (p *PlaylistsPanel) SetFocused(focused bool) {
	p.focused = focused
}

func (p *PlaylistsPanel) SelectedPlaylist() *models.Playlist {
	item := p.list.SelectedItem()
	if item == nil {
		return nil
	}
	pi, ok := item.(playlistItem)
	if !ok {
		return nil
	}
	return &pi.playlist
}

func (p *PlaylistsPanel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	p.list, cmd = p.list.Update(msg)
	return cmd
}

func (p *PlaylistsPanel) View() string {
	style := playlistBlurredStyle
	if p.focused {
		style = playlistFocusedStyle
	}
	return style.Width(p.width - 2).Height(p.height - 2).MaxWidth(p.width - 2).MaxHeight(p.height - 2).Render(p.list.View())
}
