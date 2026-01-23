package progress

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/exp/golden"
)

func TestBlend(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
		width   int
		percent float64
	}{
		{
			name: "10w-red-to-green-50perc",
			options: []Option{
				WithColors(lipgloss.Color("#FF0000"), lipgloss.Color("#00FF00")),
				WithScaled(false),
				WithoutPercentage(),
			},
			width:   10,
			percent: 0.5,
		},
		{
			name: "10w-red-to-green-50perc-full-block",
			options: []Option{
				WithColors(lipgloss.Color("#FF0000"), lipgloss.Color("#00FF00")),
				WithFillCharacters('█', DefaultEmptyCharBlock),
				WithoutPercentage(),
			},
			width:   10,
			percent: 0.5,
		},
		{
			name: "30w-red-to-green-100perc",
			options: []Option{
				WithColors(lipgloss.Color("#FF0000"), lipgloss.Color("#00FF00")),
				WithScaled(false),
				WithoutPercentage(),
			},
			width:   30,
			percent: 1.0,
		},
		{
			name: "10w-red-to-green-scaled-50perc",
			options: []Option{
				WithColors(lipgloss.Color("#FF0000"), lipgloss.Color("#00FF00")),
				WithScaled(true),
				WithoutPercentage(),
			},
			width:   10,
			percent: 0.5,
		},
		{
			name: "30w-red-to-green-scaled-100perc",
			options: []Option{
				WithColors(lipgloss.Color("#FF0000"), lipgloss.Color("#00FF00")),
				WithScaled(true),
				WithoutPercentage(),
			},
			width:   30,
			percent: 1.0,
		},
		{
			name: "30w-colorfunc-rgb-100perc",
			options: []Option{
				WithColorFunc(func(_, current float64) color.Color {
					if current <= 0.3 {
						return lipgloss.Color("#FF0000")
					}
					if current <= 0.7 {
						return lipgloss.Color("#00FF00")
					}
					return lipgloss.Color("#0000FF")
				}),
				WithoutPercentage(),
			},
			width:   30,
			percent: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.options...)
			p.SetWidth(tt.width)
			golden.RequireEqual(t, []byte(p.ViewAs(tt.percent)))
		})
	}
}
