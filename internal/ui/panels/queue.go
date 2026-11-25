package panels

import (
	"slices"

	"github.com/charmbracelet/x/ansi"

	"SR-Player/internal/models"
	"SR-Player/internal/ui/theme"
)

// QueuePanel lists the current and upcoming tracks.
type QueuePanel struct {
	tracks        []models.QueueTrack
	cur           cursor
	width, height int
	focused       bool
	revision      int // grows with every change to what View shows
}

func NewQueuePanel() QueuePanel { return QueuePanel{} }

func (q *QueuePanel) SetTracks(tracks []models.QueueTrack) {
	if slices.Equal(q.tracks, tracks) {
		return
	}
	q.tracks = tracks
	q.cur.setTotal(len(tracks))
	q.cur.set(0)
	q.revision++
}

func (q *QueuePanel) SetSize(width, height int) {
	q.width, q.height = width, height
	q.cur.setRows(height - 2)
	q.revision++
}

func (q *QueuePanel) SetFocused(focused bool) {
	if q.focused != focused {
		q.focused = focused
		q.revision++
	}
}

func (q *QueuePanel) HandleKey(key string) {
	if q.cur.handleKey(key) {
		q.revision++
	}
}

// Revision changes whenever View would render something else.
func (q *QueuePanel) Revision() int { return q.revision }

func (q *QueuePanel) View() string {
	inner := q.width - 2
	frame := Frame{
		Key: "5", Title: "Up Next", Focused: q.focused,
		Top: q.cur.top, Visible: q.cur.rows, Total: len(q.tracks),
	}
	if len(q.tracks) == 0 {
		return frame.Render(q.width, q.height, message("Nothing queued", inner, q.height-2))
	}

	from, to := q.cur.window()
	lines := make([]string, 0, to-from)
	for i := from; i < to; i++ {
		lines = append(lines, q.row(i, inner))
	}
	return frame.Render(q.width, q.height, lines)
}

func (q *QueuePanel) row(i, width int) string {
	qt := q.tracks[i]
	name, artist := theme.Text, theme.Dim
	marker := "  "
	if qt.IsCurrent {
		marker = "▶ "
		name = theme.Active
	}
	if i == q.cur.index && q.focused {
		name = name.Background(theme.SelectedBg).Bold(true)
		artist = artist.Background(theme.SelectedBg)
	}

	text := " " + marker + ansi.Truncate(qt.Track.Name, max(0, width-4), "…")
	rest := width - ansi.StringWidth(text)
	by := ""
	if rest >= 8 {
		by = ansi.Truncate("  "+qt.Track.Artist, rest-1, "…")
	}
	return name.Render(text) + artist.Render(Fit(by, max(0, rest)))
}
