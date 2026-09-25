package key

import (
	"fmt"
	"testing"
)

// testKey implements fmt.Stringer so we can exercise Matches with a generic
// key type that is not a plain string.
type testKey string

func (k testKey) String() string { return string(k) }

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

func TestBinding_Matches(t *testing.T) {
	var _ fmt.Stringer = testKey("")

	b := NewBinding(
		WithKeys("a", "b"),
		WithHelp("a/b", "do thing"),
	)
	disabled := NewBinding(
		WithKeys("a"),
		WithDisabled(),
	)

	if !Matches(testKey("a"), b) {
		t.Error("expected match for key a")
	}
	if !Matches(testKey("b"), b) {
		t.Error("expected match for key b")
	}
	if Matches(testKey("c"), b) {
		t.Error("expected no match for key c")
	}
	if Matches(testKey("a"), disabled) {
		t.Error("expected disabled binding not to match")
	}
	if Matches(testKey("a")) {
		t.Error("expected no match with no bindings")
	}
}
