package ui

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sebasreateguib/applemusictui/internal/applescript"
	"github.com/sebasreateguib/applemusictui/internal/artwork"
	"github.com/sebasreateguib/applemusictui/internal/cava"
	"github.com/sebasreateguib/applemusictui/internal/config"
	"github.com/sebasreateguib/applemusictui/internal/models"
	"github.com/sebasreateguib/applemusictui/internal/ui/panels"
	"github.com/sebasreateguib/applemusictui/internal/ui/theme"
)

type tickMsg time.Time
type nowPlayingMsg models.NowPlaying
type artworkMsg string

// imageMsg reports that the artwork at path was uploaded to the terminal.
type imageMsg struct {
	path string
	id   uint32
	err  error
}

type configSavedMsg struct{ err error }

type playlistsMsg []models.Playlist

type tracksMsg struct {
	ctx    trackContext
	tracks []models.Track
	err    error
}

// previewMsg loads the selection of a list, unless it moved meanwhile.
type previewMsg struct{ seq int }

type uniqueStringsMsg struct {
	kind  string // "artists" or "albums"
	items []string
}

// currentContextMsg is the playlist Music is playing, to shuffle it or not.
// It has no tracks when there is none.
type currentContextMsg struct {
	ctx     trackContext
	tracks  []models.Track
	shuffle bool
}

type queueMsg struct {
	seq    int
	tracks []models.QueueTrack
}

type visualizerMsg cava.Event
type seekedMsg models.NowPlaying

// polledMsg is the answer to the periodic question.
type polledMsg models.NowPlaying

const (
	// silenceNoticeAfter is how long a playing track may yield no audio before
	// the visualizer explains why it is empty.
	silenceNoticeAfter = 5 * time.Second

	// noticeTicks is how long a notice stays in the bottom line (~3 s).
	noticeTicks = 20

	// searchAbove is the fraction of the free rows above the search popup: it
	// looks too low in the exact middle.
	searchAbove = 0.28

	// previewDelay avoids asking Music for every row the cursor passes over.
	previewDelay = 120 * time.Millisecond
)

// panel is a part of the screen that can be focused.
type panel int

const (
	panelPlaylists panel = iota
	panelArtists
	panelAlbums
	panelTracks
	panelQueue
	panelCount
)

type popup int

const (
	popupNone popup = iota
	popupSearch
	popupHelp
)

const (
	kindLibrary  = "library"
	kindPlaylist = "playlist"
	kindArtist   = "artist"
	kindAlbum    = "album"
	kindSearch   = "search"
)

// trackContext identifies the contents of the tracks panel.
type trackContext struct{ kind, value string }

func (c trackContext) title() string {
	switch c.kind {
	case kindLibrary:
		return "All Songs"
	case kindSearch:
		return "Search: " + c.value
	default:
		return strings.TrimSpace(c.value)
	}
}

type App struct {
	sides           [sideLists]panels.SideList
	tracks          panels.TrackList
	queue           panels.QueuePanel
	nowPlayingPanel panels.NowPlayingPanel
	search          panels.SearchPrompt
	help            panels.Help
	layout          Layout
	cache           *viewCache

	width, height int
	tick          int
	focus         panel
	activeSide    panel // last focused list on the left
	popup         popup
	notice        string
	noticeUntil   int // tick that hides the notice

	nowPlaying models.NowPlaying
	polling    bool // waiting for a polledMsg
	lastTrack  models.Track
	seek       seek
	volume     volume

	context    trackContext // shown in the tracks panel
	loaded     bool         // the tracks of context have arrived
	trackCache map[trackContext][]models.Track
	previewSeq int

	playback           trackContext // list being played, empty if unknown
	playbackTrackIndex int          // 1-based position in it, 0 if unknown
	playbackShuffleOn  bool
	playbackOrder      []int // indexes in playing order, when shuffled
	playbackOrderPos   int
	expected           models.Track // started by the app
	nowPlayingAt       time.Time    // when nowPlaying.Position was true
	advancedAt         time.Time    // when the app last moved on to the next track
	queueSeq           int

	visualizer *cava.Visualizer
	visLevels  []byte
	visErr     error
	visHeardAt time.Time // last time audio was heard, or none was expected

	config          config.Config
	imagesSupported bool
	imageID         uint32 // image in the terminal, 0 if none
	imagePath       string // artwork uploaded, or being uploaded
}

