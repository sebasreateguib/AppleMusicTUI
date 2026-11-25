package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sebasreateguib/applemusictui/internal/models"
)

// seekMargin keeps a seek to the end of the bar from skipping to the next track.
const seekMargin = 1.0

// seek is the state of moving the playback position with the mouse.
type seek struct {
	dragging bool
	target   float64 // position under the mouse
	sent     float64 // last position sent to Music
	pending  int     // seeks Music has not confirmed yet
}

// shownNowPlaying is what the UI displays: the position under the mouse while
// the bar is dragged.
func (a App) shownNowPlaying() models.NowPlaying {
	np := a.nowPlaying
	if a.seek.dragging {
		np.Position = a.seek.target
	}
	return np
}

// playerBar locates the progress bar on the screen.
func (a App) playerBar() (x, y, width int) {
	line := newPlayerLine(a.shownNowPlaying(), a.width, !a.layout.wide())
	// 1: the frame's border.
	return 1 + line.barX, a.layout.bodyH + 1, line.barW
}

// seekTo aims at the position under column x.
func (a *App) seekTo(x int) {
	barX, _, width := a.playerBar()
	total := a.nowPlaying.Track.Duration
	cell := clamp(x-barX, 0, width-1)
	target := total * float64(cell) / float64(max(1, width-1))
	a.seek.target = max(0, min(target, total-seekMargin))
}

func (a *App) handleMouse(msg tea.MouseMsg) tea.Cmd {
	// With the mouse enabled, terminals stop sending the wheel as arrow keys.
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		return a.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	case tea.MouseButtonWheelDown:
		return a.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	}

	switch msg.Action {
	case tea.MouseActionPress:
		a.seek.dragging = false
		np := a.nowPlaying
		if msg.Button != tea.MouseButtonLeft || a.popup != popupNone || a.layout.tooSmall() ||
			np.State == models.StateStopped || np.Track.Duration <= 0 {
			return nil
		}
		if x, y, w := a.playerBar(); msg.Y == y && msg.X >= x && msg.X < x+w {
			a.seek.dragging = true
			a.seekTo(msg.X)
		}

	case tea.MouseActionMotion:
		if a.seek.dragging {
			a.seekTo(msg.X)
		}

	case tea.MouseActionRelease:
		if a.seek.dragging {
			a.seekTo(msg.X)
			a.seek.dragging = false
			a.seek.sent = a.seek.target
			a.seek.pending++
			// Right away, so that nothing acts on the old position.
			a.nowPlaying.Position = a.seek.sent
			a.nowPlayingAt = time.Now()
			return doSeek(a.seek.target)
		}
	}
	return nil
}
