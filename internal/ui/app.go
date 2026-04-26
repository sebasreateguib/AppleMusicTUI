package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"SR-Player/internal/applescript"
	"SR-Player/internal/models"
	"SR-Player/internal/ui/panels"
)

type tickMsg time.Time
type nowPlayingMsg models.NowPlaying
type artworkMsg string
type playlistsMsg []models.Playlist

type contextTracksMsg struct {
	contextType  string
	contextValue string
	offset       int
	tracks       []models.Track
}

type uniqueStringsMsg struct {
	kind  string // "artists" or "albums"
	items []string
}

type queueMsg []models.QueueTrack
type searchResultsMsg []models.Track

type focusedPanel int

const (
	focusSearchInput focusedPanel = iota
	focusCenter
	focusRight
)

type CenterView int

const (
	CenterContextTracks CenterView = iota
	CenterSearchResults
	CenterStringList
)

type TopTab int

const (
	TabAlbums TopTab = iota
	TabArtists
	TabSongs
	TabPlaylists
	TabSearch
)

type App struct {
	trackListPanel  panels.TrackListPanel
	stringListPanel panels.StringListPanel
	nowPlayingPanel panels.NowPlayingPanel
	queuePanel      panels.QueuePanel
	searchPanel     panels.SearchPanel
	statusBar       StatusBar
	layout          Layout

	nowPlaying             models.NowPlaying
	lastTrackName          string
	rightView              RightView
	centerView             CenterView
	focus                  focusedPanel
	activeTopTab           TopTab
	playlists              []models.Playlist
	tick                   int
	width                  int
	height                 int
	centerTracksHeight     int
	centerVisualizerHeight int

	currentContextType  string
	currentContextValue string
	currentContextTotal int
	fetchingTracks      bool
}

func NewApp(width, height int) App {
	tmpLayout := NewLayout(width, height)
	centerW, rightW, mainH := tmpLayout.PanelSizes()
	centerTracksH, centerVisH := splitCenterHeights(mainH)

	app := App{
		trackListPanel:  panels.NewTrackListPanel(centerW, centerTracksH),
		stringListPanel: panels.NewStringListPanel(centerW, mainH, "Library"),
		nowPlayingPanel: panels.NewNowPlayingPanel(rightW, mainH),
		queuePanel:      panels.NewQueuePanel(rightW, mainH),
		searchPanel:     panels.NewSearchPanel(width, mainH),
		statusBar:       NewStatusBar(width),
		layout:          tmpLayout,
		rightView:       ViewNowPlaying,
		centerView:      CenterContextTracks,
		focus:           focusCenter,
		activeTopTab:    TabSongs,
		width:           width,
		height:          height,

		centerTracksHeight:     centerTracksH,
		centerVisualizerHeight: centerVisH,
		currentContextType:     "library",
		currentContextValue:    "",
		currentContextTotal:    999999,
	}
	app.updateFocusStyles()
	return app
}

