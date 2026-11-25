package panels

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"SR-Player/internal/models"
	"SR-Player/internal/ui/theme"
)

// TrackList is the main panel: the tracks of the selected playlist, artist or
// album, or the results of a search.
type TrackList struct {
	title         string
	kind          string // "playlist", "artist", "album", "library", "search"
	tracks        []models.Track
	current       models.Track
	cur           cursor
	width, height int
	focused       bool
	loading       bool
	empty         string
	revision      int // grows with every change to what View shows
}

func NewTrackList() TrackList { return TrackList{loading: true} }

// SetLoading empties the list until SetTracks is called.
func (t *TrackList) SetLoading(kind, title string) {
	t.kind, t.title = kind, title
	t.tracks, t.loading = nil, true
	t.cur = cursor{rows: t.cur.rows}
	t.revision++
}

// SetTracks fills the list; empty is shown when there are no tracks.
func (t *TrackList) SetTracks(kind, title string, tracks []models.Track, empty string) {
	t.kind, t.title = kind, title
	t.tracks, t.loading, t.empty = tracks, false, empty
	t.cur = cursor{rows: t.cur.rows}
	t.cur.setTotal(len(tracks))
	t.revision++
}

// SetCurrent marks the track that is playing.
func (t *TrackList) SetCurrent(track models.Track) {
	if !sameTrack(t.current, track) {
		t.current = track
		t.revision++
	}
}

func sameTrack(a, b models.Track) bool {
	return a.Name == b.Name && a.Artist == b.Artist && a.Album == b.Album
}

func (t *TrackList) Len() int { return len(t.tracks) }

// SelectedIndex returns the 1-based position of the selected track, 0 if none.
func (t *TrackList) SelectedIndex() int {
	if len(t.tracks) == 0 {
		return 0
	}
	return t.cur.index + 1
}

// Selected returns the selected track, or nil if the list is empty.
func (t *TrackList) Selected() *models.Track {
	if len(t.tracks) == 0 {
		return nil
	}
	tr := t.tracks[t.cur.index]
	return &tr
}

// SelectCurrent selects the track that is playing, if it is in the list.
func (t *TrackList) SelectCurrent() bool {
	for i, tr := range t.tracks {
		if t.current.Name != "" && sameTrack(tr, t.current) {
			t.cur.set(i)
			t.cur.center()
			t.revision++
			return true
		}
	}
	return false
}

func (t *TrackList) SetSize(width, height int) {
	t.width, t.height = width, height
	t.cur.setRows(height - 3) // frame and column titles
	t.revision++
}

func (t *TrackList) SetFocused(focused bool) {
	if t.focused != focused {
		t.focused = focused
		t.revision++
	}
}

// HandleKey reports whether the selection changed.
func (t *TrackList) HandleKey(key string) bool {
	moved := t.cur.handleKey(key)
	if moved {
		t.revision++
	}
	return moved
}

// Revision changes whenever View would render something else.
func (t *TrackList) Revision() int { return t.revision }

func (t *TrackList) View() string {
	inner := t.width - 2
	frame := Frame{
		Key: "4", Title: "Tracks", Focused: t.focused,
		Footer: position(t.cur.index, len(t.tracks)),
		Top:    t.cur.top, Visible: t.cur.rows, Total: len(t.tracks),
	}
	if t.title != "" {
		frame.Title += " - " + t.title
	}
	if total := totalDuration(t.tracks); total != "" {
		frame.Info = theme.Dim.Render(" " + total + " ")
	}
	if len(t.tracks) == 0 {
		text := t.empty
		if t.loading {
			text = "Loading…"
		}
		return frame.Render(t.width, t.height, message(text, inner, t.height-2))
	}

	cols := t.columns(inner)
	from, to := t.cur.window()
	lines := make([]string, 0, to-from+1)
	lines = append(lines, cols.render([]string{"", "#", "Title", "Artist", "Album", "Time"}, func(int) lipgloss.Style { return theme.Dim }, theme.Dim))
	for i := from; i < to; i++ {
		lines = append(lines, t.row(i, cols))
	}
	return frame.Render(t.width, t.height, lines)
}

func (t *TrackList) row(i int, cols trackColumns) string {
	tr := t.tracks[i]
	playing := t.current.Name != "" && sameTrack(tr, t.current)

	styles := [6]lipgloss.Style{theme.Active, theme.Dim, theme.Text, theme.Dim, theme.Dim, theme.Dim}
	marker := ""
	if playing {
		marker = "▶"
		styles[1], styles[2] = theme.Active, theme.Active
	}
	gap := theme.Text
	if i == t.cur.index {
		bg := theme.ActiveBg
		if t.focused {
			bg = theme.SelectedBg
			styles[2] = styles[2].Bold(true)
			styles[3] = theme.Text
		}
		gap = gap.Background(bg)
		for j := range styles {
			styles[j] = styles[j].Background(bg)
		}
	}

	duration := ""
	if tr.Duration > 0 {
		duration = formatTrackDuration(tr.Duration)
	}
	cells := []string{marker, strconv.Itoa(i + 1), tr.Name, tr.Artist, tr.Album, duration}
	return cols.render(cells, func(c int) lipgloss.Style { return styles[c] }, gap)
}

// trackColumns holds the width of each column, 0 for the hidden ones.
type trackColumns [6]int

func (t *TrackList) columns(width int) trackColumns {
	const gaps = 2
	cols := trackColumns{1, max(2, len(strconv.Itoa(len(t.tracks)))), 0, 0, 0, 5}

	// The column every row would repeat is left out, and so is the album when
	// there is no room for it.
	artist := t.kind != "artist" && width >= 40
	album := t.kind != "album" && width >= 62

	flex := width - 1 - cols[0] - 1 - cols[1] - gaps - gaps - cols[5] - 1
	if totalDuration(t.tracks) == "" { // search results have no duration
		flex += cols[5] + 1
		cols[5] = 0
	}
	if artist {
		flex -= gaps
	}
	if album {
		flex -= gaps
	}
	flex = max(4, flex)
	switch {
	case artist && album:
		cols[3], cols[4] = flex*28/100, flex*28/100
	case artist:
		cols[3] = flex * 40 / 100
	case album:
		cols[4] = flex * 42 / 100
	}
	cols[2] = flex - cols[3] - cols[4]
	return cols
}

func (c trackColumns) render(cells []string, style func(int) lipgloss.Style, gap lipgloss.Style) string {
	var sb strings.Builder
	sb.WriteString(gap.Render(" "))
	for i, w := range c {
		if w == 0 {
			continue
		}
		text := ansi.Truncate(cells[i], w, "…")
		pad := strings.Repeat(" ", max(0, w-ansi.StringWidth(text)))
		if i == 1 || i == 5 {
			text = pad + text
		} else {
			text += pad
		}
		sb.WriteString(style(i).Render(text))
		switch i {
		case 0, 5:
			sb.WriteString(gap.Render(" "))
		default:
			sb.WriteString(gap.Render("  "))
		}
	}
	return sb.String()
}

func formatTrackDuration(secs float64) string {
	total := int(math.Round(secs))
	return fmt.Sprintf("%d:%02d", total/60, total%60)
}

// totalDuration formats the length of all tracks, or "" if it is unknown.
func totalDuration(tracks []models.Track) string {
	var secs float64
	for _, t := range tracks {
		secs += t.Duration
	}
	minutes := int(math.Round(secs / 60))
	switch {
	case minutes <= 0:
		return ""
	case minutes < 60:
		return fmt.Sprintf("%d min", minutes)
	default:
		return fmt.Sprintf("%d h %02d min", minutes/60, minutes%60)
	}
}
