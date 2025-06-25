// Package progress provides a simple progress bar for Bubble Tea applications.
package progress

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
)

// Internal ID management. Used during animating to assure that frame messages
// can only be received by progress components that sent them.
var lastID int64

func nextID() int {
	return int(atomic.AddInt64(&lastID, 1))
}

const (
	fps              = 60
	defaultWidth     = 40
	defaultFrequency = 18.0
	defaultDamping   = 1.0
)

// FillStep define each possible step for the completion
// of a single block in the progress bar. An array of FillStep
// is used to define the full range of possible completions.
// in the progress bar.
type FillStep struct {
	rune       rune
	completion float64 // 0% to 100% of that particular block
}

func defaultFillSteps() []FillStep {
	return []FillStep{
		{' ', 0.0},
		{'▏', 1.0 / 8.0},
		{'▎', 2.0 / 8.0},
		{'▍', 3.0 / 8.0},
		{'▌', 4.0 / 8.0},
		{'▋', 5.0 / 8.0},
		{'▊', 6.0 / 8.0},
		{'▉', 7.0 / 8.0},
		{'█', 1.0},
	}
}

// Option is used to set options in New. For example:
//
//	    progress := New(
//		       WithRamp("#ff0000", "#0000ff"),
//		       WithoutPercentage(),
//	    )
type Option func(*Model)

// WithDefaultGradient sets a gradient fill with default colors.
func WithDefaultGradient() Option {
	return WithGradient("#5A56E0", "#EE6FF8")
}

