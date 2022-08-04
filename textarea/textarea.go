package textarea

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	rw "github.com/mattn/go-runewidth"
)

const (
	minHeight        = 1
	minWidth         = 2
	defaultHeight    = 6
	defaultWidth     = 40
	defaultCharLimit = 400
	maxHeight        = 99
	maxWidth         = 500
)

// Internal messages for clipboard operations.
type pasteMsg string
type pasteErrMsg struct{ error }

// KeyMap is the key bindings for different actions within the textarea.
type KeyMap struct {
	CharacterBackward       key.Binding
	CharacterForward        key.Binding
	DeleteAfterCursor       key.Binding
	DeleteBeforeCursor      key.Binding
	DeleteCharacterBackward key.Binding
	DeleteCharacterForward  key.Binding
	DeleteWordBackward      key.Binding
	DeleteWordForward       key.Binding
	InsertNewline           key.Binding
	LineEnd                 key.Binding
	LineNext                key.Binding
	LinePrevious            key.Binding
	LineStart               key.Binding
	Paste                   key.Binding
	WordBackward            key.Binding
	WordForward             key.Binding
}

// DefaultKeyMap is the default set of key bindings for navigating and acting
// upon the textarea.
var DefaultKeyMap = KeyMap{
	CharacterForward:        key.NewBinding(key.WithKeys("right", "ctrl+f")),
	CharacterBackward:       key.NewBinding(key.WithKeys("left", "ctrl+b")),
	WordForward:             key.NewBinding(key.WithKeys("alt+right", "alt+f")),
	WordBackward:            key.NewBinding(key.WithKeys("alt+left", "alt+b")),
	LineNext:                key.NewBinding(key.WithKeys("down", "ctrl+n")),
	LinePrevious:            key.NewBinding(key.WithKeys("up", "ctrl+p")),
	DeleteWordBackward:      key.NewBinding(key.WithKeys("alt+backspace", "ctrl+w")),
	DeleteWordForward:       key.NewBinding(key.WithKeys("alt+delete", "alt+d")),
	DeleteAfterCursor:       key.NewBinding(key.WithKeys("ctrl+k")),
	DeleteBeforeCursor:      key.NewBinding(key.WithKeys("ctrl+u")),
	InsertNewline:           key.NewBinding(key.WithKeys("enter", "ctrl+m")),
	DeleteCharacterBackward: key.NewBinding(key.WithKeys("backspace", "ctrl+h")),
	DeleteCharacterForward:  key.NewBinding(key.WithKeys("delete", "ctrl+d")),
	LineStart:               key.NewBinding(key.WithKeys("home", "ctrl+a")),
	LineEnd:                 key.NewBinding(key.WithKeys("end", "ctrl+e")),
	Paste:                   key.NewBinding(key.WithKeys("ctrl+v")),
}

// LineInfo is a helper for keeping track of line information regarding
// soft-wrapped lines.
type LineInfo struct {
	// Width is the number of columns in the line.
	Width int
	// CharWidth is the number of characters in the line to account for
	// double-width runes.
	CharWidth int
	// Height is the number of rows in the line.
	Height int
	// StartColumn is the index of the first column of the line.
	StartColumn int
	// ColumnOffset is the number of columns that the cursor is offset from the
	// start of the line.
	ColumnOffset int
	// RowOffset is the number of rows that the cursor is offset from the start
	// of the line.
	RowOffset int
	// CharOffset is the number of characters that the cursor is offset
	// from the start of the line. This will generally be equivalent to
	// ColumnOffset, but will be different there are double-width runes before
	// the cursor.
	CharOffset int
}

// Style that will be applied to the text area.
//
// Style can be applied to focused and unfocused states to change the styles
// depending on the focus state.
//
// For an introduction to styling with Lip Gloss see:
// https://github.com/charmbracelet/lipgloss
type Style struct {
	Base             lipgloss.Style
	CursorLine       lipgloss.Style
	CursorLineNumber lipgloss.Style
	EndOfBuffer      lipgloss.Style
	LineNumber       lipgloss.Style
	Placeholder      lipgloss.Style
	Prompt           lipgloss.Style
	Text             lipgloss.Style
}

