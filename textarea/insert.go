package textarea

import (
	tea "github.com/charmbracelet/bubbletea"
	rw "github.com/mattn/go-runewidth"
)

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
