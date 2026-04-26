package panels

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"SR-Player/internal/models"
)

type trackListItem struct {
	track     models.Track
	index     int
	isCurrent bool
	titleW    int
	artistW   int
	rowW      int
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
	prefixStyle := lipgloss.NewStyle().Width(2).Foreground(lipgloss.Color("#4F545B"))
	prefix := "· "
	if t.isCurrent {
		prefixStyle = lipgloss.NewStyle().Width(2).Foreground(lipgloss.Color("#FF4D67")).Bold(true)
		prefix = "▶ "
	}

	indexStyle := lipgloss.NewStyle().
		Width(4).
		Align(lipgloss.Right).
		Foreground(lipgloss.Color("#72777E"))

	titleStyle := lipgloss.NewStyle().
		Width(t.titleW).
		MaxWidth(t.titleW).
		Foreground(lipgloss.Color("#E6E8EB"))
	if t.isCurrent {
		titleStyle = titleStyle.Foreground(lipgloss.Color("#FFF0C2")).Bold(true)
	}

	artistStyle := lipgloss.NewStyle().
		Width(t.artistW).
		MaxWidth(t.artistW).
		Foreground(lipgloss.Color("#7F858C"))

	durStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9EA4AB"))

	idx := fmt.Sprintf("%d", t.index)
	title := padTruncate(t.track.Name, t.titleW)
	artist := padTruncate(t.track.Artist, t.artistW)
	dur := formatTrackDuration(t.track.Duration)

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		prefixStyle.Render(prefix),
		indexStyle.Render(idx),
		"  ",
		titleStyle.Render(title),
		"  ",
		artistStyle.Render(artist),
		"  ",
		durStyle.Render(dur),
	)
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
	trackListBlurredBorder = lipgloss.Color("#2E3237")
)

func NewTrackListPanel(width, height int) TrackListPanel {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
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
	l.Styles.StatusBar = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666B72"))
	l.SetShowTitle(false)
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
	titleW, artistW := t.columnWidths()

	var cmds []tea.Cmd
	for i, tr := range tracks {
		item := trackListItem{
			track:     tr,
			index:     startIndex + i + 1,
			isCurrent: tr.Name == t.currentTrack,
			titleW:    titleW,
			artistW:   artistW,
			rowW:      t.rowWidth(),
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
	titleW, artistW := t.columnWidths()
	items := make([]list.Item, len(t.tracks))
	for i, tr := range t.tracks {
		items[i] = trackListItem{
			track:     tr,
			index:     i + 1,
			isCurrent: tr.Name == t.currentTrack,
			titleW:    titleW,
			artistW:   artistW,
			rowW:      t.rowWidth(),
		}
	}
	t.list.SetItems(items)
}

func (t *TrackListPanel) SetSize(width, height int) {
	t.width = width
	t.height = height
	t.list.SetSize(width-4, height-4)
	t.rebuildItems()
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
	header := t.renderHeader()
	body := t.list.View()
	content := lipgloss.JoinVertical(lipgloss.Left, header, "", body)

	border := trackListBlurredBorder
	if t.focused {
		border = trackListFocusedBorder
	}
	return renderRectBox(content, t.width, t.height, border, lipgloss.Top)
}

func (t *TrackListPanel) panelTitle() string {
	switch t.playlistName {
	case "", "Loading...":
		return "Loading Tracks"
	case "Songs":
		return "Library Tracks"
	default:
		return "Playlist: " + t.playlistName
	}
}

func (t *TrackListPanel) rowWidth() int {
	rowW := t.width - 12
	if rowW < 24 {
		rowW = 24
	}
	return rowW
}

func (t *TrackListPanel) columnWidths() (titleW, artistW int) {
	rowW := t.rowWidth()
	switch {
	case rowW >= 72:
		titleW = 34
		artistW = 20
	case rowW >= 58:
		titleW = 28
		artistW = 16
	default:
		titleW = 22
		artistW = 12
	}
	return titleW, artistW
}

func (t *TrackListPanel) renderHeader() string {
	rowW := t.width - 6
	if rowW < 20 {
		rowW = 20
	}

	kicker := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666B72")).
		Render(strings.ToUpper(t.panelTitle()))

	meta := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A7ADB4")).
		Bold(true).
		Render(fmt.Sprintf("%d tracks", len(t.tracks)))

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		lipgloss.NewStyle().Width(rowW-lipgloss.Width(meta)).Render(kicker),
		meta,
	)
}