// NewApp creates the app. imagesSupported tells whether the terminal can show
// real images (see artwork.Supported).
func NewApp(width, height int, cfg config.Config, imagesSupported bool) App {
	app := App{
		sides: [sideLists]panels.SideList{
			panels.NewSideList("1", "Playlists"),
			panels.NewSideList("2", "Artists"),
			panels.NewSideList("3", "Albums"),
		},
		tracks:          panels.NewTrackList(),
		queue:           panels.NewQueuePanel(),
		nowPlayingPanel: panels.NewNowPlayingPanel(),
		search:          panels.NewSearchPrompt(),
		help:            panels.NewHelp(helpBindings),
		cache:           &viewCache{},
		width:           width,
		height:          height,
		trackCache:      map[trackContext][]models.Track{},
		visualizer:      cava.New(),
		visHeardAt:      time.Now(),
		config:          cfg,
		imagesSupported: imagesSupported,
	}
	app.context = trackContext{kind: kindLibrary}
	app.tracks.SetLoading(kindLibrary, app.context.title())
	app.setPlaylists(nil)
	app.relayout()
	return app
}

// Close releases the resources held by the app (its child processes).
func (a App) Close() {
	a.visualizer.Close()
	applescript.Close()
}

func (a App) isPlaying() bool {
	return a.nowPlaying.State == models.StatePlaying
}

// updateVisualizer applies an event coming from cava.
func (a *App) updateVisualizer(e cava.Event) {
	a.visErr = e.Err
	if e.Err != nil {
		a.visLevels = nil
		return
	}
	if !cava.IsSilent(e.Levels) {
		a.visHeardAt = time.Now()
	}
	if !a.isPlaying() {
		// cava hears the whole system: while Music is paused let the bars
		// fall instead of following other apps.
		for i := range e.Levels {
			prev := 0
			if i < len(a.visLevels) {
				prev = int(a.visLevels[i]) * 7 / 10
			}
			e.Levels[i] = byte(min(int(e.Levels[i]), prev))
		}
	}
	a.visLevels = e.Levels
}

// visualizerNotice explains an empty visualizer, or returns "".
func (a App) visualizerNotice() string {
	if a.visErr != nil {
		return a.visErr.Error()
	}
	if a.isPlaying() && time.Since(a.visHeardAt) > silenceNoticeAfter {
		return "No audio captured - allow your terminal in Privacy > System Audio Recording"
	}
	return ""
}

// syncImage makes the terminal hold the image the UI has to show, if any, and
// returns the upload command when one is needed.
func (a *App) syncImage() tea.Cmd {
	path := ""
	if a.config.PhotoArtwork && a.imagesSupported {
		path = a.nowPlaying.ArtworkPath
	}
	if path == a.imagePath {
		return nil
	}

	a.imagePath = path
	a.nowPlayingPanel.SetImage(0)
	artwork.Delete(a.imageID)
	a.imageID = 0

	if path == "" {
		return nil
	}
	return func() tea.Msg {
		id, err := artwork.Upload(path)
		return imageMsg{path: path, id: id, err: err}
	}
}

// showImage handles a finished upload. The pixelated artwork stays on screen
// when it failed.
func (a *App) showImage(msg imageMsg) {
	if msg.err != nil {
		return
	}
	if msg.path != a.imagePath || a.imageID != 0 {
		artwork.Delete(msg.id) // the track or the mode changed meanwhile
		return
	}
	a.imageID = msg.id
	a.nowPlayingPanel.SetImage(msg.id)
}

