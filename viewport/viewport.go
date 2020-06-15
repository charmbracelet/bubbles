package viewport

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// MODEL

type Model struct {
	Err    error
	Width  int
	Height int

	// YOffset is the vertical scroll position.
	YOffset int

	// Y is the position of the viewport in relation to the terminal window.
	// It's used in high performance rendering.
	Y int

	// UseInternalRenderer specifies whether or not to use the pager's internal,
	// high performance renderer to paint the screen.
	UseInternalRenderer bool

	lines []string

	r renderer
}

func NewModel(yPos, width, height, terminalWidth, terminalHeight int) Model {
	return Model{
		Width:               width,
		Height:              height,
		UseInternalRenderer: true,
		r: renderer{
			Y:              yPos,
			Height:         height,
			TerminalWidth:  terminalWidth,
			TerminalHeight: terminalHeight,
			Out:            os.Stdout,
		},
	}
}

// Scrollpercent returns the amount scrolled as a float between 0 and 1.
func (m Model) ScrollPercent() float64 {
	if m.Height >= len(m.lines) {
		return 1.0
	}
	y := float64(m.YOffset)
	h := float64(m.Height)
	t := float64(len(m.lines))
	return y / (t - h)
}

// SetContent set the pager's text content.
func (m *Model) SetContent(s string) {
	s = strings.Replace(s, "\r\n", "\n", -1) // normalize line endings
	m.lines = strings.Split(s, "\n")

	lines := bounded(m.lines, m.YOffset, m.Height)
	m.r.sync(strings.Join(lines, "\n"))
}

// ViewDown moves the view down by the number of lines in the viewport.
// Basically, "page down".
func (m *Model) ViewDown() {
	m.YOffset = min(len(m.lines)-m.Height, m.YOffset+m.Height)
}

// ViewUp moves the view up by one height of the viewport. Basically, "page up".
func (m *Model) ViewUp() {
	m.YOffset = max(0, m.YOffset-m.Height)
}

// HalfViewUp moves the view up by half the height of the viewport.
func (m *Model) HalfViewUp() {
	m.YOffset = max(0, m.YOffset-m.Height/2)
}

// HalfViewDown moves the view down by half the height of the viewport.
func (m *Model) HalfViewDown() {
	m.YOffset = min(len(m.lines)-m.Height, m.YOffset+m.Height/2)
}

// LineDown moves the view up by the given number of lines.
func (m *Model) LineDown(n int) {
	if m.YOffset >= len(m.lines)-m.Height-1 {
		return
	}
	m.YOffset = min(len(m.lines)-m.Height, m.YOffset+n)

	if m.UseInternalRenderer {
		m.r.insertBottom(m.lines[m.YOffset+m.Height-1])
	}
}

// LineUp moves the view down by the given number of lines.
func (m *Model) LineUp(n int) {
	if m.YOffset == 0 {
		return
	}
	m.YOffset = max(0, m.YOffset-n)

	if m.UseInternalRenderer {
		m.r.insertTop(m.lines[m.YOffset])
	}
}

// UPDATE

// Update runs the update loop with default keybindings. To define your own
// keybindings use the methods on Model and define your own update function.
func Update(msg tea.Msg, m Model) (Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		// Down one page
		case "pgdown":
			fallthrough
		case " ": // spacebar
			fallthrough
		case "f":
			m.ViewDown()
			return m, nil

		// Up one page
		case "pgup":
			fallthrough
		case "b":
			m.ViewUp()
			return m, nil

		// Down half page
		case "d":
			m.HalfViewDown()
			return m, nil

		// Up half page
		case "u":
			m.HalfViewUp()
			return m, nil

		// Down one line
		case "down":
			fallthrough
		case "j":
			m.LineDown(1)
			return m, nil

		// Up one line
		case "up":
			fallthrough
		case "k":
			m.LineUp(1)
			return m, nil
		}
	}

	return m, nil
}

// VIEW

// View renders the viewport into a string.
func View(m Model) string {

	if m.UseInternalRenderer {
		// Skip over the area that would normally be rendered
		return cursorDownString(m.Height)
	}

	if m.Err != nil {
		return m.Err.Error()
	}

	var lines []string

	if len(m.lines) > 0 {
		top := max(0, m.YOffset)
		bottom := min(len(m.lines), m.YOffset+m.Height)
		lines = m.lines[top:bottom]
	}

	// Fill empty space with newlines
	extraLines := ""
	if len(lines) < m.Height {
		extraLines = strings.Repeat("\n", m.Height-len(lines))
	}

	return strings.Join(lines, "\n") + extraLines
}

// ETC

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clamp(val, low, high int) int {
	return max(low, min(high, val))
}

func bounded(s []string, start, end int) []string {
	return s[clamp(start, 0, len(s)):clamp(end, 0, len(s))]
}
