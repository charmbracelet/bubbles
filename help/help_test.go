package help

import (
	"fmt"
	"testing"

	"charm.land/bubbles/v2/key"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
)

func TestFullHelp(t *testing.T) {
	m := New()
	m.FullSeparator = " | "
	k := key.WithKeys("x")
	kb := [][]key.Binding{
		{
			key.NewBinding(k, key.WithHelp("enter", "continue")),
		},
		{
			key.NewBinding(k, key.WithHelp("esc", "back")),
			key.NewBinding(k, key.WithHelp("?", "help")),
		},
		{
			key.NewBinding(k, key.WithHelp("H", "home")),
			key.NewBinding(k, key.WithHelp("ctrl+c", "quit")),
			key.NewBinding(k, key.WithHelp("ctrl+l", "log")),
		},
	}

	for _, w := range []int{20, 30, 40} {
		t.Run(fmt.Sprintf("full help %d width", w), func(t *testing.T) {
			m.SetWidth(w)
			s := m.FullHelpView(kb)
			s = ansi.Strip(s)
			golden.RequireEqual(t, []byte(s))
		})
	}
}

func TestHelpEdgeCases(t *testing.T) {
	m := New()

	if got := m.ShortHelpView(nil); got != "" {
		t.Errorf("expected empty short help for nil bindings, got %q", got)
	}
	if got := m.FullHelpView(nil); got != "" {
		t.Errorf("expected empty full help for nil groups, got %q", got)
	}
	if got := m.FullHelpView([][]key.Binding{nil}); got != "" {
		t.Errorf("expected empty full help for nil group, got %q", got)
	}

	disabled := key.NewBinding(key.WithKeys("x"), key.WithDisabled())
	if got := m.FullHelpView([][]key.Binding{{disabled}}); got != "" {
		t.Errorf("expected empty full help for all-disabled group, got %q", got)
	}

	m.SetWidth(0)
	if m.Width() != 0 {
		t.Errorf("expected width 0, got %d", m.Width())
	}
	// Should not panic when width is zero.
	_ = m.ShortHelpView([]key.Binding{
		key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "do")),
	})
}
