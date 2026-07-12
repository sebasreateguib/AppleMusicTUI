package ui

import (
	"math/rand"
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

type playbackIndexMsg struct {
	contextType  string
	contextValue string
	index        int
}

type contextCountMsg struct {
	contextType  string
	contextValue string
	count        int
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

	playbackContextType  string
	playbackContextValue string
	playbackTrackIndex   int
	playbackContextCount int
	playbackShuffleOn    bool
	playbackOrder        []int
	playbackOrderPos     int
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
		currentContextTotal:    0,
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

func fetchContextTracks(cType, cVal string, offset, limit int) tea.Cmd {
	return func() tea.Msg {
		tracks, _ := applescript.GetFilteredTracks(cType, cVal, offset, limit)
		return contextTracksMsg{contextType: cType, contextValue: cVal, offset: offset, tracks: tracks}
	}
}

func fetchContextCount(cType, cVal string) tea.Cmd {
	return func() tea.Msg {
		count, _ := applescript.GetContextTrackCount(cType, cVal)
		return contextCountMsg{contextType: cType, contextValue: cVal, count: count}
	}
}

func searchLibrary(query string) tea.Cmd {
	return func() tea.Msg {
		results, _ := applescript.SearchLibrary(query)
		return searchResultsMsg(results)
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

func doPlayTrack(name string) tea.Cmd {
	return func() tea.Msg {
		_ = applescript.PlayTrackByName(name)
		time.Sleep(300 * time.Millisecond)
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

func (a App) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		fetchNowPlaying(),
		fetchPlaylists(),
		fetchContextTracks("library", "", 0, 30),
		fetchContextCount("library", ""),
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
		a.currentContextTotal = 0
		a.fetchingTracks = true
		a.trackListPanel.SetTracks("Songs", []models.Track{}, a.nowPlaying.Track.Name)
		cmds = append(cmds, fetchContextTracks("library", "", 0, 30), fetchContextCount("library", ""))

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
	a.currentContextTotal = 0
	a.fetchingTracks = true
	a.trackListPanel.SetTracks("Loading...", []models.Track{}, "")
	a.updateFocusStyles()

	return []tea.Cmd{fetchContextTracks(cType, selectedStr, 0, 30), fetchContextCount(cType, selectedStr)}
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
	if a.rightView == ViewQueue {
		cmds = append(cmds, a.queueFetchCmd())
	}
	return tea.Batch(cmds...)
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
				cmds = append(cmds, a.queueFetchCmd())
			}
		}

		if a.centerView == CenterContextTracks {
			a.trackListPanel.SetCurrentTrack(a.nowPlaying.Track.Name)
		}

	case nowPlayingMsg:
		prevTrack := a.nowPlaying.Track
		fresh := models.NowPlaying(msg)
		if a.playbackContextType != "" {
			fresh.ShuffleEnabled = a.playbackShuffleOn
		}
		fresh.ArtworkPath = a.nowPlaying.ArtworkPath
		a.nowPlaying = fresh
		a.nowPlayingPanel.SetNowPlaying(a.nowPlaying)

		if a.nowPlaying.Track.Name != a.lastTrackName {
			a.lastTrackName = a.nowPlaying.Track.Name
			a.nowPlaying.ArtworkPath = ""
			a.nowPlayingPanel.SetNowPlaying(a.nowPlaying)
			cmds = append(cmds, fetchArtwork())
		}

		if a.playbackContextType != "" &&
			(a.nowPlaying.Track.Name != prevTrack.Name ||
				a.nowPlaying.Track.Artist != prevTrack.Artist ||
				a.nowPlaying.Track.Album != prevTrack.Album) &&
			a.nowPlaying.Track.Name != "" {
			cmds = append(cmds, fetchPlaybackIndex(a.playbackContextType, a.playbackContextValue, a.nowPlaying.Track))
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

	case contextCountMsg:
		if msg.contextType == a.currentContextType && msg.contextValue == a.currentContextValue {
			a.currentContextTotal = msg.count
		}
		if msg.contextType == a.playbackContextType && msg.contextValue == a.playbackContextValue {
			a.playbackContextCount = msg.count
			if a.playbackShuffleOn {
				a.rebuildPlaybackOrder(a.playbackTrackIndex)
				if a.rightView == ViewQueue {
					cmds = append(cmds, a.queueFetchCmd())
				}
			}
		}

	case playbackIndexMsg:
		if msg.contextType == a.playbackContextType && msg.contextValue == a.playbackContextValue {
			a.syncPlaybackPosition(msg.index)
			if a.rightView == ViewQueue {
				cmds = append(cmds, a.queueFetchCmd())
			}
		}

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
			cmds = append(cmds, a.queueFetchCmd())

		case "[":
			cmds = append(cmds, a.switchTopTab(prevTopTab(a.activeTopTab))...)

		case "]":
			cmds = append(cmds, a.switchTopTab(nextTopTab(a.activeTopTab))...)

		case " ":
			cmds = append(cmds, doPlayPause())

		case "n":
			cmds = append(cmds, a.nextCmd())

		case "p":
			cmds = append(cmds, a.prevCmd())

		case "s":
			if a.hasContextualPlayback() {
				cmds = append(cmds, a.toggleContextShuffle())
			} else {
				cmds = append(cmds, doToggleShuffle())
			}

		case "r":
			cmds = append(cmds, doToggleRepeat())

		case "+", "=":
			cmds = append(cmds, doChangeVolume(5, a.nowPlaying.Volume))

		case "-":
			cmds = append(cmds, doChangeVolume(-5, a.nowPlaying.Volume))
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
						a.setPlaybackContext(a.currentContextType, a.currentContextValue, idx, a.currentContextTotal)
						cmds = append(cmds, doPlayTrackInContext(a.currentContextType, a.currentContextValue, idx, a.playbackShuffleOn))
					}
				} else if a.centerView == CenterSearchResults {
					track := a.searchPanel.SelectedTrack()
					if track != nil {
						a.setPlaybackContext("library", "", 0, a.currentContextTotal)
						cmds = append(cmds, doPlayLibraryTrack(*track))
						cmds = append(cmds, fetchPlaybackIndex("library", "", *track))
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
					if idx >= loaded-10 && (a.currentContextTotal == 0 || loaded < a.currentContextTotal) && !a.fetchingTracks {
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

	topView := renderTopBar(a.width, a.activeTopTab, a.tick)

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
