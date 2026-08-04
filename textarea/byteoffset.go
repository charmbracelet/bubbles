package textarea

// ByteOffset returns the cursor's position as a byte offset into [Model.Value].
//
// [Model.Line] and [Model.Column] locate the cursor within the value's grid of
// rows and runes, which is the wrong coordinate space for a caller that wants
// to slice, splice, or scan the value as a flat string. Value joins the rows
// with "\n", so the offset is the byte length of every preceding row plus its
// separator, then the current row up to the cursor.
//
// The returned offset always lands on a UTF-8 boundary.
func (m Model) ByteOffset() int {
	var off int
	for i, row := range m.value {
		if i == m.row {
			return off + len(string(row[:clamp(m.col, 0, len(row))]))
		}
		off += len(string(row)) + 1 // +1 for the joining newline
	}
	return off
}

// SetCursorByteOffset places the cursor at a byte offset into [Model.Value],
// the inverse of [Model.ByteOffset]. A negative offset clamps to the start and
// an offset past the end clamps to the end.
//
// An offset landing inside a multi-byte rune, or on the "\n" that Value
// inserts between rows, is not a position the cursor can occupy; it snaps
// forward to the next one that is. Round-tripping an offset produced by
// ByteOffset is therefore always exact.
func (m *Model) SetCursorByteOffset(off int) {
	if off < 0 {
		off = 0
	}

	var seen int
	for row, line := range m.value {
		s := string(line)
		if off <= seen+len(s) {
			m.row = row
			// Count runes up to the offset rather than slicing at it: a
			// slice landing mid-rune would yield a partial sequence that
			// decodes to U+FFFD and miscount the column.
			col := 0
			for i := range s {
				if i >= off-seen {
					break
				}
				col++
			}
			if off-seen >= len(s) {
				col = len([]rune(s))
			}
			m.SetCursorColumn(col)
			// Every other cursor mover repositions the viewport; without
			// it the cursor can land scrolled out of sight after an edit
			// in a long value.
			m.repositionView()
			return
		}
		seen += len(s) + 1 // +1 for the joining newline
	}

	m.MoveToEnd()
}
