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
//     )
type Option func(*Model)

// WithDefaultRamp sets a gradient fill with default colors.
func WithDefaultRamp() Option {
	return WithRamp("#00dbde", "#fc00ff")
}

// WithRamp sets a gradient fill blending between two colors.
func WithRamp(colorA, colorB string) Option {
	a, _ := colorful.Hex(colorA)
	b, _ := colorful.Hex(colorB)
	return func(m *Model) {
		m.useRamp = true
		m.colorA = a
		m.colorB = b
	}
}

type Model struct {
	// Left side color of progress bar. By default, it's #00dbde
	colorA colorful.Color

	// Left side color of progress bar. By default, it's #fc00ff
	colorB colorful.Color

	// Total width of the progress bar, including percentage, if set.
	Width int

	// Rune for "filled" sections of the progress bar.
	Full rune

	// Rune for "empty" sections of progress bar.
	Empty rune

	// if true, gradient will be setted from start to end of filled part. Instead, it'll work
	// on all proggress bar length
	FullGradientMode bool

	ShowPercent   bool
	PercentFormat string

	useRamp bool
}

// NewModel returns a model with default values.
func NewModel(opts ...Option) *Model {
	m := &Model{
		Width:         40,
		Full:          '█',
		Empty:         '░',
		ShowPercent:   true,
		PercentFormat: " %3.0f%%",
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m Model) View(percent float64) string {
	if m.ShowPercent {
		s := fmt.Sprintf(m.PercentFormat, percent*100)
		w := ansi.PrintableRuneWidth(s)
		return m.bar(percent, w) + s
	}
	return m.bar(percent, 0)
}

func (m Model) bar(percent float64, textWidth int) string {
	w := m.Width - textWidth

	ramp := make([]string, int(float64(w)*percent))
	for i := 0; i < len(ramp); i++ {
		gradientPart := float64(w)
		if m.FullGradientMode {
			gradientPart = float64(len(ramp))
		}
		percent := float64(i) / gradientPart
		c := m.colorA.BlendLuv(m.colorB, percent)
		ramp[i] = c.Hex()
	}

	var fullCells string
	for i := 0; i < len(ramp); i++ {
		fullCells += termenv.String(string(m.Full)).Foreground(color(ramp[i])).String()
	}

	fullCells += strings.Repeat(string(m.Empty), w-len(ramp))
	return fullCells
}
