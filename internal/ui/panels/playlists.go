package panels

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"SR-Player/internal/models"
)

type SidebarItem struct {
	TitleStr string
	DescStr  string
	Kind     string // "artists", "albums", "songs", "playlist"
	Value    string // The raw playlist name or nothing for library items
}

func (s SidebarItem) Title() string       { return s.TitleStr }
func (s SidebarItem) Description() string { return s.DescStr }
func (s SidebarItem) FilterValue() string { return s.TitleStr }

type PlaylistsPanel struct {
	libList    list.Model
	plList     list.Model
	width      int
	height     int
	focused    bool
	activeList int // 0 for Library, 1 for Playlists
}

var (
	sidebarFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#FA243C")).
				Padding(0, 1)

	sidebarBlurredStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#333333")).
				Padding(0, 1)
)

func createSidebarDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#FA243C")).
		BorderLeftForeground(lipgloss.Color("#FA243C"))
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#FA243C")).
		BorderLeftForeground(lipgloss.Color("#FA243C"))
	d.ShowDescription = false
	d.SetSpacing(0)
	return d
}

func NewPlaylistsPanel(width, height int) PlaylistsPanel {
	libList := list.New([]list.Item{
		SidebarItem{TitleStr: "Albums", Kind: "albums"},
		SidebarItem{TitleStr: "Artists", Kind: "artists"},
		SidebarItem{TitleStr: "Songs", Kind: "songs"},
	}, createSidebarDelegate(), width-4, 8)
	libList.Title = "Library"
	libList.Styles.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	libList.SetShowHelp(false)
	libList.SetShowStatusBar(false)
	libList.SetFilteringEnabled(false)

	plList := list.New([]list.Item{}, createSidebarDelegate(), width-4, height-8)
	plList.Title = "Playlists"
	plList.Styles.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	plList.SetShowHelp(false)
	plList.SetShowStatusBar(false)
	plList.SetFilteringEnabled(false)

	return PlaylistsPanel{
		libList:    libList,
		plList:     plList,
		width:      width,
		height:     height,
		activeList: 0, // start focus on Library
	}
}

func (p *PlaylistsPanel) SetPlaylists(playlists []models.Playlist) {
	var items []list.Item
	for _, pl := range playlists {
		items = append(items, SidebarItem{TitleStr: pl.Name, DescStr: "", Kind: "playlist", Value: pl.Name})
	}
	p.plList.SetItems(items)
}

func renderLogo(width int) string {
	colors := []string{
		"#4ade80",
		"#4ade80",
		"#facc15",
		"#f97316",
		"#FA243C",
		"#a855f7",
		"#3b82f6",
	}
	lines := []string{
		"       .:'",
		"    __ :'__",
		" .'`__`-'__``.",
		":__________.-'",
		":_________:",
		" :_________`-;",
		"  `.__.-.__.'",
	}
	var out strings.Builder
	for i, line := range lines {
		out.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colors[i])).Render(line) + "\n")
	}

	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).MarginBottom(1).Render(out.String())
}

func (p *PlaylistsPanel) SetSize(width, height int) {
	p.width = width
	p.height = height

	logoHeight := 0
	if height >= 25 {
		logoHeight = lipgloss.Height(renderLogo(width))
	}
	
	libHeight := 10 
	plHeight := height - libHeight - logoHeight
	if plHeight < 5 {
		plHeight = 5
	}
	
	p.libList.SetSize(width-4, libHeight-2)
	p.plList.SetSize(width-4, plHeight-2)
}

func (p *PlaylistsPanel) SetFocused(focused bool) {
	p.focused = focused
}

func (p *PlaylistsPanel) SelectedItem() *SidebarItem {
	var item list.Item
	if p.activeList == 0 {
		item = p.libList.SelectedItem()
	} else {
		item = p.plList.SelectedItem()
	}
	if item == nil {
		return nil
	}
	si, ok := item.(SidebarItem)
	if !ok {
		return nil
	}
	return &si
}

func (p *PlaylistsPanel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if p.activeList == 1 && p.plList.Index() == 0 {
				p.activeList = 0
				p.libList.Select(len(p.libList.Items()) - 1)
				return nil
			}
		case "down", "j":
			if p.activeList == 0 && p.libList.Index() == len(p.libList.Items())-1 {
				p.activeList = 1
				p.plList.Select(0)
				return nil
			}
		}
	}

	if p.activeList == 0 {
		p.libList, cmd = p.libList.Update(msg)
	} else {
		p.plList, cmd = p.plList.Update(msg)
	}
	return cmd
}

func (p *PlaylistsPanel) View() string {
	libStyle := sidebarBlurredStyle
	plStyle := sidebarBlurredStyle

	if p.focused {
		if p.activeList == 0 {
			libStyle = sidebarFocusedStyle
		} else {
			plStyle = sidebarFocusedStyle
		}
	}

	logoHeight := 0
	var logoView string
	if p.height >= 25 {
		logoView = renderLogo(p.width)
		logoHeight = lipgloss.Height(logoView)
	}

	libHeight := 10
	plHeight := p.height - libHeight - logoHeight

	libView := libStyle.Width(p.width - 2).Height(libHeight - 2).MaxWidth(p.width - 2).MaxHeight(libHeight - 2).Render(p.libList.View())
	plView := plStyle.Width(p.width - 2).Height(plHeight - 2).MaxWidth(p.width - 2).MaxHeight(plHeight - 2).Render(p.plList.View())

	if logoView != "" {
		return lipgloss.JoinVertical(lipgloss.Left, logoView, libView, plView)
	}

	return lipgloss.JoinVertical(lipgloss.Left, libView, plView)
}
