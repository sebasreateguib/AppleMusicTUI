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
	queueFocusedBorder = lipgloss.Color("#FA243C")
	queueBlurredBorder = lipgloss.Color("#333333")
)

func NewQueuePanel(width, height int) QueuePanel {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#FA243C")).
		BorderLeftForeground(lipgloss.Color("#FA243C"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#FA243C")).
		BorderLeftForeground(lipgloss.Color("#FA243C"))

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
	q.list.SetSize(width-4, height-4)
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
	border := queueBlurredBorder
	if q.focused {
		border = queueFocusedBorder
	}
	return renderRectBox(q.list.View(), q.width, q.height, border, lipgloss.Top)
}
