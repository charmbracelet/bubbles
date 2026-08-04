package textarea

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// newSelectionTextarea returns a focused textarea with a known geometry and
// no prompt, so tests can reason about columns directly.
func newSelectionTextarea(t *testing.T, value string) Model {
	t.Helper()
	ta := New()
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.SetWidth(20)
	ta.SetHeight(6)
	ta.Focus()
	ta.SetValue(value)
	return ta
}

func TestPositionAt(t *testing.T) {
	t.Parallel()

	ta := newSelectionTextarea(t, "hello\nworld")

	tests := []struct {
		name string
		x, y int
		want Position
	}{
		{"origin", 0, 0, Position{Row: 0, Col: 0}},
		{"mid first line", 3, 0, Position{Row: 0, Col: 3}},
		{"start second line", 0, 1, Position{Row: 1, Col: 0}},
		{"mid second line", 2, 1, Position{Row: 1, Col: 2}},
		{"past end of line clamps to line end", 40, 0, Position{Row: 0, Col: 5}},
		{"negative y clamps to buffer start", 0, -3, Position{Row: 0, Col: 0}},
		{"below last line clamps to buffer end", 0, 50, Position{Row: 1, Col: 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ta.PositionAt(tt.x, tt.y); got != tt.want {
				t.Errorf("PositionAt(%d, %d) = %+v, want %+v", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

// TestPositionAtAccountsForGutter asserts that the prompt and line-number
// columns are subtracted before mapping a click to a text column. Without
// this, every click lands N columns to the right of where the user aimed.
func TestPositionAtAccountsForGutter(t *testing.T) {
	t.Parallel()

	ta := New()
	ta.Prompt = "> " // 2 columns
	ta.ShowLineNumbers = false
	ta.SetWidth(20)
	ta.SetHeight(3)
	ta.Focus()
	ta.SetValue("hello")

	// Clicking the first text cell is x=2 because the prompt occupies 0..1.
	if got := ta.PositionAt(2, 0); got != (Position{Row: 0, Col: 0}) {
		t.Errorf("first text cell = %+v, want row 0 col 0", got)
	}
	if got := ta.PositionAt(5, 0); got != (Position{Row: 0, Col: 3}) {
		t.Errorf("fourth text cell = %+v, want row 0 col 3", got)
	}
	// A click inside the gutter clamps to the start of the line rather than
	// producing a negative column.
	if got := ta.PositionAt(0, 0); got != (Position{Row: 0, Col: 0}) {
		t.Errorf("click in gutter = %+v, want row 0 col 0", got)
	}
}

func TestSelectedTextSingleLine(t *testing.T) {
	t.Parallel()

	ta := newSelectionTextarea(t, "hello world")
	ta.BeginSelection(0, 0)
	ta.ExtendSelection(5, 0)
	ta.EndSelection()

	if !ta.HasSelection() {
		t.Fatal("expected an active selection")
	}
	if got, want := ta.SelectedText(), "hello"; got != want {
		t.Errorf("SelectedText() = %q, want %q", got, want)
	}
}

func TestSelectedTextAcrossLines(t *testing.T) {
	t.Parallel()

	ta := newSelectionTextarea(t, "one\ntwo\nthree")
	// From the "e" of "one" through the "t" of "three".
	ta.BeginSelection(2, 0)
	ta.ExtendSelection(1, 2)
	ta.EndSelection()

	if got, want := ta.SelectedText(), "e\ntwo\nt"; got != want {
		t.Errorf("SelectedText() = %q, want %q", got, want)
	}
}

// TestSelectedTextBackwardDrag asserts a right-to-left / bottom-to-top drag
// yields the same text as the equivalent forward drag: the range is
// normalized rather than returned inverted (which would slice out of order).
func TestSelectedTextBackwardDrag(t *testing.T) {
	t.Parallel()

	forward := newSelectionTextarea(t, "one\ntwo\nthree")
	forward.BeginSelection(2, 0)
	forward.ExtendSelection(1, 2)
	forward.EndSelection()

	backward := newSelectionTextarea(t, "one\ntwo\nthree")
	backward.BeginSelection(1, 2)
	backward.ExtendSelection(2, 0)
	backward.EndSelection()

	if got, want := backward.SelectedText(), forward.SelectedText(); got != want {
		t.Errorf("backward drag = %q, want %q (same as forward)", got, want)
	}
	if got, want := backward.SelectedText(), "e\ntwo\nt"; got != want {
		t.Errorf("SelectedText() = %q, want %q", got, want)
	}
}

// TestClickWithoutDragMakesNoSelection asserts a plain click (press and
// release with no movement) does not leave a zero-width selection behind,
// which would otherwise make HasSelection true with empty text.
func TestClickWithoutDragMakesNoSelection(t *testing.T) {
	t.Parallel()

	ta := newSelectionTextarea(t, "hello world")
	ta.BeginSelection(3, 0)
	ta.EndSelection()

	if ta.HasSelection() {
		t.Error("a click without a drag must not leave a selection")
	}
	if got := ta.SelectedText(); got != "" {
		t.Errorf("SelectedText() = %q, want empty", got)
	}
}

// TestExtendSelectionRequiresBegin asserts stray motion events (a drag that
// started outside the textarea) do not spontaneously create a selection.
func TestExtendSelectionRequiresBegin(t *testing.T) {
	t.Parallel()

	ta := newSelectionTextarea(t, "hello world")
	ta.ExtendSelection(5, 0)

	if ta.HasSelection() {
		t.Error("ExtendSelection without BeginSelection must not select")
	}
}

func TestClearSelection(t *testing.T) {
	t.Parallel()

	ta := newSelectionTextarea(t, "hello world")
	ta.BeginSelection(0, 0)
	ta.ExtendSelection(5, 0)
	ta.EndSelection()
	if !ta.HasSelection() {
		t.Fatal("expected a selection to clear")
	}

	ta.ClearSelection()
	if ta.HasSelection() {
		t.Error("HasSelection() = true after ClearSelection()")
	}
	if got := ta.SelectedText(); got != "" {
		t.Errorf("SelectedText() = %q, want empty", got)
	}
}

func TestSelectAll(t *testing.T) {
	t.Parallel()

	ta := newSelectionTextarea(t, "one\ntwo")
	ta.SelectAll()

	if got, want := ta.SelectedText(), "one\ntwo"; got != want {
		t.Errorf("SelectedText() = %q, want %q", got, want)
	}
}

// TestKeyPressClearsSelection asserts a pointer selection does not survive
// keyboard input, which would leave a highlight pointing at stale offsets
// after the buffer or cursor moved.
func TestKeyPressClearsSelection(t *testing.T) {
	t.Parallel()

	ta := newSelectionTextarea(t, "hello world")
	ta.BeginSelection(0, 0)
	ta.ExtendSelection(5, 0)
	ta.EndSelection()
	if !ta.HasSelection() {
		t.Fatal("expected a selection before typing")
	}

	ta, _ = ta.Update(keyPress('x'))

	if ta.HasSelection() {
		t.Error("selection must be cleared by keyboard input")
	}
}

// TestSelectionRendersWithSelectionStyle asserts selected runes are painted
// with the Selection style and unselected ones are not, so the highlight is
// visually distinct from the terminal's native selection.
func TestSelectionRendersWithSelectionStyle(t *testing.T) {
	t.Parallel()

	ta := newSelectionTextarea(t, "hello world")

	plain := ta.View()

	ta.BeginSelection(0, 0)
	ta.ExtendSelection(5, 0)
	ta.EndSelection()
	selected := ta.View()

	if plain == selected {
		t.Fatal("view must change when a selection is active")
	}

	// The text itself is unchanged — only styling differs.
	if got, want := strings.TrimSpace(ansi.Strip(selected)), strings.TrimSpace(ansi.Strip(plain)); got != want {
		t.Errorf("stripped view changed: got %q, want %q", got, want)
	}
}

// TestSelectionSurvivesSoftWrap asserts selection offsets are mapped through
// the wrapped layout, not the logical line: clicking the second visual row of
// a wrapped line must select from that row, not from the middle of row one.
func TestSelectionSurvivesSoftWrap(t *testing.T) {
	t.Parallel()

	ta := New()
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.SetWidth(10)
	ta.SetHeight(6)
	ta.Focus()
	// Wider than the 10-column width, so it soft wraps.
	ta.SetValue("aaaaabbbbbccccc")

	// The second visual row starts at rune offset 10.
	pos := ta.PositionAt(0, 1)
	if pos.Row != 0 {
		t.Fatalf("soft wrap must stay on logical row 0, got row %d", pos.Row)
	}
	if pos.Col != 10 {
		t.Errorf("second visual row starts at col 10, got %d", pos.Col)
	}

	ta.BeginSelection(0, 1)
	ta.ExtendSelection(5, 1)
	ta.EndSelection()

	if got, want := ta.SelectedText(), "ccccc"; got != want {
		t.Errorf("SelectedText() = %q, want %q", got, want)
	}
}

// TestSelectionEmptyBuffer asserts the selection API is safe on an empty
// textarea rather than panicking on an out-of-range row.
func TestSelectionEmptyBuffer(t *testing.T) {
	t.Parallel()

	ta := newSelectionTextarea(t, "")
	ta.BeginSelection(0, 0)
	ta.ExtendSelection(5, 3)
	ta.EndSelection()
	ta.SelectAll()

	if ta.HasSelection() {
		t.Error("empty buffer must not report a selection")
	}
	if got := ta.SelectedText(); got != "" {
		t.Errorf("SelectedText() = %q, want empty", got)
	}
	_ = ta.View() // must not panic
}