// Model is the Bubble Tea model for this text area element.
type Model struct {
	Err error

	// General settings.
	Prompt               string
	Placeholder          string
	ShowLineNumbers      bool
	EndOfBufferCharacter rune
	KeyMap               KeyMap

	// Styling. FocusedStyle and BlurredStyle are used to style the textarea in
	// focused and blurred states.
	FocusedStyle Style
	BlurredStyle Style
	// style is the current styling to use.
	// It is used to abstract the differences in focus state when styling the
	// model, since we can simply assign the set of styles to this variable
	// when switching focus states.
	style *Style

	// Cursor is the text area cursor.
	Cursor cursor.Model

	// CharLimit is the maximum number of characters this input element will
	// accept. If 0 or less, there's no limit.
	CharLimit int

	// width is the maximum number of characters that can be displayed at once.
	// If 0 or less this setting is ignored.
	width int

	// height is the maximum number of lines that can be displayed at once. It
	// essentially treats the text field like a vertically scrolling viewport
	// if there are more lines than the permitted height.
	height int

	// Underlying text value.
	value [][]rune

	// focus indicates whether user input focus should be on this input
	// component. When false, ignore keyboard input and hide the cursor.
	focus bool

	// Cursor column.
	col int

	// Cursor row.
	row int

	// Last character offset, used to maintain state when the cursor is moved
	// vertically such that we can maintain the same navigating position.
	lastCharOffset int

	// lineNumberFormat is the format string used to display line numbers.
	lineNumberFormat string

	// viewport is the vertically-scrollable viewport of the multi-line text
	// input.
	viewport *viewport.Model
}

// New creates a new model with default settings.
func New() Model {
	vp := viewport.New(0, 0)
	vp.KeyMap = viewport.KeyMap{}
	cur := cursor.New()

	focusedStyle, blurredStyle := DefaultStyles()

	m := Model{
		CharLimit:            defaultCharLimit,
		Prompt:               lipgloss.ThickBorder().Left + " ",
		style:                &blurredStyle,
		FocusedStyle:         focusedStyle,
		BlurredStyle:         blurredStyle,
		EndOfBufferCharacter: '~',
		ShowLineNumbers:      true,
		Cursor:               cur,
		KeyMap:               DefaultKeyMap,

		value:            make([][]rune, minHeight, maxWidth),
		focus:            false,
		col:              0,
		row:              0,
		lineNumberFormat: "%2v ",

		viewport: &vp,
	}

	m.SetHeight(defaultHeight)
	m.SetWidth(defaultWidth)

	return m
}

// DefaultStyles returns the default styles for focused and blurred states for
// the textarea.
func DefaultStyles() (Style, Style) {
	focused := Style{
		Base:             lipgloss.NewStyle(),
		CursorLine:       lipgloss.NewStyle().Background(lipgloss.AdaptiveColor{Light: "255", Dark: "0"}),
		CursorLineNumber: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "240"}),
		EndOfBuffer:      lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "254", Dark: "0"}),
		LineNumber:       lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "249", Dark: "7"}),
		Placeholder:      lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Prompt:           lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
		Text:             lipgloss.NewStyle(),
	}
	blurred := Style{
		Base:             lipgloss.NewStyle(),
		CursorLine:       lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "245", Dark: "7"}),
		CursorLineNumber: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "249", Dark: "7"}),
		EndOfBuffer:      lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "254", Dark: "0"}),
		LineNumber:       lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "249", Dark: "7"}),
		Placeholder:      lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Prompt:           lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
		Text:             lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "245", Dark: "7"}),
	}

	return focused, blurred
}

// SetValue sets the value of the text input.
func (m *Model) SetValue(s string) {
	m.Reset()
	m.InsertString(s)
}

// InsertString inserts a string at the cursor position.
func (m *Model) InsertString(s string) {
	lines := strings.Split(s, "\n")
	for l, line := range lines {
		for _, rune := range line {
			m.InsertRune(rune)
		}
		if l != len(lines)-1 {
			m.InsertRune('\n')
		}
	}
}

// InsertRune inserts a rune at the cursor position.
func (m *Model) InsertRune(r rune) {
	if r == '\n' {
		m.splitLine(m.row, m.col)
		return
	}

	m.value[m.row] = append(m.value[m.row][:m.col], append([]rune{r}, m.value[m.row][m.col:]...)...)
	m.col++
}

