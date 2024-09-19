package picker

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the structure of this component's key bindings.
type KeyMap struct {
	Next key.Binding
	Prev key.Binding
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
	}
}
