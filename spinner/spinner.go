package spinner

import (
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/ansi"
)

// Internal ID management for text inputs. Necessary for blink integrity when
// multiple text inputs are involved.
var (
	lastID int
	idMtx  sync.Mutex
)

// Return the next ID we should use on the Model.
func nextID() int {
	idMtx.Lock()
	defer idMtx.Unlock()
	lastID++
	return lastID
}

// Spinner is a set of frames used in animating the spinner.
type Spinner struct {
	Frames []string
	FPS    time.Duration
}

// Some spinners to choose from. You could also make your own.
var (
	Line = Spinner{
		Frames: []string{"|", "/", "-", "\\"},
		FPS:    time.Second / 10, //nolint:gomnd
	}
	Dot = Spinner{
		Frames: []string{"⣾ ", "⣽ ", "⣻ ", "⢿ ", "⡿ ", "⣟ ", "⣯ ", "⣷ "},
		FPS:    time.Second / 10, //nolint:gomnd
	}
	MiniDot = Spinner{
		Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		FPS:    time.Second / 12, //nolint:gomnd
	}
	Jump = Spinner{
		Frames: []string{"⢄", "⢂", "⢁", "⡁", "⡈", "⡐", "⡠"},
		FPS:    time.Second / 10, //nolint:gomnd
	}
	Pulse = Spinner{
		Frames: []string{"█", "▓", "▒", "░"},
		FPS:    time.Second / 8, //nolint:gomnd
	}
	Points = Spinner{
		Frames: []string{"∙∙∙", "●∙∙", "∙●∙", "∙∙●"},
		FPS:    time.Second / 7, //nolint:gomnd
	}
	Globe = Spinner{
		Frames: []string{"🌍", "🌎", "🌏"},
		FPS:    time.Second / 4, //nolint:gomnd
	}
	Moon = Spinner{
		Frames: []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"},
		FPS:    time.Second / 8, //nolint:gomnd
	}
	Monkey = Spinner{
		Frames: []string{"🙈", "🙉", "🙊"},
		FPS:    time.Second / 3, //nolint:gomnd
	}
)

// Model contains the state for the spinner. Use NewModel to create new models
// rather than using Model as a struct literal.
type Model struct {

	// Spinner settings to use. See type Spinner.
	Spinner Spinner

	// Style sets the styling for the spinner. Most of the time you'll just
	// want foreground and background coloring, and potentially some padding.
	//
	// For an introduction to styling with Lip Gloss see:
	// https://github.com/charmbracelet/lipgloss
	Style lipgloss.Style

	// MinimumLifetime is the minimum amount of time the spinner can run. Any
	// logic around this can be implemented in view that implements this
	// spinner. If HideFor is set MinimumLifetime will be added on top of
	// HideFor. In other words, if HideFor is 100ms and MinimumLifetime is
	// 200ms then MinimumLifetime will expire after 300ms.
	//
	// MinimumLifetime is optional.
	//
	// This is considered experimental and may not appear in future versions of
	// this library.
	MinimumLifetime time.Duration

	// HideFor can be used to wait to show the spinner until a certain amount
	// of time has passed. This can be useful for preventing flicking when load
	// times are very fast.
	// Optional.
	//
	// This is considered experimental and may not appear in future versions of
	// this library.
	HideFor time.Duration

	frame     int
	startTime time.Time
	id        int
	tag       int
}

// Start resets resets the spinner start time. For use with MinimumLifetime and
// MinimumStartTime. Optional.
//
// This function is optional and generally considered for advanced use only.
// Most of the time your application logic will obviate the need for this
// method.
//
// This is considered experimental and may not appear in future versions of
// this library.
func (m *Model) Start() {
	m.startTime = time.Now()
}

// Finish sets the internal timer to a completed state so long as the spinner
// isn't flagged to be showing. If it is showing, finish has no effect. The
// idea here is that you call Finish if your operation has completed and, if
// the spinner isn't showing yet (by virtue of HideFor) then Visible() doesn't
// show the spinner at all.
//
// This is intended to work in conjunction with MinimumLifetime and
// MinimumStartTime, is completely optional.
//
// This function is optional and generally considered for advanced use only.
// Most of the time your application logic will obviate the need for this
// method.
//
// This is considered experimental and may not appear in future versions of
// this library.
func (m *Model) Finish() {
	if m.hidden() {
		m.startTime = time.Time{}
	}
}

