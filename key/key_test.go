package key

import (
	"testing"
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

func TestBinding_Equal(t *testing.T) {
	binding1 := NewBinding(
		WithKeys("k", "up"),
		WithHelp("↑/k", "move up"),
	)
	binding2 := NewBinding(
		WithKeys("k", "up"),
		WithHelp("↑/k", "move up"),
	)
	if !binding1.Equal(binding2) {
		t.Errorf("expected bindings to be Equal")
	}
	binding2.SetEnabled(false)
	if binding1.Equal(binding2) {
		t.Errorf("expected bindings not to be Equal")
	}
	binding3 := NewBinding(
		WithKeys("j", "down"),
		WithHelp("↓/j", "move down"),
	)
	if binding1.Equal(binding3) {
		t.Errorf("expected bindings not to be Equal")
	}
}
