package textarea

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
	rw "github.com/mattn/go-runewidth"
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
		m.col = clamp(m.col-1, 0, len(m.value[m.row]))
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
	// ActionChange deletes text and enters insert mode.
	ActionChange
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

func (m *Model) insertUpdate(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m.SetMode(ModeNormal)
		case "ctrl+k":
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			if m.col >= len(m.value[m.row]) {
				m.mergeLineBelow(m.row)
				break
			}
			m.deleteAfterCursor()
		case "ctrl+u":
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			if m.col <= 0 {
				m.mergeLineAbove(m.row)
				break
			}
			m.deleteBeforeCursor()
		case "backspace", "ctrl+h":
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			if m.col <= 0 {
				m.mergeLineAbove(m.row)
				break
			}
			if len(m.value[m.row]) > 0 {
				m.value[m.row] = append(m.value[m.row][:max(0, m.col-1)], m.value[m.row][m.col:]...)
				if m.col > 0 {
					m.SetCursor(m.col - 1)
				}
			}
		case "delete", "ctrl+d":
			if len(m.value[m.row]) > 0 && m.col < len(m.value[m.row]) {
				m.value[m.row] = append(m.value[m.row][:m.col], m.value[m.row][m.col+1:]...)
			}
			if m.col >= len(m.value[m.row]) {
				m.mergeLineBelow(m.row)
				break
			}
		case "alt+backspace", "ctrl+w":
			if m.col <= 0 {
				m.mergeLineAbove(m.row)
				break
			}
			m.deleteWordLeft()
		case "alt+delete", "alt+d":
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			if m.col >= len(m.value[m.row]) {
				m.mergeLineBelow(m.row)
				break
			}
			m.deleteWordRight()
		case "enter", "ctrl+m":
			if len(m.value) >= maxHeight {
				return nil
			}
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			m.splitLine(m.row, m.col)
		case "end", "ctrl+e":
			m.CursorEnd()
		case "home", "ctrl+a":
			m.CursorStart()
		case "right", "ctrl+f":
			if m.col < len(m.value[m.row]) {
				m.SetCursor(m.col + 1)
			} else {
				if m.row < len(m.value)-1 {
					m.row++
					m.CursorStart()
				}
			}
		case "down", "ctrl+n":
			m.CursorDown()
		case "alt+right", "alt+f":
			m.wordRight()
		case "ctrl+v":
			return Paste
		case "left", "ctrl+b":
			if m.col == 0 && m.row != 0 {
				m.row--
				m.CursorEnd()
				break
			}
			if m.col > 0 {
				m.SetCursor(m.col - 1)
			}
		case "up", "ctrl+p":
			m.CursorUp()
		case "alt+left", "alt+b":
			m.wordLeft()
		default:
			if m.CharLimit > 0 && rw.StringWidth(m.Value()) >= m.CharLimit {
				break
			}

			m.col = min(m.col, len(m.value[m.row]))
			m.value[m.row] = append(m.value[m.row][:m.col], append(msg.Runes, m.value[m.row][m.col:]...)...)
			m.SetCursor(m.col + len(msg.Runes))
		}

	case pasteMsg:
		m.handlePaste(string(msg))

	case pasteErrMsg:
		m.Err = msg
	}

	return nil
}

