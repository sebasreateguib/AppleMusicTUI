package ui

import (
	"math/rand"
	"slices"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sebasreateguib/applemusictui/internal/models"
)

const (
	// queueAhead is how many upcoming tracks Up Next lists.
	queueAhead = 30

	// advanceBefore is how long before the end of a track the next one is
	// started. It covers the time Music takes to report the position and to
	// obey; otherwise Music moves on by itself, to a track of its own choice.
	advanceBefore = 800 * time.Millisecond

	// endMargin is how close to its end a track has to be for a change to count
	// as Music moving on by itself. It is the longest crossfade Music offers.
	endMargin = 12 * time.Second

	// stalePosition is the age from which a reported position is not trusted.
	stalePosition = 3 * time.Second
)

// Music does not tell scripts what it will play next, and what follows a track
// started by a script is often one of Autoplay's suggestions. Up Next therefore
// comes from the playback context: the list the current track was started from,
// with the app's own order when shuffled. The app starts each track of it.
// Without a context, the tracks that follow in Music's playlist are shown.

// setPlaybackContext makes ctx the list that is playing, from the 1-based index
// (0 if it is not known yet).
func (a *App) setPlaybackContext(ctx trackContext, index int) {
	if !a.hasContext() {
		// Until now Music was in charge: keep its setting.
		a.playbackShuffleOn = a.nowPlaying.ShuffleEnabled
	}
	a.playback = ctx
	a.playbackTrackIndex = index
	a.playbackOrder = nil
	a.expected = models.Track{}
	a.rebuildPlaybackOrder()
}

func (a App) hasContext() bool { return a.playback.kind != "" }

// hasContextualPlayback reports whether the current track is in the context.
func (a App) hasContextualPlayback() bool {
	return a.hasContext() && a.playbackTrackIndex > 0
}

func (a App) playbackTracks() []models.Track { return a.trackCache[a.playback] }

// rebuildPlaybackOrder shuffles the context, starting with the current track.
func (a *App) rebuildPlaybackOrder() {
	a.playbackOrder = nil
	a.playbackOrderPos = 0
	count := len(a.playbackTracks())
	if !a.playbackShuffleOn || a.playbackTrackIndex < 1 || a.playbackTrackIndex > count {
		return
	}
	order := []int{a.playbackTrackIndex}
	for _, i := range rand.Perm(count) {
		if i+1 != a.playbackTrackIndex {
			order = append(order, i+1)
		}
	}
	a.playbackOrder = order
}

// syncPlaybackPosition moves the context to the track with that index.
func (a *App) syncPlaybackPosition(index int) {
	a.playbackTrackIndex = index
	// Outside of the list, keep the order to come back to it.
	if !a.playbackShuffleOn || index < 1 {
		return
	}
	if pos := slices.Index(a.playbackOrder, index); pos >= 0 {
		a.playbackOrderPos = pos
		return
	}
	a.rebuildPlaybackOrder()
}

// step returns the index of the track delta positions away, 0 if there is none.
func (a App) step(delta int) int {
	if a.playbackShuffleOn && len(a.playbackOrder) > 0 {
		if pos := a.playbackOrderPos + delta; pos >= 0 && pos < len(a.playbackOrder) {
			return a.playbackOrder[pos]
		}
		return 0
	}
	if index := a.playbackTrackIndex + delta; index >= 1 && index <= len(a.playbackTracks()) {
		return index
	}
	return 0
}

// playIndex plays a track of the context.
func (a *App) playIndex(index int) tea.Cmd {
	a.syncPlaybackPosition(index)
	if tracks := a.playbackTracks(); index <= len(tracks) {
		a.expected = tracks[index-1]
	}
	return tea.Batch(
		doPlayTrackInContext(a.playback.kind, a.playback.value, index, a.playbackShuffleOn),
		a.refreshQueue(),
	)
}

func (a *App) nextCmd() tea.Cmd {
	if !a.hasContextualPlayback() {
		return doNext()
	}
	if index := a.step(1); index > 0 {
		return a.playIndex(index)
	}
	a.setNotice("End of the list")
	return nil
}

func (a *App) prevCmd() tea.Cmd {
	if !a.hasContextualPlayback() {
		return doPrev()
	}
	if index := a.step(-1); index > 0 {
		return a.playIndex(index)
	}
	return a.playIndex(a.playbackTrackIndex) // start over
}

// endsWithin reports whether the current track, if the app started it, is at
// most margin away from its end.
func (a App) endsWithin(margin time.Duration, now time.Time) bool {
	np := a.nowPlaying
	if !a.hasContextualPlayback() || !a.isPlaying() ||
		np.RepeatMode == models.RepeatOne || np.Track.Duration <= 0 {
		return false
	}
	// Leave alone what the user plays from Music itself.
	if a.expected.Name == "" || !np.Track.Same(a.expected) {
		return false
	}
	age := now.Sub(a.nowPlayingAt)
	left := time.Duration((np.Track.Duration-np.Position)*float64(time.Second)) - age
	return age <= stalePosition && left <= margin
}

