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
