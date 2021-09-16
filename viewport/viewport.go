package viewport

import (
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	spacebar        = " "
	mouseWheelDelta = 3
)

// Model is the Bubble Tea model for this viewport element.
type Model struct {
	Width  int
	Height int

	// YOffset is the vertical scroll position.
	YOffset int

	// YPosition is the position of the viewport in relation to the terminal
	// window. It's used in high performance rendering only.
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
	return m.YOffset >= len(m.lines)-m.Height
}

// PastBottom returns whether or not the viewport is scrolled beyond the last
// line. This can happen when adjusting the viewport height.
func (m Model) PastBottom() bool {
	return m.YOffset > len(m.lines)-m.Height
}

// ScrollPercent returns the amount scrolled as a float between 0 and 1.
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
	s = strings.ReplaceAll(s, "\r\n", "\n") // normalize line endings
	m.lines = strings.Split(s, "\n")

	if m.YOffset > len(m.lines)-1 {
		m.GotoBottom()
	}
}

// visibleLines returns the lines that should currently be visible in the
// viewport.
func (m Model) visibleLines() (lines []string) {
	if len(m.lines) > 0 {
		top := max(0, m.YOffset)
		bottom := clamp(m.YOffset+m.Height, top, len(m.lines))
		lines = m.lines[top:bottom]
	}
	return lines
}

// scrollArea returns the scrollable boundaries for high performance rendering.
func (m Model) scrollArea() (top, bottom int) {
	top = max(0, m.YPosition)
	bottom = max(top, top+m.Height)
	if top > 0 && bottom > top {
		bottom--
	}
	return top, bottom
}

// SetYOffset sets the Y offset.
func (m *Model) SetYOffset(n int) {
	m.YOffset = clamp(n, 0, len(m.lines)-m.Height)
}

// ViewDown moves the view down by the number of lines in the viewport.
// Basically, "page down".
func (m *Model) ViewDown() []string {
	if m.AtBottom() {
		return nil
	}

	m.SetYOffset(m.YOffset + m.Height)
	return m.visibleLines()
}

// ViewUp moves the view up by one height of the viewport. Basically, "page up".
func (m *Model) ViewUp() []string {
	if m.AtTop() {
		return nil
	}

	m.SetYOffset(m.YOffset - m.Height)
	return m.visibleLines()
}

// HalfViewDown moves the view down by half the height of the viewport.
func (m *Model) HalfViewDown() (lines []string) {
	if m.AtBottom() {
		return nil
	}

	m.SetYOffset(m.YOffset + m.Height/2)
	return m.visibleLines()
}

// HalfViewUp moves the view up by half the height of the viewport.
func (m *Model) HalfViewUp() (lines []string) {
	if m.AtTop() {
		return nil
	}

	m.SetYOffset(m.YOffset - m.Height/2)
	return m.visibleLines()
}

// LineDown moves the view down by the given number of lines.
func (m *Model) LineDown(n int) (lines []string) {
	if m.AtBottom() || n == 0 {
		return nil
	}

	// Make sure the number of lines by which we're going to scroll isn't
	// greater than the number of lines we actually have left before we reach
	// the bottom.
	m.SetYOffset(m.YOffset + n)
	return m.visibleLines()
}

// LineUp moves the view down by the given number of lines. Returns the new
// lines to show.
func (m *Model) LineUp(n int) (lines []string) {
	if m.AtTop() || n == 0 {
		return nil
	}

	// Make sure the number of lines by which we're going to scroll isn't
	// greater than the number of lines we are from the top.
	m.SetYOffset(m.YOffset - n)
	return m.visibleLines()
}

// GotoTop sets the viewport to the top position.
func (m *Model) GotoTop() (lines []string) {
	if m.AtTop() {
		return nil
	}

	m.SetYOffset(0)
	return m.visibleLines()
}

// GotoBottom sets the viewport to the bottom position.
func (m *Model) GotoBottom() (lines []string) {
	m.SetYOffset(len(m.lines) - 1 - m.Height)
	return m.visibleLines()
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
	top, bottom := m.scrollArea()
	return tea.SyncScrollArea(m.visibleLines(), top, bottom)
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
	top, bottom := m.scrollArea()
	return tea.ScrollDown(lines, top, bottom)
}

// ViewUp is a high performance command the moves the viewport down by a given
// number of lines height. Use Model.ViewUp to get the lines that should be
// rendered.
func ViewUp(m Model, lines []string) tea.Cmd {
	if len(lines) == 0 {
		return nil
	}
	top, bottom := m.scrollArea()
	return tea.ScrollUp(lines, top, bottom)
}

// UPDATE

// Update runs the update loop with default keybindings similar to popular
// pagers. To define your own keybindings use the methods on Model (i.e.
// Model.LineDown()) and define your own update function.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
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
			lines := m.LineUp(mouseWheelDelta)
			if m.HighPerformanceRendering {
				cmd = ViewUp(m, lines)
			}

		case tea.MouseWheelDown:
			lines := m.LineDown(mouseWheelDelta)
			if m.HighPerformanceRendering {
				cmd = ViewDown(m, lines)
			}
		}
	}

	return m, cmd
}

// VIEW

// View renders the viewport into a string.
func (m Model) View() string {
	if m.HighPerformanceRendering {
		// Just send newlines since we're going to be rendering the actual
		// content seprately. We still need to send something that equals the
		// height of this view so that the Bubble Tea standard renderer can
		// position anything below this view properly.
		return strings.Repeat("\n", m.Height-1)
	}

	lines := m.visibleLines()

	// Fill empty space with newlines
	extraLines := ""
	if len(lines) < m.Height {
		extraLines = strings.Repeat("\n", max(0, m.Height-len(lines)))
	}

	return strings.Join(lines, "\n") + extraLines
}

// ETC

func clamp(v, low, high int) int {
	return min(high, max(low, v))
}

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
