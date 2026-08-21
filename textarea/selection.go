package textarea

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Position is a location in the textarea's logical buffer: Row indexes a
// logical (unwrapped) line, Col indexes a rune within that line. Col may
// equal the line length, meaning "just past the last rune".
type Position struct {
	Row int
	Col int
}

// before reports whether p sorts before q in the buffer.
func (p Position) before(q Position) bool {
	if p.Row != q.Row {
		return p.Row < q.Row
	}
	return p.Col < q.Col
}

// gutterWidth returns the number of columns rendered to the left of the text
// content: the prompt, plus the line-number column when enabled. Mouse
// coordinates are offset by this much before being mapped to a column.
func (m Model) gutterWidth() int {
	w := m.promptWidth
	if m.ShowLineNumbers {
		w += len(strconv.Itoa(m.MaxHeight)) + 2
	}
	return w
}

// runeIndexForColumn returns the index of the rune occupying the given display
// column within runes, accounting for double-width runes. A column past the
// end of the line yields len(runes).
func runeIndexForColumn(runes []rune, col int) int {
	if col <= 0 {
		return 0
	}
	w := 0
	for i, r := range runes {
		rWidth := ansi.StringWidth(string(r))
		if w+rWidth > col {
			return i
		}
		w += rWidth
	}
	return len(runes)
}

// PositionAt maps coordinates within the textarea's rendered area to a buffer
// position. Coordinates are relative to the textarea itself: (0, 0) is its
// top-left cell, including the prompt/line-number gutter. Callers that render
// the textarea at an offset must subtract that offset first.
//
// Coordinates outside the content resolve to the nearest position: above the
// first line yields the start of the buffer, below the last yields its end,
// and a column past a line's text yields the end of that line.
func (m Model) PositionAt(x, y int) Position {
	if len(m.value) == 0 {
		return Position{}
	}

	targetLine := y + m.viewport.YOffset()
	if targetLine < 0 {
		return Position{}
	}

	contentX := x - m.gutterWidth()
	if contentX < 0 {
		contentX = 0
	}

	display := 0
	for row, line := range m.value {
		wrapped := m.memoizedWrap(line, m.width)
		if len(wrapped) == 0 {
			wrapped = [][]rune{{}}
		}
		base := 0
		for _, wrappedLine := range wrapped {
			if display == targetLine {
				col := base + runeIndexForColumn(wrappedLine, contentX)
				return Position{Row: row, Col: clamp(col, 0, len(line))}
			}
			base += len(wrappedLine)
			display++
		}
	}

	lastRow := len(m.value) - 1
	return Position{Row: lastRow, Col: len(m.value[lastRow])}
}

// BeginSelection starts a selection at the given textarea-relative
// coordinates, discarding any previous selection, and moves the cursor there.
// Pair it with [Model.ExtendSelection] as the pointer moves and
// [Model.EndSelection] when the drag finishes.
//
// See [Model.PositionAt] for the coordinate convention.
func (m *Model) BeginSelection(x, y int) {
	pos := m.PositionAt(x, y)
	m.selectFrom(pos, pos)
	m.selecting = true
	m.moveCursorTo(pos)
}

// ExtendSelection extends an in-progress selection to the given
// textarea-relative coordinates and moves the cursor there. It is a no-op
// unless [Model.BeginSelection] started a drag.
func (m *Model) ExtendSelection(x, y int) {
	if !m.selecting {
		return
	}
	pos := m.PositionAt(x, y)
	m.selHead = pos
	m.hasSelection = true
	m.moveCursorTo(pos)
}

// EndSelection completes an in-progress drag. The selection itself is
// retained so it can be read with [Model.SelectedText]; a zero-width
// selection (a plain click) is discarded.
func (m *Model) EndSelection() {
	m.selecting = false
	if m.selAnchor == m.selHead {
		m.ClearSelection()
	}
}

// SelectAll selects the entire buffer.
func (m *Model) SelectAll() {
	if len(m.value) == 0 {
		return
	}
	lastRow := len(m.value) - 1
	m.selectFrom(
		Position{Row: 0, Col: 0},
		Position{Row: lastRow, Col: len(m.value[lastRow])},
	)
	m.selecting = false
}

// ClearSelection removes the current selection, if any.
func (m *Model) ClearSelection() {
	m.hasSelection = false
	m.selecting = false
	m.selAnchor = Position{}
	m.selHead = Position{}
}

// HasSelection reports whether a non-empty selection is active.
func (m Model) HasSelection() bool {
	return m.hasSelection && m.selAnchor != m.selHead
}

// Selection returns the selected range, normalized so start sorts before end,
// and whether a non-empty selection is active.
func (m Model) Selection() (start, end Position, ok bool) {
	if !m.HasSelection() {
		return Position{}, Position{}, false
	}
	start, end = m.selAnchor, m.selHead
	if end.before(start) {
		start, end = end, start
	}
	return start, end, true
}

// SelectedText returns the selected text, with logical lines joined by "\n".
// It returns the empty string when nothing is selected.
func (m Model) SelectedText() string {
	start, end, ok := m.Selection()
	if !ok {
		return ""
	}

	if start.Row == end.Row {
		line := m.value[start.Row]
		return string(line[clamp(start.Col, 0, len(line)):clamp(end.Col, 0, len(line))])
	}

	var b strings.Builder
	for row := start.Row; row <= end.Row; row++ {
		line := m.value[row]
		switch row {
		case start.Row:
			b.WriteString(string(line[clamp(start.Col, 0, len(line)):]))
		case end.Row:
			b.WriteString(string(line[:clamp(end.Col, 0, len(line))]))
		default:
			b.WriteString(string(line))
		}
		if row < end.Row {
			b.WriteRune('\n')
		}
	}
	return b.String()
}

// selectFrom sets the selection anchor and head.
func (m *Model) selectFrom(anchor, head Position) {
	m.selAnchor = anchor
	m.selHead = head
	m.hasSelection = true
}

// moveCursorTo places the cursor at the given buffer position, clamped to the
// buffer's bounds.
func (m *Model) moveCursorTo(pos Position) {
	if len(m.value) == 0 {
		return
	}
	row := clamp(pos.Row, 0, len(m.value)-1)
	m.row = row
	m.col = clamp(pos.Col, 0, len(m.value[row]))
}

// selectionSpanFor returns the half-open rune range [from, to) of the given
// logical row that is selected, restricted to the wrapped segment covering
// [base, base+length). ok is false when the segment holds no selected runes.
func (m Model) selectionSpanFor(row, base, length int) (from, to int, ok bool) {
	start, end, active := m.Selection()
	if !active || row < start.Row || row > end.Row {
		return 0, 0, false
	}

	lineLen := len(m.value[row])

	rowFrom, rowTo := 0, lineLen
	if row == start.Row {
		rowFrom = clamp(start.Col, 0, lineLen)
	}
	if row == end.Row {
		rowTo = clamp(end.Col, 0, lineLen)
	}

	from = max(rowFrom, base)
	to = min(rowTo, base+length)
	if from >= to {
		return 0, 0, false
	}
	return from - base, to - base, true
}
