package textarea

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestDeleteForwardPreservesFollowingNewline(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		col   int
		want  string
	}{
		{"last character", "ab\ncd", 1, "a\ncd"},
		{"only character", "a\nb", 0, "\nb"},
		{"wide character", "a界\n다", 1, "a\n다"},
		{"middle character", "abc\ncd", 1, "ac\ncd"},
		{"line end", "ab\ncd", 2, "abcd"},
		{"empty line", "\ncd", 0, "cd"},
		{"last line character", "ab", 1, "a"},
		{"last line end", "ab", 2, "ab"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newTextArea()
			m.SetValue(tc.value)
			for m.Line() > 0 {
				m.CursorUp()
			}
			m.SetCursorColumn(tc.col)
			m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
			if got := m.Value(); got != tc.want {
				t.Errorf("Delete at column %d in %q = %q; want %q", tc.col, tc.value, got, tc.want)
			}
		})
	}
}
