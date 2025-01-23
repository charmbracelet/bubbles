package viewport

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	defaultHorizontalStep = 6
)

// New returns a new model with the given width and height as well as default
// key mappings.
func New(width, height int) (m Model) {
	m.Width = width
	m.Height = height
	m.setInitialValues()
	return m
}

// Model is the Bubble Tea model for this viewport element.
type Model struct {
	Width  int
	Height int
	KeyMap KeyMap

	// Whether or not to wrap text. If false, it'll allow horizontal scrolling
	// instead.
	SoftWrap bool

	// Whether or not to fill to the height of the viewport with empty lines.
	FillHeight bool

	// Whether or not to respond to the mouse. The mouse must be enabled in
	// Bubble Tea for this to work. For details, see the Bubble Tea docs.
	MouseWheelEnabled bool

	// The number of lines the mouse wheel will scroll. By default, this is 3.
	MouseWheelDelta int

	// YOffset is the vertical scroll position.
	YOffset int

	// xOffset is the horizontal scroll position.
	xOffset int

	// horizontalStep is the number of columns we move left or right during a
	// default horizontal scroll.
	horizontalStep int

	// YPosition is the position of the viewport in relation to the terminal
	// window. It's used in high performance rendering only.
	YPosition int

	// Style applies a lipgloss style to the viewport. Realistically, it's most
	// useful for setting borders, margins and padding.
	Style lipgloss.Style

	// HighPerformanceRendering bypasses the normal Bubble Tea renderer to
	// provide higher performance rendering. Most of the time the normal Bubble
	// Tea rendering methods will suffice, but if you're passing content with
	// a lot of ANSI escape codes you may see improved rendering in certain
	// terminals with this enabled.
	//
	// This should only be used in program occupying the entire terminal,
	// which is usually via the alternate screen buffer.
	//
	// Deprecated: high performance rendering is now deprecated in Bubble Tea.
	HighPerformanceRendering bool

	// LeftGutterFunc allows to define a [GutterFunc] that adds a column into
	// the left of the viewport, which is kept when horizontal scrolling.
	// This can be used for things like line numbers, selection indicators,
	// show statuses, etc.
	LeftGutterFunc GutterFunc

	initialized      bool
	lines            []string
	longestLineWidth int

	// HighlightStyle highlights the ranges set with [SetHighligths].
	HighlightStyle lipgloss.Style

	// SelectedHighlightStyle highlights the highlight range focused during
	// navigation.
	// Use [SetHighligths] to set the highlight ranges, and [HightlightNext]
	// and [HihglightPrevious] to navigate.
	SelectedHighlightStyle lipgloss.Style

	highlights           []highlightInfo
	hiIdx                int
	memoizedMatchedLines []string
}

// GutterFunc can be implemented and set into [Model.LeftGutterFunc].
type GutterFunc func(GutterContext) string

// LineNumberGutter return a [GutterFunc] that shows line numbers.
func LineNumberGutter(style lipgloss.Style) GutterFunc {
	return func(info GutterContext) string {
		if info.Soft {
			return style.Render("     │ ")
		}
		if info.Index >= info.TotalLines {
			return style.Render("   ~ │ ")
		}
		return style.Render(fmt.Sprintf("%4d │ ", info.Index+1))
	}
}

// NoGutter is the default gutter used.
var NoGutter = func(GutterContext) string { return "" }

// GutterContext provides context to a [GutterFunc].
type GutterContext struct {
	Index      int
	TotalLines int
	Soft       bool
}

func (m *Model) setInitialValues() {
	m.KeyMap = DefaultKeyMap()
	m.MouseWheelEnabled = true
	m.MouseWheelDelta = 3
	m.initialized = true
	m.horizontalStep = defaultHorizontalStep
	m.LeftGutterFunc = NoGutter
}

// Init exists to satisfy the tea.Model interface for composability purposes.
func (m Model) Init() tea.Cmd {
	return nil
}

// AtTop returns whether or not the viewport is at the very top position.
func (m Model) AtTop() bool {
	return m.YOffset <= 0
}

// AtBottom returns whether or not the viewport is at or past the very bottom
// position.
func (m Model) AtBottom() bool {
	return m.YOffset >= m.maxYOffset()
}

