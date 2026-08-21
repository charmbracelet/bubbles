package textarea

import (
	"testing"
	"unicode/utf8"
)

func TestByteOffsetSingleLine(t *testing.T) {
	t.Parallel()

	m := New()
	m.SetValue("hello")

	for _, col := range []int{0, 1, 5} {
		m.SetCursorColumn(col)
		if got := m.ByteOffset(); got != col {
			t.Errorf("col %d: ByteOffset() = %d, want %d", col, got, col)
		}
	}
}

func TestByteOffsetCountsPrecedingRowsAndNewlines(t *testing.T) {
	t.Parallel()

	m := New()
	m.SetValue("ab\ncd\nef")

	// Row 2 col 1 → "ab"(2) + "\n"(1) + "cd"(2) + "\n"(1) + "e"(1) = 7.
	m.CursorDown()
	m.CursorDown()
	m.SetCursorColumn(1)

	if got, want := m.ByteOffset(), 7; got != want {
		t.Errorf("ByteOffset() = %d, want %d", got, want)
	}
	if got, want := m.Value()[m.ByteOffset():], "f"; got != want {
		t.Errorf("Value()[off:] = %q, want %q", got, want)
	}
}

// TestByteOffsetIsBytesNotRunes is the distinction the API exists for:
// Column counts runes, ByteOffset counts bytes, and they diverge the moment
// a multi-byte rune precedes the cursor.
func TestByteOffsetIsBytesNotRunes(t *testing.T) {
	t.Parallel()

	m := New()
	m.SetValue("世界x")
	m.SetCursorColumn(2) // after both CJK runes

	if got, want := m.Column(), 2; got != want {
		t.Errorf("Column() = %d, want %d", got, want)
	}
	// Each CJK rune is 3 bytes in UTF-8.
	if got, want := m.ByteOffset(), 6; got != want {
		t.Errorf("ByteOffset() = %d, want %d", got, want)
	}
	if got, want := m.Value()[m.ByteOffset():], "x"; got != want {
		t.Errorf("Value()[off:] = %q, want %q", got, want)
	}
}

// TestByteOffsetSlicesValueCleanly is the property callers actually rely on:
// the offset is a valid index into Value() at every cursor position, and the
// two halves rejoin to the original.
func TestByteOffsetSlicesValueCleanly(t *testing.T) {
	t.Parallel()

	const value = "ab 世界\n👍🏽 cd\n\nef"

	m := New()
	m.SetValue(value)

	for row := range m.value {
		m.row = row
		for col := 0; col <= len(m.value[row]); col++ {
			m.SetCursorColumn(col)
			off := m.ByteOffset()

			v := m.Value()
			if off < 0 || off > len(v) {
				t.Fatalf("row %d col %d: offset %d out of range for %d bytes", row, col, off, len(v))
			}
			if !utf8.RuneStart(v[off%max(len(v), 1)]) && off != len(v) {
				t.Errorf("row %d col %d: offset %d lands mid-rune", row, col, off)
			}
			if got := v[:off] + v[off:]; got != v {
				t.Errorf("row %d col %d: split/rejoin = %q, want %q", row, col, got, v)
			}
		}
	}
}

func TestSetCursorByteOffsetRoundTrips(t *testing.T) {
	t.Parallel()

	const value = "ab 世界\n👍🏽 cd\n\nef"

	m := New()
	m.SetValue(value)

	for row := range m.value {
		m.row = row
		for col := 0; col <= len(m.value[row]); col++ {
			m.SetCursorColumn(col)
			wantRow, wantCol, off := m.Line(), m.Column(), m.ByteOffset()

			// Move somewhere else, then restore purely from the offset.
			m.MoveToBegin()
			m.SetCursorByteOffset(off)

			if m.Line() != wantRow || m.Column() != wantCol {
				t.Errorf("offset %d: restored to row %d col %d, want row %d col %d",
					off, m.Line(), m.Column(), wantRow, wantCol)
			}
			if got := m.ByteOffset(); got != off {
				t.Errorf("offset %d round-tripped to %d", off, got)
			}
		}
	}
}

func TestSetCursorByteOffsetClamps(t *testing.T) {
	t.Parallel()

	m := New()
	m.SetValue("ab\ncd")

	m.SetCursorByteOffset(-5)
	if m.Line() != 0 || m.Column() != 0 {
		t.Errorf("negative offset: row %d col %d, want row 0 col 0", m.Line(), m.Column())
	}

	m.SetCursorByteOffset(9999)
	if got, want := m.ByteOffset(), len(m.Value()); got != want {
		t.Errorf("past-end offset: ByteOffset() = %d, want %d", got, want)
	}
}

// TestSetCursorByteOffsetMidRuneSnapsForward pins the documented behavior for
// an offset that is not a cursor position. Slicing the row at a mid-rune byte
// would produce a partial UTF-8 sequence, decode it to U+FFFD, and land the
// cursor a rune short.
func TestSetCursorByteOffsetMidRuneSnapsForward(t *testing.T) {
	t.Parallel()

	m := New()
	m.SetValue("世界") // two 3-byte runes

	for _, off := range []int{1, 2} {
		m.MoveToBegin()
		m.SetCursorByteOffset(off)
		if got, want := m.Column(), 1; got != want {
			t.Errorf("mid-rune offset %d: Column() = %d, want %d", off, got, want)
		}
		if got, want := m.ByteOffset(), 3; got != want {
			t.Errorf("mid-rune offset %d: ByteOffset() = %d, want %d", off, got, want)
		}
	}
}

// TestSetCursorByteOffsetOnNewlineSnapsForward covers the other non-position:
// the "\n" Value inserts between rows is not part of any row.
func TestSetCursorByteOffsetOnNewlineSnapsForward(t *testing.T) {
	t.Parallel()

	m := New()
	m.SetValue("ab\ncd")

	// Offset 2 is the end of row 0; offset 3 is the start of row 1. The
	// separator itself occupies neither.
	m.SetCursorByteOffset(2)
	if m.Line() != 0 || m.Column() != 2 {
		t.Errorf("offset 2: row %d col %d, want row 0 col 2", m.Line(), m.Column())
	}

	m.SetCursorByteOffset(3)
	if m.Line() != 1 || m.Column() != 0 {
		t.Errorf("offset 3: row %d col %d, want row 1 col 0", m.Line(), m.Column())
	}
}

func TestSetCursorByteOffsetEmptyValue(t *testing.T) {
	t.Parallel()

	m := New()

	m.SetCursorByteOffset(0)
	if got := m.ByteOffset(); got != 0 {
		t.Errorf("ByteOffset() = %d, want 0", got)
	}

	m.SetCursorByteOffset(10)
	if got := m.ByteOffset(); got != 0 {
		t.Errorf("past-end on empty value: ByteOffset() = %d, want 0", got)
	}
}
