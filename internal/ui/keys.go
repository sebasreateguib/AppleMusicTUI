package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/sebasreateguib/applemusictui/internal/artwork"
	"github.com/sebasreateguib/applemusictui/internal/ui/panels"
)

var helpBindings = []panels.Binding{
	{Text: "Panels"},
	{Key: "1 - 5", Text: "Focus a panel"},
	{Key: "h l ← → tab", Text: "Previous / next panel"},
	{Key: "enter", Text: "Open the selection / play the track"},
	{Key: "esc", Text: "Back to the list on the left"},
	{},
	{Text: "Lists"},
	{Key: "j k ↓ ↑", Text: "Move"},
	{Key: "pgdn pgup", Text: "Move a page (also ctrl+d, ctrl+u)"},
	{Key: "g G", Text: "First / last"},
	{Key: "c", Text: "Go to the track that is playing"},
	{},
	{Text: "Playback"},
	{Key: "space", Text: "Play / pause"},
	{Key: "n p", Text: "Next / previous track"},
	{Key: "mouse", Text: "Click or drag the progress bar"},
	{Key: "+ -", Text: "Volume"},
	{Key: "s", Text: "Shuffle"},
	{Key: "r", Text: "Repeat"},
	{},
	{Text: "Other"},
	{Key: "/", Text: "Search the library"},
	{Key: "i", Text: "Artwork: pixelated / photo"},
	{Key: "R", Text: "Refresh the library"},
	{Key: "?", Text: "This help"},
	{Key: "q", Text: "Quit"},
}

func (a *App) handleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()
	if key == "ctrl+c" {
		return a.quit()
	}
	switch a.popup {
	case popupSearch:
		return a.handleSearchKey(msg)
	case popupHelp:
		switch key {
		case "esc", "q", "?", "enter":
			a.popup = popupNone
		default:
			a.help.HandleKey(key)
		}
		return nil
	}

	switch key {
	case "q", "Q":
		return a.quit()
	case "?":
		a.popup = popupHelp
		a.help.Open()
		return nil
	case "/":
		a.popup = popupSearch
		return a.search.Open()

	case "1", "2", "3", "4", "5":
		return a.focusPanel(panel(key[0] - '1'))
	case "right", "l", "tab":
		return a.focusPanel((a.focus + 1) % panelCount)
	case "left", "h", "shift+tab":
		return a.focusPanel((a.focus + panelCount - 1) % panelCount)
	case "esc":
		if a.focus >= sideLists || a.context.kind == kindSearch {
			return a.focusPanel(a.activeSide)
		}
		return nil
	case "enter":
		return a.open()
	case "c":
		if a.tracks.SelectCurrent() {
			return a.focusPanel(panelTracks)
		}
		a.setNotice("The current track is not in this list")
		return nil

	case " ":
		return doPlayPause()
	case "n":
		return a.nextCmd()
	case "p":
		return a.prevCmd()
	case "+", "=":
		return a.changeVolume(5)
	case "-":
		return a.changeVolume(-5)
	case "s":
		return a.toggleShuffle()
	case "r":
		return doToggleRepeat()
	case "i":
		return a.toggleArtwork()
	case "R":
		return a.refresh()
	}

	switch a.focus {
	case panelTracks:
		a.tracks.HandleKey(key)
	case panelQueue:
		a.queue.HandleKey(key)
	default:
		if a.sides[a.focus].HandleKey(key) {
			return a.preview()
		}
	}
	return nil
}

func (a *App) handleSearchKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		a.popup = popupNone
		a.search.Close()
		return nil
	case "enter":
		query := a.search.Query()
		if query == "" {
			return nil
		}
		a.popup = popupNone
		a.search.Close()
		a.focus = panelTracks
		cmd := a.showContext(trackContext{kindSearch, query})
		a.relayout()
		return cmd
	default:
		return a.search.Update(msg)
	}
}

// open enters the selected list, or plays the selected track.
func (a *App) open() tea.Cmd {
	switch {
	case a.focus < sideLists:
		ctx, ok := a.selectedContext()
		if !ok {
			return nil
		}
		// Do not wait for the preview: the tracks must match the selection.
		cmd := a.showContext(ctx)
		return tea.Batch(cmd, a.focusPanel(panelTracks))

	case a.focus == panelTracks:
		return a.playSelected()

	default:
		return nil
	}
}

func (a *App) playSelected() tea.Cmd {
	track := a.tracks.Selected()
	if track == nil {
		return nil
	}
	if a.context.kind == kindSearch {
		// Found again in the library once it plays, see locatePlaying.
		a.setPlaybackContext(trackContext{kind: kindLibrary}, 0)
		a.expected = *track
		return doPlayLibraryTrack(*track)
	}
	a.setPlaybackContext(a.context, 0)
	return a.playIndex(a.tracks.SelectedIndex())
}

func (a *App) quit() tea.Cmd {
	// Before leaving the alternate screen, which owns the image.
	artwork.Delete(a.imageID)
	return tea.Quit
}
