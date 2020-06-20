package viewport

import (
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// MODEL

type Model struct {
	Width  int
	Height int

	// YOffset is the vertical scroll position.
	YOffset int

	// YPosition is the position of the viewport in relation to the terminal
	// window. It's used in high performance rendering.
	YPosition int

	// HighPerformanceRendering bypasses the normal Bubble Tea renderer to
	// provide higher performance rendering. Most of the time the normal Bubble
	// Tea rendering methods will suffice, but if you're passing content with
	// a lot of ANSI escape codes you may see improved rendering in certain
	// terminals with this enabled.
	//
	// This should only be used in program occupying the entire terminal,
	// which is usually via the alternate screen buffer.
	HighPerformanceRendering bool

	lines []string
}

// TODO: do we really need this?
func NewModel(width, height int) Model {
	return Model{
		Width:  width,
		Height: height,
	}
}

// TODO: do we really need this?
func (m Model) SetSize(yPos, width, height int) {
	m.YPosition = yPos
	m.Width = width
	m.Height = height
}

// AtTop returns whether or not the viewport is in the very top position.
func (m Model) AtTop() bool {
	return m.YOffset <= 0
}

// AtBottom returns whether or not the viewport is at the very botom position.
func (m Model) AtBottom() bool {
	return m.YOffset >= len(m.lines)-m.Height-1
}

// Scrollpercent returns the amount scrolled as a float between 0 and 1.
func (m Model) ScrollPercent() float64 {
	if m.Height >= len(m.lines) {
		return 1.0
	}
	y := float64(m.YOffset)
	h := float64(m.Height)
	t := float64(len(m.lines))
	v := y / (t - h)
	return math.Max(0.0, math.Min(1.0, v))
}

// SetContent set the pager's text content. For high performance rendering the
// Sync command should also be called.
func (m *Model) SetContent(s string) {
	s = strings.Replace(s, "\r\n", "\n", -1) // normalize line endings
	m.lines = strings.Split(s, "\n")
}

// Return the lines that should currently be visible in the viewport
func (m Model) visibleLines() (lines []string) {
	if len(m.lines) > 0 {
		top := max(0, m.YOffset)
		bottom := min(len(m.lines), m.YOffset+m.Height)
		lines = m.lines[top:bottom]
	}
	return lines
}

// ViewDown moves the view down by the number of lines in the viewport.
// Basically, "page down".
func (m *Model) ViewDown() []string {
	if m.AtBottom() {
		return nil
	}

	m.YOffset = min(
		m.YOffset+m.Height,    // target
		len(m.lines)-m.Height, // fallback
	)

	return m.visibleLines()
}

// ViewUp moves the view up by one height of the viewport. Basically, "page up".
func (m *Model) ViewUp() []string {
	if m.AtTop() {
		return nil
	}

	m.YOffset = max(
		m.YOffset-m.Height, // target
		0,                  // fallback
	)

	return m.visibleLines()
}

// HalfViewDown moves the view down by half the height of the viewport.
func (m *Model) HalfViewDown() (lines []string) {
	if m.AtBottom() {
		return nil
	}

	m.YOffset = min(
		m.YOffset+m.Height/2,  // target
		len(m.lines)-m.Height, // fallback
	)

	if len(m.lines) > 0 {
		top := max(m.YOffset+m.Height/2, 0)
		bottom := min(m.YOffset+m.Height, len(m.lines)-1)
		lines = m.lines[top:bottom]
	}

	return lines
}

// HalfViewUp moves the view up by half the height of the viewport.
func (m *Model) HalfViewUp() (lines []string) {
	if m.AtTop() {
		return nil
	}

	m.YOffset = max(
		m.YOffset-m.Height/2, // target
		0,                    // fallback
	)

	if len(m.lines) > 0 {
		top := max(m.YOffset, 0)
		bottom := min(m.YOffset+m.Height/2, len(m.lines)-1)
		lines = m.lines[top:bottom]
	}

	return lines
}

// LineDown moves the view up by the given number of lines.
func (m *Model) LineDown(n int) (lines []string) {
	if m.AtBottom() || n == 0 {
		return nil
	}

	m.YOffset = min(
		m.YOffset+n,           // target
		len(m.lines)-m.Height, // fallback
	)

	if len(m.lines) > 0 {
		top := max(0, m.YOffset+m.Height-n)
		bottom := min(len(m.lines)-1, m.YOffset+m.Height)
		lines = m.lines[top:bottom]
	}

	return lines
}

// LineUp moves the view down by the given number of lines. Returns the new
// lines to show.
func (m *Model) LineUp(n int) (lines []string) {
	if m.AtTop() || n == 0 {
		return nil
	}

	m.YOffset = max(m.YOffset-n, 0)

	if len(m.lines) > 0 {
		top := max(0, m.YOffset)
		bottom := min(len(m.lines)-1, m.YOffset+n)
		lines = m.lines[top:bottom]
	}

	return lines
}

// COMMANDS

// Sync tells the renderer where the viewport will be located and requests
// a render of the current state of the viewport. It should be called for the
// first render and after a window resize.
//
// For high performance rendering only.
func Sync(m Model) tea.Cmd {
	if len(m.lines) == 0 {
		return nil
	}

	top := max(m.YOffset, 0)
	bottom := min(m.YOffset+m.Height, len(m.lines)-1)

	return tea.SyncScrollArea(
		m.lines[top:bottom],
		m.YPosition,
		m.YPosition+m.Height,
	)
}

// ViewDown is a high performance command that moves the viewport up by one
// viewport height. Use Model.ViewDown to get the lines that should be
// rendered. For example:
//
//     lines := model.ViewDown(1)
//     cmd := ViewDown(m, lines)
//
func ViewDown(m Model, lines []string) tea.Cmd {
	if len(lines) == 0 {
		return nil
	}
	return tea.ScrollDown(lines, m.YPosition, m.YPosition+m.Height)
}

// ViewUp is a high performance command the moves the viewport down by one
// viewport height. Use Model.ViewDown to get the lines that should be
// rendered.
func ViewUp(m Model, lines []string) tea.Cmd {
	if len(lines) == 0 {
		return nil
	}
	return tea.ScrollUp(lines, m.YPosition, m.YPosition+m.Height)
}

// HalfViewDown is a high performance command the moves the viewport down by
// half of the height of the viewport. Use Model.HalfViewDown to get the lines
// that should be rendered.
func HalfViewDown(m Model, lines []string) tea.Cmd {
	if len(lines) == 0 {
		return nil
	}
	return tea.ScrollDown(lines, m.YPosition, m.YPosition+m.Height)
}

// HalfViewUp is a high performance command the moves the viewport up by
// half of the height of the viewport. Use Model.HalfViewUp to get the lines
// that should be rendered.
func HalfViewUp(m Model, lines []string) tea.Cmd {
	if len(lines) == 0 {
		return nil
	}
	return tea.ScrollUp(lines, m.YPosition, m.YPosition+m.Height)
}

// LineDown is a high performance command the moves the viewport down by
// a given number of lines. Use Model.LineDown to get the lines that should be
// rendered.
func LineDown(m Model, lines []string) tea.Cmd {
	if len(lines) == 0 {
		return nil
	}
	return tea.ScrollDown(lines, m.YPosition, m.YPosition+m.Height)
}

// LineDown is a high performance command the moves the viewport up by a given
// number of lines. Use Model.LineDown to get the lines that should be
// rendered.
func LineUp(m Model, lines []string) tea.Cmd {
	if len(lines) == 0 {
		return nil
	}
	return tea.ScrollUp(lines, m.YPosition, m.YPosition+m.Height)
}

// UPDATE

// Update runs the update loop with default keybindings similar to popular
// pagers. To define your own keybindings use the methods on Model (i.e.
// Model.LineDown()) and define your own update function.
func Update(msg tea.Msg, m Model) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		// Down one page
		case "pgdown":
			fallthrough
		case " ": // spacebar
			fallthrough
		case "f":
			lines := m.ViewDown()
			if m.HighPerformanceRendering {
				cmd = ViewDown(m, lines)
			}

		// Up one page
		case "pgup":
			fallthrough
		case "b":
			lines := m.ViewUp()
			if m.HighPerformanceRendering {
				cmd = ViewUp(m, lines)
			}

		// Down half page
		case "d":
			lines := m.HalfViewDown()
			if m.HighPerformanceRendering {
				cmd = HalfViewDown(m, lines)
			}

		// Up half page
		case "u":
			lines := m.HalfViewUp()
			if m.HighPerformanceRendering {
				cmd = HalfViewUp(m, lines)
			}

		// Down one line
		case "down":
			fallthrough
		case "j":
			lines := m.LineDown(1)
			if m.HighPerformanceRendering {
				cmd = LineDown(m, lines)
			}

		// Up one line
		case "up":
			fallthrough
		case "k":
			lines := m.LineUp(1)
			if m.HighPerformanceRendering {
				cmd = LineUp(m, lines)
			}
		}
	}

	return m, cmd
}

// VIEW

// View renders the viewport into a string.
func View(m Model) string {

	if m.HighPerformanceRendering {
		// Just send newlines since we're doing to be rendering the actual
		// content seprately. We still need send something that equals the
		// height of this view so that the Bubble Tea standard renderer can
		// position anything below this view properly.
		return strings.Repeat("\n", m.Height-1)
	}

	lines := m.visibleLines()

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
