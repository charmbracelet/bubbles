package textinput

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// Cursor() must use display width of the visible prefix (value[offset:pos]),
// not the rune index. Fixes #906 (wide/CJK) and #1001 (horizontal scroll).
func TestCursorXUsesDisplayWidthAndScrollOffset(t *testing.T) {
	m := New()
	m.Focus()
	m.SetVirtualCursor(false)
	m.CharLimit = 200

	// Wide characters: three CJK runes occupy 6 terminal columns.
	m.SetValue("あいう")
	m.CursorEnd()
	cur := m.Cursor()
	if cur == nil {
		t.Fatal("Cursor() returned nil")
	}
	want := lipgloss.Width(m.promptView()) + lipgloss.Width("あいう")
	if cur.X != want {
		t.Fatalf("CJK Cursor().X = %d, want %d (display width, not rune count %d)",
			cur.X, want, m.Position())
	}

	// Overflow + scroll: ASCII so rune width == display width; verifies offset.
	m.SetWidth(20)
	m.SetValue(strings.Repeat("a", 30))
	m.CursorEnd()
	m.SetCursor(25)
	if m.offset == 0 {
		t.Fatal("expected non-zero scroll offset")
	}
	cur = m.Cursor()
	if cur == nil {
		t.Fatal("Cursor() returned nil after scroll setup")
	}
	want = lipgloss.Width(m.promptView()) + lipgloss.Width(string(m.value[m.offset:m.pos]))
	if cur.X != want {
		t.Fatalf("scrolled Cursor().X = %d, want %d (pos=%d offset=%d)",
			cur.X, want, m.pos, m.offset)
	}

	// Overflow + scroll with wide runes before the cursor.
	m.SetWidth(10)
	m.SetValue(strings.Repeat("あ", 20)) // 40 columns
	m.CursorEnd()
	m.SetCursor(15) // rune index 15 → 30 columns from start
	if m.offset == 0 {
		t.Fatal("expected non-zero scroll offset for wide runes")
	}
	cur = m.Cursor()
	if cur == nil {
		t.Fatal("Cursor() returned nil for wide+scroll")
	}
	want = lipgloss.Width(m.promptView()) + lipgloss.Width(string(m.value[m.offset:m.pos]))
	if cur.X != want {
		t.Fatalf("wide+scroll Cursor().X = %d, want %d (pos=%d offset=%d)",
			cur.X, want, m.pos, m.offset)
	}
}
