package dropdown

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// helpers ──────────────────────────────────────────────────────────────────

func testOptions() []Option {
	return []Option{
		{Label: "Alpha", Value: "a"},
		{Label: "Beta", Value: "b"},
		{Label: "Gamma", Value: "c"},
		{Label: "Delta", Value: "d"},
		{Label: "Epsilon", Value: "e"},
	}
}

// press sends a single key event and returns the updated model plus any
// emitted tea.Msg (by calling the returned Cmd immediately).
func press(m Model, code rune) (Model, tea.Msg) {
	updated, cmd := m.Update(tea.KeyPressMsg{Code: code})
	var emitted tea.Msg
	if cmd != nil {
		emitted = cmd()
	}
	return updated, emitted
}

func focused(opts []Option) Model {
	m := New(WithOptions(opts...), WithWidth(20))
	m.Focus() //nolint:errcheck
	return m
}

// ── Focus / Blur ───────────────────────────────────────────────────────────

func TestFocused(t *testing.T) {
	m := New(WithOptions(testOptions()...))
	if m.Focused() {
		t.Fatal("expected unfocused on creation")
	}
	m.Focus() //nolint:errcheck
	if !m.Focused() {
		t.Fatal("expected focused after Focus()")
	}
	m.Blur()
	if m.Focused() {
		t.Fatal("expected unfocused after Blur()")
	}
}

func TestBlurClosesDropdown(t *testing.T) {
	m := focused(testOptions())
	m, _ = press(m, tea.KeyEnter)
	if !m.open {
		t.Fatal("expected open after Enter")
	}
	m.Blur()
	if m.open {
		t.Fatal("expected closed after Blur()")
	}
}

// ── Disabled ───────────────────────────────────────────────────────────────

func TestDisabledIgnoresInput(t *testing.T) {
	m := focused(testOptions())
	m.Disabled = true
	m, emitted := press(m, tea.KeyEnter)
	if m.open {
		t.Error("disabled dropdown should not open")
	}
	if emitted != nil {
		t.Error("disabled dropdown should not emit messages")
	}
}

// ── Empty options ─────────────────────────────────────────────────────────

func TestEmptyOptionsNoOpen(t *testing.T) {
	m := New(WithWidth(20))
	m.Focus() //nolint:errcheck
	m, _ = press(m, tea.KeyEnter)
	if m.open {
		t.Fatal("empty dropdown should not open")
	}
}

func TestEmptyOptionsNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	m := New(WithWidth(20))
	m.Focus() //nolint:errcheck
	for _, k := range []rune{tea.KeyEnter, tea.KeyDown, tea.KeyUp, tea.KeyEscape} {
		m, _ = press(m, k)
	}
	_ = m.View()
}

// ── Opening / Closing ─────────────────────────────────────────────────────

func TestOpenWithEnter(t *testing.T) {
	m := focused(testOptions())
	m, _ = press(m, tea.KeyEnter)
	if !m.IsOpen() {
		t.Fatal("should be open after Enter")
	}
}

func TestOpenWithSpace(t *testing.T) {
	m := focused(testOptions())
	m, _ = press(m, tea.KeySpace)
	if !m.IsOpen() {
		t.Fatal("should be open after Space")
	}
}

func TestCloseWithEscape(t *testing.T) {
	m := focused(testOptions())
	m, _ = press(m, tea.KeyEnter)
	m, emitted := press(m, tea.KeyEscape)
	if m.IsOpen() {
		t.Fatal("should be closed after Escape")
	}
	if _, ok := emitted.(CloseMsg); !ok {
		t.Fatalf("expected CloseMsg, got %T", emitted)
	}
}

func TestEscapeDoesNotChangeSelection(t *testing.T) {
	m := focused(testOptions())
	m, _ = press(m, tea.KeyEnter)
	m, _ = press(m, tea.KeyDown)
	m, _ = press(m, tea.KeyEscape)
	if m.SelectedIndex() != -1 {
		t.Fatalf("Escape should not commit selection; got index %d", m.SelectedIndex())
	}
}

// ── Navigation ────────────────────────────────────────────────────────────

