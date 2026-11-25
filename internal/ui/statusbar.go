package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"SR-Player/internal/models"
	"SR-Player/internal/ui/panels"
	"SR-Player/internal/ui/theme"
)

const brand = "Apple Music TUI"

// renderPlayer draws the playback bar. withTrack adds the track's name, for
// when the Now Playing panel is not on screen.
func renderPlayer(np models.NowPlaying, width int, withTrack bool) string {
	title := "■ Stopped"
	switch np.State {
	case models.StatePlaying:
		title = "▶ Playing"
	case models.StatePaused:
		title = "┃┃ Paused"
	}

	inner := width - 2
	left := " "
	if withTrack && np.Track.Name != "" {
		name := ansi.Truncate(np.Track.Name, inner/4, "…")
		artist := ansi.Truncate(np.Track.Artist, inner/6, "…")
		left += theme.Bold.Render(name) + theme.Dim.Render(" "+artist) + "  "
	}
	elapsed, total := formatTime(np.Position), formatTime(np.Track.Duration)
	barW := inner - ansi.StringWidth(left) - len(elapsed) - len(total) - 3
	line := left + theme.Text.Render(elapsed) + " " +
		renderProgressBar(np.Position, np.Track.Duration, barW) + " " + theme.Dim.Render(total)

	frame := panels.Frame{Title: title, Info: playerInfo(np), Focused: np.State == models.StatePlaying}
	return frame.Render(width, playerHeight, []string{line})
}

func playerInfo(np models.NowPlaying) string {
	mode := func(icon, value string, on bool) string {
		style := theme.Dim
		if on {
			style = theme.Active
		}
		return style.Render(" " + icon + " " + value + " ")
	}

	shuffle := "off"
	if np.ShuffleEnabled {
		shuffle = "on"
	}
	repeat := np.RepeatMode
	if repeat == "" {
		repeat = models.RepeatOff
	}

	const steps = 8
	filled := clamp(int(math.Round(float64(np.Volume)/100*steps)), 0, steps)
	volume := theme.Dim.Render(" vol ") + theme.Red.Render(strings.Repeat("━", filled)) +
		theme.Faint.Render(strings.Repeat("━", steps-filled)) + theme.Dim.Render(fmt.Sprintf(" %3d%% ", np.Volume))

	return mode("⇄", shuffle, np.ShuffleEnabled) +
		mode("↺", string(repeat), repeat != models.RepeatOff) + volume
}

func renderProgressBar(elapsed, total float64, width int) string {
	if width <= 0 {
		return ""
	}
	if total <= 0 {
		return theme.Faint.Render(strings.Repeat("━", width))
	}
	ratio := max(0, min(1, elapsed/total))
	filled := clamp(int(math.Round(ratio*float64(width))), 0, width)
	if filled == 0 || filled == width {
		return theme.Red.Render(strings.Repeat("━", filled)) + theme.Faint.Render(strings.Repeat("━", width-filled))
	}
	return theme.Red.Render(strings.Repeat("━", filled-1)+"●") + theme.Faint.Render(strings.Repeat("━", width-filled))
}

// hint is an action offered in the bottom line.
type hint struct{ action, key string }

func hintsFor(focus panel) []hint {
	var hints []hint
	switch focus {
	case panelTracks:
		hints = append(hints, hint{"Play", "enter"})
	case panelPlaylists, panelArtists, panelAlbums:
		hints = append(hints, hint{"Open", "enter"})
	}
	return append(hints,
		hint{"Pause", "space"}, hint{"Search", "/"}, hint{"Keybindings", "?"},
		hint{"Next/Prev", "n/p"}, hint{"Panels", "1-5"}, hint{"Quit", "q"})
}

// renderBottom draws the last line: the notice, or the main keys, and the brand.
func renderBottom(width int, focus panel, notice, brandText string) string {
	var left string
	switch {
	case notice != "":
		left = " " + theme.Active.Render(notice)
	default:
		room := width - ansi.StringWidth(brandText) - 3
		used := 0
		for i, h := range hintsFor(focus) {
			text := h.action + ": " + h.key
			if i > 0 {
				text = " | " + text
			}
			if used+ansi.StringWidth(text) > room {
				break
			}
			used += ansi.StringWidth(text)
			if i > 0 {
				left += theme.Faint.Render(" | ")
			}
			left += theme.Dim.Render(h.action+": ") + theme.Artist.Render(h.key)
		}
		left = " " + left
	}
	brandW := ansi.StringWidth(brandText) + 1
	return panels.Fit(left, max(0, width-brandW)) + theme.Dim.Render(brandText) + " "
}

func formatTime(secs float64) string {
	total := int(math.Round(max(0, secs)))
	return fmt.Sprintf("%d:%02d", total/60, total%60)
}

// encryptedLabel scrambles base for a moment every ~10 seconds.
func encryptedLabel(base string, tick int) string {
	const periodTicks = 67 // ticks are ~150 ms, see tickCmd
	const activeTicks = 14 // ~2.1 s

	phase := tick % periodTicks
	if tick < 0 || phase >= activeTicks {
		return base
	}

	glyphs := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%&*+-=?")
	out := []rune(base)
	for i, r := range out {
		// Keep some characters so that the label can still be guessed.
		if r != ' ' && (tick+i)%6 != 0 {
			out[i] = glyphs[(tick*17+i*31+phase*13)%len(glyphs)]
		}
	}
	return string(out)
}