func splitCenterHeights(mainH int) (tracksH, visH int) {
	if mainH <= 11 {
		return mainH, 0
	}

	visH = mainH / 3
	if visH < 6 {
		visH = 6
	}
	if visH > 14 {
		visH = 14
	}

	tracksH = mainH - visH
	if tracksH < 6 {
		tracksH = 6
		visH = mainH - tracksH
	}
	if visH < 4 {
		visH = 4
		tracksH = mainH - visH
	}
	if tracksH < 1 {
		tracksH = mainH
		visH = 0
	}

	return
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

func fetchContextTracks(cType, cVal string, offset, limit int) tea.Cmd {
	return func() tea.Msg {
		tracks, _ := applescript.GetFilteredTracks(cType, cVal, offset, limit)
		return contextTracksMsg{contextType: cType, contextValue: cVal, offset: offset, tracks: tracks}
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

func doPlayTrackInContext(cType, cVal string, index int) tea.Cmd {
	return func() tea.Msg {
		_ = applescript.PlayTrackInContext(cType, cVal, index)
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

func (a App) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		fetchNowPlaying(),
		fetchPlaylists(),
		fetchContextTracks("library", "", 0, 30),
	)
}

func (a *App) topTabFromKey(key string) (TopTab, bool) {
	switch key {
	case "1":
		return TabAlbums, true
	case "2":
		return TabArtists, true
	case "3":
		return TabSongs, true
	case "4":
		return TabPlaylists, true
	case "5":
		return TabSearch, true
	default:
		return TabSongs, false
	}
}

func (a *App) playlistsAsStrings() []string {
	items := make([]string, 0, len(a.playlists))
	for _, pl := range a.playlists {
		if pl.Name != "" {
			items = append(items, pl.Name)
		}
	}
	return items
}

func (a *App) refreshPlaylistList() {
	items := a.playlistsAsStrings()
	if len(items) == 0 {
		a.stringListPanel.SetItems([]string{"No playlists"}, "Playlists")
		return
	}
	a.stringListPanel.SetItems(items, "Playlists")
}

func (a *App) switchTopTab(tab TopTab) []tea.Cmd {
	a.activeTopTab = tab

	if a.focus == focusSearchInput && tab != TabSearch {
		a.focus = focusCenter
	}

	var cmds []tea.Cmd
	switch tab {
	case TabAlbums:
		a.centerView = CenterStringList
		a.currentContextType = ""
		a.currentContextValue = ""
		a.stringListPanel.SetItems([]string{"Loading..."}, "Albums")
		cmds = append(cmds, fetchUniqueStrings("albums"))

	case TabArtists:
		a.centerView = CenterStringList
		a.currentContextType = ""
		a.currentContextValue = ""
		a.stringListPanel.SetItems([]string{"Loading..."}, "Artists")
		cmds = append(cmds, fetchUniqueStrings("artists"))

	case TabSongs:
		a.centerView = CenterContextTracks
		a.currentContextType = "library"
		a.currentContextValue = ""
		a.currentContextTotal = 999999
		a.fetchingTracks = true
		a.trackListPanel.SetTracks("Songs", []models.Track{}, a.nowPlaying.Track.Name)
		cmds = append(cmds, fetchContextTracks("library", "", 0, 30))

	case TabPlaylists:
		a.centerView = CenterStringList
		a.currentContextType = ""
		a.currentContextValue = ""
		if len(a.playlists) == 0 {
			a.stringListPanel.SetItems([]string{"Loading..."}, "Playlists")
			cmds = append(cmds, fetchPlaylists())
		} else {
			a.refreshPlaylistList()
		}

	case TabSearch:
		a.centerView = CenterSearchResults
		a.focus = focusSearchInput
	}

	a.updateFocusStyles()
	return cmds
}

func nextTopTab(tab TopTab) TopTab {
	return TopTab((int(tab) + 1) % 5)
}

func prevTopTab(tab TopTab) TopTab {
	return TopTab((int(tab) + 4) % 5)
}

func (a *App) openTracksForSelectedString() []tea.Cmd {
	selectedStr := a.stringListPanel.SelectedItem()
	if selectedStr == "" || selectedStr == "Loading..." || selectedStr == "No playlists" {
		return nil
	}

	cType := ""
	switch a.activeTopTab {
	case TabArtists:
		cType = "artist"
	case TabAlbums:
		cType = "album"
	case TabPlaylists:
		cType = "playlist"
	default:
		return nil
	}

	a.centerView = CenterContextTracks
	a.currentContextType = cType
	a.currentContextValue = selectedStr
	a.currentContextTotal = 999999
	a.fetchingTracks = true
	a.trackListPanel.SetTracks("Loading...", []models.Track{}, "")
	a.updateFocusStyles()

	return []tea.Cmd{fetchContextTracks(cType, selectedStr, 0, 30)}
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.layout.SetSize(msg.Width, msg.Height)

		centerW, rightW, mainH := a.layout.PanelSizes()
		centerTracksH, centerVisH := splitCenterHeights(mainH)
		a.centerTracksHeight = centerTracksH
		a.centerVisualizerHeight = centerVisH
		a.trackListPanel.SetSize(centerW, centerTracksH)
		a.stringListPanel.SetSize(centerW, mainH)
		a.nowPlayingPanel.SetSize(rightW, mainH)
		a.queuePanel.SetSize(rightW, mainH)
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

		if a.centerView == CenterContextTracks {
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
		a.playlists = []models.Playlist(msg)
		if a.activeTopTab == TabPlaylists && a.centerView == CenterStringList {
			a.refreshPlaylistList()
		}

	case uniqueStringsMsg:
		title := "Artists"
		if msg.kind == "albums" {
			title = "Albums"
		}
		a.stringListPanel.SetItems(msg.items, title)

	case contextTracksMsg:
		a.fetchingTracks = false
		if msg.contextType == a.currentContextType && msg.contextValue == a.currentContextValue {
			if msg.offset == 0 {
				title := msg.contextValue
				if msg.contextType == "library" {
					title = "Songs"
				}
				a.trackListPanel.SetTracks(title, msg.tracks, a.nowPlaying.Track.Name)
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
				a.focus = focusCenter
				a.updateFocusStyles()
			case "enter":
				query := a.searchPanel.GetQuery()
				if query != "" {
					a.activeTopTab = TabSearch
					a.centerView = CenterSearchResults
					a.focus = focusCenter
					a.updateFocusStyles()
					cmds = append(cmds, searchLibrary(query))
				}
			default:
				var cmd tea.Cmd
				cmd, _ = a.searchPanel.Update(msg)
				cmds = append(cmds, cmd)
			}
			break
		}

		if tab, ok := a.topTabFromKey(msg.String()); ok {
			cmds = append(cmds, a.switchTopTab(tab)...)
			return a, tea.Batch(cmds...)
		}

		switch msg.String() {
		case "q", "Q":
			return a, tea.Quit

		case "/":
			a.activeTopTab = TabSearch
			a.centerView = CenterSearchResults
			a.focus = focusSearchInput
			a.updateFocusStyles()
			return a, nil

		case "z":
			a.rightView = ViewNowPlaying
			a.updateFocusStyles()

		case "x":
			a.rightView = ViewQueue
			a.updateFocusStyles()
			cmds = append(cmds, fetchQueue())

		case "[":
			cmds = append(cmds, a.switchTopTab(prevTopTab(a.activeTopTab))...)

		case "]":
			cmds = append(cmds, a.switchTopTab(nextTopTab(a.activeTopTab))...)

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
			if a.focus == focusRight {
				a.focus = focusCenter
				a.updateFocusStyles()
			} else if a.focus == focusCenter {
				a.focus = focusSearchInput
				a.updateFocusStyles()
			} else {
				cmds = append(cmds, doSeek(-10, a.nowPlaying.Position))
			}

		case "right":
			if a.focus == focusSearchInput {
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
				a.focus = focusCenter
			case focusCenter:
				a.focus = focusRight
			case focusRight:
				a.focus = focusSearchInput
			}
			a.updateFocusStyles()

		case "esc":
			if a.focus == focusRight {
				a.focus = focusCenter
				a.updateFocusStyles()
			} else if a.focus == focusCenter && a.centerView == CenterContextTracks {
				if a.activeTopTab == TabArtists || a.activeTopTab == TabAlbums || a.activeTopTab == TabPlaylists {
					cmds = append(cmds, a.switchTopTab(a.activeTopTab)...)
				}
			}

		case "enter":
			if a.focus == focusCenter {
				if a.centerView == CenterContextTracks {
					idx := a.trackListPanel.SelectedTrackIndex()
					if idx > 0 && a.currentContextType != "" {
						cmds = append(cmds, doPlayTrackInContext(a.currentContextType, a.currentContextValue, idx))
					}
				} else if a.centerView == CenterSearchResults {
					track := a.searchPanel.SelectedTrack()
					if track != nil {
						cmds = append(cmds, doPlayTrack(track.Name))
						a.rightView = ViewNowPlaying
						a.updateFocusStyles()
					}
				} else if a.centerView == CenterStringList {
					cmds = append(cmds, a.openTracksForSelectedString()...)
				}
			}

		default:
			switch a.focus {
			case focusCenter:
				if a.centerView == CenterContextTracks {
					cmd := a.trackListPanel.Update(msg)
					cmds = append(cmds, cmd)

					idx := a.trackListPanel.SelectedTrackIndex()
					loaded := len(a.trackListPanel.GetTracks())
					if idx >= loaded-10 && loaded < a.currentContextTotal && !a.fetchingTracks {
						a.fetchingTracks = true
						cmds = append(cmds, fetchContextTracks(a.currentContextType, a.currentContextValue, loaded, 30))
					}
				} else if a.centerView == CenterSearchResults {
					cmd, _ := a.searchPanel.Update(msg)
					cmds = append(cmds, cmd)
				} else if a.centerView == CenterStringList {
					cmd := a.stringListPanel.Update(msg)
					cmds = append(cmds, cmd)
				}
			case focusRight:
				if a.rightView == ViewQueue {
					cmd := a.queuePanel.Update(msg)
					cmds = append(cmds, cmd)
				}
			}
		}
	}

	return a, tea.Batch(cmds...)
}

func (a *App) updateFocusStyles() {
	a.searchPanel.SetFocused(a.focus == focusSearchInput)
	a.trackListPanel.SetFocused(a.focus == focusCenter && a.centerView == CenterContextTracks)
	a.stringListPanel.SetFocused(a.focus == focusCenter && a.centerView == CenterStringList)
	a.nowPlayingPanel.SetFocused(a.focus == focusRight && a.rightView == ViewNowPlaying)
	a.queuePanel.SetFocused(a.focus == focusRight && a.rightView == ViewQueue)
}

func (a App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}

	topView := renderTopBar(a.width, a.activeTopTab)

	centerW, _, mainH := a.layout.PanelSizes()

	var centerView string
	if a.centerView == CenterContextTracks {
		centerView = a.trackListPanel.View()
		if a.centerVisualizerHeight > 0 {
			visView := panels.RenderVisualizerPanel(
				centerW,
				a.centerVisualizerHeight,
				a.tick,
				a.nowPlaying.State == models.StatePlaying,
				a.focus == focusCenter,
			)
			centerView = lipgloss.JoinVertical(lipgloss.Left, centerView, visView)
		}
	} else if a.centerView == CenterStringList {
		centerView = a.stringListPanel.View()
	} else {
		inputH := 3
		resultsH := mainH - inputH
		if resultsH < 6 {
			resultsH = 6
		}
		searchInput := a.searchPanel.ViewInput(centerW, a.focus == focusSearchInput)
		searchResults := a.searchPanel.ViewResults(centerW, resultsH, a.focus == focusCenter)
		centerView = lipgloss.JoinVertical(lipgloss.Left, searchInput, searchResults)
	}

	rightView := a.nowPlayingPanel.View(a.tick)
	if a.rightView == ViewQueue {
		rightView = a.queuePanel.View()
	}

	statusView := a.statusBar.View(a.nowPlaying)

	return a.layout.Render(
		topView,
		centerView,
		rightView,
		statusView,
	)
}
