package picker

import (
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	Selection lipgloss.Style
	Next      IndicatorStyles
	Previous  IndicatorStyles
}

type IndicatorStyles struct {
	Value    string
	Enabled  lipgloss.Style
	Disabled lipgloss.Style
}

func DefaultStyles() Styles {
	indEnabled := lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(7))
	indDisabled := lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(8))

	return Styles{
		Selection: lipgloss.NewStyle().Padding(0, 1),
		Next: IndicatorStyles{
			Value:    ">",
			Enabled:  indEnabled,
			Disabled: indDisabled,
		},
		Previous: IndicatorStyles{
			Value:    "<",
			Enabled:  indEnabled,
			Disabled: indDisabled,
		},
	}
}
