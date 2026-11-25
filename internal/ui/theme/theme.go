// Package theme holds the colors shared by the whole UI. Ordinary text uses the
// terminal's own foreground, so light and dark terminals both stay readable.
package theme

import "github.com/charmbracelet/lipgloss"

var (
	Accent     = lipgloss.Color("#FA243C")
	AccentSoft = lipgloss.AdaptiveColor{Light: "#D7142E", Dark: "#FF4D67"}
	DimColor   = lipgloss.AdaptiveColor{Light: "#6B7077", Dark: "#7D838B"}
	FaintColor = lipgloss.AdaptiveColor{Light: "#C9CDD2", Dark: "#3A3F46"}
	// Selection backgrounds, for the focused panel and for the others.
	SelectedBg = lipgloss.AdaptiveColor{Light: "#FFD3DA", Dark: "#5A1623"}
	ActiveBg   = lipgloss.AdaptiveColor{Light: "#ECEDEF", Dark: "#26292E"}
	// Popups: their own background, and what is left of the screen behind them.
	PopupBg       = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#15171B"}
	BackdropColor = lipgloss.AdaptiveColor{Light: "#CDD1D6", Dark: "#474B53"}
)

var (
	Text     = lipgloss.NewStyle()
	Bold     = lipgloss.NewStyle().Bold(true)
	Backdrop = lipgloss.NewStyle().Foreground(BackdropColor)

	Dim    = lipgloss.NewStyle().Foreground(DimColor)
	Faint  = lipgloss.NewStyle().Foreground(FaintColor)
	Red    = lipgloss.NewStyle().Foreground(Accent)
	Artist = lipgloss.NewStyle().Foreground(AccentSoft)
	Active = lipgloss.NewStyle().Foreground(AccentSoft).Bold(true)
)
