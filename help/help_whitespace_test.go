package help

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/golden"
)

func TestWhitespaceStyle(t *testing.T) {
	m := New()
	m.FullSeparator = " | "

	// Set a distinctive background color for whitespace to make it visible in tests
	whitespaceBg := lipgloss.Color("#FF0000")
	m.Styles.ShortWhitespace = m.Styles.ShortWhitespace.Background(whitespaceBg)
	m.Styles.FullWhitespace = m.Styles.FullWhitespace.Background(whitespaceBg)

	// Standard keys setup
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

	// Test both views at different widths
	for _, w := range []int{20, 30, 40} {
		t.Run(fmt.Sprintf("full_help_width_%d", w), func(t *testing.T) {
			m.Width = w
			s := m.FullHelpView(kb)
			golden.RequireEqual(t, []byte(s))
		})

		t.Run(fmt.Sprintf("short_help_width_%d", w), func(t *testing.T) {
			m.Width = w
			// Flatten the bindings for short help
			var shortBindings []key.Binding
			for _, group := range kb {
				shortBindings = append(shortBindings, group...)
			}
			s := m.ShortHelpView(shortBindings)
			golden.RequireEqual(t, []byte(s))
		})
	}

// Test with a disabled item and custom style
for _, tc := range []struct {
    name     string
    setupFn  func()
    bindings [][]key.Binding
}{
    {
        name: "disabled_item",
        setupFn: func() {
            m.Width = 40
        },
        bindings: [][]key.Binding{{
            key.NewBinding(k, key.WithHelp("enter", "continue")),
            key.NewBinding(k, key.WithHelp("ctrl+c", "quit"), key.WithDisabled()),
        }},
    },
    {
        name: "custom_style",
        setupFn: func() {
            m.Width = 40
            customBg := lipgloss.Color("#00FF00")
            m.Styles.FullWhitespace = m.Styles.FullWhitespace.Background(customBg)
            m.Styles.ShortWhitespace = m.Styles.ShortWhitespace.Background(customBg)
        },
        bindings: kb,
    },
} {
    t.Run(tc.name+"_full", func(t *testing.T) {
        tc.setupFn()
        s := m.FullHelpView(tc.bindings)
        golden.RequireEqual(t, []byte(s))
    })

    t.Run(tc.name+"_short", func(t *testing.T) {
        tc.setupFn()
        // Flatten the bindings for short help
        var shortBindings []key.Binding
        for _, group := range tc.bindings {
            shortBindings = append(shortBindings, group...)
        }
        s := m.ShortHelpView(shortBindings)
        golden.RequireEqual(t, []byte(s))
    })
}
}
