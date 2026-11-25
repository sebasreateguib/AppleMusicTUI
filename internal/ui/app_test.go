package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"

	"github.com/sebasreateguib/applemusictui/internal/cava"
	"github.com/sebasreateguib/applemusictui/internal/config"
	"github.com/sebasreateguib/applemusictui/internal/models"
	"github.com/sebasreateguib/applemusictui/internal/ui/theme"
)

func send(t *testing.T, a App, msgs ...tea.Msg) App {
	t.Helper()
	for _, msg := range msgs {
		m, _ := a.Update(msg)
		a = m.(App)
	}
	return a
}

func press(k string) tea.KeyMsg {
	switch k {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func filledApp(t *testing.T) App {
	var tracks []models.Track
	for i := range 300 {
		tracks = append(tracks, models.Track{
			Name: fmt.Sprintf("日本語の曲 🎵 %d with a rather long name", i), Artist: "Beyoncé & JAŸ-Z", Album: "Álbum 👩‍🚀", Duration: 200,
		})
	}
	var names []string
	for i := range 80 {
		names = append(names, fmt.Sprintf("Artista número %d with a very long name indeed", i))
	}
	a := send(t, newTestApp(t),
		tea.WindowSizeMsg{Width: 140, Height: 45},
		tracksMsg{ctx: trackContext{kind: kindLibrary}, tracks: tracks},
		playlistsMsg{{Name: "Mix", Count: 3}, {Name: "Otra", Count: 1200}},
		uniqueStringsMsg{kind: "artists", items: names},
		uniqueStringsMsg{kind: "albums", items: names},
		nowPlayingMsg{Track: tracks[7], State: models.StatePlaying, Position: 20, Volume: 70},
		visualizerMsg(cava.Event{Levels: []byte{10, 200, 90, 255, 0, 40}}),
	)
	return send(t, a, queueMsg{seq: a.queueSeq, tracks: []models.QueueTrack{{Index: 1, Track: tracks[7], IsCurrent: true}, {Index: 2, Track: tracks[8]}}})
}

// Every screen must fill the terminal exactly, or lines wrap and panels break.
func TestViewFillsTheTerminal(t *testing.T) {
	a := filledApp(t)
	for _, k := range []string{"1", "2", "3", "4", "5", "?", "/"} {
		a := send(t, a, press(k))
		for w := 20; w <= 240; w += 3 {
			for h := 6; h <= 80; h += 3 {
				a := send(t, a, tea.WindowSizeMsg{Width: w, Height: h})
				lines := strings.Split(a.View(), "\n")
				if len(lines) != h {
					t.Fatalf("key %s, %dx%d: %d lines", k, w, h, len(lines))
				}
				for i, l := range lines {
					if got := ansi.StringWidth(l); got != w {
						t.Fatalf("key %s, %dx%d: line %d is %d wide: %q", k, w, h, i, got, ansi.Strip(l))
					}
				}
			}
		}
	}
}

// Playing must never use the tracks of one context and the name of another.
func TestOpenBeforeThePreviewArrives(t *testing.T) {
	a := send(t, filledApp(t), press("2"), press("down"))
	if a.context.kind != kindLibrary {
		t.Fatalf("context changed before its tracks were asked for: %v", a.context)
	}
	stale := previewMsg{a.previewSeq}

	a = send(t, a, press("enter"))
	want := trackContext{kindArtist, a.sides[panelArtists].Selected().Value}
	if a.context != want || a.focus != panelTracks || a.tracks.Len() != 0 {
		t.Fatalf("context = %v, focus = %v, %d tracks", a.context, a.focus, a.tracks.Len())
	}
	a = send(t, a, press("enter")) // nothing to play yet
	if a.hasContext() {
		t.Fatal("played a track that has not arrived")
	}

	// Late replies for other contexts are kept, but not shown.
	a = send(t, a, stale, tracksMsg{ctx: trackContext{kindAlbum, "other"}, tracks: make([]models.Track, 4)})
	if a.context != want || a.tracks.Len() != 0 {
		t.Fatal("showed the tracks of another context")
	}
	a = send(t, a, tracksMsg{ctx: want, tracks: []models.Track{{Name: "A"}, {Name: "B"}}}, press("down"), press("enter"))
	if a.playback != want || a.playbackTrackIndex != 2 {
		t.Fatalf("played %v %d", a.playback, a.playbackTrackIndex)
	}
}

func TestSearch(t *testing.T) {
	a := send(t, filledApp(t), press("/"), press("q"), press("1"), press("enter"))
	ctx := trackContext{kindSearch, "q1"}
	if a.context != ctx || a.popup != popupNone || a.focus != panelTracks {
		t.Fatalf("context = %v, popup = %v, focus = %v", a.context, a.popup, a.focus)
	}
	a = send(t, a, tracksMsg{ctx: ctx, tracks: []models.Track{{Name: "Found"}}})
	if !strings.Contains(ansi.Strip(a.View()), "Found") {
		t.Fatal("results are not shown")
	}
	a = send(t, a, press("esc"))
	if a.context.kind != kindLibrary || a.tracks.Len() != 300 {
		t.Fatalf("esc did not go back to the library: %v", a.context)
	}
}

// Moving over many rows only asks for the one the cursor stops at.
func TestPreviewIsDebounced(t *testing.T) {
	a := send(t, filledApp(t), press("2"))
	var old []previewMsg
	for range 5 {
		a = send(t, a, press("down"))
		old = append(old, previewMsg{a.previewSeq})
	}
	for _, msg := range old[:4] {
		a = send(t, a, msg)
	}
	if a.context.kind != kindLibrary {
		t.Fatalf("an outdated preview was shown: %v", a.context)
	}
	if a = send(t, a, old[4]); a.context.value != a.sides[panelArtists].Selected().Value {
		t.Fatalf("context = %v", a.context)
	}
}

// Behind a popup nothing keeps its color, so that the popup stands out.
func TestPopupFadesTheScreen(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	accent, _, _ := strings.Cut(lipgloss.NewStyle().Foreground(theme.Accent).Render("x"), "x")
	a := filledApp(t)
	normal := a.View()
	if !strings.Contains(normal, accent) {
		t.Fatal("no accent color without a popup")
	}
	for _, k := range []string{"/", "?"} {
		a := send(t, a, press(k))
		for i, l := range strings.Split(a.View(), "\n") {
			first, last := strings.Index(l, accent), strings.LastIndex(l, accent)
			if first < 0 {
				continue
			}
			// On a 140 cell screen, no popup goes beyond these columns.
			if ansi.StringWidth(l[:first]) < 28 || ansi.StringWidth(l[:last]) >= 112 {
				t.Fatalf("key %s: line %d keeps its color outside the popup", k, i)
			}
		}
		if a = send(t, a, press("esc")); a.View() != normal {
			t.Fatalf("key %s: closing the popup did not restore the screen", k)
		}
	}
}

func BenchmarkSearchFrame(b *testing.B) {
	b.Setenv("PATH", "")
	a := NewApp(0, 0, config.Config{}, false)
	defer a.Close()
	for _, msg := range []tea.Msg{tea.WindowSizeMsg{Width: 180, Height: 50}, press("/")} {
		m, _ := a.Update(msg)
		a = m.(App)
	}
	for b.Loop() {
		a.View()
	}
}

func BenchmarkVisualizerFrame(b *testing.B) {
	b.Setenv("PATH", "")
	a := NewApp(0, 0, config.Config{}, false)
	defer a.Close()
	m, _ := a.Update(tea.WindowSizeMsg{Width: 180, Height: 50})
	a = m.(App)
	for i := 0; b.Loop(); i++ {
		m, _ := a.Update(visualizerMsg(cava.Event{Levels: []byte{byte(i), 200, 90}}))
		a = m.(App)
		a.View()
	}
}