func (m *Model) normalUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	var execute bool

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.command.Action == ActionReplace {
			for i := m.col; i < m.col+max(m.command.Count, 1); i++ {
				if i >= len(m.value[m.row]) || len(msg.Runes) <= 0 {
					break
				}
				m.value[m.row][i] = msg.Runes[0]
			}
			m.command = &NormalCommand{}
			return nil
		}
		switch msg.String() {
		case "esc":
			m.command = &NormalCommand{}
			return m.SetMode(ModeNormal)
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			if m.command.Count == 0 && msg.String() == "0" {
				m.command.Range = Range{
					Start: Position{Row: m.row, Col: m.col},
					End:   Position{Row: m.row, Col: 0},
				}
				execute = true
				break
			}

			v := m.command.Buffer + msg.String()
			count, err := strconv.Atoi(v)
			if err != nil {
				count, _ = strconv.Atoi(msg.String())
				m.command.Buffer = msg.String()
				m.command.Count = count
			} else {
				m.command.Buffer = v
				m.command.Count = count
			}
		case "G":
			var row int
			if m.command.Count > 0 {
				row = m.command.Count - 1
			} else {
				row = len(m.value) - 1
			}
			m.row = clamp(row, 0, len(m.value)-1)
			return nil
		case "g":
			if m.command.Buffer == "g" {
				m.command = &NormalCommand{}
				m.row = clamp(m.command.Count-1, 0, len(m.value)-1)
			} else {
				m.command = &NormalCommand{Buffer: "g"}
			}
			return nil
		case "x":
			m.command.Action = ActionDelete
			m.command.Range = Range{
				Start: Position{Row: m.row, Col: m.col},
				End:   Position{Row: m.row, Col: m.col + max(m.command.Count, 1)},
			}
		case "X":
			m.command.Action = ActionDelete
			m.command.Range = Range{
				Start: Position{Row: m.row, Col: m.col},
				End:   Position{Row: m.row, Col: m.col - max(m.command.Count, 1)},
			}
		case "c":
			if m.command.Action == ActionChange {
				m.CursorStart()
				m.deleteAfterCursor()
				m.command = &NormalCommand{}
				return m.SetMode(ModeInsert)
			}
			m.command.Action = ActionChange
		case "d":
			if m.command.Action == ActionDelete {
				for i := 0; i < max(m.command.Count, 1); i++ {
					m.value[m.row] = []rune{}
					if m.row < len(m.value)-1 {
						m.mergeLineBelow(m.row)
					} else {
						m.mergeLineAbove(m.row)
					}
				}
				m.command = &NormalCommand{}
				return nil
			}
			m.command.Action = ActionDelete
		case "y":
			m.command.Action = ActionYank
		case "r":
			m.command.Action = ActionReplace
		case "i":
			m.command.Range = Range{
				Start: Position{Row: m.row, Col: m.col},
				End:   Position{Row: m.row, Col: m.col},
			}
			cmd = m.SetMode(ModeInsert)
		case "I":
			m.command.Range = Range{
				Start: Position{Row: m.row, Col: m.col},
				End:   Position{Row: m.row, Col: 0},
			}
			cmd = m.SetMode(ModeInsert)
		case "a":
			m.command.Range = Range{
				Start: Position{Row: m.row, Col: m.col},
				End:   Position{Row: m.row, Col: m.col + 1},
			}
			cmd = m.SetMode(ModeInsert)
		case "A":
			m.command.Range = Range{
				Start: Position{Row: m.row, Col: m.col},
				End:   Position{Row: m.row, Col: len(m.value[m.row]) + 1},
			}
			cmd = m.SetMode(ModeInsert)
		case "^":
			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   Position{m.row, 0},
			}
		case "$":
			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   Position{m.row, len(m.value[m.row])},
			}
		case "e", "E":
			end := m.findWordEndRight(max(m.command.Count, 1), msg.String() == "E")
			if m.command.Action == ActionDelete {
				end.Col = min(end.Col+1, len(m.value[end.Row]))
			}
			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   end,
			}
		case "w", "W":
			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   m.findWordStartRight(max(m.command.Count, 1), msg.String() == "W"),
			}
		case "b", "B":
			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   m.findWordLeft(max(m.command.Count, 1), msg.String() == "B"),
			}
		case "h", "l":
			direction := 1
			if msg.String() == "h" {
				direction = -1
			}
			m.lastCharOffset = 0
			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   Position{m.row, clamp(m.col+(direction*max(m.command.Count, 1)), 0, len(m.value[m.row]))},
			}
		case "j", "k":
			direction := 1
			if msg.String() == "k" {
				direction = -1
			}
			row := clamp(m.row+(direction*max(m.command.Count, 1)), 0, len(m.value)-1)
			li := m.LineInfo()
			charOffset := max(m.lastCharOffset, li.CharOffset)
			m.lastCharOffset = charOffset

			rowContent := m.value[row]
			charWidth := rw.StringWidth(string(rowContent))

			col := 0
			offset := 0

			for offset < charOffset {
				if col > len(m.value[row]) || offset >= charWidth-1 {
					break
				}
				offset += rw.RuneWidth(m.value[row][col])
				col++
			}

			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   Position{row, col},
			}
		case "C":
			m.deleteAfterCursor()
			m.command = &NormalCommand{}
			return m.SetMode(ModeInsert)
		case "D":
			m.deleteAfterCursor()
			m.command = &NormalCommand{}
			return nil
		case "J":
			m.CursorEnd()
			m.mergeLineBelow(m.row)
			return nil
		case "p":
			cmd = Paste
		}

		if !strings.ContainsAny(msg.String(), "jk") {
			m.lastCharOffset = 0
		}

		if strings.ContainsAny(msg.String(), "iIaAeEwWbBhjklp$^xX") || execute {
			switch m.command.Action {
			case ActionDelete:
				m.deleteRange(m.command.Range)
			case ActionChange:
				m.deleteRange(m.command.Range)
				cmd = m.SetMode(ModeInsert)
			case ActionMove:
				m.row = clamp(m.command.Range.End.Row, 0, len(m.value)-1)
				m.col = clamp(m.command.Range.End.Col, 0, len(m.value[m.row]))
			}
			m.command = &NormalCommand{}
		}

	case pasteMsg:
		m.handlePaste(string(msg))
	}

	return cmd
}

