package spinner

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

const defaultFPS = time.Second / 10

// Spinner is a set of frames used in animating the spinner.
type Spinner struct {
	Frames []string
	FPS    time.Duration
}

var (
	// Some spinners to choose from. You could also make your own.
	Line = Spinner{
		Frames: []string{"|", "/", "-", "\\"},
		FPS:    time.Second / 10,
	}
	Dot = Spinner{
		Frames: []string{"⣾ ", "⣽ ", "⣻ ", "⢿ ", "⡿ ", "⣟ ", "⣯ ", "⣷ "},
		FPS:    time.Second / 10,
	}
	MiniDot = Spinner{
		Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		FPS:    time.Second / 12,
	}
	Jump = Spinner{
		Frames: []string{"⢄", "⢂", "⢁", "⡁", "⡈", "⡐", "⡠"},
		FPS:    time.Second / 10,
	}
	Bit8 = Spinner{
		Frames: []string{
			"⠀", "⠁", "⠂", "⠃", "⠄", "⠅", "⠆", "⠇", "⡀", "⡁", "⡂", "⡃", "⡄", "⡅", "⡆", "⡇",
			"⠈", "⠉", "⠊", "⠋", "⠌", "⠍", "⠎", "⠏", "⡈", "⡉", "⡊", "⡋", "⡌", "⡍", "⡎", "⡏",
			"⠐", "⠑", "⠒", "⠓", "⠔", "⠕", "⠖", "⠗", "⡐", "⡑", "⡒", "⡓", "⡔", "⡕", "⡖", "⡗",
			"⠘", "⠙", "⠚", "⠛", "⠜", "⠝", "⠞", "⠟", "⡘", "⡙", "⡚", "⡛", "⡜", "⡝", "⡞", "⡟",
			"⠠", "⠡", "⠢", "⠣", "⠤", "⠥", "⠦", "⠧", "⡠", "⡡", "⡢", "⡣", "⡤", "⡥", "⡦", "⡧",
			"⠨", "⠩", "⠪", "⠫", "⠬", "⠭", "⠮", "⠯", "⡨", "⡩", "⡪", "⡫", "⡬", "⡭", "⡮", "⡯",
			"⠰", "⠱", "⠲", "⠳", "⠴", "⠵", "⠶", "⠷", "⡰", "⡱", "⡲", "⡳", "⡴", "⡵", "⡶", "⡷",
			"⠸", "⠹", "⠺", "⠻", "⠼", "⠽", "⠾", "⠿", "⡸", "⡹", "⡺", "⡻", "⡼", "⡽", "⡾", "⡿",
			"⢀", "⢁", "⢂", "⢃", "⢄", "⢅", "⢆", "⢇", "⣀", "⣁", "⣂", "⣃", "⣄", "⣅", "⣆", "⣇",
			"⢈", "⢉", "⢊", "⢋", "⢌", "⢍", "⢎", "⢏", "⣈", "⣉", "⣊", "⣋", "⣌", "⣍", "⣎", "⣏",
			"⢐", "⢑", "⢒", "⢓", "⢔", "⢕", "⢖", "⢗", "⣐", "⣑", "⣒", "⣓", "⣔", "⣕", "⣖", "⣗",
			"⢘", "⢙", "⢚", "⢛", "⢜", "⢝", "⢞", "⢟", "⣘", "⣙", "⣚", "⣛", "⣜", "⣝", "⣞", "⣟",
			"⢠", "⢡", "⢢", "⢣", "⢤", "⢥", "⢦", "⢧", "⣠", "⣡", "⣢", "⣣", "⣤", "⣥", "⣦", "⣧",
			"⢨", "⢩", "⢪", "⢫", "⢬", "⢭", "⢮", "⢯", "⣨", "⣩", "⣪", "⣫", "⣬", "⣭", "⣮", "⣯",
			"⢰", "⢱", "⢲", "⢳", "⢴", "⢵", "⢶", "⢷", "⣰", "⣱", "⣲", "⣳", "⣴", "⣵", "⣶", "⣷",
			"⢸", "⢹", "⢺", "⢻", "⢼", "⢽", "⢾", "⢿", "⣸", "⣹", "⣺", "⣻", "⣼", "⣽", "⣾", "⣿",
		},
		FPS: time.Second / 16,
	}
	Globe = Spinner{
		Frames: []string{"🌍 ", "🌎 ", "🌏 "},
		FPS:    time.Second / 4,
	}
	Moon = Spinner{
		Frames: []string{"🌑 ", "🌒 ", "🌓 ", "🌔 ", "🌕 ", "🌖 ", "🌗 ", "🌘 "},
		FPS:    time.Second / 8,
	}
	Monkey = Spinner{
		Frames: []string{"🙈 ", "🙉 ", "🙊 "},
		FPS:    time.Second / 3,
	}

	color = termenv.ColorProfile().Color
)

