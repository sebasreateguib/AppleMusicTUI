package panels

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"SR-Player/internal/ui/theme"
)

// SearchPrompt is the popup that asks what to search for.
type SearchPrompt struct {
	input textinput.Model
}

func NewSearchPrompt() SearchPrompt {
	ti := textinput.New()
	ti.Prompt = "❯ "
	ti.PromptStyle = theme.Active
	ti.Placeholder = "Song, artist or album…"
	ti.TextStyle = theme.Bold
	ti.PlaceholderStyle = theme.Dim
	ti.Cursor.Style = theme.Red
	ti.CharLimit = 100
	return SearchPrompt{input: ti}
}

// Open empties and focuses the prompt.
func (s *SearchPrompt) Open() tea.Cmd {
	s.input.SetValue("")
	return s.input.Focus()
}

func (s *SearchPrompt) Close() { s.input.Blur() }

func (s *SearchPrompt) Query() string { return strings.TrimSpace(s.input.Value()) }

func (s *SearchPrompt) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return cmd
}

func (s *SearchPrompt) View(screenW int) string {
	width := max(44, min(84, screenW*60/100))
	inner := width - 2
	const margin = "   "
	s.input.Width = inner - 2*len(margin) - 3 // prompt and cursor

	keys := theme.Active.Render("enter") + theme.Dim.Render(" search   ") +
		theme.Active.Render("esc") + theme.Dim.Render(" cancel")
	frame := Frame{Title: "Search library", Focused: true}
	return fill(frame.Render(width, 7, []string{
		"",
		margin + s.input.View(),
		margin + theme.Red.Render(strings.Repeat("━", inner-2*len(margin))),
		"",
		Center(keys, inner),
	}), theme.PopupBg)
}

// Binding is a line of the help popup. An empty Key makes it a section title.
type Binding struct {
	Key, Text string
}

// Help is the popup that lists every key.
type Help struct {
	bindings []Binding
	top      int
	rows     int
}

func NewHelp(bindings []Binding) Help { return Help{bindings: bindings} }

func (h *Help) Open() { h.top = 0 }

func (h *Help) HandleKey(key string) {
	switch key {
	case "up", "k":
		h.top--
	case "down", "j":
		h.top++
	}
	h.top = max(0, min(h.top, len(h.bindings)-h.rows))
}

func (h *Help) View(screenW, screenH int) string {
	width := max(30, min(56, screenW-4))
	h.rows = max(1, min(len(h.bindings), screenH-6))
	h.top = max(0, min(h.top, len(h.bindings)-h.rows))

	lines := make([]string, 0, h.rows)
	for _, b := range h.bindings[h.top : h.top+h.rows] {
		if b.Key == "" {
			lines = append(lines, " "+theme.Bold.Render(b.Text))
			continue
		}
		key := strings.Repeat(" ", max(0, 14-ansi.StringWidth(b.Key))) + b.Key
		lines = append(lines, theme.Active.Render(key)+"  "+b.Text)
	}
	frame := Frame{
		Title: "Keybindings", Footer: "esc close", Focused: true,
		Top: h.top, Visible: h.rows, Total: len(h.bindings),
	}
	return fill(frame.Render(width, h.rows+2, lines), theme.PopupBg)
}