func (a *App) toggleArtwork() tea.Cmd {
	if !a.imagesSupported {
		a.setNotice("Photo artwork needs Ghostty or Kitty (outside a multiplexer)")
		return nil
	}

	a.config.PhotoArtwork = !a.config.PhotoArtwork
	if a.config.PhotoArtwork {
		a.setNotice("Artwork: photo")
	} else {
		a.setNotice("Artwork: pixelated")
	}

	cmds := []tea.Cmd{a.syncImage()}
	if path := a.nowPlaying.ArtworkPath; a.config.PhotoArtwork && path != "" {
		// Another instance deletes the file when it starts: export it again.
		if _, err := os.Stat(path); err != nil {
			cmds = append(cmds, fetchArtwork())
		}
	}

	cfg := a.config
	cmds = append(cmds, func() tea.Msg {
		return configSavedMsg{config.Save(cfg)}
	})
	return tea.Batch(cmds...)
}

func (a *App) setNotice(text string) {
	a.notice = text
	a.noticeUntil = a.tick + noticeTicks
}

func (a App) visualizerVisible() bool {
	return a.layout.visH > 0 && !a.queueInMain() && !a.layout.tooSmall()
}

// queueInMain reports whether Up Next replaces the tracks panel.
func (a App) queueInMain() bool {
	return !a.layout.wide() && a.focus == panelQueue
}

// relayout recomputes the size and state of every panel.
func (a *App) relayout() {
	a.layout = newLayout(a.width, a.height, int(a.activeSide), a.sides[panelPlaylists].Len())
	for i := range a.sides {
		a.sides[i].SetSize(a.layout.leftW, a.layout.sideH[i])
		a.sides[i].SetState(a.focus == panel(i), a.activeSide == panel(i) && a.context.kind != kindSearch)
	}
	a.tracks.SetSize(a.layout.mainW, a.layout.tracksH)
	a.tracks.SetFocused(a.focus == panelTracks)
	a.queue.SetSize(a.layout.queueSize())
	a.queue.SetFocused(a.focus == panelQueue)
	if a.layout.wide() {
		a.nowPlayingPanel.SetSize(a.layout.rightW, a.layout.nowPlayingH)
	}
}

// focusPanel moves the focus; focusing a list shows its selection.
func (a *App) focusPanel(p panel) tea.Cmd {
	a.focus = p
	var cmd tea.Cmd
	if p < sideLists {
		a.activeSide = p
		cmd = a.preview()
	}
	a.relayout()
	return cmd
}

// setPlaylists fills the first list, which starts with the whole library.
func (a *App) setPlaylists(playlists []models.Playlist) {
	songs := panels.SideItem{Label: "All Songs"}
	if tracks, ok := a.trackCache[trackContext{kind: kindLibrary}]; ok {
		songs.Count = len(tracks)
	}
	items := []panels.SideItem{songs}
	for _, pl := range playlists {
		if pl.Name != "" {
			items = append(items, panels.SideItem{Label: strings.TrimSpace(pl.Name), Value: pl.Name, Count: pl.Count})
		}
	}
	a.sides[panelPlaylists].SetItems(items)
}

func sideItems(values []string) []panels.SideItem {
	items := make([]panels.SideItem, len(values))
	for i, v := range values {
		items[i] = panels.SideItem{Label: strings.TrimSpace(v), Value: v}
	}
	slices.SortStableFunc(items, func(a, b panels.SideItem) int {
		return strings.Compare(strings.ToLower(a.Label), strings.ToLower(b.Label))
	})
	return items
}

// selectedContext returns the context of the active list's selection.
func (a *App) selectedContext() (trackContext, bool) {
	it := a.sides[a.activeSide].Selected()
	if it == nil {
		return trackContext{}, false
	}
	switch a.activeSide {
	case panelArtists:
		return trackContext{kindArtist, it.Value}, true
	case panelAlbums:
		return trackContext{kindAlbum, it.Value}, true
	default:
		if it.Value == "" {
			return trackContext{kind: kindLibrary}, true
		}
		return trackContext{kindPlaylist, it.Value}, true
	}
}

