package list

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	bullet   = "•"
	ellipsis = "…"
)

// Styles contains style definitions for this list component. By default, these
// values are generated by DefaultStyles.
type Styles struct {
	TitleBar     lipgloss.Style
	Title        lipgloss.Style
	Spinner      lipgloss.Style
	FilterPrompt lipgloss.Style
	FilterCursor lipgloss.Style

	// Default styling for matched characters in a filter. This can be
	// overridden by delegates.
	DefaultFilterCharacterMatch lipgloss.Style

	StatusBar             lipgloss.Style
	StatusEmpty           lipgloss.Style
	StatusBarActiveFilter lipgloss.Style
	StatusBarFilterCount  lipgloss.Style

	NoItems lipgloss.Style

	PaginationStyle lipgloss.Style
	HelpStyle       lipgloss.Style

	// Styled characters.
	ActivePaginationDot   lipgloss.Style
	InactivePaginationDot lipgloss.Style
	ArabicPagination      lipgloss.Style
	DividerDot            lipgloss.Style
}

// BaseStyles returns a set of base styles for this list component. No color
// will be applied. You can use this as a starting point for your own styles,
// if you like.
func BaseStyles() Styles {
	var s Styles

	s.TitleBar = lipgloss.NewStyle().
		Padding(0, 0, 1, 2)
	s.Title = lipgloss.NewStyle().
		Padding(0, 1)
	s.Spinner = lipgloss.NewStyle()
	s.FilterPrompt = lipgloss.NewStyle()
	s.FilterCursor = lipgloss.NewStyle()
	s.DefaultFilterCharacterMatch = lipgloss.NewStyle().
		Underline(true)
	s.StatusBar = lipgloss.NewStyle().
		Padding(0, 0, 1, 2)
	s.StatusEmpty = lipgloss.NewStyle()
	s.StatusBarActiveFilter = lipgloss.NewStyle()
	s.StatusBarFilterCount = lipgloss.NewStyle()
	s.NoItems = lipgloss.NewStyle()
	s.ArabicPagination = lipgloss.NewStyle()
	s.PaginationStyle = lipgloss.NewStyle().
		PaddingLeft(2) //nolint:gomnd
	s.HelpStyle = lipgloss.NewStyle().
		Padding(1, 0, 0, 2)
	s.ActivePaginationDot = lipgloss.NewStyle().
		SetString(bullet)
	s.InactivePaginationDot = lipgloss.NewStyle().
		SetString(bullet)
	s.DividerDot = lipgloss.NewStyle().
		SetString(" " + bullet + " ")

	return s
}

// LightStyles returns a set of light style definitions for this list Bubble.
func LightStyles() Styles {
	return newStyles(false)
}

// DarkStyles returns a set of dark style definitions for this list Bubble.
func DarkStyles() Styles {
	return newStyles(true)
}

// DefaultStyles returns a set of default style definitions for this list
// component.
func newStyles(isDark bool) (s Styles) {
	lightDark := lipgloss.LightDark(isDark)

	verySubduedColor := lightDark("#DDDADA", "#3C3C3C")
	subduedColor := lightDark("#9B9B9B", "#5C5C5C")

	s.Title = s.Title.
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230"))
	s.Spinner = s.Spinner.
		Foreground(lightDark("#8E8E8E", "#747373"))
	s.FilterPrompt = s.FilterPrompt.
		Foreground(lightDark("#04B575", "#ECFD65"))
	s.FilterCursor = s.FilterCursor.
		Foreground(lightDark("#EE6FF8", "#EE6FF8"))
	s.StatusBar = s.StatusBar.
		Foreground(lightDark("#A49FA5", "#777777"))
	s.StatusEmpty = s.StatusEmpty.
		Foreground(subduedColor)
	s.StatusBarActiveFilter = s.StatusBarActiveFilter.
		Foreground(lightDark("#1a1a1a", "#dddddd"))
	s.StatusBarFilterCount = s.StatusBarFilterCount.
		Foreground(verySubduedColor)
	s.NoItems = s.NoItems.
		Foreground(lightDark("#909090", "#626262"))
	s.ArabicPagination = s.ArabicPagination.
		Foreground(subduedColor)
	s.HelpStyle = s.HelpStyle.
		Padding(1, 0, 0, 2)
	s.ActivePaginationDot = s.ActivePaginationDot.
		Foreground(lightDark("#847A85", "#979797")).
		SetString(bullet)
	s.InactivePaginationDot = s.InactivePaginationDot.
		Foreground(verySubduedColor).
		SetString(bullet)
	s.DividerDot = s.DividerDot.
		Foreground(verySubduedColor).
		SetString(" " + bullet + " ")

	return s
}
