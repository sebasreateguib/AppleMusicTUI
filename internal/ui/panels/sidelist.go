package panels

import (
	"strconv"

	"github.com/charmbracelet/x/ansi"

	"SR-Player/internal/ui/theme"
)

// SideItem is a row of a SideList.
type SideItem struct {
	Label string
	Value string // what the app identifies the row with
	Count int    // shown on the right when positive
}

// SideList is one of the stacked lists on the left: playlists, artists, albums.
type SideList struct {
	key, title    string
	items         []SideItem
	cur           cursor
	width, height int
	focused       bool
	active        bool // its selection is the one shown in the tracks panel
	loading       bool
	revision      int // grows with every change to what View shows
}

func NewSideList(key, title string) SideList {
	return SideList{key: key, title: title, loading: true}
}

// SetItems replaces the rows, keeping the selected one if it still exists.
func (s *SideList) SetItems(items []SideItem) {
	selected := s.Selected()
	s.items = items
	s.loading = false
	s.cur.setTotal(len(items))
	for i, it := range items {
		if selected != nil && it.Value == selected.Value {
			s.cur.set(i)
			break
		}
	}
	s.revision++
}

// SetCount changes the number next to the row with that value.
func (s *SideList) SetCount(value string, count int) {
	for i := range s.items {
		if s.items[i].Value == value {
			s.items[i].Count = count
			s.revision++
		}
	}
}

func (s *SideList) Len() int { return len(s.items) }

// Selected returns the selected row, or nil if the list is empty.
func (s *SideList) Selected() *SideItem {
	if s.cur.index < 0 || s.cur.index >= len(s.items) {
		return nil
	}
	it := s.items[s.cur.index]
	return &it
}

func (s *SideList) SetSize(width, height int) {
	s.width, s.height = width, height
	s.cur.setRows(height - 2)
	s.revision++
}

func (s *SideList) SetState(focused, active bool) {
	if s.focused != focused || s.active != active {
		s.focused, s.active = focused, active
		s.revision++
	}
}

// HandleKey reports whether the selection changed.
func (s *SideList) HandleKey(key string) bool {
	moved := s.cur.handleKey(key)
	if moved {
		s.revision++
	}
	return moved
}

// Revision changes whenever View would render something else.
func (s *SideList) Revision() int { return s.revision }

func (s *SideList) View() string {
	inner := s.width - 2
	frame := Frame{
		Key: s.key, Title: s.title, Focused: s.focused,
		Footer: position(s.cur.index, len(s.items)),
		Top:    s.cur.top, Visible: s.cur.rows, Total: len(s.items),
	}
	if len(s.items) == 0 {
		text := "Empty"
		if s.loading {
			text = "Loading…"
		}
		return frame.Render(s.width, s.height, message(text, inner, s.height-2))
	}

	from, to := s.cur.window()
	lines := make([]string, 0, to-from)
	for i := from; i < to; i++ {
		lines = append(lines, s.row(i, inner))
	}
	return frame.Render(s.width, s.height, lines)
}

func (s *SideList) row(i, width int) string {
	it := s.items[i]
	label, count := theme.Text, theme.Dim
	if i == s.cur.index && (s.focused || s.active) {
		bg := theme.ActiveBg
		if s.focused {
			bg = theme.SelectedBg
			label = label.Bold(true)
		}
		label, count = label.Background(bg), count.Background(bg)
	}

	right := ""
	if it.Count > 0 {
		right = strconv.Itoa(it.Count) + " "
	}
	labelW := max(0, width-ansi.StringWidth(right))
	return label.Render(Fit(" "+ansi.Truncate(it.Label, max(0, labelW-2), "…"), labelW)) + count.Render(right)
}