// Value returns the value of the text input.
func (m Model) Value() string {
	if m.value == nil {
		return ""
	}

	var v string
	for _, l := range m.value {
		v += string(l)
		v += "\n"
	}

	return strings.TrimSuffix(v, "\n")
}

// Length returns the number of characters currently in the text input.
func (m *Model) Length() int {
	var l int
	for _, row := range m.value {
		l += rw.StringWidth(string(row))
	}
	return l
}

// Line returns the line position.
func (m Model) Line() int {
	return m.row
}

// CursorDown moves the cursor down by one line.
// Returns whether or not the cursor blink should be reset.
func (m *Model) CursorDown() {
	li := m.LineInfo()
	charOffset := max(m.lastCharOffset, li.CharOffset)
	m.lastCharOffset = charOffset

	if li.RowOffset+1 >= li.Height && m.row < len(m.value)-1 {
		m.row++
		m.col = 0
	} else {
		// Move the cursor to the start of the next line. So that we can get
		// the line information. We need to add 2 columns to account for the
		// trailing space wrapping.
		m.col = min(li.StartColumn+li.Width+2, len(m.value[m.row])-1)
	}

	nli := m.LineInfo()
	m.col = nli.StartColumn

	if nli.Width <= 0 {
		return
	}

	offset := 0
	for offset < charOffset {
		if m.col > len(m.value[m.row]) || offset >= nli.CharWidth-1 {
			break
		}
		offset += rw.RuneWidth(m.value[m.row][m.col])
		m.col++
	}
}

// CursorUp moves the cursor up by one line.
func (m *Model) CursorUp() {
	li := m.LineInfo()
	charOffset := max(m.lastCharOffset, li.CharOffset)
	m.lastCharOffset = charOffset

	if li.RowOffset <= 0 && m.row > 0 {
		m.row--
		m.col = len(m.value[m.row])
	} else {
		// Move the cursor to the end of the previous line.
		// This can be done by moving the cursor to the start of the line and
		// then subtracting 2 to account for the trailing space we keep on
		// soft-wrapped lines.
		m.col = li.StartColumn - 2
	}

	nli := m.LineInfo()
	m.col = nli.StartColumn

	if nli.Width <= 0 {
		return
	}

	offset := 0
	for offset < charOffset {
		if m.col >= len(m.value[m.row]) || offset >= nli.CharWidth-1 {
			break
		}
		offset += rw.RuneWidth(m.value[m.row][m.col])
		m.col++
	}
}

// SetCursor moves the cursor to the given position. If the position is
// out of bounds the cursor will be moved to the start or end accordingly.
func (m *Model) SetCursor(col int) {
	m.col = clamp(col, 0, len(m.value[m.row]))
	// Any time that we move the cursor horizontally we need to reset the last
	// offset so that the horizontal position when navigating is adjusted.
	m.lastCharOffset = 0
}

// CursorStart moves the cursor to the start of the input field.
func (m *Model) CursorStart() {
	m.SetCursor(0)
}

// CursorEnd moves the cursor to the end of the input field.
func (m *Model) CursorEnd() {
	m.SetCursor(len(m.value[m.row]))
}

// Focused returns the focus state on the model.
func (m Model) Focused() bool {
	return m.focus
}

// Focus sets the focus state on the model. When the model is in focus it can
// receive keyboard input and the cursor will be hidden.
func (m *Model) Focus() tea.Cmd {
	m.focus = true
	m.style = &m.FocusedStyle
	return m.Cursor.Focus()
}

// Blur removes the focus state on the model.  When the model is blurred it can
// not receive keyboard input and the cursor will be hidden.
func (m *Model) Blur() {
	m.focus = false
	m.style = &m.BlurredStyle
	m.Cursor.Blur()
}

// Reset sets the input to its default state with no input.
func (m *Model) Reset() {
	m.value = make([][]rune, minHeight, maxWidth)
	m.col = 0
	m.row = 0
	m.viewport.GotoTop()
	m.SetCursor(0)
}

