package progress

import (
	"image/color"
	"math"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/exp/golden"
)

func TestPercentBounds(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  float64
	}{
		{"NaN", math.NaN(), 0},
		{"negative infinity", math.Inf(-1), 0},
		{"positive infinity", math.Inf(1), 1},
		{"below zero", -0.5, 0},
		{"above one", 1.5, 1},
		{"fraction", 0.5, 0.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			m.SetPercent(tt.input)
			if got := m.Percent(); got != tt.want {
				t.Fatalf("Percent() = %v, want %v", got, tt.want)
			}
			if got, want := m.ViewAs(tt.input), m.ViewAs(tt.want); got != want {
				t.Fatalf("ViewAs(%v) = %q, want %q", tt.input, got, want)
			}
		})
	}
}

func TestNaNAnimationRecovery(t *testing.T) {
	m := New()
	m.SetPercent(0.5)
	for range 20 {
		m, _ = m.Update(FrameMsg{id: m.id, tag: m.tag})
	}
	for _, target := range []float64{math.NaN(), 0.75} {
		m.SetPercent(target)
		stopped := false
		for range 600 {
			var cmd tea.Cmd
			m, cmd = m.Update(FrameMsg{id: m.id, tag: m.tag})
			if cmd == nil {
				stopped = true
				break
			}
		}
		if !stopped || m.IsAnimating() {
			t.Fatalf("animation did not stop after SetPercent(%v)", target)
		}
		want := target
		if math.IsNaN(want) {
			want = 0
		}
		if math.IsNaN(m.percentShown) || math.Abs(m.percentShown-want) >= 0.001 {
			t.Fatalf("shown percent = %v, want approximately %v", m.percentShown, want)
		}
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
