package ui

import (
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (a *App) setPlaybackContext(contextType, contextValue string, index, count int) {
	a.playbackContextType = contextType
	a.playbackContextValue = contextValue
	a.playbackTrackIndex = index
	if count > 0 {
		a.playbackContextCount = count
	} else {
		a.playbackContextCount = 0
	}
	if a.playbackShuffleOn {
		a.rebuildPlaybackOrder(index)
	} else {
		a.playbackOrder = nil
		a.playbackOrderPos = 0
	}
}

func (a App) hasContextualPlayback() bool {
	return a.playbackContextType != "" && a.playbackTrackIndex > 0
}

func (a App) queueFetchCmd() tea.Cmd {
	if a.hasContextualPlayback() {
		return fetchContextQueue(a.playbackContextType, a.playbackContextValue, a.playbackTrackIndex)
	}
	return fetchQueue()
}

func (a App) nextCmd() tea.Cmd {
	if a.hasContextualPlayback() {
		if a.playbackShuffleOn && len(a.playbackOrder) > 0 {
			nextPos := a.playbackOrderPos + 1
			if nextPos >= len(a.playbackOrder) {
				nextPos = len(a.playbackOrder) - 1
			}
			if nextPos >= 0 && nextPos < len(a.playbackOrder) {
				return doPlayTrackInContext(a.playbackContextType, a.playbackContextValue, a.playbackOrder[nextPos], true)
			}
		}
		return doPlayTrackInContext(a.playbackContextType, a.playbackContextValue, a.playbackTrackIndex+1, false)
	}
	return doNext()
}

func (a App) prevCmd() tea.Cmd {
	if a.hasContextualPlayback() {
		if a.playbackShuffleOn && len(a.playbackOrder) > 0 {
			targetPos := a.playbackOrderPos - 1
			if targetPos < 0 {
				targetPos = 0
			}
			if targetPos < len(a.playbackOrder) {
				return doPlayTrackInContext(a.playbackContextType, a.playbackContextValue, a.playbackOrder[targetPos], true)
			}
		}
		targetIndex := a.playbackTrackIndex - 1
		if targetIndex < 1 {
			targetIndex = 1
		}
		return doPlayTrackInContext(a.playbackContextType, a.playbackContextValue, targetIndex, false)
	}
	return doPrev()
}

func (a *App) rebuildPlaybackOrder(currentIndex int) {
	if !a.playbackShuffleOn || a.playbackContextCount < 1 || currentIndex < 1 {
		a.playbackOrder = nil
		a.playbackOrderPos = 0
		return
	}

	order := make([]int, 0, a.playbackContextCount)
	order = append(order, currentIndex)
	for i := 1; i <= a.playbackContextCount; i++ {
		if i != currentIndex {
			order = append(order, i)
		}
	}

	if len(order) > 1 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		r.Shuffle(len(order)-1, func(i, j int) {
			i++
			j++
			order[i], order[j] = order[j], order[i]
		})
	}

	a.playbackOrder = order
	a.playbackOrderPos = 0
}

func (a *App) syncPlaybackPosition(index int) {
	a.playbackTrackIndex = index
	if !a.playbackShuffleOn {
		return
	}
	for pos, trackIndex := range a.playbackOrder {
		if trackIndex == index {
			a.playbackOrderPos = pos
			return
		}
	}
	a.rebuildPlaybackOrder(index)
}

func (a *App) toggleContextShuffle() tea.Cmd {
	a.playbackShuffleOn = !a.playbackShuffleOn
	if a.playbackShuffleOn {
		a.rebuildPlaybackOrder(a.playbackTrackIndex)
	} else {
		a.playbackOrder = nil
		a.playbackOrderPos = 0
	}
	a.nowPlaying.ShuffleEnabled = a.playbackShuffleOn
	a.nowPlayingPanel.SetNowPlaying(a.nowPlaying)

	cmds := []tea.Cmd{doSetShuffle(a.playbackShuffleOn)}
	if a.playbackShuffleOn && a.playbackContextCount == 0 {
		cmds = append(cmds, fetchContextCount(a.playbackContextType, a.playbackContextValue))
	}
	cmds = append(cmds, a.queueFetchCmd())
	return tea.Batch(cmds...)
}
