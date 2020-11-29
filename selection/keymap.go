package selection

// DefaultKeyMap is a KeyMap with sensible default key mappings
// that can also be used as a starting point for customization.
var DefaultKeyMap = KeyMap{
	Down:        []string{"down"},
	Up:          []string{"up"},
	Select:      []string{"enter"},
	Abort:       []string{"ctrl+c"},
	ClearFilter: []string{"esc"},
	ScrollDown:  []string{"pgdown", "right"},
	ScrollUp:    []string{"pgup", "left"},
}

// KeyMap defines the keys that trigger certain actions.
type KeyMap struct {
	Down        []string
	Up          []string
	Select      []string
	Abort       []string
	ClearFilter []string
	ScrollDown  []string
	ScrollUp    []string
}

func keyMatches(key string, mapping []string) bool {
	for _, m := range mapping {
		if m == key {
			return true
		}
	}

	return false
}

// validateKeyMap returns true if the given key map contains at
// least the bare minimum set of key bindings for the functional
// prompt and false otherwise.
func validateKeyMap(km KeyMap) bool {
	if len(km.Up) == 0 {
		return false
	}

	if len(km.Down) == 0 {
		return false
	}

	if len(km.Select) == 0 {
		return false
	}

	return true
}
