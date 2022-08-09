package textarea

import (
	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
)

// Mode is the possible modes of the textarea for modal editing.
type Mode string

const (
	// ModeNormal is the normal mode for navigating around text.
	ModeNormal Mode = "normal"
	// ModeInsert is the insert mode for inserting text.
	ModeInsert Mode = "insert"
)

// SetMode sets the mode of the textarea.
func (m *Model) SetMode(mode Mode) tea.Cmd {
	switch mode {
	case ModeInsert:
		m.mode = ModeInsert
		return m.Cursor.SetCursorMode(cursor.CursorBlink)
	case ModeNormal:
		m.mode = ModeNormal
		return m.Cursor.SetCursorMode(cursor.CursorStatic)
	}
	return nil
}

// Action is the type of action that will be performed when the user completes
// a key combination.
type Action int

const (
	// ActionMove moves the cursor.
	ActionMove Action = iota
	// ActionSeek seeks the cursor to the desired character.
	// Used in conjunction with f/F/t/T.
	ActionSeek
	// ActionReplace replaces text.
	ActionReplace
	// ActionDelete deletes text.
	ActionDelete
	// ActionYank yanks text.
	ActionYank
)

// Position is a (row, column) pair representing a position of the cursor or
// any character.
type Position struct {
	Row int
	Col int
}

// Range is a range of characters in the text area.
type Range struct {
	Start Position
	End   Position
}

// NormalCommand is a helper for keeping track of the various relevant information
// when performing vim motions in the textarea.
type NormalCommand struct {
	// Buffer is the buffer of keys that have been press for the current
	// command.
	Buffer string
	// Count is the number of times to replay the action. This is usually
	// optional and defaults to 1.
	Count int
	// Action is the action to be performed.
	Action Action
	// Range is the range of characters to perform the action on.
	Range Range
}
