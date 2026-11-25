package ui

import (
	"strings"

	"github.com/sebasreateguib/applemusictui/internal/ui/panels"
)

const (
	minWidth  = 60
	minHeight = 16

	// wideWidth is the width from which Now Playing and Up Next get a column.
	wideWidth = 100

	playerHeight = 3
	bottomHeight = 1
	sideLists    = 3
	minQueueRows = 6
)

// Layout is the size of every panel for a given terminal size.
type Layout struct {
	width, height int
	bodyH         int

	leftW, mainW, rightW int // rightW is 0 on narrow terminals

	tracksH, visH       int
	nowPlayingH, queueH int
	sideH               [sideLists]int
}

// newLayout computes the layout. activeSide is the list on the left that gets
// the most room, and playlists the number of rows of the first one.
func newLayout(width, height, activeSide, playlists int) Layout {
	l := Layout{width: width, height: height}
	l.bodyH = max(0, height-playerHeight-bottomHeight)

	l.leftW = clamp(width*24/100, 22, 38)
	if width >= wideWidth {
		l.rightW = clamp(width*27/100, 28, 44)
	}
	l.mainW = width - l.leftW - l.rightW

	if l.bodyH >= 20 {
		l.visH = clamp(l.bodyH/4, 6, 12)
	}
	l.tracksH = l.bodyH - l.visH

	l.nowPlayingH = panels.NowPlayingHeight(l.rightW, l.bodyH, minQueueRows)
	l.queueH = l.bodyH - l.nowPlayingH
	l.sideH = sideHeights(l.bodyH, activeSide, playlists)
	return l
}

// sideHeights splits the left column. The active list grows; on short
// terminals the others shrink to their selected row.
func sideHeights(height, active, playlists int) [sideLists]int {
	var h [sideLists]int
	if height < 18 {
		for i := range h {
			h[i] = 3
		}
		h[active] = height - 3*(sideLists-1)
		return h
	}

	limit := height / 4
	if active == 0 {
		limit = height / 2
	}
	h[0] = clamp(playlists+2, 4, limit)
	rest := height - h[0]
	switch active {
	case 1:
		h[1] = rest * 2 / 3
	case 2:
		h[1] = rest / 3
	default:
		h[1] = rest / 2
	}
	h[2] = rest - h[1]
	return h
}

func (l Layout) wide() bool { return l.rightW > 0 }

func (l Layout) tooSmall() bool { return l.width < minWidth || l.height < minHeight }

// queueSize returns the size of the Up Next panel, which takes the place of the
// main column on narrow terminals.
func (l Layout) queueSize() (width, height int) {
	if l.wide() {
		return l.rightW, l.queueH
	}
	return l.mainW, l.bodyH
}

// joinColumns puts columns with the same number of lines side by side.
func joinColumns(columns ...[]string) []string {
	lines := make([]string, len(columns[0]))
	var sb strings.Builder
	for i := range lines {
		sb.Reset()
		for _, c := range columns {
			sb.WriteString(c[i])
		}
		lines[i] = sb.String()
	}
	return lines
}

func clamp(v, lo, hi int) int { return max(lo, min(v, hi)) }