// following returns the index of the track to play when the current one ends,
// 0 at the end of the list.
func (a App) following() int {
	if index := a.step(1); index > 0 {
		return index
	}
	if a.nowPlaying.RepeatMode != models.RepeatAll || len(a.playbackTracks()) < 2 {
		return 0
	}
	if a.playbackShuffleOn && len(a.playbackOrder) > 0 {
		return a.playbackOrder[0]
	}
	return 1
}

// advanceCmd starts the next track when the current one is about to end. It
// returns nil the rest of the time.
func (a *App) advanceCmd(now time.Time) tea.Cmd {
	// The position of the previous track can still arrive after advancing, and
	// a list can have the same track twice in a row.
	if now.Sub(a.advancedAt) < stalePosition || !a.endsWithin(advanceBefore, now) {
		return nil
	}
	index := a.following()
	if index == 0 {
		a.expected = models.Track{}
		return nil
	}
	a.advancedAt = now
	return a.playIndex(index)
}

// takeBackCmd is for when Music moved on from the current track to another one
// before the app did, because of a crossfade or a slow reply. It returns nil if
// the app does not have to correct it.
func (a *App) takeBackCmd(fresh models.NowPlaying, now time.Time) tea.Cmd {
	margin := endMargin
	if fresh.Track.Name == "" { // stopped, or it did not answer
		margin = advanceBefore
	}
	if fresh.State == models.StatePaused || !a.endsWithin(margin, now) {
		return nil
	}
	index := a.following()
	if index == 0 || index > len(a.playbackTracks()) {
		return nil
	}
	if next := a.playbackTracks()[index-1]; fresh.Track.Same(next) {
		a.expected = next // Music agrees
		return nil
	}
	a.advancedAt = now
	return a.playIndex(index)
}

// refreshQueue updates Up Next, with a command if something has to be fetched.
func (a *App) refreshQueue() tea.Cmd {
	a.queueSeq++ // whatever is being fetched is outdated
	if !a.hasContextualPlayback() {
		return fetchQueue(a.queueSeq)
	}
	tracks := a.playbackTracks()
	if tracks == nil {
		return fetchTracks(a.playback) // see receiveTracks
	}

	var queue []models.QueueTrack
	for d := 0; d <= queueAhead; d++ {
		index := a.step(d)
		if index < 1 || index > len(tracks) {
			break
		}
		queue = append(queue, models.QueueTrack{Index: index, Track: tracks[index-1], IsCurrent: d == 0})
	}
	a.queue.SetTracks(queue, "")
	return nil
}

// receiveQueue shows what follows in Music's playlist.
func (a *App) receiveQueue(msg queueMsg) {
	if msg.seq != a.queueSeq || a.hasContextualPlayback() {
		return
	}
	if a.nowPlaying.ShuffleEnabled && len(msg.tracks) > 1 {
		a.queue.SetTracks(msg.tracks[:1], "Shuffled by Music, which does not share its order")
		return
	}
	a.queue.SetTracks(msg.tracks, "")
}

// locatePlaying finds the current track in the context and updates Up Next.
func (a *App) locatePlaying() tea.Cmd {
	tracks, ok := a.trackCache[a.playback]
	if !ok {
		return fetchTracks(a.playback) // see receiveTracks
	}
	track := a.nowPlaying.Track
	index := a.playbackTrackIndex
	// A track can be more than once in a playlist: prefer where we are.
	if index < 1 || index > len(tracks) || !tracks[index-1].Same(track) {
		index = slices.IndexFunc(tracks, track.Same) + 1
	}
	a.syncPlaybackPosition(index)
	return a.refreshQueue()
}

// toggleShuffle turns shuffle on or off. Without a context it first adopts the
// playlist Music is playing, so that the order is known.
func (a *App) toggleShuffle() tea.Cmd {
	on := !a.nowPlaying.ShuffleEnabled
	if a.hasContextualPlayback() {
		return a.setShuffle(on)
	}
	return fetchCurrentContext(on)
}

func (a *App) setShuffle(on bool) tea.Cmd {
	a.playbackShuffleOn = on
	a.nowPlaying.ShuffleEnabled = on
	a.nowPlayingPanel.SetNowPlaying(a.nowPlaying)
	a.expected = a.nowPlaying.Track
	a.rebuildPlaybackOrder()
	return tea.Batch(doSetShuffle(on), a.refreshQueue())
}

// receiveCurrentContext finishes toggleShuffle.
func (a *App) receiveCurrentContext(msg currentContextMsg) tea.Cmd {
	if a.hasContextualPlayback() { // something was played meanwhile
		return a.setShuffle(msg.shuffle)
	}
	index := slices.IndexFunc(msg.tracks, a.nowPlaying.Track.Same) + 1
	if index == 0 {
		a.nowPlaying.ShuffleEnabled = msg.shuffle
		a.playback = trackContext{}
		return tea.Batch(doSetShuffle(msg.shuffle), a.refreshQueue())
	}
	a.trackCache[msg.ctx] = msg.tracks
	a.setPlaybackContext(msg.ctx, index)
	return a.setShuffle(msg.shuffle)
}
