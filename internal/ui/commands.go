package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sebasreateguib/applemusictui/internal/applescript"
	"github.com/sebasreateguib/applemusictui/internal/models"
)

// Music takes a moment to report what it was just told.
const (
	settleDelay = 300 * time.Millisecond // after changing a setting
	trackDelay  = 600 * time.Millisecond // after changing the track
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

func pollNowPlaying() tea.Cmd {
	return func() tea.Msg {
		np, _ := applescript.GetNowPlaying()
		return polledMsg(np)
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

func fetchQueue(seq int) tea.Cmd {
	return func() tea.Msg {
		tracks, _ := applescript.GetQueueTracks()
		return queueMsg{seq: seq, tracks: tracks}
	}
}

// fetchCurrentContext asks what Music is playing, to then set shuffle.
func fetchCurrentContext(shuffle bool) tea.Cmd {
	return func() tea.Msg {
		msg := currentContextMsg{shuffle: shuffle}
		kind, value, err := applescript.GetCurrentContext()
		if err != nil {
			return msg
		}
		tracks, err := applescript.GetTracks(kind, value)
		if err != nil {
			return msg
		}
		msg.ctx, msg.tracks = trackContext{kind, value}, tracks
		return msg
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

func doPlayPause() tea.Cmd {
	return func() tea.Msg {
		_ = applescript.PlayPause()
		time.Sleep(settleDelay)
		np, _ := applescript.GetNowPlaying()
		return nowPlayingMsg(np)
	}
}

func doNext() tea.Cmd {
	return func() tea.Msg {
		_ = applescript.NextTrack()
		time.Sleep(trackDelay)
		np, _ := applescript.GetNowPlaying()
		return nowPlayingMsg(np)
	}
}

func doPrev() tea.Cmd {
	return func() tea.Msg {
		_ = applescript.PreviousTrack()
		time.Sleep(trackDelay)
		np, _ := applescript.GetNowPlaying()
		return nowPlayingMsg(np)
	}
}

func doSetShuffle(enabled bool) tea.Cmd {
	return func() tea.Msg {
		_ = applescript.SetShuffle(enabled)
		time.Sleep(settleDelay)
		np, _ := applescript.GetNowPlaying()
		return nowPlayingMsg(np)
	}
}

func doToggleRepeat() tea.Cmd {
	return func() tea.Msg {
		_ = applescript.ToggleRepeat()
		time.Sleep(settleDelay)
		np, _ := applescript.GetNowPlaying()
		return nowPlayingMsg(np)
	}
}

func doSeek(position float64) tea.Cmd {
	return func() tea.Msg {
		_ = applescript.SetPlayerPosition(position)
		time.Sleep(settleDelay)
		np, _ := applescript.GetNowPlaying()
		return seekedMsg(np)
	}
}

func doPlayTrackInContext(cType, cVal string, index int, shuffleAfter bool) tea.Cmd {
	return func() tea.Msg {
		_ = applescript.PlayTrackInContextWithShuffle(cType, cVal, index, shuffleAfter)
		time.Sleep(trackDelay)
		np, _ := applescript.GetNowPlaying()
		return nowPlayingMsg(np)
	}
}

func doPlayLibraryTrack(track models.Track) tea.Cmd {
	return func() tea.Msg {
		_ = applescript.PlayTrackInLibrary(track)
		time.Sleep(trackDelay)
		np, _ := applescript.GetNowPlaying()
		return nowPlayingMsg(np)
	}
}

func doSetVolume(volume int) tea.Cmd {
	return func() tea.Msg {
		_ = applescript.SetVolume(volume) // reported right away
		np, _ := applescript.GetNowPlaying()
		return volumeMsg{sent: volume, now: np}
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