// PastBottom returns whether or not the viewport is scrolled beyond the last
// line. This can happen when adjusting the viewport height.
func (m Model) PastBottom() bool {
	return m.YOffset > m.maxYOffset()
}

// ScrollPercent returns the amount scrolled as a float between 0 and 1.
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

// HorizontalScrollPercent returns the amount horizontally scrolled as a float
// between 0 and 1.
func (m Model) HorizontalScrollPercent() float64 {
	if m.xOffset >= m.longestLineWidth-m.Width {
		return 1.0
	}
	y := float64(m.xOffset)
	h := float64(m.Width)
	t := float64(m.longestLineWidth)
	v := y / (t - h)
	return math.Max(0.0, math.Min(1.0, v))
}

// SetContent set the pager's text content.
// Line endings will be normalized to '\n'.
func (m *Model) SetContent(s string) {
	s = strings.ReplaceAll(s, "\r\n", "\n") // normalize line endings
	m.lines = strings.Split(s, "\n")
	// if there's no content, set content to actual nil instead of one empty
	// line.
	if len(m.lines) == 1 && ansi.StringWidth(m.lines[0]) == 0 {
		m.lines = nil
	}
	m.longestLineWidth = maxLineWidth(m.lines)
	m.ClearHighlights()

	if m.YOffset > len(m.lines)-1 {
		m.GotoBottom()
	}
}

// GetContent returns the entire content as a single string.
// Line endings are normalized to '\n'.
func (m Model) GetContent() string {
	return strings.Join(m.lines, "\n")
}

// maxYOffset returns the maximum possible value of the y-offset based on the
// viewport's content and set height.
func (m Model) maxYOffset() int {
	return max(0, len(m.lines)-m.Height+m.Style.GetVerticalFrameSize())
}

// maxXOffset returns the maximum possible value of the x-offset based on the
// viewport's content and set width.
func (m Model) maxXOffset() int {
	return max(0, m.longestLineWidth-m.Width)
}

func (m Model) maxWidth() int {
	return m.Width -
		m.Style.GetHorizontalFrameSize() -
		lipgloss.Width(m.LeftGutterFunc(GutterContext{}))
}

func (m Model) maxHeight() int {
	return m.Height - m.Style.GetVerticalFrameSize()
}

// visibleLines returns the lines that should currently be visible in the
// viewport.
func (m Model) visibleLines() (lines []string) {
	maxHeight := m.maxHeight()
	maxWidth := m.maxWidth()

	if len(m.lines) > 0 {
		top := max(0, m.YOffset)
		bottom := clamp(m.YOffset+maxHeight, top, len(m.lines))
		lines = make([]string, bottom-top)
		copy(lines, m.lines[top:bottom])
		lines = m.highlighLines(lines, top)
	}

	for m.FillHeight && len(lines) < maxHeight {
		lines = append(lines, "")
	}

	// if longest line fit within width, no need to do anything else.
	if (m.xOffset == 0 && m.longestLineWidth <= maxWidth) || maxWidth == 0 {
		return m.prependColumn(lines)
	}

	if m.SoftWrap {
		return m.softWrap(lines, maxWidth)
	}

	for i := range lines {
		lines[i] = ansi.Cut(lines[i], m.xOffset, m.xOffset+maxWidth)
	}
	return m.prependColumn(lines)
}

func (m Model) highlighLines(lines []string, offset int) []string {
	if len(m.highlights) == 0 {
		return lines
	}
	for i := range lines {
		if memoized := m.memoizedMatchedLines[i+offset]; memoized != "" {
			lines[i] = memoized
		} else {
			ranges := makeHilightRanges(
				m.highlights,
				i+offset,
				m.HighlightStyle,
			)
			lines[i] = lipgloss.StyleRanges(lines[i], ranges...)
			m.memoizedMatchedLines[i+offset] = lines[i]
		}
		if m.hiIdx < 0 {
			continue
		}
		sel := m.highlights[m.hiIdx]
		if hi, ok := sel.lines[i+offset]; ok {
			lines[i] = lipgloss.StyleRanges(lines[i], lipgloss.NewRange(
				hi[0],
				hi[1],
				m.SelectedHighlightStyle,
			))
		}
	}
	return lines
}