// handle a clipboard paste event, if supported.
func (m *Model) handlePaste(v string) {
	paste := []rune(v)

	var availSpace int
	if m.CharLimit > 0 {
		availSpace = m.CharLimit - m.Length()
	}

	// If the char limit's been reached cancel
	if m.CharLimit > 0 && availSpace <= 0 {
		return
	}

	// If there's not enough space to paste the whole thing cut the pasted
	// runes down so they'll fit
	if m.CharLimit > 0 && availSpace < len(paste) {
		paste = paste[:len(paste)-availSpace]
	}

	// Stuff before and after the cursor
	head := m.value[m.row][:m.col]
	tailSrc := m.value[m.row][m.col:]
	tail := make([]rune, len(tailSrc))
	copy(tail, tailSrc)

	// Insert pasted runes
	for _, r := range paste {
		head = append(head, r)
		m.col++
		if m.CharLimit > 0 {
			availSpace--
			if availSpace <= 0 {
				break
			}
		}
	}

	// Reset blink state if necessary and run overflow checks
	m.SetCursor(m.col + len(paste))
}

// deleteBeforeCursor deletes all text before the cursor. Returns whether or
// not the cursor blink should be reset.
func (m *Model) deleteBeforeCursor() {
	m.value[m.row] = m.value[m.row][m.col:]
	m.SetCursor(0)
}

// deleteAfterCursor deletes all text after the cursor. Returns whether or not
// the cursor blink should be reset. If input is masked delete everything after
// the cursor so as not to reveal word breaks in the masked input.
func (m *Model) deleteAfterCursor() {
	m.value[m.row] = m.value[m.row][:m.col]
	m.SetCursor(len(m.value[m.row]))
}

// deleteWordLeft deletes the word left to the cursor. Returns whether or not
// the cursor blink should be reset.
func (m *Model) deleteWordLeft() {
	if m.col == 0 || len(m.value[m.row]) == 0 {
		return
	}

	// Linter note: it's critical that we acquire the initial cursor position
	// here prior to altering it via SetCursor() below. As such, moving this
	// call into the corresponding if clause does not apply here.
	oldCol := m.col //nolint:ifshort

	m.SetCursor(m.col - 1)
	for unicode.IsSpace(m.value[m.row][m.col]) {
		if m.col <= 0 {
			break
		}
		// ignore series of whitespace before cursor
		m.SetCursor(m.col - 1)
	}

	for m.col > 0 {
		if !unicode.IsSpace(m.value[m.row][m.col]) {
			m.SetCursor(m.col - 1)
		} else {
			if m.col > 0 {
				// keep the previous space
				m.SetCursor(m.col + 1)
			}
			break
		}
	}

	if oldCol > len(m.value[m.row]) {
		m.value[m.row] = m.value[m.row][:m.col]
	} else {
		m.value[m.row] = append(m.value[m.row][:m.col], m.value[m.row][oldCol:]...)
	}
}

// deleteWordRight deletes the word right to the cursor.
func (m *Model) deleteWordRight() {
	if m.col >= len(m.value[m.row]) || len(m.value[m.row]) == 0 {
		return
	}

	oldCol := m.col
	m.SetCursor(m.col + 1)
	for unicode.IsSpace(m.value[m.row][m.col]) {
		// ignore series of whitespace after cursor
		m.SetCursor(m.col + 1)

		if m.col >= len(m.value[m.row]) {
			break
		}
	}

	for m.col < len(m.value[m.row]) {
		if !unicode.IsSpace(m.value[m.row][m.col]) {
			m.SetCursor(m.col + 1)
		} else {
			break
		}
	}

	if m.col > len(m.value[m.row]) {
		m.value[m.row] = m.value[m.row][:oldCol]
	} else {
		m.value[m.row] = append(m.value[m.row][:oldCol], m.value[m.row][m.col:]...)
	}

	m.SetCursor(oldCol)
}

