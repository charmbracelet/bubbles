package key

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestBinding_Enabled(t *testing.T) {
	binding := NewBinding(
		WithKeys("k", "up"),
		WithHelp("↑/k", "move up"),
	)
	if !binding.Enabled() {
		t.Errorf("expected key to be Enabled")
	}

	binding.SetEnabled(false)
	if binding.Enabled() {
		t.Errorf("expected key not to be Enabled")
	}

	binding.SetEnabled(true)
	binding.Unbind()
	if binding.Enabled() {
		t.Errorf("expected key not to be Enabled")
	}
}

func BenchmarkMatches(b *testing.B) {
	msg1 := tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune("c"), Alt: true})
	msg2 := tea.KeyMsg(tea.Key{Type: tea.KeyEnter})
	kb := NewBinding(WithKeys("alt+c"))
	b.Run("success", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Matches(msg1, kb)
		}
	})
	b.Run("fail", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Matches(msg2, kb)
		}
	})
}