func (m Model) softWrap(lines []string, maxWidth int) []string {
	var wrappedLines []string
	for i, line := range lines {
		idx := 0
		for ansi.StringWidth(line) >= idx {
			truncatedLine := ansi.Cut(line, idx, maxWidth+idx)
			wrappedLines = append(wrappedLines, m.LeftGutterFunc(GutterContext{
				Index:      i + m.YOffset,
				TotalLines: m.TotalLineCount(),
				Soft:       idx > 0,
			})+truncatedLine)
			idx += maxWidth
		}
	}
	return wrappedLines
}

func (m Model) prependColumn(lines []string) []string {
	result := make([]string, len(lines))
	for i, line := range lines {
		result[i] = m.LeftGutterFunc(GutterContext{
			Index:      i + m.YOffset,
			TotalLines: m.TotalLineCount(),
		}) + line
	}
	return result
}

// scrollArea returns the scrollable boundaries for high performance rendering.
//
// XXX: high performance rendering is deprecated in Bubble Tea.
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
	m.YOffset = clamp(n, 0, m.maxYOffset())
}

// SetXOffset sets the X offset.
// No-op when soft wrap is enabled.
func (m *Model) SetXOffset(n int) {
	if m.SoftWrap {
		return
	}
	m.xOffset = clamp(n, 0, m.maxXOffset())
}

// EnsureVisible ensures that the given line and column are in the viewport.
func (m *Model) EnsureVisible(line, colstart, colend int) {
	maxWidth := m.maxWidth()
	if colend <= maxWidth {
		m.SetXOffset(0)
	} else {
		m.SetXOffset(colstart - m.horizontalStep) // put one step to the left, feels more natural
	}

	if line < m.YOffset || line >= m.YOffset+m.maxHeight() {
		m.SetYOffset(line)
	}

	m.visibleLines()
}

// ViewDown moves the view down by the number of lines in the viewport.
// Basically, "page down".
func (m *Model) ViewDown() []string {
	if m.AtBottom() {
		return nil
	}

	return m.LineDown(m.Height)
}

// ViewUp moves the view up by one height of the viewport. Basically, "page up".
func (m *Model) ViewUp() []string {
	if m.AtTop() {
		return nil
	}

	return m.LineUp(m.Height)
}

// HalfViewDown moves the view down by half the height of the viewport.
func (m *Model) HalfViewDown() (lines []string) {
	if m.AtBottom() {
		return nil
	}

	return m.LineDown(m.Height / 2) //nolint:mnd
}

// HalfViewUp moves the view up by half the height of the viewport.
func (m *Model) HalfViewUp() (lines []string) {
	if m.AtTop() {
		return nil
	}

	return m.LineUp(m.Height / 2) //nolint:mnd
}

// LineDown moves the view down by the given number of lines.
func (m *Model) LineDown(n int) (lines []string) {
	if m.AtBottom() || n == 0 || len(m.lines) == 0 {
		return nil
	}

	// Make sure the number of lines by which we're going to scroll isn't
	// greater than the number of lines we actually have left before we reach
	// the bottom.
	m.SetYOffset(m.YOffset + n)
	m.hiIdx = m.findNearedtMatch()

	// Gather lines to send off for performance scrolling.
	//
	// XXX: high performance rendering is deprecated in Bubble Tea.
	bottom := clamp(m.YOffset+m.Height, 0, len(m.lines))
	top := clamp(m.YOffset+m.Height-n, 0, bottom)
	return m.lines[top:bottom]
}

// LineUp moves the view down by the given number of lines. Returns the new
// lines to show.
func (m *Model) LineUp(n int) (lines []string) {
	if m.AtTop() || n == 0 || len(m.lines) == 0 {
		return nil
	}

	// Make sure the number of lines by which we're going to scroll isn't
	// greater than the number of lines we are from the top.
	m.SetYOffset(m.YOffset - n)
	m.hiIdx = m.findNearedtMatch()

	// Gather lines to send off for performance scrolling.
	//
	// XXX: high performance rendering is deprecated in Bubble Tea.
	top := max(0, m.YOffset)
	bottom := clamp(m.YOffset+n, 0, m.maxYOffset())
	return m.lines[top:bottom]
}

// TotalLineCount returns the total number of lines (both hidden and visible) within the viewport.
func (m Model) TotalLineCount() int {
	return len(m.lines)
}

// VisibleLineCount returns the number of the visible lines within the viewport.
func (m Model) VisibleLineCount() int {
	return len(m.visibleLines())
}

