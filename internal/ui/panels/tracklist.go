package panels

import (
	"fmt"
	"math"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"SR-Player/internal/models"
)

type trackListItem struct {
	track     models.Track
	index     int
	isCurrent bool
}

func padTruncate(s string, width int) string {
	runes := []rune(s)
	if len(runes) > width {
		if width > 1 {
			return string(runes[:width-1]) + "…"
		}
		return ""
	}
	return s + string(make([]rune, width-len(runes)))
}

func (t trackListItem) Title() string {
	prefix := "  "
	if t.isCurrent {
		prefix = "▶ "
	}
	idx := fmt.Sprintf("%2d", t.index)
	title := padTruncate(t.track.Name, 35)
	artist := padTruncate(t.track.Artist, 25)
	dur := formatTrackDuration(t.track.Duration)

	// Single line format for the table row
	return fmt.Sprintf("%s%s  %s %s [%s]", prefix, idx, title, artist, dur)
}

func (t trackListItem) Description() string {
	// Empty description because we packed everything into Title.
	// This reduces list row height to 1 line, fixing layout overflows and list scrolling issues.
	return ""
}

func (t trackListItem) FilterValue() string { return t.track.Name + t.track.Artist }

func formatTrackDuration(secs float64) string {
	total := int(math.Round(secs))
	m := total / 60
	s := total % 60
	return fmt.Sprintf("%d:%02d", m, s)
}

type TrackListPanel struct {
	list         list.Model
	tracks       []models.Track
	playlistName string
	currentTrack string
	width        int
	height       int
	focused      bool
}

var (
	trackListFocusedBorder = lipgloss.Color("#FA243C")
	trackListBlurredBorder = lipgloss.Color("#333333")
)

func NewTrackListPanel(width, height int) TrackListPanel {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#e1ca72")).
		BorderLeftForeground(lipgloss.Color("#e1ca72"))

	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(lipgloss.Color("#CCCCCC"))

	// Disable description styling since we don't use it anymore
	delegate.ShowDescription = false

	l := list.New([]list.Item{}, delegate, width-4, height-4)
	l.Title = "Playlist Tracks"
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)
	l.SetShowHelp(false)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	return TrackListPanel{
		list:   l,
		width:  width,
		height: height,
	}
}

func (t *TrackListPanel) SetTracks(playlistName string, tracks []models.Track, currentTrack string) {
	t.playlistName = playlistName
	t.tracks = tracks
	t.currentTrack = currentTrack
	t.rebuildItems()
	t.list.Title = "Playlist · " + playlistName
}

func (t *TrackListPanel) SetCurrentTrack(name string) {
	if t.currentTrack == name {
		return
	}
	t.currentTrack = name
	t.rebuildItems()
}

func (t *TrackListPanel) GetTracks() []models.Track {
	return t.tracks
}

func (t *TrackListPanel) AppendTracks(tracks []models.Track) tea.Cmd {
	startIndex := len(t.tracks)
	t.tracks = append(t.tracks, tracks...)

	var cmds []tea.Cmd
	for i, tr := range tracks {
		item := trackListItem{
			track:     tr,
			index:     startIndex + i + 1,
			isCurrent: tr.Name == t.currentTrack,
		}
		// Safely add items without destroying the current cursor position
		cmd := t.list.InsertItem(len(t.list.Items()), item)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

func (t *TrackListPanel) rebuildItems() {
	items := make([]list.Item, len(t.tracks))
	for i, tr := range t.tracks {
		items[i] = trackListItem{
			track:     tr,
			index:     i + 1,
			isCurrent: tr.Name == t.currentTrack,
		}
	}
	t.list.SetItems(items)
}

func (t *TrackListPanel) SetSize(width, height int) {
	t.width = width
	t.height = height
	t.list.SetSize(width-4, height-4)
}

func (t *TrackListPanel) SetFocused(focused bool) {
	t.focused = focused
}

func (t *TrackListPanel) SelectedTrackIndex() int {
	return t.list.Index() + 1
}

func (t *TrackListPanel) SelectedTrack() *models.Track {
	item := t.list.SelectedItem()
	if item == nil {
		return nil
	}
	ti, ok := item.(trackListItem)
	if !ok {
		return nil
	}
	return &ti.track
}

func (t *TrackListPanel) PlaylistName() string {
	return t.playlistName
}

func (t *TrackListPanel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	t.list, cmd = t.list.Update(msg)
	return cmd
}

func (t *TrackListPanel) View() string {
	border := trackListBlurredBorder
	if t.focused {
		border = trackListFocusedBorder
	}
	return renderRectBox(t.list.View(), t.width, t.height, border, lipgloss.Top)
}
