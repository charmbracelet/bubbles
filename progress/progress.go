package progress

import (
	"fmt"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/termenv"
)

var color func(string) termenv.Color = termenv.ColorProfile().Color

// Option is used to set options in NewModel. For example:
//
//     progress := NewModel(
//	       WithRamp("#ff0000", "#0000ff"),
//	       WithoutPercentage(),
//     )
type Option func(*Model)

// WithDefaultRamp sets a gradient fill with default colors.
func WithDefaultRamp() Option {
	return WithRamp("#00dbde", "#fc00ff")
}

// WithRamp sets a gradient fill blending between two colors.
func WithRamp(colorA, colorB string) Option {
	return func(m *Model) {
		m.setRamp(colorA, colorB, false)
	}
}

// WithDefaultScaledRamp sets a gradient with default colors, and scales the
// gradient to fit the filled portion of the ramp.
func WithDefaultScaledRamp() Option {
	return WithScaledRamp("#00dbde", "#fc00ff")
}

// WithScaledRamp scales the gradient to fit the width of the filled portion of
// the progress bar.
func WithScaledRamp(colorA, colorB string) Option {
	return func(m *Model) {
		m.setRamp(colorA, colorB, true)
	}
}

// WithSoildFill sets the progress to use a solid fill with the given color.
func WithSolidFill(color string) Option {
	return func(m *Model) {
		m.FullColor = color
		m.useRamp = false
	}
}

// WithoutPercentage hides the numeric percentage.
func WithoutPercentage() Option {
	return func(m *Model) {
		m.ShowPercentage = false
	}
}

type Model struct {

	// Total width of the progress bar, including percentage, if set.
	Width int

	// "Filled" sections of the progress bar
	Full      rune
	FullColor string

	// "Empty" sections of progress bar
	Empty      rune
	EmptyColor string

	// Settings for rendering the numeric percentage
	ShowPercentage bool
	PercentFormat  string // a fmt string for a float

	useRamp    bool
	rampColorA colorful.Color
	rampColorB colorful.Color

	// When true, we scale the gradient to fit the width of the filled section
	// of the progress bar. When false, the width of the gradient will be set
	// to the full width of the progress bar.
	scaleRamp bool
}

// NewModel returns a model with default values.
func NewModel(opts ...Option) *Model {
	m := &Model{
		Width:          40,
		Full:           '█',
		FullColor:      "#7571F9",
		Empty:          '░',
		EmptyColor:     "#606060",
		ShowPercentage: true,
		PercentFormat:  " %3.0f%%",
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m Model) View(percent float64) string {
	if m.ShowPercentage {
		s := fmt.Sprintf(m.PercentFormat, percent*100)
		w := ansi.PrintableRuneWidth(s)
		return m.bar(percent, w) + s
	}
	return m.bar(percent, 0)
}

func (m Model) bar(percent float64, textWidth int) string {
	var (
		b  = strings.Builder{}
		tw = m.Width - textWidth        // total width
		fw = int(float64(tw) * percent) // filled width
		p  float64
	)

	if m.useRamp {
		// Gradient fill
		for i := 0; i < fw; i++ {
			if m.scaleRamp {
				p = float64(i) / float64(fw)
			} else {
				p = float64(i) / float64(tw)
			}
			c := m.rampColorA.BlendLuv(m.rampColorB, p).Hex()
			b.WriteString(termenv.
				String(string(m.Full)).
				Foreground(color(c)).
				String(),
			)
		}
	} else {
		// Solid fill
		s := termenv.String(string(m.Full)).Foreground(color(m.FullColor)).String()
		b.WriteString(strings.Repeat(s, fw))
	}

	// Empty fill
	e := termenv.String(string(m.Empty)).Foreground(color(m.EmptyColor)).String()
	b.WriteString(strings.Repeat(e, tw-fw))

	return b.String()
}

func (m *Model) setRamp(colorA, colorB string, scaled bool) {
	a, _ := colorful.Hex(colorA)
	b, _ := colorful.Hex(colorB)
	m.useRamp = true
	m.scaleRamp = scaled
	m.rampColorA = a
	m.rampColorB = b
}