// WithGradient sets a gradient fill blending between two colors.
func WithGradient(colorA, colorB string) Option {
	return func(m *Model) {
		m.setRamp(colorA, colorB, false)
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
	return func(m *Model) {
		m.setRamp(colorA, colorB, true)
	}
}

// WithSolidFill sets the progress to use a solid fill with the given color.
func WithSolidFill(color string) Option {
	return func(m *Model) {
		m.FullColor = color
		m.useRamp = false
	}
}

// WithFillCharacters sets the characters used to construct the full and empty components of the progress bar.
func WithFillCharacters(full rune, empty rune) Option {
	return func(m *Model) {
		m.FillSteps = []FillStep{
			{empty, 0.0},
			{full, 1.0},
		}
	}
}

// WithGranularFill sets the characters used to construct the full and empty components of the progress bar.
func WithGranularFill(steps []FillStep) Option {
	sort.Slice(steps, func(i, j int) bool {
		return steps[i].completion < steps[j].completion
	})
	return func(m *Model) {
		m.FillSteps = steps
	}
}

// WithBinaryFill results in a less granular but possible more widely compatible
// progress bar as only two characters are used to represent completion of a
// single block (full/complete and empty/incomplete).
func WithBinaryFill() Option {
	return func(m *Model) {
		m.FillSteps = []FillStep{
			{' ', 0.0},
			{'█', 1.0},
		}
	}
}

// WithDefaultFill sets the progress bar to use the default fill resolution/characters.
func WithDefaultFill() Option {
	return WithGranularFill(defaultFillSteps())
}

// WithoutPercentage hides the numeric percentage.
func WithoutPercentage() Option {
	return func(m *Model) {
		m.ShowPercentage = false
	}
}

// WithWidth sets the initial width of the progress bar. Note that you can also
// set the width via the Width property, which can come in handy if you're
// waiting for a tea.WindowSizeMsg.
func WithWidth(w int) Option {
	return func(m *Model) {
		m.Width = w
	}
}

// WithSpringOptions sets the initial frequency and damping options for the
// progress bar's built-in spring-based animation. Frequency corresponds to
// speed, and damping to bounciness. For details see:
//
// https://github.com/charmbracelet/harmonica
func WithSpringOptions(frequency, damping float64) Option {
	return func(m *Model) {
		m.SetSpringOptions(frequency, damping)
		m.springCustomized = true
	}
}

// WithColorProfile sets the color profile to use for the progress bar.
func WithColorProfile(p termenv.Profile) Option {
	return func(m *Model) {
		m.colorProfile = p
	}
}

// FrameMsg indicates that an animation step should occur.
type FrameMsg struct {
	id  int
	tag int
}

// Model stores values we'll use when rendering the progress bar.
type Model struct {
	// An identifier to keep us from receiving messages intended for other
	// progress bars.
	id int

	// An identifier to keep us from receiving frame messages too quickly.
	tag int

	// Total width of the progress bar, including percentage, if set.
	Width int

	FillSteps []FillStep

	// "Filled" sections of the progress bar.
	FullColor string

	EmptyColor string

	// Settings for rendering the numeric percentage.
	ShowPercentage  bool
	PercentFormat   string // a fmt string for a float
	PercentageStyle lipgloss.Style

	// Members for animated transitions.
	spring           harmonica.Spring
	springCustomized bool
	percentShown     float64 // percent currently displaying
	targetPercent    float64 // percent to which we're animating
	velocity         float64

	// Gradient settings
	useRamp    bool
	rampColorA colorful.Color
	rampColorB colorful.Color

	// When true, we scale the gradient to fit the width of the filled section
	// of the progress bar. When false, the width of the gradient will be set
	// to the full width of the progress bar.
	scaleRamp bool

	// Color profile for the progress bar.
	colorProfile termenv.Profile
}

// New returns a model with default values.
func New(opts ...Option) Model {
	m := Model{
		id:             nextID(),
		Width:          defaultWidth,
		FillSteps:      defaultFillSteps(),
		FullColor:      "#7571F9",
		EmptyColor:     "#606060",
		ShowPercentage: true,
		PercentFormat:  " %3.0f%%",
		colorProfile:   termenv.ColorProfile(),
	}

	for _, opt := range opts {
		opt(&m)
	}

	if !m.springCustomized {
		m.SetSpringOptions(defaultFrequency, defaultDamping)
	}

	return m
}

// NewModel returns a model with default values.
//
// Deprecated: use [New] instead.
var NewModel = New

// Init exists to satisfy the tea.Model interface.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update is used to animate the progress bar during transitions. Use
// SetPercent to create the command you'll need to trigger the animation.
//
// If you're rendering with ViewAs you won't need this.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case FrameMsg:
		if msg.id != m.id || msg.tag != m.tag {
			return m, nil
		}

		// If we've more or less reached equilibrium, stop updating.
		if !m.IsAnimating() {
			return m, nil
		}

		m.percentShown, m.velocity = m.spring.Update(m.percentShown, m.velocity, m.targetPercent)
		return m, m.nextFrame()

	default:
		return m, nil
	}
}

// SetSpringOptions sets the frequency and damping for the current spring.
// Frequency corresponds to speed, and damping to bounciness. For details see:
//
// https://github.com/charmbracelet/harmonica
func (m *Model) SetSpringOptions(frequency, damping float64) {
	m.spring = harmonica.NewSpring(harmonica.FPS(fps), frequency, damping)
}

// Percent returns the current visible percentage on the model. This is only
// relevant when you're animating the progress bar.
//
// If you're rendering with ViewAs you won't need this.
func (m Model) Percent() float64 {
	return m.targetPercent
}

// SetPercent sets the percentage state of the model as well as a command
// necessary for animating the progress bar to this new percentage.
//
// If you're rendering with ViewAs you won't need this.
func (m *Model) SetPercent(p float64) tea.Cmd {
	m.targetPercent = math.Max(0, math.Min(1, p))
	m.tag++
	return m.nextFrame()
}

// IncrPercent increments the percentage by a given amount, returning a command
// necessary to animate the progress bar to the new percentage.
//
// If you're rendering with ViewAs you won't need this.
func (m *Model) IncrPercent(v float64) tea.Cmd {
	return m.SetPercent(m.Percent() + v)
}

