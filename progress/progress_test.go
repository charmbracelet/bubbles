package progress

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
)

func TestPartialBlocks(t *testing.T) {
	// The bar is 10 cells wide and the percentage is hidden, so the visible
	// (ANSI-stripped) output should always be exactly 10 cells wide with the
	// leading edge rendered at eighth-of-a-cell resolution.
	tests := []struct {
		name    string
		percent float64
		want    string
	}{
		{"empty", 0, "░░░░░░░░░░"},
		{"whole-cell", 0.5, "█████░░░░░"},
		{"half-partial", 0.55, "█████▌░░░░"},  // 5.5 cells -> 5 full + 4/8
		{"one-eighth", 0.01, "▏░░░░░░░░░"},    // 0.1 cells -> 1/8
		{"seven-eighths", 0.99, "█████████▉"}, // 9.9 cells -> 9 full + 7/8
		{"rounds-up-to-full", 0.999, "██████████"},
		{"full", 1, "██████████"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(WithPartialBlocks(), WithoutPercentage(), WithWidth(10))
			got := ansi.Strip(p.ViewAs(tt.percent))
			if got != tt.want {
				t.Errorf("ViewAs(%v) = %q, want %q", tt.percent, got, tt.want)
			}
		})
	}
}

func TestPartialBlocksColored(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
		width   int
		percent float64
	}{
		{
			name: "solid-10w-55perc",
			options: []Option{
				WithColors(lipgloss.Color("#FF0000")),
				WithoutPercentage(),
			},
			width:   10,
			percent: 0.55,
		},
		{
			name: "blend-10w-55perc",
			options: []Option{
				WithColors(lipgloss.Color("#FF0000"), lipgloss.Color("#00FF00")),
				WithScaled(false),
				WithoutPercentage(),
			},
			width:   10,
			percent: 0.55,
		},
		{
			name: "blend-scaled-10w-55perc",
			options: []Option{
				WithColors(lipgloss.Color("#FF0000"), lipgloss.Color("#00FF00")),
				WithScaled(true),
				WithoutPercentage(),
			},
			width:   10,
			percent: 0.55,
		},
		{
			name: "colorfunc-30w-65perc",
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
			percent: 0.65,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := append([]Option{WithPartialBlocks()}, tt.options...)
			p := New(opts...)
			p.SetWidth(tt.width)
			golden.RequireEqual(t, []byte(p.ViewAs(tt.percent)))
		})
	}
}

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
