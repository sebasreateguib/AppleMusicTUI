package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"SR-Player/internal/cava"
	"SR-Player/internal/config"
	"SR-Player/internal/models"
)

func newTestApp(t *testing.T) App {
	t.Setenv("PATH", "") // no cava
	a := NewApp(0, 0, config.Config{}, false)
	t.Cleanup(a.Close)
	return a
}

// TestViewCacheIsNeverStale checks, after every message, that the screen drawn
// from the cache is the one that would be drawn without it.
func TestViewCacheIsNeverStale(t *testing.T) {
	a := newTestApp(t)
	step := func(name string, msg tea.Msg) {
		t.Helper()
		m, _ := a.Update(msg)
		a = m.(App)
		fresh := a
		fresh.cache = &viewCache{}
		if a.View() != fresh.View() {
			t.Errorf("%s: the cached screen is stale", name)
		}
	}
	key := func(k string) {
		t.Helper()
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		switch k {
		case "down":
			msg = tea.KeyMsg{Type: tea.KeyDown}
		case "tab":
			msg = tea.KeyMsg{Type: tea.KeyTab}
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		}
		step("key "+k, msg)
	}
	playing := func(name string, position float64, volume int) nowPlayingMsg {
		return nowPlayingMsg{
			Track:    models.Track{Name: name, Artist: "Artist", Album: "Album", Duration: 200},
			State:    models.StatePlaying,
			Position: position,
			Volume:   volume,
		}
	}
	tracks := []models.Track{{Name: "One"}, {Name: "Two"}, {Name: "Three"}}

	step("size", tea.WindowSizeMsg{Width: 140, Height: 45})
	step("tracks", tracksMsg{ctx: trackContext{kind: kindLibrary}, tracks: tracks})
	step("playlists", playlistsMsg{{Name: "Mix", Count: 3}})
	step("artists", uniqueStringsMsg{kind: "artists", items: []string{"b", "A", "c"}})
	step("albums", uniqueStringsMsg{kind: "albums", items: []string{"x", "y"}})
	step("playing", playing("One", 1, 50))
	step("position", playing("One", 40, 50))
	step("volume", playing("One", 40, 80))
	step("tick", tickMsg{})
	step("next track", playing("Two", 0, 80))
	step("bars", visualizerMsg(cava.Event{Levels: []byte{10, 200, 90}}))
	step("long title", playing(strings.Repeat("long title ", 12), 0, 80))
	for range 100 { // marquee and brand animations
		step("tick", tickMsg{})
	}
	for _, k := range []string{"down", "2", "down", "tab", "down", "4", "down", "down", "G", "g", "c", "5", "esc"} {
		key(k)
	}
	step("preview", previewMsg{a.previewSeq})
	step("preview tracks", tracksMsg{ctx: a.context, tracks: tracks[:1]})
	step("queue", queueMsg{{Index: 1, Track: models.Track{Name: "Queued"}}})
	step("queue again", queueMsg{{Index: 1, Track: models.Track{Name: "Other"}, IsCurrent: true}})
	for _, k := range []string{"5", "down", "i", "s", "r", "R", "?", "down", "esc", "/", "a", "esc", "1", "h", "l"} {
		key(k)
	}
	step("narrow", tea.WindowSizeMsg{Width: 80, Height: 24})
	for _, k := range []string{"5", "4", "1", "2", "3", "?", "esc"} {
		key(k)
	}
	step("tiny", tea.WindowSizeMsg{Width: 40, Height: 10})
	step("wide again", tea.WindowSizeMsg{Width: 200, Height: 60})
	step("cava error", visualizerMsg(cava.Event{Err: cava.ErrNotInstalled}))
}

// TestViewCacheIsUsed checks that new bars alone do not render the panels.
func TestViewCacheIsUsed(t *testing.T) {
	a := newTestApp(t)
	m, _ := a.Update(tea.WindowSizeMsg{Width: 140, Height: 70})
	a = m.(App)
	a.View()
	before := *a.cache

	m, _ = a.Update(visualizerMsg(cava.Event{Levels: []byte{10, 200, 90}}))
	a = m.(App)
	a.View()

	same := func(x, y section) bool { return &x.lines[0] == &y.lines[0] }
	for i := range before.sides {
		if !same(before.sides[i], a.cache.sides[i]) {
			t.Errorf("a visualizer frame rendered side list %d", i)
		}
	}
	if !same(before.tracks, a.cache.tracks) || !same(before.nowPlaying, a.cache.nowPlaying) ||
		!same(before.queue, a.cache.queue) || !same(before.player, a.cache.player) ||
		!same(before.bottom, a.cache.bottom) {
		t.Error("a visualizer frame rendered panels that did not change")
	}
}