// preview shows the active list's selection, after a short delay when its
// tracks have to be fetched.
func (a *App) preview() tea.Cmd {
	a.previewSeq++
	ctx, ok := a.selectedContext()
	if !ok || ctx == a.context {
		return nil
	}
	if _, cached := a.trackCache[ctx]; cached {
		return a.showContext(ctx)
	}
	seq := a.previewSeq
	return tea.Tick(previewDelay, func(time.Time) tea.Msg { return previewMsg{seq} })
}

// showContext shows ctx in the tracks panel, fetching its tracks if needed.
func (a *App) showContext(ctx trackContext) tea.Cmd {
	a.previewSeq++ // whatever was waiting is superseded
	if ctx == a.context && a.loaded {
		return nil
	}
	a.context = ctx
	if tracks, ok := a.trackCache[ctx]; ok {
		a.loaded = true
		a.tracks.SetTracks(ctx.kind, ctx.title(), tracks, "No tracks")
		a.relayout()
		return nil
	}
	a.loaded = false
	a.tracks.SetLoading(ctx.kind, ctx.title())
	a.relayout()
	if ctx.kind == kindSearch {
		return searchLibrary(ctx.value)
	}
	return fetchTracks(ctx)
}

func (a *App) receiveTracks(msg tracksMsg) tea.Cmd {
	if msg.err == nil && msg.ctx.kind != kindSearch {
		a.trackCache[msg.ctx] = msg.tracks
		if msg.ctx.kind == kindLibrary {
			a.setLibraryCount(len(msg.tracks))
		}
	}
	var cmd tea.Cmd
	if msg.err == nil && a.hasContext() && msg.ctx == a.playback {
		if len(a.playbackOrder) != len(msg.tracks) {
			a.playbackOrder = nil // the list changed: shuffle it again
		}
		cmd = a.locatePlaying()
	}
	if msg.ctx != a.context || a.loaded {
		return cmd
	}
	empty := "No tracks"
	switch {
	case msg.err != nil:
		empty = "Music did not answer - press R to retry"
	case msg.ctx.kind == kindSearch:
		empty = fmt.Sprintf("Nothing found for %q", msg.ctx.value)
	}
	a.loaded = true
	a.tracks.SetTracks(msg.ctx.kind, msg.ctx.title(), msg.tracks, empty)
	return cmd
}

func (a *App) setLibraryCount(count int) {
	a.sides[panelPlaylists].SetCount("", count)
}

// refresh forgets what is known about the library and asks for it again.
func (a *App) refresh() tea.Cmd {
	a.setNotice("Refreshing library…")
	playing, ok := a.trackCache[a.playback]
	clear(a.trackCache)
	var fetchPlaying tea.Cmd
	if ok { // needed until the new one arrives
		a.trackCache[a.playback] = playing
		if a.playback != a.context {
			fetchPlaying = fetchTracks(a.playback)
		}
	}
	a.loaded = false
	a.tracks.SetLoading(a.context.kind, a.context.title())
	fetch := fetchTracks(a.context)
	if a.context.kind == kindSearch {
		fetch = searchLibrary(a.context.value)
	}
	return tea.Batch(fetch, fetchPlaying, fetchPlaylists(), fetchUniqueStrings("artists"), fetchUniqueStrings("albums"))
}

