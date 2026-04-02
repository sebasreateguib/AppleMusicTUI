package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"SR-Player/internal/applescript"
	"SR-Player/internal/models"
	"SR-Player/internal/ui/panels"
)

type tickMsg time.Time
type nowPlayingMsg models.NowPlaying
type artworkMsg string
type playlistsMsg []models.Playlist
type playlistTracksMsg struct {
	name   string
	offset int
	tracks []models.Track
}
type queueMsg []models.QueueTrack
type searchResultsMsg []models.Track

type focusedPanel int

const (
	focusSearchInput focusedPanel = iota
	focusPlaylists
	focusCenter
	focusRight
)

type CenterView int

const (
	CenterPlaylistTracks CenterView = iota
	CenterSearchResults
)

type App struct {
	playlistPanel   panels.PlaylistsPanel
	trackListPanel  panels.TrackListPanel
	nowPlayingPanel panels.NowPlayingPanel
	queuePanel      panels.QueuePanel
	searchPanel     panels.SearchPanel
	statusBar       StatusBar
	layout          Layout

	nowPlaying           models.NowPlaying
	lastTrackName        string
	rightView            RightView
	centerView           CenterView
	focus                focusedPanel
	tick                 int
	width                int
	height               int
	currentPlaylist      string
	currentPlaylistTotal int
	fetchingTracks       bool
}

func NewApp(width, height int) App {
	tmpLayout := NewLayout(width, height)
	leftW, centerW, rightW, mainH := tmpLayout.PanelSizes()

	app := App{
		playlistPanel:   panels.NewPlaylistsPanel(leftW, mainH),
		trackListPanel:  panels.NewTrackListPanel(centerW, mainH),
		nowPlayingPanel: panels.NewNowPlayingPanel(rightW, mainH-1),
		queuePanel:      panels.NewQueuePanel(rightW, mainH-1),
		searchPanel:     panels.NewSearchPanel(width, mainH),
		statusBar:       NewStatusBar(width),
		layout:          tmpLayout,
		rightView:       ViewNowPlaying,
		centerView:      CenterPlaylistTracks,
		focus:           focusPlaylists,
		width:           width,
		height:          height,
	}
	app.playlistPanel.SetFocused(true)
	return app
}

