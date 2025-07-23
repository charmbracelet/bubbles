package key

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea/v2"
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

func TestMatches(t *testing.T) {
	cases := []struct {
		key      tea.KeyMsg
		bindings []Binding
	}{
		{
			key: tea.KeyPressMsg{Code: 'k', Text: "k"},
			bindings: []Binding{
				NewBinding(
					WithKeys("k", "up"),
					WithHelp("↑/k", "move up"),
				),
			},
		},
		{
			key: tea.KeyPressMsg{Code: '/', Mod: tea.ModShift, Text: "?"},
			bindings: []Binding{
				NewBinding(
					WithKeys("?"),
					WithHelp("?", "search"),
				),
			},
		},
		{
			key: tea.KeyPressMsg{Code: 'a', Mod: tea.ModCtrl},
			bindings: []Binding{
				NewBinding(
					WithKeys("ctrl+a"),
					WithHelp("ctrl+a", "select all"),
				),
			},
		},
	}

	for i, c := range cases {
		if !Matches(c.key, c.bindings...) {
			t.Errorf("case %d: expected key (%q) to match binding(s)", i+1, c.key.String())
		}
	}
}