func TestCursorDownUp(t *testing.T) {
	m := focused(testOptions())
	m, _ = press(m, tea.KeyEnter)
	m, _ = press(m, tea.KeyDown)
	m, _ = press(m, tea.KeyDown)
	if m.cursor != 2 {
		t.Fatalf("expected cursor=2, got %d", m.cursor)
	}
	m, _ = press(m, tea.KeyUp)
	if m.cursor != 1 {
		t.Fatalf("expected cursor=1 after Up, got %d", m.cursor)
	}
}

func TestCursorNoWrapAtTop(t *testing.T) {
	m := focused(testOptions())
	m, _ = press(m, tea.KeyEnter)
	m, _ = press(m, tea.KeyUp)
	if m.cursor != 0 {
		t.Fatalf("cursor should stay at 0, got %d", m.cursor)
	}
}

func TestCursorNoWrapAtBottom(t *testing.T) {
	opts := testOptions()
	m := focused(opts)
	m, _ = press(m, tea.KeyEnter)
	for i := 0; i < len(opts)-1; i++ {
		m, _ = press(m, tea.KeyDown)
	}
	last := m.cursor
	m, _ = press(m, tea.KeyDown)
	if m.cursor != last {
		t.Fatalf("cursor should not go past last option; got %d", m.cursor)
	}
}

func TestVimKeys(t *testing.T) {
	m := focused(testOptions())
	m, _ = press(m, tea.KeyEnter)
	m, _ = m.Update(tea.KeyPressMsg{Text: "j"})
	if m.cursor != 1 {
		t.Fatalf("expected cursor=1 after 'j', got %d", m.cursor)
	}
	m, _ = m.Update(tea.KeyPressMsg{Text: "k"})
	if m.cursor != 0 {
		t.Fatalf("expected cursor=0 after 'k', got %d", m.cursor)
	}
}

// ── Selection (SelectMsg) ─────────────────────────────────────────────────

func TestSelectEmitsSelectMsg(t *testing.T) {
	m := focused(testOptions())
	m, _ = press(m, tea.KeyEnter) // open
	m, _ = press(m, tea.KeyDown)  // cursor → 1 "Beta"
	m, emitted := press(m, tea.KeyEnter)

	sel, ok := emitted.(SelectMsg)
	if !ok {
		t.Fatalf("expected SelectMsg, got %T", emitted)
	}
	if sel.Index != 1 {
		t.Errorf("expected index 1, got %d", sel.Index)
	}
	if sel.Option.Value != "b" {
		t.Errorf("expected value 'b', got %q", sel.Option.Value)
	}
}

func TestSelectCommitsSelection(t *testing.T) {
	m := focused(testOptions())
	m, _ = press(m, tea.KeyEnter)
	m, _ = press(m, tea.KeyDown)
	m, _ = press(m, tea.KeyDown)
	m, _ = press(m, tea.KeyEnter)

	if m.SelectedIndex() != 2 {
		t.Fatalf("expected selected=2, got %d", m.SelectedIndex())
	}
	opt, ok := m.SelectedItem()
	if !ok {
		t.Fatal("SelectedItem should return true after selection")
	}
	if opt.Label != "Gamma" {
		t.Errorf("expected 'Gamma', got %q", opt.Label)
	}
}

func TestSelectClosesDropdown(t *testing.T) {
	m := focused(testOptions())
	m, _ = press(m, tea.KeyEnter)
	m, _ = press(m, tea.KeyEnter)
	if m.IsOpen() {
		t.Fatal("dropdown should close after selecting")
	}
}

func TestSelectedItemNoneInitially(t *testing.T) {
	m := New(WithOptions(testOptions()...))
	_, ok := m.SelectedItem()
	if ok {
		t.Fatal("expected no selection initially")
	}
	if m.SelectedIndex() != -1 {
		t.Fatalf("expected SelectedIndex=-1, got %d", m.SelectedIndex())
	}
}

// ── Scroll offset ─────────────────────────────────────────────────────────