// Model contains the state for the spinner. Use NewModel to create new models
// rather than using Model as a struct literal.
type Model struct {

	// Spinner settings to use. See type Spinner.
	Spinner Spinner

	// ForegroundColor sets the background color of the spinner. It can be
	// a hex code or one of the 256 ANSI colors. If the terminal emulator can't
	// support the color specified it will automatically degrade (per
	// github.com/muesli/termenv).
	ForegroundColor string

	// BackgroundColor sets the background color of the spinner. It can be
	// a hex code or one of the 256 ANSI colors. If the terminal emulator can't
	// support the color specified it will automatically degrade (per
	// github.com/muesli/termenv).
	BackgroundColor string

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
}

// Start resets resets the spinner start time. For use with MinimumLifetime and
// MinimumStartTime. Optional.
//
// This is considered experimental and may not appear in future versions of
// this library.
func (m *Model) Start() {
	m.startTime = time.Now()
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

// finished returns whether Model.MinimumLifetimeReached has been met.
func (m Model) finished() bool {
	if m.startTime.IsZero() {
		return true
	}
	if m.MinimumLifetime == 0 {
		return true
	}
	return m.startTime.Add(m.HideFor).Add(m.MinimumLifetime).Before(time.Now())
}

// Visible returns whether or not the view should be rendered. Works in
// conjunction with Model.HideFor and Model.MinimumLifetimeReached. You should
// use this message directly to determine whether or not to render this view in
// the parent view and whether to continue sending spin messaging in the
// parent update function.
//
// Also note that using this function is optional and generally considered for
// advanced use only. Most of the time your application logic will determine
// whether or not this view should be used.
//
// This is considered experimental and may not appear in future versions of
// this library.
func (m Model) Visible() bool {
	return !m.hidden() && !m.finished()
}

// NewModel returns a model with default values.
func NewModel() Model {
	return Model{Spinner: Line}
}

// TickMsg indicates that the timer has ticked and we should render a frame.
type TickMsg struct {
	Time time.Time
}

// Update is the Tea update function. This will advance the spinner one frame
// every time it's called, regardless the message passed, so be sure the logic
// is setup so as not to call this Update needlessly.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg.(type) {
	case TickMsg:
		m.frame++
		if m.frame >= len(m.Spinner.Frames) {
			m.frame = 0
		}
		return m, m.tick()
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

	if m.ForegroundColor != "" || m.BackgroundColor != "" {
		return termenv.
			String(frame).
			Foreground(color(m.ForegroundColor)).
			Background(color(m.BackgroundColor)).
			String()
	}

	return frame
}

// Tick is the command used to advance the spinner one frame.
func Tick() tea.Msg {
	return TickMsg{Time: time.Now()}
}

func (m Model) tick() tea.Cmd {
	return tea.Tick(m.Spinner.FPS, func(t time.Time) tea.Msg {
		return TickMsg{
			Time: t,
		}
	})
}