// wordLeft moves the cursor one word to the left. Returns whether or not the
// cursor blink should be reset. If input is masked, move input to the start
// so as not to reveal word breaks in the masked input.
func (m *Model) wordLeft() {
	if m.col == 0 || len(m.value[m.row]) == 0 {
		return
	}

	i := m.col - 1
	for i >= 0 {
		if unicode.IsSpace(m.value[m.row][min(i, len(m.value[m.row])-1)]) {
			m.SetCursor(m.col - 1)
			i--
		} else {
			break
		}
	}

	for i >= 0 {
		if !unicode.IsSpace(m.value[m.row][min(i, len(m.value[m.row])-1)]) {
			m.SetCursor(m.col - 1)
			i--
		} else {
			break
		}
	}
}

// wordRight moves the cursor one word to the right. Returns whether or not the
// cursor blink should be reset. If the input is masked, move input to the end
// so as not to reveal word breaks in the masked input.
func (m *Model) wordRight() {
	if m.col >= len(m.value[m.row]) || len(m.value[m.row]) == 0 {
		return
	}

	i := m.col
	for i < len(m.value[m.row]) {
		if unicode.IsSpace(m.value[m.row][i]) {
			m.SetCursor(m.col + 1)
			i++
		} else {
			break
		}
	}

	for i < len(m.value[m.row]) {
		if !unicode.IsSpace(m.value[m.row][i]) {
			m.SetCursor(m.col + 1)
			i++
		} else {
			break
		}
	}
}

// LineInfo returns the number of characters from the start of the
// (soft-wrapped) line and the (soft-wrapped) line width.
func (m Model) LineInfo() LineInfo {
	grid := wrap(m.value[m.row], m.width)

	// Find out which line we are currently on. This can be determined by the
	// m.col and counting the number of runes that we need to skip.
	var counter int
	for i, line := range grid {
		// We've found the line that we are on
		if counter+len(line) == m.col && i+1 < len(grid) {
			// We wrap around to the next line if we are at the end of the
			// previous line so that we can be at the very beginning of the row
			return LineInfo{
				CharOffset:   0,
				ColumnOffset: 0,
				Height:       len(grid),
				RowOffset:    i + 1,
				StartColumn:  m.col,
				Width:        len(grid[i+1]),
				CharWidth:    rw.StringWidth(string(line)),
			}
		}

		if counter+len(line) >= m.col {
			return LineInfo{
				CharOffset:   rw.StringWidth(string(line[:max(0, m.col-counter)])),
				ColumnOffset: m.col - counter,
				Height:       len(grid),
				RowOffset:    i,
				StartColumn:  counter,
				Width:        len(line),
				CharWidth:    rw.StringWidth(string(line)),
			}
		}

		counter += len(line)
	}
	return LineInfo{}
}

// repositionView repositions the view of the viewport based on the defined
// scrolling behavior.
func (m *Model) repositionView() {
	min := m.viewport.YOffset
	max := min + m.viewport.Height - 1

	if row := m.cursorLineNumber(); row < min {
		m.viewport.LineUp(min - row)
	} else if row > max {
		m.viewport.LineDown(row - max)
	}
}

// Width returns the width of the textarea.
func (m Model) Width() int {
	return m.width
}

// SetWidth sets the width of the textarea to fit exactly within the given width.
// This means that the textarea will account for the width of the prompt and
// whether or not line numbers are being shown.
//
// Ensure that SetWidth is called after setting the Prompt and ShowLineNumbers,
// If it important that the width of the textarea be exactly the given width
// and no more.
func (m *Model) SetWidth(w int) {
	m.viewport.Width = clamp(w, minWidth, maxWidth)

	// Since the width of the textarea input is dependant on the width of the
	// prompt and line numbers, we need to calculate it by subtracting.
	inputWidth := w
	if m.ShowLineNumbers {
		inputWidth -= rw.StringWidth(fmt.Sprintf(m.lineNumberFormat, 0))
	}

	// Account for base style borders and padding.
	inputWidth -= m.style.Base.GetHorizontalFrameSize()

	inputWidth -= rw.StringWidth(m.Prompt)
	m.width = clamp(inputWidth, minWidth, maxWidth)
}

// Height returns the current height of the textarea.
func (m Model) Height() int {
	return m.height
}

// SetHeight sets the height of the textarea.
func (m *Model) SetHeight(h int) {
	m.height = clamp(h, minHeight, maxHeight)
	m.viewport.Height = clamp(h, minHeight, maxHeight)
}

