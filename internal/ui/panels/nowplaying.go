package panels

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"

	"SR-Player/internal/models"
)

type NowPlayingPanel struct {
	nowPlaying  models.NowPlaying
	artCache    string
	artRendered string
	artWidth    int
	width       int
	height      int
	focused     bool
}

var (
	npFocusedBorder = lipgloss.Color("#FA243C")
	npBlurredBorder = lipgloss.Color("#333333")

	appleAccentOnce    sync.Once
	appleAccentContent string

	trackTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	trackArtistStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FA243C"))

	trackAlbumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)

	infoLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	infoValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC"))

	volumeFilledStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FA243C"))

	volumeEmptyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#333333"))
)

func NewNowPlayingPanel(width, height int) NowPlayingPanel {
	return NowPlayingPanel{
		width:  width,
		height: height,
	}
}

func (n *NowPlayingPanel) SetNowPlaying(np models.NowPlaying) {
	n.nowPlaying = np
	n.ensureArtwork()
}

func (n *NowPlayingPanel) SetSize(width, height int) {
	if width != n.width || height != n.height {
		n.width = width
		n.height = height
		n.ensureArtwork()
	}
}

func (n *NowPlayingPanel) SetFocused(focused bool) {
	n.focused = focused
}

func (n *NowPlayingPanel) targetArtworkWidth() int {
	innerWidth := n.width - 6
	if innerWidth < 12 {
		innerWidth = 12
	}
	// Reserve extra vertical space for top offset + metadata + state + vol/format + full ascii + footer line.
	maxByHeight := (n.height - 20) * 2
	if maxByHeight < 20 {
		maxByHeight = 20
	}
	if innerWidth > maxByHeight {
		innerWidth = maxByHeight
	}
	if innerWidth > 38 {
		innerWidth = 38
	}
	return innerWidth
}

func (n *NowPlayingPanel) ensureArtwork() {
	target := n.targetArtworkWidth()
	if n.artRendered == "" || n.artCache != n.nowPlaying.ArtworkPath || n.artWidth != target {
		n.artCache = n.nowPlaying.ArtworkPath
		n.artWidth = target
		n.artRendered = renderArtwork(n.nowPlaying.ArtworkPath, target)
	}
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
			topSY := (cy * 2) * srcH / pixH
			botSY := (cy*2 + 1) * srcH / pixH

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

func renderAppleAccent(width int) string {
	appleAccentOnce.Do(func() {
		content, err := os.ReadFile("assets/am-ascii.txt")
		if err != nil {
			return
		}
		appleAccentContent = strings.TrimRight(string(content), "\n")
	})

	if appleAccentContent == "" {
		return ""
	}

	lines := strings.Split(appleAccentContent, "\n")
	for i, line := range lines {
		for j := 0; j < 4 && strings.HasPrefix(line, " "); j++ {
			line = strings.TrimPrefix(line, " ")
		}
		lines[i] = line
	}
	shifted := strings.Join(lines, "\n")

	colored := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FA243C")).
		Render(shifted)

	return lipgloss.Place(width, lipgloss.Height(shifted), lipgloss.Center, lipgloss.Top, colored)
}

func renderVolumeLine(width int, volume int) string {
	if volume < 0 {
		volume = 0
	}
	if volume > 100 {
		volume = 100
	}

	label := "Vol "
	pct := fmt.Sprintf(" %d%%", volume)
	barW := width - lipgloss.Width(label) - lipgloss.Width(pct) - 2
	if barW < 8 {
		barW = 8
	}

	filled := int(math.Round((float64(volume) / 100.0) * float64(barW)))
	if filled < 0 {
		filled = 0
	}
	if filled > barW {
		filled = barW
	}
	empty := barW - filled

	bar := volumeFilledStyle.Render(strings.Repeat("█", filled)) +
		volumeEmptyStyle.Render(strings.Repeat("░", empty))

	return infoLabelStyle.Render(label) + bar + infoValueStyle.Render(pct)
}

func normalizeFormat(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	if strings.Contains(v, "dolby atmos") || (strings.Contains(v, "dolby") && strings.Contains(v, "atmos")) {
		return "Dolby Atmos"
	}
	if strings.Contains(v, "lossless") || strings.Contains(v, "apple lossless") || strings.Contains(v, "alac") {
		return "Lossless"
	}
	return "Standard"
}

func encryptedFooterLabel(base string, tick int) string {
	// tick is emitted every ~150ms (see ui/tickCmd), so ~67 ticks ~= 10 seconds.
	const periodTicks = 67
	const activeTicks = 14 // ~2.1s of encrypted animation every 10s

	if periodTicks <= 0 || tick < 0 {
		return base
	}

	phase := tick % periodTicks
	if phase >= activeTicks {
		return base
	}

	glyphs := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%&*+-=?")
	src := []rune(base)
	out := make([]rune, len(src))

	for i, r := range src {
		if r == ' ' {
			out[i] = r
			continue
		}
		// Keep some original chars so the encrypted effect still hints at the label.
		if (tick+i)%6 == 0 {
			out[i] = r
			continue
		}
		idx := (tick*17 + i*31 + phase*13) % len(glyphs)
		out[i] = glyphs[idx]
	}

	return string(out)
}

// renderVisualizer displays distinct vertical columns dancing like a traditional EQ.
func renderVisualizer(tick int, width int, height int, isPlaying bool) string {
	if height <= 0 || width <= 0 {
		return ""
	}
	// Blocks growing from bottom to top vertically
	blocks := []string{" ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

	var out strings.Builder
	for y := 0; y < height; y++ {
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

				val := (v1+v2+2.0)/4.0 + noise // roughly 0.0 to 1.0
				if val > 1.0 {
					val = 1.0
				}
				if val < 0 {
					val = 0
				}

				h = int(val * float64(height*8))
			}

			// Invertimos la lógica para que la barra crezca desde la base (la base está en la línea 'height-1')
			baseVal := (height - 1 - y) * 8
			cellH := h - baseVal

			if cellH < 0 {
				cellH = 0
			}
			if cellH > 7 {
				cellH = 7
			}

			color := lipgloss.Color("#FA243C")

			if cellH == 0 {
				out.WriteString(" ")
			} else {
				out.WriteString(lipgloss.NewStyle().Foreground(color).Render(blocks[cellH]))
			}
		}
		if y < height-1 {
			out.WriteString("\n")
		}
	}
	return out.String()
}

