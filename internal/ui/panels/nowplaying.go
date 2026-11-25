package panels

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/sebasreateguib/applemusictui/assets"
	"github.com/sebasreateguib/applemusictui/internal/artwork"
	"github.com/sebasreateguib/applemusictui/internal/models"
	"github.com/sebasreateguib/applemusictui/internal/ui/theme"
)

const (
	maxArtRows = 20
	minArtRows = 5
	// nowPlayingChrome is every row but the artwork: the frame, a gap, the gap
	// below the artwork, title, artist, album, a gap and the format.
	nowPlayingChrome = 9
)

// NowPlayingPanel shows the artwork and details of the current track.
type NowPlayingPanel struct {
	nowPlaying  models.NowPlaying
	artCache    string
	artRendered string
	artBlurred  string // artwork behind a popup
	blurredFor  blurKey
	artWidth    int
	imageID     uint32 // image uploaded to the terminal, 0 to draw pixels
	artImageID  uint32 // image used by artRendered
	width       int
	height      int
}

// blurKey is what artBlurred depends on.
type blurKey struct {
	path  string
	width int
}

func NewNowPlayingPanel() NowPlayingPanel { return NowPlayingPanel{} }

// NowPlayingHeight returns the height the panel needs for the largest artwork
// that fits, leaving at least reserved rows of maxHeight free.
func NowPlayingHeight(width, maxHeight, reserved int) int {
	rows := min((width-4)/2, maxArtRows, maxHeight-reserved-nowPlayingChrome)
	if rows < minArtRows {
		rows = -1 // no artwork, nor the gap below it
	}
	return rows + nowPlayingChrome
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

// SetImage shows an image already uploaded to the terminal, or goes back to
// the pixelated artwork when id is 0.
func (n *NowPlayingPanel) SetImage(id uint32) {
	n.imageID = id
	n.ensureArtwork()
}

// artworkWidth returns the width of the artwork in cells, 0 if it does not fit.
func (n *NowPlayingPanel) artworkWidth() int {
	rows := min((n.width-4)/2, maxArtRows, n.height-nowPlayingChrome)
	if rows < minArtRows {
		return 0
	}
	return rows * 2
}

func (n *NowPlayingPanel) ensureArtwork() {
	target := n.artworkWidth()
	if n.artCache == n.nowPlaying.ArtworkPath && n.artWidth == target && n.artImageID == n.imageID {
		return
	}
	n.artCache = n.nowPlaying.ArtworkPath
	n.artWidth = target
	n.artImageID = n.imageID

	// Not when a popup opens: the file may be gone by then.
	if key := (blurKey{n.artCache, target}); key != n.blurredFor {
		n.blurredFor = key
		n.artBlurred = ""
		if target > 0 {
			n.artBlurred = renderBlurredArtwork(n.artCache, target)
		}
	}

	switch {
	case target == 0:
		n.artRendered = ""
	case n.imageID != 0:
		rows := artworkRows(target)
		artwork.Place(n.imageID, target, rows)
		n.artRendered = artwork.Cells(n.imageID, target, rows)
	default:
		n.artRendered = renderArtwork(n.nowPlaying.ArtworkPath, target)
	}
}

// title returns the only text that depends on the animation tick.
func (n *NowPlayingPanel) title(tick int) string {
	return marqueeText(n.nowPlaying.Track.Name, max(1, n.width-4), tick)
}

// NowPlayingKey identifies what NowPlayingPanel.View renders.
type NowPlayingKey struct {
	track         models.Track
	artCache      string
	artWidth      int
	artImageID    uint32
	width, height int
	title         string
	faded         bool
}

// ViewKey returns equal keys when View would render the same.
func (n *NowPlayingPanel) ViewKey(tick int, faded bool) NowPlayingKey {
	return NowPlayingKey{
		track:      n.nowPlaying.Track,
		artCache:   n.artCache,
		artWidth:   n.artWidth,
		artImageID: n.artImageID,
		width:      n.width,
		height:     n.height,
		title:      n.title(tick),
		faded:      faded,
	}
}

// View renders the panel. Faded is for the screen behind a popup: it has no
// colors, and the artwork is out of focus.
func (n *NowPlayingPanel) View(tick int, faded bool) string {
	inner := n.width - 2
	frame := Frame{Title: "Now Playing", Faded: faded}
	fade := func(lines ...string) []string {
		if faded {
			return Fade(lines)
		}
		return lines
	}
	track := n.nowPlaying.Track

	if track.Name == "" {
		lines := []string{""}
		if logo := appleLogo(); n.height >= len(logo)+6 {
			for _, l := range logo {
				lines = append(lines, Center(theme.Red.Render(l), inner))
			}
			lines = append(lines, "")
		}
		lines = append(lines, Center(theme.Dim.Render("Nothing playing"), inner))
		return frame.Render(n.width, n.height, fade(lines...))
	}

	lines := []string{""}
	art := n.artRendered
	if faded {
		art = n.artBlurred
	}
	if art != "" {
		for _, l := range strings.Split(art, "\n") {
			lines = append(lines, Center(l, inner))
		}
		lines = append(lines, "")
	}
	textW := max(0, inner-2)
	lines = append(lines, fade(
		Center(theme.Bold.Render(n.title(tick)), inner),
		Center(theme.Artist.Render(ansi.Truncate(track.Artist, textW, "…")), inner),
		Center(theme.Dim.Render(ansi.Truncate(track.Album, textW, "…")), inner),
		"",
		Center(renderFormatBadge(normalizeFormat(track.Format)), inner),
	)...)
	return frame.Render(n.width, n.height, lines)
}

var (
	appleLogoOnce  sync.Once
	appleLogoLines []string
)

// appleLogo returns the ASCII logo as lines of equal width.
func appleLogo() []string {
	appleLogoOnce.Do(func() {
		lines := strings.Split(strings.TrimRight(assets.Logo, "\n"), "\n")
		width := 0
		for i, l := range lines {
			lines[i] = strings.TrimPrefix(l, "     ")
			width = max(width, len(lines[i]))
		}
		for i := range lines {
			lines[i] = Fit(lines[i], width)
		}
		appleLogoLines = lines
	})
	return appleLogoLines
}

// artworkRows returns the height, in cells, of a square charWidth cells wide.
func artworkRows(charWidth int) int {
	return max(5, charWidth/2)
}

func renderArtwork(path string, charWidth int) string {
	img, err := decodeImage(path)
	if err != nil {
		return placeholderArt(charWidth)
	}

	bounds := img.Bounds()
	srcW := bounds.Max.X - bounds.Min.X
	srcH := bounds.Max.Y - bounds.Min.Y

	charHeight := artworkRows(charWidth)
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

// renderBlurredArtwork draws the artwork out of focus and darkened.
func renderBlurredArtwork(path string, charWidth int) string {
	const (
		samples = 6 // per side of each pixel drawn
		passes  = 2
		light   = 0.38
	)
	rows := artworkRows(charWidth)
	img, err := decodeImage(path)
	if err != nil {
		return strings.Join(Fade(strings.Split(placeholderArt(charWidth), "\n")), "\n")
	}

	// One pixel for each half of a cell, averaging the part of the image it covers.
	type rgb [3]float64
	w, h := charWidth, rows*2
	b := img.Bounds()
	pix := make([]rgb, w*h)
	for y := range h {
		for x := range w {
			var sum rgb
			for s := range samples * samples {
				sx := b.Min.X + (x*samples+s%samples)*b.Dx()/(w*samples)
				sy := b.Min.Y + (y*samples+s/samples)*b.Dy()/(h*samples)
				pr, pg, pb, _ := img.At(sx, sy).RGBA()
				sum[0] += float64(pr >> 8)
				sum[1] += float64(pg >> 8)
				sum[2] += float64(pb >> 8)
			}
			for c := range sum {
				pix[y*w+x][c] = sum[c] / (samples * samples)
			}
		}
	}

	for range passes {
		blurred := make([]rgb, len(pix))
		for y := range h {
			for x := range w {
				var sum rgb
				n := 0.0
				for ny := max(0, y-1); ny <= min(h-1, y+1); ny++ {
					for nx := max(0, x-1); nx <= min(w-1, x+1); nx++ {
						for c := range sum {
							sum[c] += pix[ny*w+nx][c]
						}
						n++
					}
				}
				for c := range sum {
					blurred[y*w+x][c] = sum[c] / n
				}
			}
		}
		pix = blurred
	}

	hex := func(p rgb) lipgloss.Color {
		return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", int(p[0]*light), int(p[1]*light), int(p[2]*light)))
	}
	lines := make([]string, rows)
	for y := range lines {
		var sb strings.Builder
		for x := range w {
			top, bottom := pix[2*y*w+x], pix[(2*y+1)*w+x]
			sb.WriteString(lipgloss.NewStyle().Background(hex(top)).Foreground(hex(bottom)).Render("▄"))
		}
		lines[y] = sb.String()
	}
	return strings.Join(lines, "\n")
}

func decodeImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func placeholderArt(charWidth int) string {
	if charWidth < 10 {
		charWidth = 10
	}
	charHeight := artworkRows(charWidth)

	line := strings.Repeat("░", charWidth)
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
			row := strings.Repeat("░", pad) + label + strings.Repeat("░", right)
			lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")).Render(row))
		} else {
			lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render(line))
		}
	}
	return strings.Join(lines, "\n")
}

