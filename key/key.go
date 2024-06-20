// Package key provides some types and functions for generating user-definable
// keymappings useful in Bubble Tea components. There are a few different ways
// you can define a keymapping with this package. Here's one example:
//
//	type KeyMap struct {
//	    Up key.Binding
//	    Down key.Binding
//	}
//
//	var DefaultKeyMap = KeyMap{
//	    Up: key.NewBinding(
//	        key.WithKeys("k", "up"),        // actual keybindings
//	        key.WithHelp("↑/k", "move up"), // corresponding help text
//	    ),
//	    Down: key.NewBinding(
//	        key.WithKeys("j", "down"),
//	        key.WithHelp("↓/j", "move down"),
//	    ),
//	}
//
//	func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	    switch msg := msg.(type) {
//	    case tea.KeyMsg:
//	        switch {
//	        case key.Matches(msg, DefaultKeyMap.Up):
//	            // The user pressed up
//	        case key.Matches(msg, DefaultKeyMap.Down):
//	            // The user pressed down
//	        }
//	    }
//
//	    // ...
//	}
//
// The help information, which is not used in the example above, can be used
// to render help text for keystrokes in your views.
package key

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Binding describes a set of keybindings and, optionally, their associated
// help text.
type Binding struct {
	keys     []tea.Key
	help     Help
	disabled bool
}

// BindingOpt is an initialization option for a keybinding. It's used as an
// argument to NewBinding.
type BindingOpt func(*Binding)

// NewBinding returns a new keybinding from a set of BindingOpt options.
func NewBinding(opts ...BindingOpt) Binding {
	b := &Binding{}
	for _, opt := range opts {
		opt(b)
	}
	return *b
}

// WithKeys initializes a keybinding with the given keystrokes.
func WithKeys(keys ...string) BindingOpt {
	return func(b *Binding) {
		b.SetKeys(keys...)
	}
}

// WithHelp initializes a keybinding with the given help text.
func WithHelp(key, desc string) BindingOpt {
	return func(b *Binding) {
		b.help = Help{Key: key, Desc: desc}
	}
}

// WithDisabled initializes a disabled keybinding.
func WithDisabled() BindingOpt {
	return func(b *Binding) {
		b.disabled = true
	}
}

// SetKeys sets the keys for the keybinding.
func (b *Binding) SetKeys(keys ...string) {
	b.keys = make([]tea.Key, 0, len(keys))
	for _, k := range keys {
		if tk, ok := MakeKey(k); ok {
			b.keys = append(b.keys, tk)
		}
	}
}

// Keys returns the keys for the keybinding.
func (b Binding) Keys() []string {
	kn := make([]string, len(b.keys))
	for i, tk := range b.keys {
		kn[i] = tk.String()
	}
	return kn
}

// SetHelp sets the help text for the keybinding.
func (b *Binding) SetHelp(key, desc string) {
	b.help = Help{Key: key, Desc: desc}
}

// Help returns the Help information for the keybinding.
func (b Binding) Help() Help {
	return b.help
}

// Enabled returns whether or not the keybinding is enabled. Disabled
// keybindings won't be activated and won't show up in help. Keybindings are
// enabled by default.
func (b Binding) Enabled() bool {
	return !b.disabled && b.keys != nil
}

// SetEnabled enables or disables the keybinding.
func (b *Binding) SetEnabled(v bool) {
	b.disabled = !v
}

// Unbind removes the keys and help from this binding, effectively nullifying
// it. This is a step beyond disabling it, since applications can enable
// or disable key bindings based on application state.
func (b *Binding) Unbind() {
	b.keys = nil
	b.help = Help{}
}

// Help is help information for a given keybinding.
type Help struct {
	Key  string
	Desc string
}

// Matches checks if the given KeyMsg matches the given bindings.
func Matches(k tea.KeyMsg, b ...Binding) bool {
	for _, binding := range b {
		for _, v := range binding.keys {
			if keyEq(v, tea.Key(k)) && binding.Enabled() {
				return true
			}
		}
	}
	return false
}

func keyEq(a, b tea.Key) bool {
	if a.Type != b.Type {
		return false
	}
	if a.Alt != b.Alt {
		return false
	}
	if len(a.Runes) != len(b.Runes) {
		return false
	}
	for i, ar := range a.Runes {
		if b.Runes[i] != ar {
			return false
		}
	}
	return true
}

// MakeKey returns a tea.Key for the given keyName.
func MakeKey(keyName string) (tea.Key, bool) {
	alt := false
	if strings.HasPrefix(keyName, "alt+") {
		alt = true
		keyName = keyName[4:]
	}
	// Is this a special key?
	k, ok := allKeys[keyName]
	if ok {
		k.Alt = alt
		return k, true
	}
	// Not a special key: either a simple key "a" or with an alt
	// modifier "alt+a".
	r := []rune(keyName)
	if len(r) != 1 {
		// Caller used a key name which we don't understand, bail.
		return tea.Key{}, false
	}
	return tea.Key{
		Type:  tea.KeyRunes,
		Runes: r,
		Alt:   alt,
	}, true
}

// allKeys contains the map of all "special" keys and their
// ctrl/shift/alt combinations.
var allKeys = func() map[string]tea.Key {
	result := make(map[string]tea.Key)
	for i := 0; ; i++ {
		k := tea.Key{Type: tea.KeyType(i)}
		keyName := k.String()
		// fmt.Println("found key:", keyName)
		if keyName == "" {
			break
		}
		result[keyName] = k
	}
	for i := -2; ; i-- {
		k := tea.Key{Type: tea.KeyType(i)}
		keyName := k.String()
		// fmt.Println("found key:", keyName)
		if keyName == "" {
			break
		}
		result[keyName] = k
	}
	return result
}()
