package ui

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/sebasreateguib/applemusictui/internal/models"
)

var mix = trackContext{kindPlaylist, " Mix"}

func mixTracks() []models.Track {
	var tracks []models.Track
	for i := range 40 {
		tracks = append(tracks, models.Track{Name: fmt.Sprintf("Song%02d", i+1), Artist: "A", Duration: 200})
	}
	return tracks
}

func playing(track models.Track, position float64, shuffle bool) nowPlayingMsg {
	return nowPlayingMsg{Track: track, State: models.StatePlaying, Position: position, ShuffleEnabled: shuffle}
}

// upNext returns the names on the Up Next panel.
func upNext(a App) []string {
	var names []string
	for _, l := range strings.Split(ansi.Strip(a.queue.View()), "\n") {
		if _, rest, ok := strings.Cut(l, "Song"); ok {
			names = append(names, "Song"+rest[:2])
		}
	}
	return names
}

// playingMix is playing track 10 of the playlist, started from the app.
func playingMix(t *testing.T) App {
	a := send(t, newTestApp(t), tea.WindowSizeMsg{Width: 140, Height: 60},
		playlistsMsg{{Name: mix.value, Count: 40}}, press("1"), press("down"))
	a = send(t, a, previewMsg{a.previewSeq}, tracksMsg{ctx: mix, tracks: mixTracks()}, press("4"))
	for range 9 {
		a = send(t, a, press("down"))
	}
	return send(t, a, press("enter"), playing(mixTracks()[9], 5, false))
}

func isSequential(names []string) bool { return slices.IsSorted(names) }

func TestShuffleUpdatesUpNext(t *testing.T) {
	a := playingMix(t)
	if got := upNext(a); len(got) < 5 || got[0] != "Song10" || got[1] != "Song11" || !isSequential(got) {
		t.Fatalf("not shuffled: %v", got)
	}

	a = send(t, a, press("s"))
	first := upNext(a)
	if first[0] != "Song10" || isSequential(first) {
		t.Fatalf("shuffled: %v", first)
	}
	// What is announced is what plays.
	a = send(t, a, press("n"))
	if got := upNext(a); got[0] != first[1] || got[1] != first[2] {
		t.Fatalf("after next: %v, want it to follow %v", got, first)
	}
	a = send(t, a, press("p"))
	if got := upNext(a); !slices.Equal(got, first) {
		t.Fatalf("after previous: %v, want %v", got, first)
	}

	a = send(t, a, press("s"))
	if got := upNext(a); got[0] != "Song10" || got[1] != "Song11" || !isSequential(got) {
		t.Fatalf("shuffle off: %v", got)
	}
	a = send(t, a, press("s"))
	if got := upNext(a); slices.Equal(got, first) {
		t.Fatal("the same order again")
	}
	if n := len(a.playbackOrder); n != 40 || len(slices.Compact(slices.Sorted(slices.Values(a.playbackOrder)))) != 40 {
		t.Fatalf("the order is not a permutation: %v", a.playbackOrder)
	}
}

func TestRefreshKeepsUpNext(t *testing.T) {
	a := send(t, playingMix(t), press("s"))
	want := upNext(a)
	a = send(t, a, press("R"), press("1"))
	if a = send(t, a, tracksMsg{ctx: mix, tracks: mixTracks()}); !slices.Equal(upNext(a), want) {
		t.Fatalf("an unchanged list was shuffled again: %v, want %v", upNext(a), want)
	}
	a = send(t, a, tracksMsg{ctx: mix, tracks: mixTracks()[:20]})
	if len(a.playbackOrder) != 20 || upNext(a)[0] != "Song10" {
		t.Fatalf("order = %v", a.playbackOrder)
	}
}

// Started from the Music app: shuffling adopts its playlist.
func TestShuffleAdoptsMusicsPlaylist(t *testing.T) {
	a := send(t, newTestApp(t), tea.WindowSizeMsg{Width: 140, Height: 60}, playing(mixTracks()[4], 5, false), press("s"),
		currentContextMsg{ctx: mix, tracks: mixTracks(), shuffle: true})
	if got := upNext(a); got[0] != "Song05" || isSequential(got) || !a.nowPlaying.ShuffleEnabled {
		t.Fatalf("got %v", got)
	}
	// A reply from before must not replace it.
	a = send(t, a, queueMsg{seq: 1, tracks: []models.QueueTrack{{Track: mixTracks()[4], IsCurrent: true}, {Track: mixTracks()[5]}}})
	if got := upNext(a); isSequential(got) {
		t.Fatalf("an old queue was shown: %v", got)
	}
}

// Without a playlist the order is unknown, and not made up.
func TestShuffleWithoutPlaylist(t *testing.T) {
	a := send(t, newTestApp(t), tea.WindowSizeMsg{Width: 140, Height: 60}, playing(mixTracks()[4], 5, false), press("s"),
		currentContextMsg{shuffle: true})
	a = send(t, a, queueMsg{seq: a.queueSeq, tracks: []models.QueueTrack{{Track: mixTracks()[4], IsCurrent: true}, {Track: mixTracks()[5]}}})
	if got := upNext(a); !slices.Equal(got, []string{"Song05"}) || !strings.Contains(ansi.Strip(a.queue.View()), "Shuffled by Music") {
		t.Fatalf("got %v", got)
	}
}

// Playing something used to turn off a shuffle set in Music.
func TestPlayingKeepsMusicsShuffle(t *testing.T) {
	a := send(t, newTestApp(t), tea.WindowSizeMsg{Width: 140, Height: 60}, playing(models.Track{Name: "Other"}, 5, true),
		tracksMsg{ctx: trackContext{kind: kindLibrary}, tracks: mixTracks()}, press("4"), press("enter"))
	if !a.playbackShuffleOn || isSequential(upNext(a)) {
		t.Fatalf("shuffle = %v, up next = %v", a.playbackShuffleOn, upNext(a))
	}
}

