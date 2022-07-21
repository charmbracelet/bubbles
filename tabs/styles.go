package tabs

import (
	"github.com/charmbracelet/lipgloss"
)

type style struct {
	highlightLight string
	highlightDark  string
	highlight      lipgloss.AdaptiveColor
	tabdesign      lipgloss.Style
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
	s.highlight = lipgloss.AdaptiveColor{Light: s.highlightLight, Dark: s.highlightDark}
	s.tabdesign = lipgloss.NewStyle().
		Border(tabBorder, true).
		BorderForeground(s.highlight).
		Padding(0, 1)

}
func (s style) Render(title string, active bool) string {
	if active == true {
		activeTab := s.tabdesign.Copy().Border(activeTabBorder, true)
		return activeTab.Render(title)
	} else {
		return s.tabdesign.Render(title)
	}
}
func (s style) GetTabGap() lipgloss.Style {
	tabGap := s.tabdesign.Copy().
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)
	return tabGap
}