func TestScrollOffsetAdvancesWithCursor(t *testing.T) {
	m := New(WithOptions(testOptions()...), WithWidth(20), WithMaxVisible(3))
	m.Focus() //nolint:errcheck
	m, _ = press(m, tea.KeyEnter)
	for i := 0; i < 4; i++ {
		m, _ = press(m, tea.KeyDown)
	}
	if m.scrollOffset < 2 {
		t.Fatalf("expected scrollOffset>=2 at cursor=4 with MaxVisible=3, got %d", m.scrollOffset)
	}
	end := m.scrollOffset + m.MaxVisible
	if m.cursor < m.scrollOffset || m.cursor >= end {
		t.Fatalf("cursor %d outside visible window [%d, %d)", m.cursor, m.scrollOffset, end)
	}
}

func TestScrollOffsetRetreatsWithCursor(t *testing.T) {
	opts := testOptions()
	m := New(WithOptions(opts...), WithWidth(20), WithMaxVisible(3))
	m.Focus() //nolint:errcheck
	m, _ = press(m, tea.KeyEnter)
	for i := 0; i < len(opts)-1; i++ {
		m, _ = press(m, tea.KeyDown)
	}
	prevOffset := m.scrollOffset
	for i := 0; i < len(opts)-1; i++ {
		m, _ = press(m, tea.KeyUp)
	}
	if m.scrollOffset >= prevOffset {
		t.Fatalf("scrollOffset should have decreased (was %d, now %d)", prevOffset, m.scrollOffset)
	}
	if m.scrollOffset != 0 {
		t.Fatalf("expected scrollOffset=0 at top, got %d", m.scrollOffset)
	}
}

// ── SetOptions ────────────────────────────────────────────────────────────

func TestSetOptionsResetsState(t *testing.T) {
	m := focused(testOptions())
	m, _ = press(m, tea.KeyEnter)
	m, _ = press(m, tea.KeyDown)
	m, _ = press(m, tea.KeyEnter)
	if m.SelectedIndex() == -1 {
		t.Fatal("expected a selection before SetOptions")
	}
	m.SetOptions([]Option{{Label: "X", Value: "x"}})
	if m.SelectedIndex() != -1 {
		t.Fatal("SetOptions should reset selection to -1")
	}
	if m.cursor != 0 {
		t.Fatal("SetOptions should reset cursor to 0")
	}
	if m.scrollOffset != 0 {
		t.Fatal("SetOptions should reset scrollOffset to 0")
	}
}

// ── Option funcs ──────────────────────────────────────────────────────────

func TestWithKeyMap(t *testing.T) {
	km := DefaultKeyMap()
	m := New(WithKeyMap(km))
	if m.KeyMap.Up.Keys()[0] != km.Up.Keys()[0] {
		t.Fatal("WithKeyMap should set the KeyMap")
	}
}

func TestWithStyles(t *testing.T) {
	s := DefaultStyles()
	m := New(WithStyles(s))
	_ = m.View() // should not panic
}

// ── View ──────────────────────────────────────────────────────────────────

func TestViewDoesNotPanic(t *testing.T) {
	cases := []struct {
		name  string
		model Model
	}{
		{"default", New()},
		{"focused", func() Model { m := New(WithOptions(testOptions()...)); m.Focus(); return m }()},
		{"open", func() Model {
			m := New(WithOptions(testOptions()...))
			m.Focus()
			m, _ = press(m, tea.KeyEnter)
			return m
		}()},
		{"disabled", func() Model {
			m := New(WithOptions(testOptions()...))
			m.Disabled = true
			return m
		}()},
		{"empty", New()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("View() panicked: %v", r)
				}
			}()
			_ = tc.model.View()
		})
	}
}

func TestViewContainsSelectedLabel(t *testing.T) {
	m := focused(testOptions())
	m, _ = press(m, tea.KeyEnter)
	m, _ = press(m, tea.KeyDown) // cursor → 1 "Beta"
	m, _ = press(m, tea.KeyEnter)
	view := m.View()
	if !containsSubstring(view, "Beta") {
		t.Errorf("expected 'Beta' in view, got:\n%s", view)
	}
}

func TestViewContainsPlaceholder(t *testing.T) {
	m := New(WithOptions(testOptions()...), WithPlaceholder("Pick one"))
	view := m.View()
	if !containsSubstring(view, "Pick one") {
		t.Errorf("expected placeholder in view, got:\n%s", view)
	}
}

// containsSubstring is an inlined helper to avoid importing strings.
func containsSubstring(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