// DecrPercent decrements the percentage by a given amount, returning a command
// necessary to animate the progress bar to the new percentage.
//
// If you're rendering with ViewAs you won't need this.
func (m *Model) DecrPercent(v float64) tea.Cmd {
	return m.SetPercent(m.Percent() - v)
}

// View renders an animated progress bar in its current state. To render
// a static progress bar based on your own calculations use ViewAs instead.
func (m Model) View() string {
	return m.ViewAs(m.percentShown)
}

// ViewAs renders the progress bar with a given percentage.
func (m Model) ViewAs(percent float64) string {
	b := strings.Builder{}
	percentView := m.percentageView(percent)
	m.barView(&b, percent, ansi.StringWidth(percentView))
	b.WriteString(percentView)
	return b.String()
}

func (m *Model) nextFrame() tea.Cmd {
	return tea.Tick(time.Second/time.Duration(fps), func(time.Time) tea.Msg {
		return FrameMsg{id: m.id, tag: m.tag}
	})
}

func (m Model) barView(b *strings.Builder, percent float64, textWidth int) {
	var (
		tw = max(0, m.Width-textWidth) // total width of the progress bar
		fw = percent * float64(tw)     // filled width in exact units
	)

	for i := 0; i < tw; i++ {
		cellPercent := float64(i) / float64(tw) // percentage of each cell
		if cellPercent < percent {
			// Filled cell: calculate the closest FillStep
			step := interpolateFillStep(m.FillSteps, fw-float64(i))
			color := m.FullColor
			if m.useRamp {
				color = m.interpolateRamp(i, tw, true)
			}
			b.WriteString(
				termenv.String(string(step.rune)).
					Foreground(m.color(color)).
					Background(m.color(m.EmptyColor)).
					String(),
			)
		} else {
			// Empty cell
			emptyStep := m.FillSteps[0]
			b.WriteString(
				termenv.String(string(emptyStep.rune)).
					Foreground(m.color(m.EmptyColor)).
					Background(m.color(m.EmptyColor)).
					String(),
			)
		}
	}
}

// Helper: Interpolate between FillSteps
func interpolateFillStep(steps []FillStep, remaining float64) FillStep {
	for i := len(steps) - 1; i >= 0; i-- {
		if remaining >= steps[i].completion {
			return steps[i]
		}
	}
	return steps[0]
}

// Helper: Interpolate ramp color
func (m Model) interpolateRamp(pos, total int, isFilled bool) string {
	p := float64(pos) / float64(total-1)
	if m.scaleRamp && isFilled {
		p = float64(pos) / float64(total-1)
	}
	return m.rampColorA.BlendLuv(m.rampColorB, p).Hex()
}

func (m Model) percentageView(percent float64) string {
	if !m.ShowPercentage {
		return ""
	}
	percent = math.Max(0, math.Min(1, percent))
	percentage := fmt.Sprintf(m.PercentFormat, percent*100) //nolint:mnd
	percentage = m.PercentageStyle.Inline(true).Render(percentage)
	return percentage
}

func (m *Model) setRamp(colorA, colorB string, scaled bool) {
	// In the event of an error colors here will default to black. For
	// usability's sake, and because such an error is only cosmetic, we're
	// ignoring the error.
	a, _ := colorful.Hex(colorA)
	b, _ := colorful.Hex(colorB)

	m.useRamp = true
	m.scaleRamp = scaled
	m.rampColorA = a
	m.rampColorB = b
}

func (m Model) color(c string) termenv.Color {
	return m.colorProfile.Color(c)
}

// IsAnimating returns false if the progress bar reached equilibrium and is no longer animating.
func (m *Model) IsAnimating() bool {
	dist := math.Abs(m.percentShown - m.targetPercent)
	return !(dist < 0.001 && m.velocity < 0.01)
}