// marqueeText scrolls a long string through a fixed-width window using the
// animation tick. Strings that fit within maxWidth are returned unchanged.
// The animation pauses at each end before reversing direction.
func marqueeText(s string, maxWidth int, tick int) string {
	runes := []rune(s)
	if len(runes) <= maxWidth || maxWidth <= 0 {
		return s
	}

	scrollLen := len(runes) - maxWidth
	const pauseTicks = 20  // ~3 s pause at each end (tick = 150 ms)
	const ticksPerStep = 5 // ~750 ms per character step

	totalCycle := (pauseTicks * 2) + (scrollLen * ticksPerStep * 2)
	if totalCycle <= 0 {
		return string(runes[:maxWidth])
	}
	phase := tick % totalCycle

	var offset int
	switch {
	case phase < pauseTicks:
		offset = 0
	case phase < pauseTicks+scrollLen*ticksPerStep:
		offset = (phase - pauseTicks) / ticksPerStep
	case phase < pauseTicks*2+scrollLen*ticksPerStep:
		offset = scrollLen
	default:
		offset = scrollLen - (phase-pauseTicks*2-scrollLen*ticksPerStep)/ticksPerStep
	}

	if offset < 0 {
		offset = 0
	}
	if offset > scrollLen {
		offset = scrollLen
	}
	return string(runes[offset : offset+maxWidth])
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

// renderFormatBadge returns a styled badge for the given audio format string.
func renderFormatBadge(format string) string {
	switch format {
	case "Dolby Atmos":
		return lipgloss.NewStyle().
			Background(lipgloss.Color("#2D0A5A")).
			Foreground(lipgloss.Color("#D4AAFF")).
			Bold(true).
			Padding(0, 1).
			Render("✦ Dolby Atmos")
	case "Lossless":
		return lipgloss.NewStyle().
			Background(lipgloss.Color("#0A3D1F")).
			Foreground(lipgloss.Color("#6EE89A")).
			Bold(true).
			Padding(0, 1).
			Render("◈ Lossless")
	default:
		return lipgloss.NewStyle().
			Background(lipgloss.Color("#1A1D21")).
			Foreground(lipgloss.Color("#6B7280")).
			Padding(0, 1).
			Render("◦ Standard")
	}
}
