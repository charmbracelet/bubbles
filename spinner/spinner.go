package spinner

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

const (
	defaultFPS = time.Second / 10
)

// Spinner is a set of frames used in animating the spinner.
type Spinner = []string

var (
	// Some spinners to choose from. You could also make your own.
	Line = Spinner([]string{"|", "/", "-", "\\"})
	Dot  = Spinner([]string{"⣾ ", "⣽ ", "⣻ ", "⢿ ", "⡿ ", "⣟ ", "⣯ ", "⣷ "})

	color = termenv.ColorProfile().Color
)

// Model contains the state for the spinner. Use NewModel to create new models
// rather than using Model as a struct literal.
type Model struct {

	// Type is the set of frames to use. See Spinner.
	Frames Spinner

	// FPS is the speed at which the ticker should tick
	FPS time.Duration

	// ForegroundColor sets the background color of the spinner. It can be a
	// hex code or one of the 256 ANSI colors. If the terminal emulator can't
	// doesn't support the color specified it will automatically degrade
	// (per github.com/muesli/termenv).
	ForegroundColor string

	// BackgroundColor sets the background color of the spinner. It can be a
	// hex code or one of the 256 ANSI colors. If the terminal emulator can't
	// doesn't support the color specified it will automatically degrade
	// (per github.com/muesli/termenv).
	BackgroundColor string

	// Minimum amount of time the spinner can run. Any logic around this can
	// be implemented in view that implements this spinner. Optional.
	MinimumLifetime time.Duration

	// HideFor can be used to wait to show the spinner until a certain amount
	// of time has passed. This can be useful for preventing flicking when load
	// times are very fast. The hidden state can be set with HiddenState.
	// Optional.
	HideFor time.Duration

	// HiddenState is the
	HiddenState string

	frame     int
	startTime time.Time
}

// Start resets resets the spinner start time. For use with MinimumLifetime and
// MinimumStartTime. Optional.
func (m *Model) Start() {
	m.frame = 0
	m.startTime = time.Now()
}

// MinimumLifetimeReached returns whether or not the spinner has run for the
// minimum specified duration, if any. If no minimum lifetime has been set, or
// if Model.Start() hasn't been called this function returns true.
func (m Model) MinimumLifetimeReached() bool {
	if m.startTime.IsZero() {
		return true
	}
	if m.MinimumLifetime == 0 {
		return true
	}
	return m.startTime.Add(m.MinimumLifetime).Before(time.Now())
}

// Hidden returns whether or not the view should be rendered. Works in
// conjunction with Model.HideFor. You can perform this message directly to
// Do additional logic on your views.
func (m Model) Hidden() bool {
	if m.startTime.IsZero() {
		return false
	}
	if m.HideFor == 0 {
		return false
	}
	return m.startTime.Add(m.HideFor).After(time.Now())
}

// NewModel returns a model with default values.
func NewModel() Model {
	return Model{
		Frames: Line,
		FPS:    defaultFPS,
	}
}

// TickMsg indicates that the timer has ticked and we should render a frame.
type TickMsg struct {
	Time  time.Time
	Frame string
}

// Update is the Tea update function. This will advance the spinner one frame
// every time it's called, regardless the message passed, so be sure the logic
// is setup so as not to call this Update needlessly.
func Update(msg tea.Msg, m Model) (Model, tea.Cmd) {
	if _, ok := msg.(TickMsg); ok {
		m.frame++
		if m.frame >= len(m.Frames) {
			m.frame = 0
		}
		return m, Tick(m)
	}
	return m, nil
}

// View renders the model's view.
func View(model Model) string {
	if model.frame >= len(model.Frames) {
		return "error"
	}

	if model.Hidden() {
		return termenv.String(model.HiddenState).
			Background(color(model.BackgroundColor)).
			String()
	}

	frame := model.Frames[model.frame]

	if model.ForegroundColor != "" || model.BackgroundColor != "" {
		return termenv.
			String(frame).
			Foreground(color(model.ForegroundColor)).
			Background(color(model.BackgroundColor)).
			String()
	}

	return frame
}

// Tick is the command used to advance the spinner one frame.
func Tick(m Model) tea.Cmd {
	return tea.Tick(m.FPS, func(t time.Time) tea.Msg {
		return TickMsg{
			Time:  t,
			Frame: m.Frames[m.frame],
		}
	})
}
