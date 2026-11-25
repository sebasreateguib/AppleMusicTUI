package ui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	"SR-Player/internal/models"
)

// section keeps the lines a part of the screen was last rendered to. The
// visualizer redraws the screen 60 times per second, while everything around it
// only changes a few times per second.
type section struct {
	key   any
	lines []string
}

// get returns the section's lines, rendering them again only when key, a
// comparable value with everything the output depends on, has changed.
func (s *section) get(key any, render func() []string) []string {
	if s.lines == nil || s.key != key {
		s.key = key
		s.lines = render()
	}
	return s.lines
}

// viewCache is shared by the successive copies of App.
type viewCache struct {
	sides                                     [sideLists]section
	tracks, nowPlaying, queue, player, bottom section
}

// The keys below identify what each section shows; size is part of them
// because it also decides how the lines are cut and padded.
type listKey struct {
	width, height, revision int
}

type playerKey struct {
	width     int
	np        models.NowPlaying
	withTrack bool
}

type bottomKey struct {
	width             int
	focus             panel
	notice, brandText string
}

// fitLines splits s in lines exactly width cells wide. With a positive height
// it also returns exactly that many lines.
func fitLines(s string, width, height int) []string {
	lines := strings.Split(s, "\n")
	if height > 0 {
		if len(lines) > height {
			lines = lines[:height]
		}
		for len(lines) < height {
			lines = append(lines, "")
		}
	}
	for i, line := range lines {
		w := ansi.StringWidth(line)
		if w > width {
			line = ansi.Truncate(line, width, "")
			w = ansi.StringWidth(line)
		}
		if w < width {
			line += strings.Repeat(" ", width-w)
		}
		lines[i] = line
	}
	return lines
}
