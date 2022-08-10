package textarea

import (
	"strconv"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) normalUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	var execute bool

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.command = &NormalCommand{}
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
			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   m.findWordRight(max(m.command.Count, 1), msg.String() == "E"),
			}
		case "w", "W":
			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   m.findWordRight(max(m.command.Count, 1), msg.String() == "W"),
			}
		case "b", "B":
			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   m.findWordLeft(max(m.command.Count, 1), msg.String() == "B"),
			}
		case "h":
			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   Position{m.row, m.col - max(m.command.Count, 1)},
			}
		case "j":
			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   Position{min(m.row+max(m.command.Count, 1), len(m.value)-1), m.col},
			}
		case "k":
			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   Position{max(m.row-max(m.command.Count, 1), 0), m.col},
			}
		case "l":
			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   Position{m.row, m.col + max(m.command.Count, 1)},
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

		if strings.ContainsAny(msg.String(), "iIaAewWbBhjklp$^xX") || execute {
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

// findWordRight locates the end of the current word or the end of the next word
// if already at the end of the current word. It takes whether or not to break
// words on spaces or any non-alpha-numeric character as an argument.
func (m *Model) findWordRight(count int, _ bool) Position {
	row, col := m.row, m.col

	for count > 0 {
		if col < len(m.value[row])-1 && unicode.IsSpace(m.value[row][col+1]) {
			col++
		} else if col >= len(m.value[row])-1 && row < len(m.value)-1 {
			row++
			col = 0
		}

		for col < len(m.value[row])-1 {
			if !unicode.IsSpace(m.value[row][col+1]) {
				col++
			} else {
				count--
				break
			}
		}

		if row >= len(m.value)-1 && col >= len(m.value[row])-1 {
			break
		}
	}

	return Position{Row: row, Col: col}
}

// findWordLeft locates the start of the next word. It takes whether or not to
// break words on spaces or any non-alpha-numeric character as an argument.
func (m *Model) findWordLeft(count int, onlySpaces bool) Position {
	_ = onlySpaces
	row, col := m.row, m.col

	for count > 0 {
		if col > 0 && unicode.IsSpace(m.value[row][col-1]) {
			col--
		} else if col <= 0 && row > 0 {
			row--
			col = len(m.value[row]) - 1
		}

		for col > 0 {
			if !unicode.IsSpace(m.value[row][col-1]) {
				col--
			} else {
				count--
				break
			}
		}

		if row <= 0 && col <= 0 {
			break
		}
	}

	return Position{Row: row, Col: col}
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

	m.row = clamp(0, minRow, len(m.value)-1)
}
