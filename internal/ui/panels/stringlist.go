package panels

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type stringItem struct {
	value string
}

func (s stringItem) Title() string       { return "  " + s.value }
func (s stringItem) Description() string { return "" }
func (s stringItem) FilterValue() string { return s.value }

type StringListPanel struct {
	list    list.Model
	width   int
	height  int
	focused bool
	title   string
}

var (
	slFocusedBorder = lipgloss.Color("#FA243C")
	slBlurredBorder = lipgloss.Color("#2E3237")
)

func NewStringListPanel(width, height int, title string) StringListPanel {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		BorderLeftForeground(lipgloss.Color("#e1ca72"))

	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(lipgloss.Color("#C6CBD1"))

	delegate.ShowDescription = false

	l := list.New([]list.Item{}, delegate, width-4, height-2)
	l.Title = title
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)
	l.Styles.StatusBar = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666B72"))
	l.SetShowHelp(false)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	return StringListPanel{
		list:   l,
		width:  width,
		height: height,
		title:  title,
	}
}

func (s *StringListPanel) SetItems(items []string, title string) {
	var listItems []list.Item
	for _, str := range items {
		listItems = append(listItems, stringItem{value: str})
	}
	s.list.Title = title
	s.title = title
	s.list.SetItems(listItems)
	s.list.Select(0)
}

func (s *StringListPanel) SelectedItem() string {
	item := s.list.SelectedItem()
	if item == nil {
		return ""
	}
	si, ok := item.(stringItem)
	if !ok {
		return ""
	}
	return si.value
}

func (s *StringListPanel) SetSize(width, height int) {
	s.width = width
	s.height = height
	s.list.SetSize(width-4, height-2)
}

func (s *StringListPanel) SetFocused(focused bool) {
	s.focused = focused
}

func (s *StringListPanel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return cmd
}

func (s *StringListPanel) View() string {
	border := slBlurredBorder
	if s.focused {
		border = slFocusedBorder
	}
	return renderRectBox(s.list.View(), s.width, s.height, border, lipgloss.Top)
}
