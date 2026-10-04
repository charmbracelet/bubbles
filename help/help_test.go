package help

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
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

func TestFullHelpStylesColumnPadding(t *testing.T) {
	m := New()
	m.FullSeparator = " | "
	m.SetWidth(80)
	m.Styles.FullSeparator = lipgloss.NewStyle().Background(lipgloss.Color("2")).Foreground(lipgloss.Color("0"))
	m.Styles.FullKey = lipgloss.NewStyle().Background(lipgloss.Color("1")).Foreground(lipgloss.Color("0"))
	m.Styles.FullDesc = lipgloss.NewStyle().Background(lipgloss.Color("4")).Foreground(lipgloss.Color("0"))

	k := key.WithKeys("x")
	kb := [][]key.Binding{
		{
			key.NewBinding(k, key.WithHelp("a", "alpha")),
			key.NewBinding(k, key.WithHelp("b", "beta")),
		},
		{
			key.NewBinding(k, key.WithHelp("c", "gamma")),
			key.NewBinding(k, key.WithHelp("d", "delta")),
		},
	}

	rendered := m.FullHelpView(kb)
	if !strings.Contains(rendered, "\x1b[30;42m   ") {
		t.Fatalf("expected separator style to cover column padding, got %q", rendered)
	}
}

func TestFullHelpStylesColumnPaddingSkipsDisabledBindings(t *testing.T) {
	m := New()
	m.FullSeparator = " | "
	m.SetWidth(80)
	m.Styles.FullSeparator = lipgloss.NewStyle().Background(lipgloss.Color("2")).Foreground(lipgloss.Color("0"))

	k := key.WithKeys("x")
	disabled := key.NewBinding(k, key.WithHelp("z", "zeta"))
	disabled.SetEnabled(false)
	kb := [][]key.Binding{
		{
			key.NewBinding(k, key.WithHelp("a", "alpha")),
			key.NewBinding(k, key.WithHelp("b", "beta")),
		},
		{
			key.NewBinding(k, key.WithHelp("c", "gamma")),
			disabled,
			key.NewBinding(k, key.WithHelp("d", "delta")),
		},
	}

	rendered := ansi.Strip(m.FullHelpView(kb))
	if strings.Count(rendered, "\n") != 1 {
		t.Fatalf("expected exactly two rendered rows when one binding is disabled, got %q", rendered)
	}
}
