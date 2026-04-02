package panels

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"SR-Player/internal/models"
)

type NowPlayingPanel struct {
	nowPlaying  models.NowPlaying
	artCache    string
	artRendered string
	width       int
	height      int
	focused     bool
}

var (
	npFocusedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#1DB954")).
			Padding(0, 1)

	npBlurredStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#333333")).
			Padding(0, 1)

	trackTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	trackArtistStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1DB954"))

	trackAlbumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)
)

func NewNowPlayingPanel(width, height int) NowPlayingPanel {
	return NowPlayingPanel{
		width:  width,
		height: height,
	}
}

func (n *NowPlayingPanel) SetNowPlaying(np models.NowPlaying) {
	n.nowPlaying = np
	if np.ArtworkPath != n.artCache {
		n.artCache = np.ArtworkPath
		n.artRendered = renderArtwork(np.ArtworkPath, 28) // Fixed width 28 chars
	}
}

func (n *NowPlayingPanel) SetSize(width, height int) {
	if width != n.width || height != n.height {
		n.width = width
		n.height = height
		if n.artRendered == "" {
			n.artRendered = renderArtwork(n.artCache, 28)
		}
	}
}

func (n *NowPlayingPanel) SetFocused(focused bool) {
	n.focused = focused
}

func renderArtwork(path string, charWidth int) string {
	if path == "" {
		return placeholderArt(charWidth)
	}

	f, err := os.Open(path)
	if err != nil {
		return placeholderArt(charWidth)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return placeholderArt(charWidth)
	}

	bounds := img.Bounds()
	srcW := bounds.Max.X - bounds.Min.X
	srcH := bounds.Max.Y - bounds.Min.Y

	charHeight := charWidth / 2
	if charHeight < 5 {
		charHeight = 5
	}
	pixW := charWidth
	pixH := charHeight * 2

	var sb strings.Builder

	for cy := 0; cy < charHeight; cy++ {
		for cx := 0; cx < pixW; cx++ {
			sx := cx * srcW / pixW
			topSY := (cy*2) * srcH / pixH
			botSY := (cy*2+1) * srcH / pixH

			topR, topG, topB, _ := img.At(bounds.Min.X+sx, bounds.Min.Y+topSY).RGBA()
			botR, botG, botB, _ := img.At(bounds.Min.X+sx, bounds.Min.Y+botSY).RGBA()

			tr, tg, tb := topR>>8, topG>>8, topB>>8
			br, bg, bb := botR>>8, botG>>8, botB>>8

			bgHex := fmt.Sprintf("#%02x%02x%02x", tr, tg, tb)
			fgHex := fmt.Sprintf("#%02x%02x%02x", br, bg, bb)

			cell := lipgloss.NewStyle().
				Foreground(lipgloss.Color(fgHex)).
				Background(lipgloss.Color(bgHex)).
				Render("▄")

			sb.WriteString(cell)
		}
		if cy < charHeight-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func safeRepeat(s string, count int) string {
	if count <= 0 {
		return ""
	}
	return strings.Repeat(s, count)
}

func placeholderArt(charWidth int) string {
	if charWidth < 10 {
		charWidth = 10
	}
	charHeight := charWidth / 2
	if charHeight < 5 {
		charHeight = 5
	}

	line := safeRepeat("░", charWidth)
	mid := charHeight / 2
	label := "  No Artwork  "
	labelLen := len([]rune(label))

	var lines []string
	for i := 0; i < charHeight; i++ {
		if i == mid {
			pad := (charWidth - labelLen) / 2
			if pad < 0 {
				pad = 0
			}
			right := charWidth - pad - labelLen
			if right < 0 {
				right = 0
			}
			row := safeRepeat("░", pad) + label + safeRepeat("░", right)
			lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")).Render(row))
		} else {
			lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render(line))
		}
	}
	return strings.Join(lines, "\n")
}

func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}

// renderVisualizer displays distinct vertical columns dancing like a traditional EQ.
func renderVisualizer(tick int, width int, height int, isPlaying bool) string {
	if height <= 0 || width <= 0 {
		return ""
	}
	// Blocks growing from bottom to top vertically
	blocks := []string{" ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

	var out strings.Builder
	for y := height - 1; y >= 0; y-- {
		for x := 0; x < width; x++ {
			// Leave a space between bars so they look like distinct vertical columns
			if x%2 != 0 {
				out.WriteString(" ")
				continue
			}

			h := 0
			if isPlaying {
				v1 := math.Sin(float64(x)*0.8 + float64(tick)*0.6)
				v2 := math.Cos(float64(x)*1.3 - float64(tick)*0.4)

				noise := float64((tick*x)%5) * 0.1

				val := (v1 + v2 + 2.0) / 4.0 + noise // roughly 0.0 to 1.0
				if val > 1.0 {
					val = 1.0
				}
				if val < 0 {
					val = 0
				}

				h = int(val * float64(height*8))
			}

			cellH := h - (height-1-y)*8
			if cellH < 0 {
				cellH = 0
			}
			if cellH > 7 {
				cellH = 7
			}

			color := lipgloss.Color("#1DB954")

			if cellH == 0 {
				out.WriteString(" ")
			} else {
				out.WriteString(lipgloss.NewStyle().Foreground(color).Render(blocks[cellH]))
			}
		}
		if y > 0 {
			out.WriteString("\n")
		}
	}
	return out.String()
}

func (n *NowPlayingPanel) View(tick int) string {
	style := npBlurredStyle
	if n.focused {
		style = npFocusedStyle
	}

	innerWidth := n.width - 4
	if innerWidth < 20 {
		innerWidth = 20
	}

	np := n.nowPlaying
	track := np.Track
	isPlaying := np.State == models.StatePlaying

	art := n.artRendered
	if art == "" {
		art = renderArtwork(n.artCache, 28) // fixed cover width
	}

	title := truncate(track.Name, innerWidth)
	if title == "" {
		title = "No track playing"
	}
	artist := truncate(track.Artist, innerWidth)
	album := truncate(track.Album, innerWidth)

	titleLine := trackTitleStyle.Render(title)
	artistLine := trackArtistStyle.Render(artist)
	albumLine := trackAlbumStyle.Render(album)

	stateIcon := "⏸"
	if isPlaying {
		stateIcon = "▶"
	}
	stateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1DB954")).Bold(true)
	stateLine := stateStyle.Render(stateIcon + "  " + string(np.State))

	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))
	keys := keyStyle.Render("[space] play/pause  [n/p] next/prev")

	// Visualizer calculation:
	// Total internal height = n.height - 2
	// Static contents: Art(14), Empties(4), Text(3), State(1), Keys(1) -> 23 lines used
	visHeight := n.height - 25
	var vis string
	if visHeight > 0 {
		visWidth := innerWidth
		if visWidth > 40 {
			visWidth = 40
		}
		vis = renderVisualizer(tick, visWidth, visHeight, isPlaying)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		art,
		"",
		titleLine,
		artistLine,
		albumLine,
		"",
		stateLine,
		"",
		vis,
		"",
		keys,
	)

	return style.Width(n.width - 2).Height(n.height - 2).MaxWidth(n.width - 2).MaxHeight(n.height - 2).Render(content)
}