// Update is the Bubble Tea update loop.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focus {
		m.Cursor.Blur()
		return m, nil
	}

	// Used to determine if the cursor should blink.
	oldRow, oldCol := m.cursorLineNumber(), m.col

	var cmds []tea.Cmd

	if m.value[m.row] == nil {
		m.value[m.row] = make([]rune, 0)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.DeleteAfterCursor):
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			if m.col >= len(m.value[m.row]) {
				m.mergeLineBelow(m.row)
				break
			}
			m.deleteAfterCursor()
		case key.Matches(msg, m.KeyMap.DeleteBeforeCursor):
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			if m.col <= 0 {
				m.mergeLineAbove(m.row)
				break
			}
			m.deleteBeforeCursor()
		case key.Matches(msg, m.KeyMap.DeleteCharacterBackward):
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
		case key.Matches(msg, m.KeyMap.DeleteCharacterForward):
			if len(m.value[m.row]) > 0 && m.col < len(m.value[m.row]) {
				m.value[m.row] = append(m.value[m.row][:m.col], m.value[m.row][m.col+1:]...)
			}
			if m.col >= len(m.value[m.row]) {
				m.mergeLineBelow(m.row)
				break
			}
		case key.Matches(msg, m.KeyMap.DeleteWordBackward):
			if m.col <= 0 {
				m.mergeLineAbove(m.row)
				break
			}
			m.deleteWordLeft()
		case key.Matches(msg, m.KeyMap.DeleteWordForward):
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			if m.col >= len(m.value[m.row]) {
				m.mergeLineBelow(m.row)
				break
			}
			m.deleteWordRight()
		case key.Matches(msg, m.KeyMap.InsertNewline):
			if len(m.value) >= maxHeight {
				return m, nil
			}
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			m.splitLine(m.row, m.col)
		case key.Matches(msg, m.KeyMap.LineEnd):
			m.CursorEnd()
		case key.Matches(msg, m.KeyMap.LineStart):
			m.CursorStart()
		case key.Matches(msg, m.KeyMap.CharacterForward):
			if m.col < len(m.value[m.row]) {
				m.SetCursor(m.col + 1)
			} else {
				if m.row < len(m.value)-1 {
					m.row++
					m.CursorStart()
				}
			}
		case key.Matches(msg, m.KeyMap.LineNext):
			m.CursorDown()
		case key.Matches(msg, m.KeyMap.WordForward):
			m.wordRight()
		case key.Matches(msg, m.KeyMap.Paste):
			return m, Paste
		case key.Matches(msg, m.KeyMap.CharacterBackward):
			if m.col == 0 && m.row != 0 {
				m.row--
				m.CursorEnd()
				break
			}
			if m.col > 0 {
				m.SetCursor(m.col - 1)
			}
		case key.Matches(msg, m.KeyMap.LinePrevious):
			m.CursorUp()
		case key.Matches(msg, m.KeyMap.WordBackward):
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

	vp, cmd := m.viewport.Update(msg)
	m.viewport = &vp
	cmds = append(cmds, cmd)

	newRow, newCol := m.cursorLineNumber(), m.col
	m.Cursor, cmd = m.Cursor.Update(msg)
	if newRow != oldRow || newCol != oldCol {
		m.Cursor.Blink = false
		cmd = m.Cursor.BlinkCmd()
	}
	cmds = append(cmds, cmd)

	m.repositionView()

	return m, tea.Batch(cmds...)
}

