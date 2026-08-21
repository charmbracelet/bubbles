package help

import (
	"fmt"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"

	"charm.land/bubbles/v2/key"
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

	sep := m.FullSeparator
	sepWidth := lipgloss.Width(sep)

	widths := make([]int, 8)
	widths[0] = 1                                         // no usable width
	widths[1] = len("enter continue")                     // sufficient width for column, but not with traling ellipsis+space
	widths[2] = widths[1] + 1                             // sufficient width for column, but not with traling ellipsis+space
	widths[3] = widths[2] + 1                             // sufficient width for column and ellipsis+space
	widths[4] = widths[1] + sepWidth + len("esc back")    // sufficient width for 2 columns, but not with for elipsis+space
	widths[5] = widths[4] + 1                             // sufficient width for 2 columns, but not with for elipsis+space
	widths[6] = widths[5] + 1                             // sufficient width for 2 columns and elipsis+space
	widths[7] = widths[4] + sepWidth + len("ctrl+c quit") // sufficient width for all items

	for i, w := range widths {
		t.Run(fmt.Sprintf("full help width #%d", i), func(t *testing.T) {
			m.SetWidth(w)
			s := m.FullHelpView(kb)
			s = ansi.Strip(s)
			golden.RequireEqual(t, []byte(s))
		})
	}
}

func TestShortHelp(t *testing.T) {
	m := New()
	k := key.WithKeys("x")
	kb := []key.Binding{
		key.NewBinding(k, key.WithHelp("enter", "continue")),
		key.NewBinding(k, key.WithHelp("esc", "back")),
		key.NewBinding(k, key.WithHelp("?", "help")),
	}

	m.Ellipsis = "…"
	sep := m.ShortSeparator
	sepWidth := lipgloss.Width(sep)

	widths := make([]int, 8)
	widths[0] = 1                                      // no usable width
	widths[1] = len("enter continue")                  // sufficient width for item, but not with traling ellipsis+space
	widths[2] = widths[1] + 1                          // sufficient width for item, but not with traling ellipsis+space
	widths[3] = widths[2] + 1                          // sufficient width for item and ellipsis+space
	widths[4] = widths[1] + sepWidth + len("esc back") // sufficient width for 2 items, but not with for elipsis+space
	widths[5] = widths[4] + 1                          // sufficient width for 2 items, but not with for elipsis+space
	widths[6] = widths[5] + 1                          // sufficient width for 2 items and elipsis+space
	widths[7] = widths[4] + sepWidth + len("? help")   // sufficient width for all items

	for i, w := range widths {
		t.Run(fmt.Sprintf("short help width #%d", i), func(t *testing.T) {
			m.SetWidth(w)
			s := m.ShortHelpView(kb)
			s = ansi.Strip(s)
			golden.RequireEqual(t, []byte(s))
		})
	}
}
