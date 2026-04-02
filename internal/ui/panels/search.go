package panels

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"SR-Player/internal/models"
)

type searchResultItem struct {
	track models.Track
}

func padSearchTruncate(s string, width int) string {
	runes := []rune(s)
	if len(runes) > width {
		if width > 1 {
			return string(runes[:width-1]) + "…"
		}
		return ""
	}
	return s + string(make([]rune, width-len(runes)))
}

func (s searchResultItem) Title() string {
	title := padSearchTruncate(s.track.Name, 35)
	artist := padSearchTruncate(s.track.Artist, 25)
	album := padSearchTruncate(s.track.Album, 25)
	return "  " + title + " " + artist + " " + album
}
func (s searchResultItem) Description() string { return "" }
func (s searchResultItem) FilterValue() string { return s.track.Name }

type SearchPanel struct {
	input   textinput.Model
	results list.Model
	width   int
	height  int
	focused bool
	query   string
}

func NewSearchPanel(width, height int) SearchPanel {
	ti := textinput.New()
	ti.Placeholder = "Search tracks, artists, albums... (Press / to focus)"
	ti.CharLimit = 100

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#e1ca72")).
		BorderLeftForeground(lipgloss.Color("#e1ca72"))

	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(lipgloss.Color("#CCCCCC"))
	delegate.ShowDescription = false

	l := list.New([]list.Item{}, delegate, width-4, height-2)
	l.Title = "Search Results"
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return SearchPanel{
		input:   ti,
		results: l,
		width:   width,
		height:  height,
	}
}

func (s *SearchPanel) SetResults(tracks []models.Track) {
	items := make([]list.Item, len(tracks))
	for i, t := range tracks {
		items[i] = searchResultItem{track: t}
	}
	s.results.SetItems(items)
}

func (s *SearchPanel) SetSize(width, height int) {
	s.width = width
	s.height = height
}

func (s *SearchPanel) SetFocused(focused bool) {
	s.focused = focused
	if focused {
		s.input.Focus()
	} else {
		s.input.Blur()
	}
}

func (s *SearchPanel) GetQuery() string {
	return s.input.Value()
}

func (s *SearchPanel) IsTyping() bool {
	return s.input.Focused()
}

func (s *SearchPanel) SelectedTrack() *models.Track {
	item := s.results.SelectedItem()
	if item == nil {
		return nil
	}
	si, ok := item.(searchResultItem)
	if !ok {
		return nil
	}
	return &si.track
}

func (s *SearchPanel) Update(msg tea.Msg) (tea.Cmd, bool) {
	var cmd tea.Cmd
	var searchTriggered bool

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if s.input.Focused() {
				searchTriggered = true
				s.input.Blur()
			}
		default:
			if s.input.Focused() {
				s.input, cmd = s.input.Update(msg)
			} else {
				s.results, cmd = s.results.Update(msg)
			}
		}
	default:
		if s.input.Focused() {
			s.input, cmd = s.input.Update(msg)
		} else {
			s.results, cmd = s.results.Update(msg)
		}
	}

	return cmd, searchTriggered
}

// ViewInput renders ONLY the top search bar.
func (s *SearchPanel) ViewInput(width int, isFocused bool) string {
	s.input.Width = width - 4
	style := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#333333"))
	if isFocused {
		style = style.BorderForeground(lipgloss.Color("#1DB954"))
	}
	return style.Width(width - 2).Height(1).MaxWidth(width - 2).MaxHeight(3).Render(s.input.View())
}

// ViewResults renders ONLY the center list of results.
func (s *SearchPanel) ViewResults(width, height int, isFocused bool) string {
	s.results.SetSize(width-4, height-2)
	style := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#333333"))
	if isFocused {
		style = style.BorderForeground(lipgloss.Color("#1DB954"))
	}
	return style.Width(width - 2).Height(height - 2).MaxWidth(width - 2).MaxHeight(height - 2).Render(s.results.View())
}