func (a App) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		fetchNowPlaying(),
		fetchPlaylists(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
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

func fetchPlaylistTracks(name string, offset, limit int) tea.Cmd {
	return func() tea.Msg {
		tracks, _ := applescript.GetPlaylistTracks(name, offset, limit)
		return playlistTracksMsg{name: name, offset: offset, tracks: tracks}
	}
}

func searchLibrary(query string) tea.Cmd {
	return func() tea.Msg {
		results, _ := applescript.SearchLibrary(query)
		return searchResultsMsg(results)
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

func doPlayPlaylist(name string) tea.Cmd {
	return func() tea.Msg {
		_ = applescript.PlayPlaylist(name)
		time.Sleep(500 * time.Millisecond)
		np, _ := applescript.GetNowPlaying()
		return nowPlayingMsg(np)
	}
}

func doPlayTrackInPlaylist(playlist string, index int) tea.Cmd {
	return func() tea.Msg {
		_ = applescript.PlayTrackInPlaylist(playlist, index)
		time.Sleep(400 * time.Millisecond)
		np, _ := applescript.GetNowPlaying()
		return nowPlayingMsg(np)
	}
}

func doPlayTrack(name string) tea.Cmd {
	return func() tea.Msg {
		_ = applescript.PlayTrackByName(name)
		time.Sleep(300 * time.Millisecond)
		np, _ := applescript.GetNowPlaying()
		return nowPlayingMsg(np)
	}
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.layout.SetSize(msg.Width, msg.Height)

		leftW, centerW, rightW, mainH := a.layout.PanelSizes()
		a.playlistPanel.SetSize(leftW, mainH)
		a.trackListPanel.SetSize(centerW, mainH)
		a.nowPlayingPanel.SetSize(rightW, mainH-1)
		a.queuePanel.SetSize(rightW, mainH-1)
		a.statusBar.SetWidth(msg.Width)

	case tickMsg:
		a.tick++
		cmds = append(cmds, tickCmd())

		if a.tick%6 == 0 {
			cmds = append(cmds, fetchNowPlaying())
			if a.rightView == ViewQueue {
				cmds = append(cmds, fetchQueue())
			}
		}

		if a.centerView == CenterPlaylistTracks {
			a.trackListPanel.SetCurrentTrack(a.nowPlaying.Track.Name)
		}

	case nowPlayingMsg:
		fresh := models.NowPlaying(msg)
		fresh.ArtworkPath = a.nowPlaying.ArtworkPath
		a.nowPlaying = fresh
		a.nowPlayingPanel.SetNowPlaying(a.nowPlaying)

		if a.nowPlaying.Track.Name != a.lastTrackName {
			a.lastTrackName = a.nowPlaying.Track.Name
			a.nowPlaying.ArtworkPath = ""
			a.nowPlayingPanel.SetNowPlaying(a.nowPlaying)
			cmds = append(cmds, fetchArtwork())
		}

	case artworkMsg:
		a.nowPlaying.ArtworkPath = string(msg)
		a.nowPlayingPanel.SetNowPlaying(a.nowPlaying)

	case playlistsMsg:
		a.playlistPanel.SetPlaylists([]models.Playlist(msg))

	case playlistTracksMsg:
		a.fetchingTracks = false
		if msg.name == a.currentPlaylist {
			if msg.offset == 0 {
				a.trackListPanel.SetTracks(msg.name, msg.tracks, a.nowPlaying.Track.Name)
			} else {
				cmd := a.trackListPanel.AppendTracks(msg.tracks)
				cmds = append(cmds, cmd)
			}
		}

	case queueMsg:
		a.queuePanel.SetTracks([]models.QueueTrack(msg))

	case searchResultsMsg:
		a.searchPanel.SetResults([]models.Track(msg))

	case tea.KeyMsg:
		if a.focus == focusSearchInput {
			switch msg.String() {
			case "esc":
				a.focus = focusPlaylists
				a.searchPanel.SetFocused(false)
				a.updateFocusStyles()
			case "enter":
				query := a.searchPanel.GetQuery()
				if query != "" {
					a.centerView = CenterSearchResults
					a.focus = focusCenter
					a.searchPanel.SetFocused(false)
					a.updateFocusStyles()
					cmds = append(cmds, searchLibrary(query))
				}
			default:
				var cmd tea.Cmd
				// Propagate to search input
				cmd, _ = a.searchPanel.Update(msg)
				cmds = append(cmds, cmd)
			}
			break
		}

		switch msg.String() {
		case "Q":
			return a, tea.Quit

		case "/":
			if a.focus != focusSearchInput {
				a.focus = focusSearchInput
				a.searchPanel.SetFocused(true)
				a.updateFocusStyles()
				return a, nil
			}

		case "1":
			a.rightView = ViewNowPlaying
			a.updateFocusStyles()

		case "2":
			a.rightView = ViewQueue
			a.updateFocusStyles()
			cmds = append(cmds, fetchQueue())

		case " ":
			cmds = append(cmds, doPlayPause())

		case "n":
			cmds = append(cmds, doNext())

		case "p":
			cmds = append(cmds, doPrev())

		case "s":
			cmds = append(cmds, doToggleShuffle())

		case "r":
			cmds = append(cmds, doToggleRepeat())

		case "left":
			if a.focus == focusCenter {
				a.focus = focusPlaylists
				a.updateFocusStyles()
			} else if a.focus == focusRight {
				a.focus = focusCenter
				a.updateFocusStyles()
			} else {
				cmds = append(cmds, doSeek(-10, a.nowPlaying.Position))
			}

		case "right":
			if a.focus == focusPlaylists {
				a.focus = focusCenter
				a.updateFocusStyles()
			} else if a.focus == focusCenter {
				a.focus = focusRight
				a.updateFocusStyles()
			} else {
				cmds = append(cmds, doSeek(10, a.nowPlaying.Position))
			}

		case "tab":
			switch a.focus {
			case focusSearchInput:
				a.searchPanel.SetFocused(false)
				a.focus = focusPlaylists
			case focusPlaylists:
				a.focus = focusCenter
			case focusCenter:
				a.focus = focusRight
			case focusRight:
				a.focus = focusSearchInput
				a.searchPanel.SetFocused(true)
			}
			a.updateFocusStyles()

		case "esc":
			if a.focus == focusRight || a.focus == focusCenter {
				a.focus = focusPlaylists
			}
			a.updateFocusStyles()

		case "enter":
			switch a.focus {
			case focusPlaylists:
				pl := a.playlistPanel.SelectedPlaylist()
				if pl != nil {
					a.focus = focusCenter
					a.centerView = CenterPlaylistTracks
					a.updateFocusStyles()

					a.currentPlaylist = pl.Name
					a.currentPlaylistTotal = pl.Count
					a.fetchingTracks = true
					a.trackListPanel.SetTracks(pl.Name, []models.Track{}, a.nowPlaying.Track.Name)
					cmds = append(cmds, fetchPlaylistTracks(pl.Name, 0, 50))
				}

			case focusCenter:
				if a.centerView == CenterPlaylistTracks {
					idx := a.trackListPanel.SelectedTrackIndex()
					playlist := a.trackListPanel.PlaylistName()
					if idx > 0 && playlist != "" {
						cmds = append(cmds, doPlayTrackInPlaylist(playlist, idx))
					}
				} else if a.centerView == CenterSearchResults {
					track := a.searchPanel.SelectedTrack()
					if track != nil {
						cmds = append(cmds, doPlayTrack(track.Name))
						a.rightView = ViewNowPlaying
						a.updateFocusStyles()
					}
				}
			}

		default:
			switch a.focus {
			case focusPlaylists:
				cmd := a.playlistPanel.Update(msg)
				cmds = append(cmds, cmd)
			case focusCenter:
				if a.centerView == CenterPlaylistTracks {
					cmd := a.trackListPanel.Update(msg)
					cmds = append(cmds, cmd)

					idx := a.trackListPanel.SelectedTrackIndex()
					loaded := len(a.trackListPanel.GetTracks())
					if idx >= loaded-10 && loaded < a.currentPlaylistTotal && !a.fetchingTracks {
						a.fetchingTracks = true
						cmds = append(cmds, fetchPlaylistTracks(a.currentPlaylist, loaded, 50))
					}
				} else if a.centerView == CenterSearchResults {
					// Route navigation to the search results
					cmd, _ := a.searchPanel.Update(msg)
					cmds = append(cmds, cmd)
				}
			case focusRight:
				switch a.rightView {
				case ViewQueue:
					cmd := a.queuePanel.Update(msg)
					cmds = append(cmds, cmd)
				}
			}
		}
	}

	return a, tea.Batch(cmds...)
}

func (a *App) updateFocusStyles() {
	a.playlistPanel.SetFocused(a.focus == focusPlaylists)
	a.trackListPanel.SetFocused(a.focus == focusCenter && a.centerView == CenterPlaylistTracks)
	
	// Para mantener los estilos coordinados
	a.nowPlayingPanel.SetFocused(a.focus == focusRight && a.rightView == ViewNowPlaying)
	a.queuePanel.SetFocused(a.focus == focusRight && a.rightView == ViewQueue)
}

func (a App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}

	topView := a.searchPanel.ViewInput(a.width, a.focus == focusSearchInput)
	leftView := a.playlistPanel.View()

	_, centerW, _, mainH := a.layout.PanelSizes()

	var centerView string
	if a.centerView == CenterPlaylistTracks {
		centerView = a.trackListPanel.View()
	} else {
		centerView = a.searchPanel.ViewResults(centerW, mainH, a.focus == focusCenter)
	}

	var rightView string
	switch a.rightView {
	case ViewNowPlaying:
		rightView = a.nowPlayingPanel.View(a.tick)
	case ViewQueue:
		rightView = a.queuePanel.View()
	}

	statusView := a.statusBar.View(a.nowPlaying)

	return a.layout.Render(
		topView,
		leftView,
		centerView,
		rightView,
		statusView,
		a.rightView,
	)
}
