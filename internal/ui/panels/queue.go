package panels

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"SR-Player/internal/models"
)

type queueItem struct {
	qt models.QueueTrack
}

func (q queueItem) Title() string {
	prefix := "  "
	if q.qt.IsCurrent {
		prefix = "▶ "
	}
	return prefix + q.qt.Track.Name
}

func (q queueItem) Description() string {
	return fmt.Sprintf("  %s", q.qt.Track.Artist)
}

func (q queueItem) FilterValue() string { return q.qt.Track.Name }

type QueuePanel struct {
	list    list.Model
	width   int
	height  int
	focused bool
}

var (
	queueFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#1DB954")).
				Padding(0, 1)

	queueBlurredStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#333333")).
				Padding(0, 1)
)

func NewQueuePanel(width, height int) QueuePanel {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#1DB954")).
		BorderLeftForeground(lipgloss.Color("#1DB954"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#1DB954")).
		BorderLeftForeground(lipgloss.Color("#1DB954"))

	// Highlight current track in green
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(lipgloss.Color("#CCCCCC"))

	l := list.New([]list.Item{}, delegate, width-4, height-4)
	l.Title = "Queue"
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return QueuePanel{
		list:   l,
		width:  width,
		height: height,
	}
}

func (q *QueuePanel) SetTracks(tracks []models.QueueTrack) {
	items := make([]list.Item, len(tracks))
	for i, qt := range tracks {
		items[i] = queueItem{qt: qt}
	}
	q.list.SetItems(items)

	for i, qt := range tracks {
		if qt.IsCurrent {
			q.list.Select(i)
			break
		}
	}
}

func (q *QueuePanel) SetSize(width, height int) {
	q.width = width
	q.height = height
	q.list.SetSize(width-4, height-2)
}

func (q *QueuePanel) SetFocused(focused bool) {
	q.focused = focused
}

func (q *QueuePanel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	q.list, cmd = q.list.Update(msg)
	return cmd
}

func (q *QueuePanel) View() string {
	style := queueBlurredStyle
	if q.focused {
		style = queueFocusedStyle
	}
	return style.Width(q.width - 2).Height(q.height - 2).MaxWidth(q.width - 2).MaxHeight(q.height - 2).Render(q.list.View())
}
