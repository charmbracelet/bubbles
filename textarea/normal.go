package textarea

import (
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) normalUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.command = &NormalCommand{}
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			v := m.command.Buffer + msg.String()
			count, err := strconv.Atoi(v)
			if err != nil {
				count, _ = strconv.Atoi(msg.String())
				m.command = &NormalCommand{Buffer: msg.String(), Count: count}
			} else {
				m.command = &NormalCommand{Buffer: v, Count: count}
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
		case "d":
			if m.command.Action == ActionDelete {
				for i := 0; i < max(m.command.Count, 1); i++ {
					m.CursorStart()
					m.deleteAfterCursor()
					m.mergeLineBelow(m.row)
				}
				m.command = &NormalCommand{}
				break
			}
			m.command.Action = ActionDelete
		case "y":
			m.command.Action = ActionYank
		case "r":
			m.command.Action = ActionReplace
		case "i":
			return m.SetMode(ModeInsert)
		case "I":
			m.CursorStart()
			return m.SetMode(ModeInsert)
		case "a":
			m.SetCursor(m.col + 1)
			return m.SetMode(ModeInsert)
		case "A":
			m.CursorEnd()
			return m.SetMode(ModeInsert)
		case "^":
			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   Position{m.row, 0},
			}
		case "$":
			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   Position{m.row, len(m.value[m.row]) - 1},
			}
		case "e", "E":
			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   m.findWordEnd(m.command.Count, msg.String() == "E"),
			}
		case "w", "W":
			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   m.findWordStart(m.command.Count, msg.String() == "W"),
			}
		case "b", "B":
			direction := -1
			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   m.findWordStart(direction*m.command.Count, msg.String() == "B"),
			}
		case "h":
			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   Position{m.row, m.col - max(m.command.Count, 1)},
			}
		case "j":
			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   Position{m.row + max(m.command.Count, 1), m.col},
			}
		case "k":
			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   Position{m.row - max(m.command.Count, 1), m.col},
			}
		case "l":
			m.command.Range = Range{
				Start: Position{m.row, m.col},
				End:   Position{m.row, m.col + max(m.command.Count, 1)},
			}
		case "J":
			m.CursorEnd()
			m.mergeLineBelow(m.row)
			return nil
		case "p":
			cmd = Paste
		}

		switch msg.String() {
		case "i", "I", "a", "A", "e", "w", "W", "b", "B", "h", "j", "k", "l", "p", "$", "^":
			switch m.command.Action {
			case ActionDelete:
				m.deleteRange(m.command.Range)
			case ActionMove:
				rowDelta := m.command.Range.End.Row - m.command.Range.Start.Row
				if rowDelta > 0 {
					for i := 0; i < rowDelta; i++ {
						m.CursorDown()
					}
				} else if rowDelta < 0 {
					for i := 0; i < -rowDelta; i++ {
						m.CursorUp()
					}
				} else {
					m.SetCursor(m.command.Range.End.Col)
				}
			}
			m.command = &NormalCommand{}
		}

	case pasteMsg:
		m.handlePaste(string(msg))
	}

	return cmd
}
