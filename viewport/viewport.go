package viewport

import (
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	spacebar = " "
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

// AtTop returns whether or not the viewport is in the very top position.
func (m Model) AtTop() bool {
	return m.YOffset <= 0
}

// AtBottom returns whether or not the viewport is at or past the very bottom
// position.
func (m Model) AtBottom() bool {
	return m.YOffset >= len(m.lines)-1-m.Height
}

// PastBottom returns whether or not the viewport is scrolled beyond the last
// line. This can happen when adjusting the viewport height.
func (m Model) PastBottom() bool {
	return m.YOffset > len(m.lines)-1-m.Height
}

// Scrollpercent returns the amount scrolled as a float between 0 and 1.
func (m Model) ScrollPercent() float64 {
	if m.Height >= len(m.lines) {
		return 1.0
	}
	y := float64(m.YOffset)
	h := float64(m.Height)
	t := float64(len(m.lines) - 1)
	v := y / (t - h)
	return math.Max(0.0, math.Min(1.0, v))
}

// SetContent set the pager's text content. For high performance rendering the
// Sync command should also be called.
func (m *Model) SetContent(s string) {
	s = strings.Replace(s, "\r\n", "\n", -1) // normalize line endings
	m.lines = strings.Split(s, "\n")
}

// Return the lines that should currently be visible in the viewport.
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
		m.YOffset+m.Height,      // target
		len(m.lines)-1-m.Height, // fallback
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
		m.YOffset+m.Height/2,    // target
		len(m.lines)-1-m.Height, // fallback
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

// LineDown moves the view down by the given number of lines.
func (m *Model) LineDown(n int) (lines []string) {
	if m.AtBottom() || n == 0 {
		return nil
	}

	// Make sure the number of lines by which we're going to scroll isn't
	// greater than the number of lines we actually have left before we reach
	// the bottom.
	maxDelta := (len(m.lines) - 1) - (m.YOffset + m.Height) // number of lines - viewport bottom edge
	n = min(n, maxDelta)

	m.YOffset = min(
		m.YOffset+n,             // target
		len(m.lines)-1-m.Height, // fallback
	)

	if len(m.lines) > 0 {
		top := max(m.YOffset+m.Height-n, 0)
		bottom := min(m.YOffset+m.Height, len(m.lines)-1)
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

	// Make sure the number of lines by which we're going to scroll isn't
	// greater than the number of lines we are from the top.
	n = min(n, m.YOffset)

	m.YOffset = max(m.YOffset-n, 0)

	if len(m.lines) > 0 {
		top := max(0, m.YOffset)
		bottom := min(m.YOffset+n, len(m.lines)-1)
		lines = m.lines[top:bottom]
	}

	return lines
}

// GotoTop sets the viewport to the top position.
func (m *Model) GotoTop() (lines []string) {
	if m.AtTop() {
		return nil
	}

	m.YOffset = 0

	if len(m.lines) > 0 {
		top := m.YOffset
		bottom := min(m.YOffset+m.Height, len(m.lines)-1)
		lines = m.lines[top:bottom]
	}

	return lines
}

// GotoTop sets the viewport to the bottom position.
func (m *Model) GotoBottom() (lines []string) {
	m.YOffset = max(len(m.lines)-1-m.Height, 0)

	if len(m.lines) > 0 {
		top := m.YOffset
		bottom := max(len(m.lines)-1, 0)
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

	// TODO: we should probably use m.visibleLines() rather than these two
	// expressions.
	top := max(m.YOffset, 0)
	bottom := min(m.YOffset+m.Height, len(m.lines)-1)

	return tea.SyncScrollArea(
		m.lines[top:bottom],
		m.YPosition,
		m.YPosition+m.Height,
	)
}

// ViewDown is a high performance command that moves the viewport up by a given
// numer of lines. Use Model.ViewDown to get the lines that should be rendered.
// For example:
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

// ViewUp is a high performance command the moves the viewport down by a given
// number of lines height. Use Model.ViewDown to get the lines that should be
// rendered.
func ViewUp(m Model, lines []string) tea.Cmd {
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
		case "pgdown", spacebar, "f":
			lines := m.ViewDown()
			if m.HighPerformanceRendering {
				cmd = ViewDown(m, lines)
			}

		// Up one page
		case "pgup", "b":
			lines := m.ViewUp()
			if m.HighPerformanceRendering {
				cmd = ViewUp(m, lines)
			}

		// Down half page
		case "d", "ctrl+d":
			lines := m.HalfViewDown()
			if m.HighPerformanceRendering {
				cmd = ViewDown(m, lines)
			}

		// Up half page
		case "u", "ctrl+u":
			lines := m.HalfViewUp()
			if m.HighPerformanceRendering {
				cmd = ViewUp(m, lines)
			}

		// Down one line
		case "down", "j":
			lines := m.LineDown(1)
			if m.HighPerformanceRendering {
				cmd = ViewDown(m, lines)
			}

		// Up one line
		case "up", "k":
			lines := m.LineUp(1)
			if m.HighPerformanceRendering {
				cmd = ViewUp(m, lines)
			}
		}

	case tea.MouseMsg:
		switch msg.Type {
		case tea.MouseWheelUp:
			lines := m.LineUp(3)
			if m.HighPerformanceRendering {
				cmd = ViewUp(m, lines)
			}

		case tea.MouseWheelDown:
			lines := m.LineDown(3)
			if m.HighPerformanceRendering {
				cmd = ViewDown(m, lines)
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
