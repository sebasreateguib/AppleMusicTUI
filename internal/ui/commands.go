package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"SR-Player/internal/applescript"
	"SR-Player/internal/models"
)

func tickCmd() tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (a App) waitVisualizer() tea.Cmd {
	events := a.visualizer.Events()
	return func() tea.Msg {
		return visualizerMsg(<-events)
	}
}

func fetchNowPlaying() tea.Cmd {
	return func() tea.Msg {
		np, _ := applescript.GetNowPlaying()
		return nowPlayingMsg(np)
	}
}

func fetchPlaylists() tea.Cmd {
	return func() tea.Msg {
		pl, _ := applescript.GetPlaylists()
		return playlistsMsg(pl)
	}
}

func fetchArtwork() tea.Cmd {
	return func() tea.Msg {
		path := applescript.GetArtworkPath()
		return artworkMsg(path)
	}
}

func fetchQueue() tea.Cmd {
	return func() tea.Msg {
		tracks, _ := applescript.GetQueueTracks()
		return queueMsg(tracks)
	}
}

func fetchContextQueue(cType, cVal string, currentIndex int) tea.Cmd {
	return func() tea.Msg {
		tracks, _ := applescript.GetQueueTracksForContext(cType, cVal, currentIndex, 10)
		return queueMsg(tracks)
	}
}

func fetchUniqueStrings(kind string) tea.Cmd {
	return func() tea.Msg {
		var items []string
		if kind == "artists" {
			items, _ = applescript.GetUniqueArtists()
		} else if kind == "albums" {
			items, _ = applescript.GetUniqueAlbums()
		}
		return uniqueStringsMsg{kind: kind, items: items}
	}
}

func fetchContextCount(cType, cVal string) tea.Cmd {
	return func() tea.Msg {
		count, _ := applescript.GetContextTrackCount(cType, cVal)
		return contextCountMsg{contextType: cType, contextValue: cVal, count: count}
	}
}

func fetchPlaybackIndex(cType, cVal string, track models.Track) tea.Cmd {
	return func() tea.Msg {
		idx, _ := applescript.FindTrackIndexInContext(cType, cVal, track)
		return playbackIndexMsg{
			contextType:  cType,
			contextValue: cVal,
			index:        idx,
		}
	}
}

func doPlayPause() tea.Cmd {
	return func() tea.Msg {
		_ = applescript.PlayPause()
		np, _ := applescript.GetNowPlaying()
		return nowPlayingMsg(np)
	}
}

func doNext() tea.Cmd {
	return func() tea.Msg {
		_ = applescript.NextTrack()
		time.Sleep(300 * time.Millisecond)
		np, _ := applescript.GetNowPlaying()
		return nowPlayingMsg(np)
	}
}

func doPrev() tea.Cmd {
	return func() tea.Msg {
		_ = applescript.PreviousTrack()
		time.Sleep(300 * time.Millisecond)
		np, _ := applescript.GetNowPlaying()
		return nowPlayingMsg(np)
	}
}

func doToggleShuffle() tea.Cmd {
	return func() tea.Msg {
		_ = applescript.ToggleShuffle()
		np, _ := applescript.GetNowPlaying()
		return nowPlayingMsg(np)
	}
}

func doSetShuffle(enabled bool) tea.Cmd {
	return func() tea.Msg {
		_ = applescript.SetShuffle(enabled)
		np, _ := applescript.GetNowPlaying()
		return nowPlayingMsg(np)
	}
}

func doToggleRepeat() tea.Cmd {
	return func() tea.Msg {
		_ = applescript.ToggleRepeat()
		np, _ := applescript.GetNowPlaying()
		return nowPlayingMsg(np)
	}
}

func doSeek(delta float64, current float64) tea.Cmd {
	return func() tea.Msg {
		pos := current + delta
		if pos < 0 {
			pos = 0
		}
		_ = applescript.SetPlayerPosition(pos)
		np, _ := applescript.GetNowPlaying()
		return nowPlayingMsg(np)
	}
}

func doPlayTrackInContext(cType, cVal string, index int, shuffleAfter bool) tea.Cmd {
	return func() tea.Msg {
		_ = applescript.PlayTrackInContextWithShuffle(cType, cVal, index, shuffleAfter)
		time.Sleep(400 * time.Millisecond)
		np, _ := applescript.GetNowPlaying()
		return nowPlayingMsg(np)
	}
}

func doPlayLibraryTrack(track models.Track) tea.Cmd {
	return func() tea.Msg {
		_ = applescript.PlayTrackInLibrary(track)
		time.Sleep(300 * time.Millisecond)
		np, _ := applescript.GetNowPlaying()
		return nowPlayingMsg(np)
	}
}

func doChangeVolume(delta int, current int) tea.Cmd {
	return func() tea.Msg {
		newVol := current + delta
		if newVol < 0 {
			newVol = 0
		}
		if newVol > 100 {
			newVol = 100
		}
		_ = applescript.SetVolume(newVol)
		np, _ := applescript.GetNowPlaying()
		return nowPlayingMsg(np)
	}
}

func fetchTracks(ctx trackContext) tea.Cmd {
	return func() tea.Msg {
		tracks, err := applescript.GetTracks(ctx.kind, ctx.value)
		return tracksMsg{ctx: ctx, tracks: tracks, err: err}
	}
}

func searchLibrary(query string) tea.Cmd {
	return func() tea.Msg {
		tracks, err := applescript.SearchLibrary(query)
		return tracksMsg{ctx: trackContext{kind: kindSearch, value: query}, tracks: tracks, err: err}
	}
}
