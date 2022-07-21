package tabs

import (
	"github.com/charmbracelet/lipgloss"
)

type style struct {
	highlightLight, highlightDark string
}

var (
	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	themes = map[string]style{
		"default": style{highlightLight: "874BFD", highlightDark: "#7D56F4"},
	}
)

func NewStyle() style {
	s := style{}
	s.SetTheme("default")
	return s
}

func (s *style) SetTheme(theme_key string) {
	theme := themes[theme_key]
	s.highlightLight = theme.highlightLight
	s.highlightDark = theme.highlightDark
}
func (s style) Render(title string, active bool) string {
	tabdesign := s.getTabDesign()
	if active == true {
		activeTab := tabdesign.Copy().Border(activeTabBorder, true)
		return activeTab.Render(title)
	} else {
		return tabdesign.Render(title)
	}
}
func (s style) GetTabGap() lipgloss.Style {
	tabGap := s.getTabDesign().Copy().
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)
	return tabGap
}
func (s style) getTabDesign() lipgloss.Style {
	highlight := lipgloss.AdaptiveColor{Light: s.highlightLight, Dark: s.highlightDark}
	tabdesign := lipgloss.NewStyle().
		Border(tabBorder, true).
		BorderForeground(highlight).
		Padding(0, 1)
	return tabdesign
}