// GotoTop sets the viewport to the top position.
func (m *Model) GotoTop() (lines []string) {
	if m.AtTop() {
		return nil
	}

	m.SetYOffset(0)
	m.hiIdx = m.findNearedtMatch()
	return m.visibleLines()
}

// GotoBottom sets the viewport to the bottom position.
func (m *Model) GotoBottom() (lines []string) {
	m.SetYOffset(m.maxYOffset())
	m.hiIdx = m.findNearedtMatch()
	return m.visibleLines()
}

// Sync tells the renderer where the viewport will be located and requests
// a render of the current state of the viewport. It should be called for the
// first render and after a window resize.
//
// For high performance rendering only.
//
// Deprecated: high performance rendering is deprecated in Bubble Tea.
func Sync(m Model) tea.Cmd {
	if len(m.lines) == 0 {
		return nil
	}
	top, bottom := m.scrollArea()
	return tea.SyncScrollArea(m.visibleLines(), top, bottom)
}

// ViewDown is a high performance command that moves the viewport up by a given
// number of lines. Use Model.ViewDown to get the lines that should be rendered.
// For example:
//
//	lines := model.ViewDown(1)
//	cmd := ViewDown(m, lines)
func ViewDown(m Model, lines []string) tea.Cmd {
	if len(lines) == 0 {
		return nil
	}
	top, bottom := m.scrollArea()

	// XXX: high performance rendering is deprecated in Bubble Tea. In a v2 we
	// won't need to return a command here.
	return tea.ScrollDown(lines, top, bottom) //nolint:staticcheck
}

// ViewUp is a high performance command the moves the viewport down by a given
// number of lines height. Use Model.ViewUp to get the lines that should be
// rendered.
func ViewUp(m Model, lines []string) tea.Cmd {
	if len(lines) == 0 {
		return nil
	}
	top, bottom := m.scrollArea()

	// XXX: high performance rendering is deprecated in Bubble Tea. In a v2 we
	// won't need to return a command here.
	return tea.ScrollUp(lines, top, bottom) //nolint:staticcheck
}

// SetHorizontalStep sets the amount of cells that the viewport moves in the
// default viewport keymapping. If set to 0 or less, horizontal scrolling is
// disabled.
func (m *Model) SetHorizontalStep(n int) {
	if n < 0 {
		n = 0
	}

	m.horizontalStep = n
}

// MoveLeft moves the viewport to the left by the given number of columns.
func (m *Model) MoveLeft(cols int) {
	m.xOffset -= cols
	if m.xOffset < 0 {
		m.xOffset = 0
	}
}

// MoveRight moves viewport to the right by the given number of columns.
func (m *Model) MoveRight(cols int) {
	// prevents over scrolling to the right
	w := m.maxWidth()
	if m.xOffset > m.longestLineWidth-w {
		return
	}
	m.xOffset += cols
}

// Resets lines indent to zero.
func (m *Model) ResetIndent() {
	m.xOffset = 0
}

// SetHighlights sets ranges of characters to highlight.
// For instance, `[]int{[]int{2, 10}, []int{20, 30}}` will highlight characters
// 2 to 10 and 20 to 30.
// Note that highlights are not expected to transpose each other, and are also
// expected to be in order.
// Use [Model.SetHighlights] to set the highlight ranges, and
// [Model.HighlightNext] and [Model.HighlightPrevious] to navigate.
// Use [Model.ClearHighlights] to remove all highlights.
func (m *Model) SetHighlights(matches [][]int) {
	if len(matches) == 0 || len(m.lines) == 0 {
		return
	}
	m.memoizedMatchedLines = make([]string, len(m.lines))
	m.highlights = parseMatches(m.GetContent(), matches)
	m.hiIdx = m.findNearedtMatch()
	m.showHighlight()
}

// ClearHighlights clears previously set highlights.
func (m *Model) ClearHighlights() {
	m.memoizedMatchedLines = nil
	m.highlights = nil
	m.hiIdx = -1
}

func (m *Model) showHighlight() {
	if m.hiIdx == -1 {
		return
	}
	line, colstart, colend := m.highlights[m.hiIdx].coords()
	m.EnsureVisible(line, colstart, colend)
}

