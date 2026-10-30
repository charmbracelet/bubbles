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

func TestShortHelpFitsWidth(t *testing.T) {
	k := key.WithKeys("x")
	kb := []key.Binding{
		key.NewBinding(k, key.WithHelp("enter", "continue")),
		key.NewBinding(k, key.WithHelp("esc", "back")),
		key.NewBinding(k, key.WithHelp("ctrl+c", "quit")),
	}

	m := New()
	full := ansi.StringWidth(m.ShortHelpView(kb))

	for w := 1; w <= full; w++ {
		t.Run(fmt.Sprintf("width %d", w), func(t *testing.T) {
			m.SetWidth(w)
			if got := ansi.StringWidth(m.ShortHelpView(kb)); got > w {
				t.Errorf("ShortHelpView is %d cells wide, want at most %d", got, w)
			}
		})
	}
}