// View renders the text area in its current state.
func (m Model) View() string {
	if m.Value() == "" && m.row == 0 && m.col == 0 && m.Placeholder != "" {
		return m.placeholderView()
	}
	m.Cursor.TextStyle = m.style.CursorLine

	var s strings.Builder
	var style lipgloss.Style
	lineInfo := m.LineInfo()

	var newLines int

	for l, line := range m.value {
		wrappedLines := wrap(line, m.width)

		if m.row == l {
			style = m.style.CursorLine
		} else {
			style = m.style.Text
		}

		for wl, wrappedLine := range wrappedLines {
			s.WriteString(style.Render(m.style.Prompt.Render(m.Prompt)))

			if m.ShowLineNumbers {
				if wl == 0 {
					if m.row == l {
						s.WriteString(style.Render(m.style.CursorLineNumber.Render(fmt.Sprintf(m.lineNumberFormat, l+1))))
					} else {
						s.WriteString(style.Render(m.style.LineNumber.Render(fmt.Sprintf(m.lineNumberFormat, l+1))))
					}
				} else {
					s.WriteString(m.style.LineNumber.Render(style.Render("   ")))
				}
			}

			strwidth := rw.StringWidth(string(wrappedLine))
			padding := m.width - strwidth
			// If the trailing space causes the line to be wider than the
			// width, we should not draw it to the screen since it will result
			// in an extra space at the end of the line which can look off when
			// the cursor line is showing.
			if strwidth > m.width {
				// The character causing the line to be wider than the width is
				// guaranteed to be a space since any other character would
				// have been wrapped.
				wrappedLine = []rune(strings.TrimSuffix(string(wrappedLine), " "))
				padding -= m.width - strwidth
			}
			if m.row == l && lineInfo.RowOffset == wl {
				s.WriteString(style.Render(string(wrappedLine[:lineInfo.ColumnOffset])))
				if m.col >= len(line) && lineInfo.CharOffset >= m.width {
					m.Cursor.SetChar(" ")
					s.WriteString(m.Cursor.View())
				} else {
					m.Cursor.SetChar(string(wrappedLine[lineInfo.ColumnOffset]))
					s.WriteString(style.Render(m.Cursor.View()))
					s.WriteString(style.Render(string(wrappedLine[lineInfo.ColumnOffset+1:])))
				}
			} else {
				s.WriteString(style.Render(string(wrappedLine)))
			}
			s.WriteString(style.Render(strings.Repeat(" ", max(0, padding))))
			s.WriteRune('\n')
			newLines++
		}
	}

	// Always show at least `m.Height` lines at all times.
	// To do this we can simply pad out a few extra new lines in the view.
	for i := 0; i < m.height; i++ {
		s.WriteString(m.style.Prompt.Render(m.Prompt))

		if m.ShowLineNumbers {
			lineNumber := m.style.EndOfBuffer.Render((fmt.Sprintf(m.lineNumberFormat, string(m.EndOfBufferCharacter))))
			s.WriteString(lineNumber)
		}
		s.WriteRune('\n')
	}

	m.viewport.SetContent(s.String())
	return m.style.Base.Render(m.viewport.View())
}

// placeholderView returns the prompt and placeholder view, if any.
func (m Model) placeholderView() string {
	var (
		s     strings.Builder
		p     = rw.Truncate(m.Placeholder, m.width, "...")
		style = m.style.Placeholder.Inline(true)
	)

	prompt := m.style.Prompt.Render(m.Prompt)
	s.WriteString(m.style.CursorLine.Render(prompt))

	if m.ShowLineNumbers {
		s.WriteString(m.style.CursorLine.Render(m.style.CursorLineNumber.Render((fmt.Sprintf(m.lineNumberFormat, 1)))))
	}

	m.Cursor.TextStyle = m.style.Placeholder
	m.Cursor.SetChar(string(p[0]))
	s.WriteString(m.style.CursorLine.Render(m.Cursor.View()))

	// The rest of the placeholder text
	s.WriteString(m.style.CursorLine.Render(style.Render(p[1:] + strings.Repeat(" ", max(0, m.width-rw.StringWidth(p))))))

	// The rest of the new lines
	for i := 1; i < m.height; i++ {
		s.WriteRune('\n')
		s.WriteString(prompt)

		if m.ShowLineNumbers {
			eob := m.style.EndOfBuffer.Render((fmt.Sprintf(m.lineNumberFormat, string(m.EndOfBufferCharacter))))
			s.WriteString(eob)
		}
	}

	m.viewport.SetContent(s.String())
	return m.style.Base.Render(m.viewport.View())
}

// Blink returns the blink command for the cursor.
func Blink() tea.Msg {
	return cursor.Blink()
}

// cursorLineNumber returns the line number that the cursor is on.
// This accounts for soft wrapped lines.
func (m Model) cursorLineNumber() int {
	line := 0
	for i := 0; i < m.row; i++ {
		// Calculate the number of lines that the current line will be split
		// into.
		line += len(wrap(m.value[i], m.width))
	}
	line += m.LineInfo().RowOffset
	return line
}