// findWordLeft locates the start of the word on the left of the current word.
// It takes whether or not to break words on spaces or any non-alpha-numeric
// character as an argument.
func (m *Model) findWordLeft(count int, ignorePunctuation bool) Position {
	wordBreak := isSoftWordBreak
	if ignorePunctuation {
		wordBreak = isWordBreak
	}

	row, col := m.row, m.col

	for count > 0 {
		if col <= 0 && row > 0 {
			row--
			col = len(m.value[row]) - 1
		}

		// Skip all spaces (and punctuation) to the left of the cursor.
		for col > 0 && wordBreak(m.value[row][col-1]) {
			col--
		}

		// Then skip all non-spaces to the left of the cursor.
		for col > 0 && !wordBreak(m.value[row][col-1]) {
			col--
		}

		count--

		if row <= 0 && col <= 0 {
			break
		}
	}

	return Position{Row: row, Col: col}
}

// findWordStartRight locates the start of the next word. It takes whether or not to
// break words on spaces or any non-alpha-numeric character as an argument.
func (m *Model) findWordStartRight(count int, ignorePunctuation bool) Position {
	wordBreak := isSoftWordBreak
	if ignorePunctuation {
		wordBreak = isWordBreak
	}

	row, col := m.row, m.col

	for count > 0 {
		if col >= len(m.value[row])-1 && row < len(m.value)-1 {
			row++
			col = 0
		}

		// Skip until the start of a word is found.
		for col < len(m.value[row]) && !wordBreak(m.value[row][col]) {
			col++
		}
		// Skip all spaces to the right of the cursor.
		for col < len(m.value[row])-1 && wordBreak(m.value[row][col]) {
			col++
		}
		count--

		if row >= len(m.value)-1 && col >= len(m.value[row])-1 {
			break
		}
	}

	return Position{Row: row, Col: col}
}

// findWordEndRight locates the start of the next word. It takes whether or not to
// break words on spaces or any non-alpha-numeric character as an argument.
func (m *Model) findWordEndRight(count int, ignorePunctuation bool) Position {
	wordBreak := isSoftWordBreak
	if ignorePunctuation {
		wordBreak = isWordBreak
	}

	row, col := m.row, m.col

	for count > 0 {
		if col > len(m.value[row]) && row < len(m.value)-1 {
			row++
			col = 0
		}

		// Skip all spaces to the right of the cursor.
		for col < len(m.value[row])-1 && wordBreak(m.value[row][col+1]) {
			col++
		}

		// Then skip all non-spaces to the right of the cursor.
		for col < len(m.value[row])-1 && !wordBreak(m.value[row][col+1]) {
			col++
		}

		count--

		if row <= 0 && col <= 0 {
			break
		}
	}

	return Position{Row: row, Col: col}
}

func isWordBreak(char rune) bool {
	return unicode.IsSpace(char)
}

func isSoftWordBreak(char rune) bool {
	return unicode.IsSpace(char) || unicode.IsPunct(char)
}

func (m *Model) deleteRange(r Range) {
	if r.Start.Row == r.End.Row && r.Start.Col == r.End.Col {
		return
	}

	minCol, maxCol := min(r.Start.Col, r.End.Col), max(r.Start.Col, r.End.Col)

	minCol = clamp(minCol, 0, len(m.value[r.Start.Row]))
	maxCol = clamp(maxCol, 0, len(m.value[r.Start.Row]))

	if r.Start.Row == r.End.Row {
		m.value[r.Start.Row] = append(m.value[r.Start.Row][:minCol], m.value[r.Start.Row][maxCol:]...)
		m.SetCursor(minCol)
		return
	}

	minRow, maxRow := min(r.Start.Row, r.End.Row), max(r.Start.Row, r.End.Row)

	for i := max(minRow, 0); i <= min(maxRow, len(m.value)-1); i++ {
		m.value[i] = []rune{}
	}

	m.value = append(m.value[:minRow], m.value[maxRow:]...)

	m.row = clamp(0, minRow, len(m.value))
}