// HighlightNext highlights the next match.
func (m *Model) HighlightNext() {
	if m.highlights == nil {
		return
	}

	m.hiIdx = (m.hiIdx + 1) % len(m.highlights)
	m.showHighlight()
}

// HighlightPrevious highlights the previous match.
func (m *Model) HighlightPrevious() {
	if m.highlights == nil {
		return
	}

	m.hiIdx = (m.hiIdx - 1 + len(m.highlights)) % len(m.highlights)
	m.showHighlight()
}

func (m Model) findNearedtMatch() int {
	for i, match := range m.highlights {
		if match.lineStart >= m.YOffset {
			return i
		}
	}
	return -1
}

// Update handles standard message-based viewport updates.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m, cmd = m.updateAsModel(msg)
	return m, cmd
}

// Author's note: this method has been broken out to make it easier to
// potentially transition Update to satisfy tea.Model.
func (m Model) updateAsModel(msg tea.Msg) (Model, tea.Cmd) {
	if !m.initialized {
		m.setInitialValues()
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.PageDown):
			lines := m.ViewDown()
			if m.HighPerformanceRendering {
				cmd = ViewDown(m, lines)
			}

		case key.Matches(msg, m.KeyMap.PageUp):
			lines := m.ViewUp()
			if m.HighPerformanceRendering {
				cmd = ViewUp(m, lines)
			}

		case key.Matches(msg, m.KeyMap.HalfPageDown):
			lines := m.HalfViewDown()
			if m.HighPerformanceRendering {
				cmd = ViewDown(m, lines)
			}

		case key.Matches(msg, m.KeyMap.HalfPageUp):
			lines := m.HalfViewUp()
			if m.HighPerformanceRendering {
				cmd = ViewUp(m, lines)
			}

		case key.Matches(msg, m.KeyMap.Down):
			lines := m.LineDown(1)
			if m.HighPerformanceRendering {
				cmd = ViewDown(m, lines)
			}

		case key.Matches(msg, m.KeyMap.Up):
			lines := m.LineUp(1)
			if m.HighPerformanceRendering {
				cmd = ViewUp(m, lines)
			}

		case key.Matches(msg, m.KeyMap.Left):
			m.MoveLeft(m.horizontalStep)

		case key.Matches(msg, m.KeyMap.Right):
			m.MoveRight(m.horizontalStep)
		}

	case tea.MouseMsg:
		if !m.MouseWheelEnabled || msg.Action != tea.MouseActionPress {
			break
		}
		switch msg.Button { //nolint:exhaustive
		case tea.MouseButtonWheelUp:
			lines := m.LineUp(m.MouseWheelDelta)
			if m.HighPerformanceRendering {
				cmd = ViewUp(m, lines)
			}

		case tea.MouseButtonWheelDown:
			lines := m.LineDown(m.MouseWheelDelta)
			if m.HighPerformanceRendering {
				cmd = ViewDown(m, lines)
			}
		}
	}

	return m, cmd
}

// View renders the viewport into a string.
func (m Model) View() string {
	if m.HighPerformanceRendering {
		// Just send newlines since we're going to be rendering the actual
		// content separately. We still need to send something that equals the
		// height of this view so that the Bubble Tea standard renderer can
		// position anything below this view properly.
		return strings.Repeat("\n", max(0, m.Height-1))
	}

	w, h := m.Width, m.Height
	if sw := m.Style.GetWidth(); sw != 0 {
		w = min(w, sw)
	}
	if sh := m.Style.GetHeight(); sh != 0 {
		h = min(h, sh)
	}
	contentWidth := w - m.Style.GetHorizontalFrameSize()
	contentHeight := h - m.Style.GetVerticalFrameSize()
	contents := lipgloss.NewStyle().
		Width(contentWidth).      // pad to width.
		Height(contentHeight).    // pad to height.
		MaxHeight(contentHeight). // truncate height if taller.
		MaxWidth(contentWidth).   // truncate width if wider.
		Render(strings.Join(m.visibleLines(), "\n"))
	return m.Style.
		UnsetWidth().UnsetHeight(). // Style size already applied in contents.
		Render(contents)
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
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

func maxLineWidth(lines []string) int {
	maxlen := 0
	for _, line := range lines {
		llen := ansi.StringWidth(line)
		if llen > maxlen {
			maxlen = llen
		}
	}
	return maxlen
}