// mergeLineBelow merges the current line with the line below.
func (m *Model) mergeLineBelow(row int) {
	if row >= len(m.value)-1 {
		return
	}

	// To perform a merge, we will need to combine the two lines and then
	m.value[row] = append(m.value[row], m.value[row+1]...)

	// Shift all lines up by one
	for i := row + 1; i < len(m.value)-1; i++ {
		m.value[i] = m.value[i+1]
	}

	// And, remove the last line
	if len(m.value) > 0 {
		m.value = m.value[:len(m.value)-1]
	}
}

// mergeLineAbove merges the current line the cursor is on with the line above.
func (m *Model) mergeLineAbove(row int) {
	if row <= 0 {
		return
	}

	m.col = len(m.value[row-1])
	m.row = m.row - 1

	// To perform a merge, we will need to combine the two lines and then
	m.value[row-1] = append(m.value[row-1], m.value[row]...)

	// Shift all lines up by one
	for i := row; i < len(m.value)-1; i++ {
		m.value[i] = m.value[i+1]
	}

	// And, remove the last line
	if len(m.value) > 0 {
		m.value = m.value[:len(m.value)-1]
	}
}

func (m *Model) splitLine(row, col int) {
	// To perform a split, take the current line and keep the content before
	// the cursor, take the content after the cursor and make it the content of
	// the line underneath, and shift the remaining lines down by one
	head, tailSrc := m.value[row][:col], m.value[row][col:]
	tail := make([]rune, len(tailSrc))
	copy(tail, tailSrc)

	m.value = append(m.value[:row+1], m.value[row:]...)

	m.value[row] = head
	m.value[row+1] = tail

	m.col = 0
	m.row++
}

// Paste is a command for pasting from the clipboard into the text input.
func Paste() tea.Msg {
	str, err := clipboard.ReadAll()
	if err != nil {
		return pasteErrMsg{err}
	}
	return pasteMsg(str)
}

func wrap(runes []rune, width int) [][]rune {
	var (
		lines  = [][]rune{{}}
		word   = []rune{}
		row    int
		spaces int
	)

	// Word wrap the runes
	for _, r := range runes {
		if unicode.IsSpace(r) {
			spaces++
		} else {
			word = append(word, r)
		}

		if spaces > 0 {
			if rw.StringWidth(string(lines[row]))+rw.StringWidth(string(word))+spaces > width {
				row++
				lines = append(lines, []rune{})
				lines[row] = append(lines[row], word...)
				lines[row] = append(lines[row], repeatSpaces(spaces)...)
				spaces = 0
				word = nil
			} else {
				lines[row] = append(lines[row], word...)
				lines[row] = append(lines[row], repeatSpaces(spaces)...)
				spaces = 0
				word = nil
			}
		} else {
			// If the last character is a double-width rune, then we may not be able to add it to this line
			// as it might cause us to go past the width.
			lastCharLen := rw.RuneWidth(word[len(word)-1])
			if rw.StringWidth(string(word))+lastCharLen > width {
				// If the current line has any content, let's move to the next
				// line because the current word fills up the entire line.
				if len(lines[row]) > 0 {
					row++
					lines = append(lines, []rune{})
				}
				lines[row] = append(lines[row], word...)
				word = nil
			}
		}
	}

	if rw.StringWidth(string(lines[row]))+rw.StringWidth(string(word))+spaces >= width {
		lines = append(lines, []rune{})
		lines[row+1] = append(lines[row+1], word...)
		// We add an extra space at the end of the line to account for the
		// trailing space at the end of the previous soft-wrapped lines so that
		// behaviour when navigating is consistent and so that we don't need to
		// continually add edges to handle the last line of the wrapped input.
		spaces++
		lines[row+1] = append(lines[row+1], repeatSpaces(spaces)...)
	} else {
		lines[row] = append(lines[row], word...)
		spaces++
		lines[row] = append(lines[row], repeatSpaces(spaces)...)
	}

	return lines
}

func repeatSpaces(n int) []rune {
	return []rune(strings.Repeat(string(' '), n))
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