func (a App) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		a.waitVisualizer(),
		fetchNowPlaying(),
		fetchPlaylists(),
		fetchUniqueStrings("artists"),
		fetchUniqueStrings("albums"),
		fetchTracks(a.context),
		fetchQueue(a.queueSeq),
	)
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
		a.relayout()
		a.visualizer.Resize(panels.VisualizerBars(a.layout.mainW))

	case visualizerMsg:
		a.updateVisualizer(cava.Event(msg))
		cmds = append(cmds, a.waitVisualizer())

	case tickMsg:
		a.tick++
		if a.tick >= a.noticeUntil {
			a.notice = ""
		}
		cmds = append(cmds, tickCmd())

		a.visualizer.SetActive(a.visualizerVisible())
		if !a.visualizerVisible() || !a.isPlaying() || a.nowPlaying.Volume == 0 {
			a.visHeardAt = time.Now()
		}

		// Not while Music is slow to answer: the questions would pile up.
		if a.tick%6 == 0 && !a.polling {
			a.polling = true
			cmds = append(cmds, pollNowPlaying())
		}
		// Also refreshed when the track changes.
		if a.tick%40 == 0 {
			cmds = append(cmds, a.refreshQueue())
		}
		cmds = append(cmds, a.advanceCmd(time.Now()))

	case nowPlayingMsg:
		cmds = append(cmds, a.receiveNowPlaying(models.NowPlaying(msg))...)

	case polledMsg:
		a.polling = false
		cmds = append(cmds, a.receiveNowPlaying(models.NowPlaying(msg))...)

	case seekedMsg:
		a.seek.pending--
		cmds = append(cmds, a.receiveNowPlaying(models.NowPlaying(msg))...)

	case volumeMsg:
		cmds = append(cmds, a.receiveVolume(msg)...)

	case artworkMsg:
		a.nowPlaying.ArtworkPath = string(msg)
		a.nowPlayingPanel.SetNowPlaying(a.nowPlaying)
		cmds = append(cmds, a.syncImage())

	case imageMsg:
		a.showImage(msg)

	case configSavedMsg:
		if msg.err != nil {
			a.setNotice("Could not save preferences: " + msg.err.Error())
		}

	case playlistsMsg:
		a.setPlaylists(msg)
		a.relayout()

	case uniqueStringsMsg:
		side := panelArtists
		if msg.kind == "albums" {
			side = panelAlbums
		}
		a.sides[side].SetItems(sideItems(msg.items))
		if a.focus == side {
			cmds = append(cmds, a.preview())
		}

	case previewMsg:
		if ctx, ok := a.selectedContext(); ok && msg.seq == a.previewSeq {
			cmds = append(cmds, a.showContext(ctx))
		}

	case tracksMsg:
		cmds = append(cmds, a.receiveTracks(msg))

	case currentContextMsg:
		cmds = append(cmds, a.receiveCurrentContext(msg))

	case queueMsg:
		a.receiveQueue(msg)

	case tea.KeyMsg:
		// Not in the return statement: it would copy a before the key changes it.
		cmds = append(cmds, a.handleKey(msg))

	case tea.MouseMsg:
		cmds = append(cmds, a.handleMouse(msg))

	default:
		if a.popup == popupSearch { // cursor blink
			cmds = append(cmds, a.search.Update(msg))
		}
	}

	return a, tea.Batch(cmds...)
}

func (a *App) receiveNowPlaying(fresh models.NowPlaying) []tea.Cmd {
	if a.hasContextualPlayback() {
		fresh.ShuffleEnabled = a.playbackShuffleOn
	}
	fresh.ArtworkPath = a.nowPlaying.ArtworkPath
	if a.volume.sending {
		fresh.Volume = a.volume.target
	}

	prev := a.lastTrack
	track := fresh.Track
	changed := track.Name != prev.Name || track.Artist != prev.Artist || track.Album != prev.Album
	if changed {
		fresh.ArtworkPath = ""
		a.seek.dragging = false
	} else if a.seek.pending > 0 {
		fresh.Position = a.seek.sent // reported before the seek
	}
	now := time.Now()
	var takeBack tea.Cmd
	if changed { // before forgetting where the previous track was
		takeBack = a.takeBackCmd(fresh, now)
	}
	a.lastTrack = track
	a.nowPlaying = fresh
	a.nowPlayingAt = now
	if !fresh.PositionAt.IsZero() && a.seek.pending == 0 {
		a.nowPlayingAt = fresh.PositionAt
	}
	a.nowPlayingPanel.SetNowPlaying(fresh)
	a.tracks.SetCurrent(track)
	if !changed {
		return nil
	}

	if takeBack != nil {
		return []tea.Cmd{a.syncImage(), takeBack}
	}
	cmds := []tea.Cmd{a.syncImage(), fetchArtwork()}
	if a.hasContext() && track.Name != "" {
		return append(cmds, a.locatePlaying())
	}
	return append(cmds, a.refreshQueue())
}

