package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/sebasreateguib/applemusictui/internal/models"
)

func mouse(x, y int, action tea.MouseAction) tea.MouseMsg {
	return tea.MouseMsg{X: x, Y: y, Action: action, Button: tea.MouseButtonLeft}
}

// headColumn returns the column of the progress bar's head on the screen.
func headColumn(t *testing.T, a App) int {
	t.Helper()
	_, y, _ := a.playerBar()
	line := ansi.Strip(strings.Split(a.View(), "\n")[y])
	before, _, found := strings.Cut(line, "●")
	if !found {
		t.Fatalf("no progress bar in line %d: %q", y, line)
	}
	return ansi.StringWidth(before)
}

// The head of the bar must follow the mouse exactly, on wide and narrow screens.
func TestSeekHeadFollowsTheMouse(t *testing.T) {
	for _, size := range []tea.WindowSizeMsg{{Width: 140, Height: 45}, {Width: 80, Height: 24}, {Width: 60, Height: 16}} {
		for _, duration := range []float64{95, 252, 754, 4000} {
			a := filledApp(t)
			np := a.nowPlaying
			np.Track.Duration = duration
			a = send(t, a, size, nowPlayingMsg(np))
			x, y, w := a.playerBar()
			a = send(t, a, mouse(x, y, tea.MouseActionPress))
			if a.seek.target != 0 {
				t.Fatalf("%v: the first cell is not the start: %v", size, a.seek.target)
			}
			for col := x; col < x+w; col++ {
				// The vertical position does not matter once dragging.
				a = send(t, a, mouse(col, col%7, tea.MouseActionMotion))
				want := col
				if a.seek.target == duration-seekMargin {
					continue // the last cells may be less than a second
				}
				if got := headColumn(t, a); got != want {
					t.Fatalf("%v, %vs: mouse at %d, head at %d", size, duration, want, got)
				}
			}
			a = send(t, a, mouse(9999, 0, tea.MouseActionMotion))
			if a.seek.target != duration-seekMargin {
				t.Fatalf("seeking past the end: %v", a.seek.target)
			}
			a = send(t, a, mouse(-5, 0, tea.MouseActionMotion))
			if a.seek.target != 0 {
				t.Fatalf("seeking before the start: %v", a.seek.target)
			}
		}
	}
}

// Positions reported while seeking are old: the bar must not jump back to them.
func TestSeekIgnoresOldPositions(t *testing.T) {
	a := filledApp(t)
	np := a.nowPlaying
	x, y, w := a.playerBar()
	col := x + w/2

	a = send(t, a, mouse(x, y, tea.MouseActionPress), mouse(col, y, tea.MouseActionMotion), nowPlayingMsg(np))
	if got := headColumn(t, a); got != col {
		t.Fatalf("while dragging: head at %d, want %d", got, col)
	}
	m, cmd := a.Update(mouse(col, y, tea.MouseActionRelease))
	a = m.(App)
	if cmd == nil || a.seek.dragging || a.seek.pending != 1 {
		t.Fatalf("releasing did not seek: %+v", a.seek)
	}
	if a.nowPlaying.Position != a.seek.target {
		t.Fatalf("after releasing, the position is still %v", a.nowPlaying.Position)
	}
	a = send(t, a, nowPlayingMsg(np))
	if a.nowPlaying.Position != a.seek.target {
		t.Fatalf("an old position replaced the new one: %v", a.nowPlaying.Position)
	}
	// Grabbing the bar again must not change what was already sent.
	b := send(t, a, mouse(x, y, tea.MouseActionPress), nowPlayingMsg(np))
	if b.nowPlaying.Position != a.seek.sent || headColumn(t, b) != x {
		t.Fatalf("second drag: position %v, head at %d", b.nowPlaying.Position, headColumn(t, b))
	}
	if got := headColumn(t, a); got != col {
		t.Fatalf("after releasing: head at %d, want %d", got, col)
	}
	np.Position = a.seek.target
	a = send(t, a, seekedMsg(np))
	if got := headColumn(t, a); got != col {
		t.Fatalf("after seeking: head at %d, want %d", got, col)
	}
	np.Position = 0
	a = send(t, a, nowPlayingMsg(np))
	if got := headColumn(t, a); got != x {
		t.Fatalf("the bar no longer follows Music: head at %d, want %d", got, x)
	}
}

func TestSeekOnlyFromTheBar(t *testing.T) {
	a := filledApp(t)
	x, y, w := a.playerBar()
	click := func(a App, msg tea.MouseMsg) bool { return send(t, a, msg).seek.dragging }

	if !click(a, mouse(x+w-1, y, tea.MouseActionPress)) {
		t.Error("the last cell of the bar does not seek")
	}
	for _, p := range [][2]int{{x - 1, y}, {x + w, y}, {x, y - 1}, {x, y + 1}, {5, 5}} {
		if click(a, mouse(p[0], p[1], tea.MouseActionPress)) {
			t.Errorf("a click at %v seeks", p)
		}
	}
	right := mouse(x, y, tea.MouseActionPress)
	right.Button = tea.MouseButtonRight
	if click(a, right) || click(a, mouse(x, y, tea.MouseActionMotion)) {
		t.Error("only the left button seeks")
	}
	if click(send(t, a, press("?")), mouse(x, y, tea.MouseActionPress)) {
		t.Error("seeks behind a popup")
	}
	if click(send(t, a, nowPlayingMsg{State: models.StateStopped}), mouse(x, y, tea.MouseActionPress)) {
		t.Error("seeks with nothing playing")
	}
	a = send(t, a, mouse(x, y, tea.MouseActionPress), nowPlayingMsg{
		Track: models.Track{Name: "next", Duration: 100}, State: models.StatePlaying,
	})
	if a.seek.dragging {
		t.Error("the drag goes on in the next track")
	}
}

// Enabling the mouse must not take the wheel away from the lists.
func TestWheelMovesInTheList(t *testing.T) {
	a := send(t, filledApp(t), press("4"))
	before := a.tracks.SelectedIndex()
	a = send(t, a, tea.MouseMsg{Button: tea.MouseButtonWheelDown, Action: tea.MouseActionPress})
	if got := a.tracks.SelectedIndex(); got != before+1 {
		t.Fatalf("wheel down: %d, want %d", got, before+1)
	}
	a = send(t, a, tea.MouseMsg{Button: tea.MouseButtonWheelUp, Action: tea.MouseActionPress})
	if got := a.tracks.SelectedIndex(); got != before {
		t.Fatalf("wheel up: %d, want %d", got, before)
	}
}

func TestOldSeekKeysAreGone(t *testing.T) {
	a := filledApp(t)
	for _, k := range []string{",", "."} {
		if cmd := a.handleKey(press(k)); cmd != nil {
			t.Errorf("%q still does something", k)
		}
	}
}
