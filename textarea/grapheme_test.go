package textarea

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestThaiGraphemeNavigationAndDeletion(t *testing.T) {
	for _, value := range []string{"นำ", "น้ำ"} {
		t.Run(value, func(t *testing.T) {
			m := newTextArea()
			m.SetValue(value)
			if got := m.LineInfo().CharOffset; got != 2 {
				t.Fatalf("Sara Am cluster has terminal offset %d; want 2", got)
			}
			m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
			if m.col != 0 {
				t.Fatalf("left from end stopped inside %q at rune column %d", value, m.col)
			}

			m.SetValue(value)
			m.CursorStart()
			m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
			if want := len([]rune(value)); m.col != want {
				t.Fatalf("right from start stopped inside %q at rune column %d; want %d", value, m.col, want)
			}

			m.SetValue(value)
			m.CursorStart()
			m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
			if got := m.Value(); got != "" {
				t.Fatalf("forward delete left an orphaned part of %q: %q", value, got)
			}

			m.SetValue(value)
			m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
			if got := m.Value(); got != "" {
				t.Fatalf("backspace left an orphaned part of %q: %q", value, got)
			}
		})
	}
}

func TestGraphemeNavigationKeepsASCIIStepsAndSkipsEmojiCluster(t *testing.T) {
	m := newTextArea()
	m.SetValue("ab")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	if m.col != 1 {
		t.Fatalf("ASCII navigation moved to rune column %d; want 1", m.col)
	}

	value := "A👩‍💻B"
	m.SetValue(value)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	if want := len([]rune(value)) - 1; m.col != want {
		t.Fatalf("left from after the emoji stopped at rune column %d; want %d", m.col, want)
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	if m.col != 1 {
		t.Fatalf("left stopped inside the emoji cluster at rune column %d; want 1", m.col)
	}
}