func (a App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}
	if a.layout.tooSmall() {
		text := theme.Dim.Render(fmt.Sprintf("Terminal too small (needs %dx%d)", minWidth, minHeight))
		lines := fitLines("", a.width, a.height)
		lines[a.height/2] = panels.Center(text, a.width)
		return strings.Join(lines, "\n")
	}

	lines := joinColumns(a.columns()...)
	withTrack := !a.layout.wide()
	np := a.shownNowPlaying()
	lines = append(lines, a.fade(a.cache.player.get(playerKey{a.width, np, withTrack}, func() []string {
		return fitLines(renderPlayer(np, a.width, withTrack), a.width, playerHeight)
	}))...)
	brandText := encryptedLabel(brand, a.tick)
	lines = append(lines, a.fade(a.cache.bottom.get(bottomKey{a.width, a.focus, a.notice, brandText}, func() []string {
		return fitLines(renderBottom(a.width, a.focus, a.notice, brandText), a.width, bottomHeight)
	}))...)

	switch a.popup {
	case popupSearch:
		return panels.Overlay(strings.Join(lines, "\n"), a.search.View(a.width), a.width, a.height, searchAbove)
	case popupHelp:
		return panels.Overlay(strings.Join(lines, "\n"), a.help.View(a.width, a.height), a.width, a.height, 0.5)
	default:
		return strings.Join(lines, "\n")
	}
}

// fade takes the colors off lines while a popup is open.
func (a App) fade(lines []string) []string {
	if a.popup == popupNone {
		return lines
	}
	return panels.Fade(lines)
}

// columns renders the body, from left to right.
func (a App) columns() [][]string {
	l := a.layout

	var left []string
	for i := range a.sides {
		key := listKey{l.leftW, l.sideH[i], a.sides[i].Revision()}
		left = append(left, a.cache.sides[i].get(key, func() []string {
			return fitLines(a.sides[i].View(), l.leftW, l.sideH[i])
		})...)
	}

	queueW, queueH := l.queueSize()
	queue := func() []string {
		return a.cache.queue.get(listKey{queueW, queueH, a.queue.Revision()}, func() []string {
			return fitLines(a.queue.View(), queueW, queueH)
		})
	}
	if a.queueInMain() {
		return [][]string{a.fade(left), a.fade(queue())}
	}

	main := slices.Clone(a.cache.tracks.get(listKey{l.mainW, l.tracksH, a.tracks.Revision()}, func() []string {
		return fitLines(a.tracks.View(), l.mainW, l.tracksH)
	}))
	if l.visH > 0 {
		vis := panels.RenderVisualizerPanel(l.mainW, l.visH, a.visLevels, a.visualizerNotice())
		main = append(main, strings.Split(vis, "\n")...)
	}
	if !l.wide() {
		return [][]string{a.fade(left), a.fade(main)}
	}

	faded := a.popup != popupNone
	right := slices.Clone(a.cache.nowPlaying.get(a.nowPlayingPanel.ViewKey(a.tick, faded), func() []string {
		return fitLines(a.nowPlayingPanel.View(a.tick, faded), l.rightW, l.nowPlayingH)
	}))
	return [][]string{a.fade(left), a.fade(main), append(right, a.fade(queue())...)}
}