func TestAdvanceFollowsUpNext(t *testing.T) {
	a := playingMix(t)
	at := func(a App, left time.Duration) (App, bool) {
		cmd := a.advanceCmd(a.nowPlayingAt.Add(195*time.Second - left))
		return a, cmd != nil
	}
	near := func(a App) App { // poll just before the end
		return send(t, a, playing(a.nowPlaying.Track, 199.5, false))
	}

	b := near(a)
	if b.advanceCmd(b.nowPlayingAt) == nil || upNext(b)[0] != "Song11" {
		t.Fatalf("not shuffled: did not advance, up next = %v", upNext(b))
	}

	a = send(t, a, press("s"))
	want := upNext(a)[1]
	if _, ok := at(a, 2*time.Second); ok {
		t.Fatal("advanced too early")
	}
	if _, ok := at(a, 500*time.Millisecond); ok {
		t.Fatal("trusted a position from minutes ago")
	}
	a = near(a)
	if a.advanceCmd(a.nowPlayingAt) == nil {
		t.Fatal("did not advance")
	}
	if got := upNext(a)[0]; got != want {
		t.Fatalf("plays %s, Up Next said %s", got, want)
	}
	if a.advanceCmd(a.nowPlayingAt) != nil {
		t.Fatal("skipped a track: advanced twice")
	}

	// Another track chosen in Music is left alone, even if it is in the list.
	b = send(t, playingMix(t), press("s"), playing(mixTracks()[30], 199.5, true))
	if b.advanceCmd(b.nowPlayingAt) != nil {
		t.Fatal("took over what the user plays from Music")
	}
	// So are a paused track and repeat one.
	for _, change := range []func(*models.NowPlaying){
		func(np *models.NowPlaying) { np.State = models.StatePaused },
		func(np *models.NowPlaying) { np.RepeatMode = models.RepeatOne },
	} {
		np := models.NowPlaying(playing(mixTracks()[9], 199.5, true))
		change(&np)
		b = send(t, playingMix(t), press("s"), nowPlayingMsg(np))
		if b.advanceCmd(b.nowPlayingAt) != nil {
			t.Fatalf("advanced with %+v", np)
		}
	}
}

func TestAdvanceAtTheEndOfTheList(t *testing.T) {
	for repeat, want := range map[models.RepeatMode]string{models.RepeatOff: "", models.RepeatAll: "Song01"} {
		a := playingMix(t)
		a.playIndex(40)
		np := models.NowPlaying(playing(mixTracks()[39], 199.5, false))
		np.RepeatMode = repeat
		a = send(t, a, nowPlayingMsg(np))
		if cmd := a.advanceCmd(a.nowPlayingAt); (cmd != nil) != (want != "") {
			t.Fatalf("repeat %s: advanced = %v", repeat, cmd != nil)
		}
		if want != "" && upNext(a)[0] != want {
			t.Fatalf("repeat %s: plays %v", repeat, upNext(a))
		}
	}
}

// The same track twice in a row: its old position is not the new one's.
func TestAdvanceOnceOnRepeatedTracks(t *testing.T) {
	tracks := mixTracks()
	tracks[10] = tracks[9]
	a := send(t, playingMix(t), tracksMsg{ctx: mix, tracks: tracks}, playing(tracks[9], 199.5, false))
	if a.advanceCmd(a.nowPlayingAt) == nil {
		t.Fatal("did not advance")
	}
	a = send(t, a, playing(tracks[9], 199.6, false))
	if a.advanceCmd(a.nowPlayingAt) != nil {
		t.Fatal("skipped a track")
	}
}

// Music got ahead of the app, to a track of Autoplay's.
func TestTakesBackFromAutoplay(t *testing.T) {
	autoplay := models.Track{Name: "Suggested", Artist: "B", Duration: 180}
	for _, shuffle := range []bool{false, true} {
		a := playingMix(t)
		if shuffle {
			a = send(t, a, press("s"))
		}
		want := upNext(a)[1]
		a = send(t, a, playing(mixTracks()[9], 195, shuffle), playing(autoplay, 0.2, shuffle))
		if got := upNext(a)[0]; got != want || a.expected.Name != want {
			t.Fatalf("shuffle %v: plays %s, want %s", shuffle, got, want)
		}
	}

	// Music's own choice is the right one, so the app keeps following it.
	a := send(t, playingMix(t), playing(mixTracks()[9], 195, false), playing(mixTracks()[10], 0.2, false),
		playing(mixTracks()[10], 199.5, false))
	if a.advanceCmd(a.nowPlayingAt) == nil || upNext(a)[0] != "Song12" {
		t.Fatalf("stopped following: %v", upNext(a))
	}

	// In the middle of a track it is the user who chose another one, and one
	// that is not known could be a reply that failed.
	for _, change := range []struct {
		position float64
		to       models.Track
	}{{100, autoplay}, {195, models.Track{}}} {
		a = send(t, playingMix(t), playing(mixTracks()[9], change.position, false), playing(change.to, 0.2, false))
		if a.expected.Name != "Song10" {
			t.Fatalf("took over at %v from %q", change.position, change.to.Name)
		}
	}
}

// A late reply from Autoplay must not shuffle the list again.
func TestLeavingTheListKeepsItsOrder(t *testing.T) {
	a := send(t, playingMix(t), press("s"))
	want := upNext(a)
	a = send(t, a, playing(models.Track{Name: "Suggested", Duration: 180}, 1, true), playing(mixTracks()[9], 1, true))
	if got := upNext(a); !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
