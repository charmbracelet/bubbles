package picker

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the structure of this component's key bindings.
type KeyMap struct {
	Next         key.Binding
	Prev         key.Binding
	JumpForward  key.Binding
	JumpBackward key.Binding
}

// DefaultKeyMap returns a default set of key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Next: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "next"),
		),
		Prev: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "previous"),
		),
		JumpForward: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "jump forward"),
		),
		JumpBackward: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "jump backward"),
		),
	}
}
