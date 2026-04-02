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
	slFocusedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FA243C")).
			Padding(0, 1)

	slBlurredStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#333333")).
			Padding(0, 1)
)

func NewStringListPanel(width, height int, title string) StringListPanel {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#e1ca72")).
		BorderLeftForeground(lipgloss.Color("#e1ca72"))
	
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(lipgloss.Color("#CCCCCC"))
	
	delegate.ShowDescription = false

	l := list.New([]list.Item{}, delegate, width-4, height-2)
	l.Title = title
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)
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
	style := slBlurredStyle
	if s.focused {
		style = slFocusedStyle
	}
	return style.Width(s.width - 2).Height(s.height - 2).MaxWidth(s.width - 2).MaxHeight(s.height - 2).Render(s.list.View())
}
