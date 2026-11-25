package panels

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/sebasreateguib/applemusictui/internal/ui/theme"
)

// Frame is the box around every panel. Its title, position and scrollbar live
// in the border so that the whole inside belongs to the content.
type Frame struct {
	Key     string // key that focuses the panel
	Title   string
	Info    string // right side of the top border, already styled
	Footer  string // right side of the bottom border
	Focused bool
	Faded   bool // behind a popup

	// Scrollbar: first visible item, visible items and number of items.
	Top, Visible, Total int
}

// Render draws the frame with exactly width x height cells.
func (f Frame) Render(width, height int, lines []string) string {
	if width < 4 || height < 2 {
		return Blank(width, height)
	}
	border, title := theme.Faint, theme.Text
	if f.Focused {
		border, title = theme.Red, theme.Active
	}
	if f.Faded {
		border, title = theme.Backdrop, theme.Backdrop
	}
	inner := width - 2

	out := make([]string, 0, height)
	out = append(out, f.topBorder(inner, border, title))

	side := border.Render("│")
	thumb := theme.Dim.Render("▐")
	if f.Focused {
		thumb = theme.Red.Render("▐")
	}
	from, to := f.scrollThumb(height - 2)
	for i := 0; i < height-2; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		right := side
		if i >= from && i < to {
			right = thumb
		}
		out = append(out, side+Fit(line, inner)+right)
	}

	footer := f.Footer
	if footer != "" {
		footer = " " + ansi.Truncate(footer, max(0, inner-4), "…") + " "
	}
	fill := max(0, inner-1-ansi.StringWidth(footer))
	out = append(out, border.Render("╰"+strings.Repeat("─", fill))+theme.Dim.Render(footer)+border.Render("─╯"))
	return strings.Join(out, "\n")
}

func (f Frame) topBorder(inner int, border, title lipgloss.Style) string {
	var sb strings.Builder
	used := 1
	sb.WriteString(border.Render("╭─"))
	if f.Key != "" {
		key := "[" + f.Key + "]"
		sb.WriteString(title.Render(key) + border.Render("─"))
		used += len(key) + 1
	}

	info, infoW := f.Info, ansi.StringWidth(f.Info)
	if infoW > 0 {
		info += border.Render("─")
		infoW++
	}
	if used+infoW+8 > inner { // not worth squeezing the title for it
		info, infoW = "", 0
	}
	text := ansi.Truncate(f.Title, max(0, inner-used-infoW-1), "…")
	sb.WriteString(title.Render(text))
	used += ansi.StringWidth(text)

	sb.WriteString(border.Render(strings.Repeat("─", max(0, inner-used-infoW))))
	sb.WriteString(info)
	sb.WriteString(border.Render("╮"))
	return sb.String()
}

// scrollThumb returns the rows [from, to) of the scrollbar's thumb, empty when
// everything is visible.
func (f Frame) scrollThumb(rows int) (from, to int) {
	if rows <= 0 || f.Visible <= 0 || f.Total <= f.Visible {
		return 0, 0
	}
	size := max(1, rows*f.Visible/f.Total)
	from = f.Top * (rows - size) / (f.Total - f.Visible)
	return from, from + size
}

// Fit cuts or pads s, which may be styled, to exactly width cells.
func Fit(s string, width int) string {
	w := ansi.StringWidth(s)
	if w > width {
		s = ansi.Truncate(s, width, "…")
		w = ansi.StringWidth(s)
	}
	if w < width {
		s += strings.Repeat(" ", width-w)
	}
	return s
}

// Center places s, which may be styled, in the middle of width cells.
func Center(s string, width int) string {
	w := ansi.StringWidth(s)
	if w >= width {
		return Fit(s, width)
	}
	return Fit(strings.Repeat(" ", (width-w)/2)+s, width)
}

// Blank returns an empty area.
func Blank(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	row := strings.Repeat(" ", width)
	return strings.TrimSuffix(strings.Repeat(row+"\n", height), "\n")
}

// message centers a dim text in an otherwise empty panel.
func message(text string, width, rows int) []string {
	lines := make([]string, max(0, rows))
	if len(lines) > 0 {
		lines[(len(lines)-1)/2] = Center(theme.Dim.Render(ansi.Truncate(text, max(0, width-2), "…")), width)
	}
	return lines
}

// position is the "3 of 40" shown in the bottom border of lists.
func position(index, total int) string {
	if total == 0 {
		return ""
	}
	return fmt.Sprintf("%d of %d", index+1, total)
}

// Fade strips the colors of lines, for the screen behind a popup.
func Fade(lines []string) []string {
	faded := make([]string, len(lines))
	for i, l := range lines {
		faded[i] = theme.Backdrop.Render(ansi.Strip(l))
	}
	return faded
}

// fill paints the background of a whole block that has styled text inside.
func fill(block string, color lipgloss.TerminalColor) string {
	on, _, ok := strings.Cut(lipgloss.NewStyle().Background(color).Render("x"), "x")
	if !ok || on == "" {
		return block
	}
	// Every style ends with a reset, which also ends the background.
	block = strings.NewReplacer("\x1b[0m", "\x1b[0m"+on, "\x1b[m", "\x1b[m"+on).Replace(block)
	return on + strings.ReplaceAll(block, "\n", "\x1b[0m\n"+on) + "\x1b[0m"
}

// Overlay draws popup over base, which is width x height. The popup is centered
// horizontally, with the given fraction of the free rows above it.
func Overlay(base, popup string, width, height int, above float64) string {
	rows := strings.Split(popup, "\n")
	popupW := ansi.StringWidth(rows[0])
	x, y := max(0, (width-popupW)/2), max(0, int(float64(height-len(rows))*above))

	lines := strings.Split(base, "\n")
	for i, row := range rows {
		if y+i >= len(lines) {
			break
		}
		line := lines[y+i]
		left := ansi.Truncate(line, x, "")
		left += "\x1b[0m" + strings.Repeat(" ", max(0, x-ansi.StringWidth(left)))

		// A wide character under an edge of the popup is replaced by a space.
		rest := max(0, width-x-popupW)
		right := ansi.TruncateLeft(line, x+popupW, "")
		if ansi.StringWidth(right) > rest {
			right = ansi.TruncateLeft(line, x+popupW+1, "")
		}
		right = strings.Repeat(" ", max(0, rest-ansi.StringWidth(right))) + right

		lines[y+i] = left + row + "\x1b[0m" + right
	}
	return strings.Join(lines, "\n")
}
