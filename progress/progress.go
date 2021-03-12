package progress

import (
	"fmt"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/termenv"
)

const defaultWidth = 40

var color func(string) termenv.Color = termenv.ColorProfile().Color

// Option is used to set options in NewModel. For example:
//
//     progress := NewModel(
//	       WithRamp("#ff0000", "#0000ff"),
//	       WithoutPercentage(),
//     )
type Option func(*Model) error

// WithDefaultGradient sets a gradient fill with default colors.
func WithDefaultGradient() Option {
	return WithGradient("#5A56E0", "#EE6FF8")
}

// WithGradient sets a gradient fill blending between two colors.
func WithGradient(colorA, colorB string) Option {
	return func(m *Model) error {
		return m.setRamp(colorA, colorB, false)
	}
}

// WithDefaultScaledGradient sets a gradient with default colors, and scales the
// gradient to fit the filled portion of the ramp.
func WithDefaultScaledGradient() Option {
	return WithScaledGradient("#5A56E0", "#EE6FF8")
}

// WithScaledGradient scales the gradient to fit the width of the filled portion of
// the progress bar.
func WithScaledGradient(colorA, colorB string) Option {
	return func(m *Model) error {
		return m.setRamp(colorA, colorB, true)
	}
}

// WithSolidFill sets the progress to use a solid fill with the given color.
func WithSolidFill(color string) Option {
	return func(m *Model) error {
		m.FullColor = color
		m.useRamp = false
		return nil
	}
}

// WithoutPercentage hides the numeric percentage.
func WithoutPercentage() Option {
	return func(m *Model) error {
		m.ShowPercentage = false
		return nil
	}
}

// WithWidth sets the initial width of the progress bar. Note that you can also
// set the width via the Width property, which can come in handy if you're
// waiting for a tea.WindowSizeMsg.
func WithWidth(w int) Option {
	return func(m *Model) error {
		m.Width = w
		return nil
	}
}

// Model stores values we'll use when rendering the progress bar.
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
	ShowPercentage  bool
	PercentFormat   string // a fmt string for a float
	PercentageStyle *termenv.Style

	useRamp    bool
	rampColorA colorful.Color
	rampColorB colorful.Color

	// When true, we scale the gradient to fit the width of the filled section
	// of the progress bar. When false, the width of the gradient will be set
	// to the full width of the progress bar.
	scaleRamp bool
}

// NewModel returns a model with default values.
func NewModel(opts ...Option) (*Model, error) {
	m := &Model{
		Width:          defaultWidth,
		Full:           '█',
		FullColor:      "#7571F9",
		Empty:          '░',
		EmptyColor:     "#606060",
		ShowPercentage: true,
		PercentFormat:  " %3.0f%%",
	}

	for _, opt := range opts {
		if err := opt(m); err != nil {
			return nil, err
		}
	}

	return m, nil
}

// View renders the progress bar as a given percentage.
func (m Model) View(percent float64) string {
	b := strings.Builder{}
	if m.ShowPercentage {
		percentage := fmt.Sprintf(m.PercentFormat, percent*100) //nolint:gomnd
		if m.PercentageStyle != nil {
			percentage = m.PercentageStyle.Styled(percentage)
		}
		m.bar(&b, percent, ansi.PrintableRuneWidth(percentage))
		b.WriteString(percentage)
	} else {
		m.bar(&b, percent, 0)
	}
	return b.String()
}

func (m Model) bar(b *strings.Builder, percent float64, textWidth int) {
	var (
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
}

func (m *Model) setRamp(colorA, colorB string, scaled bool) error {
	a, err := colorful.Hex(colorA)
	if err != nil {
		return err
	}

	b, err := colorful.Hex(colorB)
	if err != nil {
		return err
	}

	m.useRamp = true
	m.scaleRamp = scaled
	m.rampColorA = a
	m.rampColorB = b
	return nil
}
