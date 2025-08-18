package picker

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the structure of this component's key bindings.
type KeyMap struct {
	Next         key.Binding
	Prev         key.Binding
	StepForward  key.Binding
	StepBackward key.Binding
	JumpForward  key.Binding
	JumpBackward key.Binding
}

// DefaultKeyMap returns a default set of key bindings.
func DefaultKeyMap() *KeyMap {
	return &KeyMap{
		Next: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "next"),
		),
		Prev: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "previous"),
		),
		StepForward: key.NewBinding(
			key.WithKeys("shift+right", "shift+l"),
			key.WithHelp("shift + →/l", "step forward"),
		),
		StepBackward: key.NewBinding(
			key.WithKeys("shift+left", "shift+h"),
			key.WithHelp("shift + ←/h", "step backward"),
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
