package textarea

import (
	tea "charm.land/bubbletea/v2"
	"testing"
)

func TestSetCursorColumnSnapsToGraphemeStart(t *testing.T) {
	m := newTextArea()
	m.SetValue("A👩‍💻B")
	m.SetCursorColumn(2)
	if m.col != 1 {
		t.Fatalf("cursor inside emoji: got %d want 1", m.col)
	}
}
func TestVerticalCursorSnapsToGraphemeStart(t *testing.T) {
	m := newTextArea()
	m.SetValue("ab\n👩‍💻x")
	m.SetCursorColumn(0)
	m.CursorUp()
	m.SetCursorColumn(1)
	m.CursorDown()
	if start, _ := graphemeAt(m.value[m.row], m.col); start != m.col {
		t.Fatalf("vertical cursor split cluster at %d", m.col)
	}
}
func TestBackspaceRemovesMarkWithoutConsonant(t *testing.T) {
	for _, value := range []string{"น้", "कि"} {
		m := newTextArea()
		m.SetValue(value)
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
		want := string([]rune(value)[:len([]rune(value))-1])
		if m.Value() != want {
			t.Fatalf("backspace %q: got %q want %q", value, m.Value(), want)
		}
	}
}

func TestWordOperationsAdvanceOverGraphemes(t *testing.T) {
	for _, value := range []string{"नमस्ते rest", "น้ำ rest", "👩‍💻 rest", "e\u0301 rest"} {
		t.Run(value, func(t *testing.T) {
			m := newTextArea()
			m.SetValue(value)
			m.CursorStart()
			m.wordRight()
			if m.col == 0 {
				t.Fatal("wordRight made no progress")
			}
			if start, _ := graphemeAt(m.value[m.row], m.col); start != m.col {
				t.Fatalf("wordRight split cluster at %d", m.col)
			}
			m.CursorStart()
			m.deleteWordRight()
			if m.Value() != " rest" {
				t.Fatalf("deleteWordRight: got %q", m.Value())
			}
		})
	}
}