func (n *NowPlayingPanel) View(tick int) string {
	innerWidth := n.width - 4
	if innerWidth < 20 {
		innerWidth = 20
	}

	np := n.nowPlaying
	track := np.Track
	isPlaying := np.State == models.StatePlaying

	art := n.artRendered
	if art == "" {
		n.ensureArtwork()
		art = n.artRendered
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
	apple := renderAppleAccent(innerWidth)
	volumeLine := renderVolumeLine(innerWidth, np.Volume)
	formatText := normalizeFormat(track.Format)
	formatLine := infoLabelStyle.Render("Format ") + infoValueStyle.Render(truncate(formatText, innerWidth-8))
	footerText := encryptedFooterLabel("Apple Music TUI", tick)
	footerLine := lipgloss.NewStyle().
		Width(innerWidth).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Render(footerText)

	stateIcon := "⏸"
	if isPlaying {
		stateIcon = "▶"
	}
	stateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FA243C")).Bold(true)
	stateLine := stateStyle.Render(stateIcon + "  " + string(np.State))

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		"",
		"",
		art,
		"",
		titleLine,
		artistLine,
		albumLine,
		"",
		stateLine,
		"",
		volumeLine,
		"",
		formatLine,
		"",
		"",
		apple,
		"",
		"",
		footerLine,
	)

	border := npBlurredBorder
	if n.focused {
		border = npFocusedBorder
	}
	return renderRectBoxAligned(content, n.width, n.height, border, lipgloss.Center, lipgloss.Top)
}

func RenderVisualizerPanel(width, height, tick int, isPlaying bool, focused bool) string {
	if height <= 0 || width <= 0 {
		return ""
	}

	border := lipgloss.Color("#333333")
	if focused {
		border = lipgloss.Color("#FA243C")
	}

	// renderRectBox reserves the inner drawing area; use that for visualizer generation.
	panelWidth := width - 2
	panelHeight := height - 2
	innerWidth := panelWidth - 2
	innerHeight := panelHeight - 2
	if innerWidth < 1 || innerHeight < 1 {
		return ""
	}
	vis := renderVisualizer(tick, innerWidth, innerHeight, isPlaying)
	return renderRectBox(vis, width, height, border, lipgloss.Bottom)
}