// ID returns the spinner's unique ID.
func (m Model) ID() int {
	return m.id
}

// advancedMode returns whether or not the user is making use of HideFor and
// MinimumLifetime properties.
func (m Model) advancedMode() bool {
	return m.HideFor > 0 && m.MinimumLifetime > 0
}

// hidden returns whether or not Model.HideFor is in effect.
func (m Model) hidden() bool {
	if m.startTime.IsZero() {
		return false
	}
	if m.HideFor == 0 {
		return false
	}
	return m.startTime.Add(m.HideFor).After(time.Now())
}

// finished returns whether the minimum lifetime of this spinner has been
// exceeded.
func (m Model) finished() bool {
	if m.startTime.IsZero() || m.MinimumLifetime == 0 {
		return true
	}
	return m.startTime.Add(m.HideFor).Add(m.MinimumLifetime).Before(time.Now())
}

// Visible returns whether or not the view should be rendered. Works in
// conjunction with Model.HideFor and Model.MinimumLifetimeReached. You should
// use this method directly to determine whether or not to render this view in
// the parent view and whether to continue sending spin messaging in the
// parent update function.
//
// This function is optional and generally considered for advanced use only.
// Most of the time your application logic will obviate the need for this
// method.
//
// This is considered experimental and may not appear in future versions of
// this library.
func (m Model) Visible() bool {
	return !m.hidden() && !m.finished()
}

// New returns a model with default values.
func New() Model {
	return Model{
		Spinner: Line,
		id:      nextID(),
	}
}

// NewModel returns a model with default values.
//
// Deprecated. Use New instead.
var NewModel = New

// TickMsg indicates that the timer has ticked and we should render a frame.
type TickMsg struct {
	Time time.Time
	tag  int
	ID   int
}

// Update is the Tea update function. This will advance the spinner one frame
// every time it's called, regardless the message passed, so be sure the logic
// is setup so as not to call this Update needlessly.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		// If an ID is set, and the ID doesn't belong to this spinner, reject
		// the message.
		if msg.ID > 0 && msg.ID != m.id {
			return m, nil
		}

		// If a tag is set, and it's not the one we expect, reject the message.
		// This prevents the spinner from receiving too many messages and
		// thus spinning too fast.
		if msg.tag > 0 && msg.tag != m.tag {
			return m, nil
		}

		m.frame++
		if m.frame >= len(m.Spinner.Frames) {
			m.frame = 0
		}

		m.tag++
		return m, m.tick(m.id, m.tag)
	default:
		return m, nil
	}
}

// View renders the model's view.
func (m Model) View() string {
	if m.frame >= len(m.Spinner.Frames) {
		return "(error)"
	}

	frame := m.Spinner.Frames[m.frame]

	// If we're using the fine-grained hide/show spinner rules and those rules
	// deem that the spinner should be hidden, draw an empty space in place of
	// the spinner.
	if m.advancedMode() && !m.Visible() {
		frame = strings.Repeat(" ", ansi.PrintableRuneWidth(frame))
	}

	return m.Style.Render(frame)
}

// Tick is the command used to advance the spinner one frame. Use this command
// to effectively start the spinner.
func (m Model) Tick() tea.Msg {
	return TickMsg{
		// The time at which the tick occurred.
		Time: time.Now(),

		// The ID of the spinner that this message belongs to. This can be
		// helpful when routing messages, however bear in mind that spinners
		// will ignore messages that don't contain ID by default.
		ID: m.id,

		tag: m.tag,
	}
}

func (m Model) tick(id, tag int) tea.Cmd {
	return tea.Tick(m.Spinner.FPS, func(t time.Time) tea.Msg {
		return TickMsg{
			Time: t,
			ID:   id,
			tag:  tag,
		}
	})
}

// Tick is the command used to advance the spinner one frame. Use this command
// to effectively start the spinner.
//
// This method is deprecated. Use Model.Tick instead.
func Tick() tea.Msg {
	return TickMsg{Time: time.Now()}
}
