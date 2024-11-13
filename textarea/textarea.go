package textarea

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/v2/cursor"
	"github.com/charmbracelet/bubbles/v2/internal/runeutil"
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/wcwidth"
	rw "github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
)

const (
	minHeight        = 1
	defaultHeight    = 6
	defaultWidth     = 40
	defaultCharLimit = 400
	defaultMaxHeight = 99
	defaultMaxWidth  = 500
)

// Internal messages for clipboard operations.
type (
	pasteMsg    string
	pasteErrMsg struct{ error }
)

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
	InputBegin              key.Binding
	InputEnd                key.Binding

	UppercaseWordForward  key.Binding
	LowercaseWordForward  key.Binding
	CapitalizeWordForward key.Binding

	TransposeCharacterBackward key.Binding
}

// DefaultKeyMap returns the default set of key bindings for navigating and acting
// upon the textarea.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		CharacterForward:        key.NewBinding(key.WithKeys("right", "ctrl+f"), key.WithHelp("right", "character forward")),
		CharacterBackward:       key.NewBinding(key.WithKeys("left", "ctrl+b"), key.WithHelp("left", "character backward")),
		WordForward:             key.NewBinding(key.WithKeys("alt+right", "alt+f"), key.WithHelp("alt+right", "word forward")),
		WordBackward:            key.NewBinding(key.WithKeys("alt+left", "alt+b"), key.WithHelp("alt+left", "word backward")),
		LineNext:                key.NewBinding(key.WithKeys("down", "ctrl+n"), key.WithHelp("down", "next line")),
		LinePrevious:            key.NewBinding(key.WithKeys("up", "ctrl+p"), key.WithHelp("up", "previous line")),
		DeleteWordBackward:      key.NewBinding(key.WithKeys("alt+backspace", "ctrl+w"), key.WithHelp("alt+backspace", "delete word backward")),
		DeleteWordForward:       key.NewBinding(key.WithKeys("alt+delete", "alt+d"), key.WithHelp("alt+delete", "delete word forward")),
		DeleteAfterCursor:       key.NewBinding(key.WithKeys("ctrl+k"), key.WithHelp("ctrl+k", "delete after cursor")),
		DeleteBeforeCursor:      key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "delete before cursor")),
		InsertNewline:           key.NewBinding(key.WithKeys("enter", "ctrl+m"), key.WithHelp("enter", "insert newline")),
		DeleteCharacterBackward: key.NewBinding(key.WithKeys("backspace", "ctrl+h"), key.WithHelp("backspace", "delete character backward")),
		DeleteCharacterForward:  key.NewBinding(key.WithKeys("delete", "ctrl+d"), key.WithHelp("delete", "delete character forward")),
		LineStart:               key.NewBinding(key.WithKeys("home", "ctrl+a"), key.WithHelp("home", "line start")),
		LineEnd:                 key.NewBinding(key.WithKeys("end", "ctrl+e"), key.WithHelp("end", "line end")),
		Paste:                   key.NewBinding(key.WithKeys("ctrl+v"), key.WithHelp("ctrl+v", "paste")),
		InputBegin:              key.NewBinding(key.WithKeys("alt+<", "ctrl+home"), key.WithHelp("alt+<", "input begin")),
		InputEnd:                key.NewBinding(key.WithKeys("alt+>", "ctrl+end"), key.WithHelp("alt+>", "input end")),

		CapitalizeWordForward: key.NewBinding(key.WithKeys("alt+c"), key.WithHelp("alt+c", "capitalize word forward")),
		LowercaseWordForward:  key.NewBinding(key.WithKeys("alt+l"), key.WithHelp("alt+l", "lowercase word forward")),
		UppercaseWordForward:  key.NewBinding(key.WithKeys("alt+u"), key.WithHelp("alt+u", "uppercase word forward")),

		TransposeCharacterBackward: key.NewBinding(key.WithKeys("ctrl+t"), key.WithHelp("ctrl+t", "transpose character backward")),
	}
}

// Cell represents a single cell in the textarea.
type Cell struct {
	Content string
	Width   int
}

// Line represents a single line in the textarea.
type Line []Cell

// Width returns the number of columns this line occupies wrapping at the
// textarea's width.
func (l Line) Width(width int) (w int) {
	for _, c := range l {
		if w+c.Width > width {
			return width
		}
		w += c.Width
	}
	return
}

// Height returns the number of rows in the line based on the textarea's width.
func (l Line) Height(width int) (h int) {
	lw := l.Width(width)
	if lw == 0 {
		return 1
	}
	return (lw / width) + 1
}

// Len returns the number of bytes in the line.
func (l Line) Len() (n int) {
	for _, c := range l {
		n += len(c.Content)
	}
	return
}

// RuneLen returns the number of runes in the line.
func (l Line) RuneLen() (n int) {
	for _, c := range l {
		n += utf8.RuneCountInString(c.Content)
	}
	return
}

// Insert inserts a cell at the given column.
func (l *Line) Insert(x int, c Cell) {
	_, i := l.At(x)
	*l = append((*l)[:i], append([]Cell{c}, (*l)[i:]...)...)
}

// Delete delete cells from the line between the start and end columns.
func (l *Line) Delete(start, end int) {
	_, start = l.At(start)
	_, end = l.At(end)
	*l = append((*l)[:start], (*l)[end:]...)
}

// Append appends a cell to the line.
func (l *Line) Append(c Cell) {
	*l = append(*l, c)
}

// Split splits the line at the given column.
func (l Line) Split(x int) (Line, Line) {
	_, i := l.At(x)
	return l[:i], l[i:]
}

// String returns the string representation of the line.
func (l Line) String() string {
	var s bytes.Buffer
	for _, c := range l {
		s.WriteString(c.Content)
	}
	return s.String()
}

// At returns the cell along with its index at the given column.
func (l Line) At(col int) (Cell, int) {
	var i, w int
	for i = 0; i < len(l); i++ {
		w += l[i].Width
		if w > col {
			break
		}
	}

	if i < len(l) {
		return l[i], i
	}

	return Cell{}, i
}

// TrimSuffix trims the suffix from the line.
func (l *Line) TrimSuffix(suffix string) {
	// XXX: This is incorrect. We should be trimming the suffix from the last
	// cells in the line.
	if strings.HasSuffix(l.String(), suffix) {
		*l = (*l)[:len(*l)-1]
	}
}

// Lines represents a collection of lines.
type Lines []Line

// String returns the string representation of the lines.
func (l Lines) String() string {
	var s bytes.Buffer
	for i, line := range l {
		s.WriteString(line.String())
		if i < len(l)-1 {
			s.WriteRune('\n')
		}
	}
	return s.String()
}

// Len returns the number of lines.
func (l Lines) Len() int {
	return len(l)
}

// Insert inserts a line at the given position.
func (l *Lines) Insert(y int, line Line) {
	*l = append((*l)[:y], append(Lines{line}, (*l)[y:]...)...)
}

// Delete deletes a line at the given position.
func (l *Lines) Delete(y int) {
	*l = append((*l)[:y], (*l)[y+1:]...)
}

// Append appends a line to the lines.
func (l *Lines) Append(line Line) {
	*l = append(*l, line)
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

// Styles are the styles for the textarea, separated into focused and blurred
// states. The appropriate styles will be chosen based on the focus state of
// the textarea.
type Styles struct {
	Focused StyleState
	Blurred StyleState
}

// StyleState that will be applied to the text area.
//
// StyleState can be applied to focused and unfocused states to change the styles
// depending on the focus state.
//
// For an introduction to styling with Lip Gloss see:
// https://github.com/charmbracelet/lipgloss
type StyleState struct {
	Base             lipgloss.Style
	CursorLine       lipgloss.Style
	CursorLineNumber lipgloss.Style
	EndOfBuffer      lipgloss.Style
	LineNumber       lipgloss.Style
	Placeholder      lipgloss.Style
	Prompt           lipgloss.Style
	Text             lipgloss.Style
}

func (s StyleState) computedCursorLine() lipgloss.Style {
	return s.CursorLine.Inherit(s.Base).Inline(true)
}

func (s StyleState) computedCursorLineNumber() lipgloss.Style {
	return s.CursorLineNumber.
		Inherit(s.CursorLine).
		Inherit(s.Base).
		Inline(true)
}

func (s StyleState) computedEndOfBuffer() lipgloss.Style {
	return s.EndOfBuffer.Inherit(s.Base).Inline(true)
}

func (s StyleState) computedLineNumber() lipgloss.Style {
	return s.LineNumber.Inherit(s.Base).Inline(true)
}

func (s StyleState) computedPlaceholder() lipgloss.Style {
	return s.Placeholder.Inherit(s.Base).Inline(true)
}

func (s StyleState) computedPrompt() lipgloss.Style {
	return s.Prompt.Inherit(s.Base).Inline(true)
}

func (s StyleState) computedText() lipgloss.Style {
	return s.Text.Inherit(s.Base).Inline(true)
}

// line is the input to the text wrapping function. This is stored in a struct
// so that it can be hashed and memoized.
type line struct {
	runes []rune
	width int
}

// Hash returns a hash of the line.
func (w line) Hash() string {
	v := fmt.Sprintf("%s:%d", string(w.runes), w.width)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(v)))
}

// Model is the Bubble Tea model for this text area element.
type Model struct {
	Err error

	// General settings.
	// cache *memoization.MemoCache[line, [][]rune]

	// Prompt is printed at the beginning of each line.
	//
	// When changing the value of Prompt after the model has been
	// initialized, ensure that SetWidth() gets called afterwards.
	//
	// See also SetPromptFunc().
	Prompt string

	// Placeholder is the text displayed when the user
	// hasn't entered anything yet.
	Placeholder string

	// ShowLineNumbers, if enabled, causes line numbers to be printed
	// after the prompt.
	ShowLineNumbers bool

	// EndOfBufferCharacter is displayed at the end of the input.
	EndOfBufferCharacter rune

	// KeyMap encodes the keybindings recognized by the widget.
	KeyMap KeyMap

	// Styling. FocusedStyle and BlurredStyle are used to style the textarea in
	// focused and blurred states.
	Styles Styles

	// activeStyle is the current styling to use.
	// It is used to abstract the differences in focus state when styling the
	// model, since we can simply assign the set of activeStyle to this variable
	// when switching focus states.
	activeStyle *StyleState

	// Cursor is the text area cursor.
	Cursor cursor.Model

	// CharLimit is the maximum number of characters this input element will
	// accept. If 0 or less, there's no limit.
	CharLimit int

	// MaxHeight is the maximum height of the text area in rows. If 0 or less,
	// there's no limit.
	MaxHeight int

	// MaxWidth is the maximum width of the text area in columns. If 0 or less,
	// there's no limit.
	MaxWidth int

	// If promptFunc is set, it replaces Prompt as a generator for
	// prompt strings at the beginning of each line.
	promptFunc func(line int) string

	// promptWidth is the width of the prompt.
	promptWidth int

	// width is the maximum number of characters that can be displayed at once.
	// If 0 or less this setting is ignored.
	width int

	// height is the maximum number of lines that can be displayed at once. It
	// essentially treats the text field like a vertically scrolling viewport
	// if there are more lines than the permitted height.
	height int

	// Underlying text value.
	value Lines

	// focus indicates whether user input focus should be on this input
	// component. When false, ignore keyboard input and hide the cursor.
	focus bool

	// Cursor row and column.
	y, x int

	// The bubble offset relative to the parent bubble.
	offsetX, offsetY int

	// viewport is the vertically-scrollable viewport of the multi-line text
	// input.
	viewport *viewport.Model

	// rune sanitizer for input.
	rsan runeutil.Sanitizer
}

// New creates a new model with default settings.
func New() Model {
	vp := viewport.New()
	vp.KeyMap = viewport.KeyMap{}
	cur := cursor.New()

	styles := DefaultDarkStyles()

	m := Model{
		CharLimit:   defaultCharLimit,
		MaxHeight:   defaultMaxHeight,
		MaxWidth:    defaultMaxWidth,
		Prompt:      lipgloss.ThickBorder().Left + " ",
		Styles:      styles,
		activeStyle: &styles.Blurred,
		// cache:                memoization.NewMemoCache[line, [][]rune](defaultMaxHeight),
		EndOfBufferCharacter: ' ',
		ShowLineNumbers:      true,
		Cursor:               cur,
		KeyMap:               DefaultKeyMap(),

		value: make(Lines, minHeight, defaultMaxHeight),
		focus: false,
		x:     0,
		y:     0,

		viewport: &vp,
	}

	m.SetHeight(defaultHeight)
	m.SetWidth(defaultWidth)

	return m
}

// DefaultStyles returns the default styles for focused and blurred states for
// the textarea.
func DefaultStyles(isDark bool) Styles {
	lightDark := lipgloss.LightDark(isDark)

	var s Styles
	s.Focused = StyleState{
		Base:             lipgloss.NewStyle(),
		CursorLine:       lipgloss.NewStyle().Background(lightDark("255", "0")),
		CursorLineNumber: lipgloss.NewStyle().Foreground(lightDark("240", "240")),
		EndOfBuffer:      lipgloss.NewStyle().Foreground(lightDark("254", "0")),
		LineNumber:       lipgloss.NewStyle().Foreground(lightDark("249", "7")),
		Placeholder:      lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Prompt:           lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
		Text:             lipgloss.NewStyle(),
	}
	s.Blurred = StyleState{
		Base:             lipgloss.NewStyle(),
		CursorLine:       lipgloss.NewStyle().Foreground(lightDark("245", "7")),
		CursorLineNumber: lipgloss.NewStyle().Foreground(lightDark("249", "7")),
		EndOfBuffer:      lipgloss.NewStyle().Foreground(lightDark("254", "0")),
		LineNumber:       lipgloss.NewStyle().Foreground(lightDark("249", "7")),
		Placeholder:      lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Prompt:           lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
		Text:             lipgloss.NewStyle().Foreground(lightDark("245", "7")),
	}
	return s
}

// DefaultLightStyles returns the default styles for a light background.
func DefaultLightStyles() Styles {
	return DefaultStyles(false)
}

// DefaultDarkStyles returns the default styles for a dark background.
func DefaultDarkStyles() Styles {
	return DefaultStyles(true)
}

// SetValue sets the value of the text input.
func (m *Model) SetValue(s string) {
	m.Reset()
	m.InsertString(s)
}

// InsertString inserts a string at the cursor position.
func (m *Model) InsertString(s string) {
	m.insertRunesFromUserInput(s)
}

// InsertRune inserts a rune at the cursor position.
func (m *Model) InsertRune(r rune) {
	m.insertRunesFromUserInput(string(r))
}

// insertRunesFromUserInput inserts runes at the current cursor position.
func (m *Model) insertRunesFromUserInput(s string) {
	// TODO: Can we preserve these line breaks?
	s = strings.ReplaceAll(s, "\r\n", "\n") // normalize line endings
	for _, r := range s {
		if r == '\n' {
			m.value.Insert(m.y, Line{})
			m.x = 0
			m.y++
		} else {
			// We're only using WcWidth as of now. In case of Grapheme
			// clusters, we'd need to use the uniseg package, iterate over the
			// graphemes in the input string, and append/insert the cells with
			// the grapheme widths.
			// See [tea.EnableGraphemeClustering].
			c := Cell{Content: string(r), Width: wcwidth.RuneWidth(r)}
			lastCell := m.x == m.value[m.y].Width(m.width) && m.y == m.value[m.y].Height(m.width)-1
			if len(m.value[m.y]) == 0 || lastCell {
				m.value[m.y].Append(c)
			} else {
				m.value[m.y].Insert(m.x, c)
			}

			m.x += c.Width
			if m.x >= m.value[m.y].Width(m.width) {
				m.x = 0
				m.y++
			}
		}
	}

	// // Clean up any special characters in the input provided by the
	// // clipboard. This avoids bugs due to e.g. tab characters and
	// // whatnot.
	// runes = m.san().Sanitize(runes)
	//
	// var availSpace int
	// if m.CharLimit > 0 {
	// 	availSpace = m.CharLimit - m.Length()
	// 	// If the char limit's been reached, cancel.
	// 	if availSpace <= 0 {
	// 		return
	// 	}
	// 	// If there's not enough space to paste the whole thing cut the pasted
	// 	// runes down so they'll fit.
	// 	if availSpace < len(runes) {
	// 		runes = runes[:availSpace]
	// 	}
	// }
	//
	// // Split the input into lines.
	// var lines [][]rune
	// lstart := 0
	// for i := 0; i < len(runes); i++ {
	// 	if runes[i] == '\n' {
	// 		// Queue a line to become a new row in the text area below.
	// 		// Beware to clamp the max capacity of the slice, to ensure no
	// 		// data from different rows get overwritten when later edits
	// 		// will modify this line.
	// 		lines = append(lines, runes[lstart:i:i])
	// 		lstart = i + 1
	// 	}
	// }
	// if lstart <= len(runes) {
	// 	// The last line did not end with a newline character.
	// 	// Take it now.
	// 	lines = append(lines, runes[lstart:])
	// }

	// var availSpace int
	// if m.CharLimit > 0 {
	// 	availSpace = m.CharLimit - m.Length()
	// 	// If the char limit's been reached, cancel.
	// 	if availSpace <= 0 {
	// 		return
	// 	}
	// 	// If there's not enough space to paste the whole thing cut the pasted
	// 	// runes down so they'll fit.
	// 	if availSpace < len(runes) {
	// 		runes = runes[:availSpace]
	// 	}
	// }
	//
	// // Split the input into lines.
	// var lines [][]rune
	// lstart := 0
	// for i := 0; i < len(runes); i++ {
	// 	if runes[i] == '\n' {
	// 		// Queue a line to become a new row in the text area below.
	// 		// Beware to clamp the max capacity of the slice, to ensure no
	// 		// data from different rows get overwritten when later edits
	// 		// will modify this line.
	// 		lines = append(lines, runes[lstart:i:i])
	// 		lstart = i + 1
	// 	}
	// }
	// if lstart <= len(runes) {
	// 	// The last line did not end with a newline character.
	// 	// Take it now.
	// 	lines = append(lines, runes[lstart:])
	// }
	//
	// // Obey the maximum height limit.
	// if m.MaxHeight > 0 && len(m.value)+len(lines)-1 > m.MaxHeight {
	// 	allowedHeight := max(0, m.MaxHeight-len(m.value)+1)
	// 	lines = lines[:allowedHeight]
	// }
	//
	// if len(lines) == 0 {
	// 	// Nothing left to insert.
	// 	return
	// }
	//
	// // Save the remainder of the original line at the current
	// // cursor position.
	// tail := make([]rune, len(m.value[m.row][m.col:]))
	// copy(tail, m.value[m.row][m.col:])
	//
	// // Paste the first line at the current cursor position.
	// m.value[m.row] = append(m.value[m.row][:m.col], lines[0]...)
	// m.col += len(lines[0])
	//
	// if numExtraLines := len(lines) - 1; numExtraLines > 0 {
	// 	// Add the new lines.
	// 	// We try to reuse the slice if there's already space.
	// 	var newGrid [][]rune
	// 	if cap(m.value) >= len(m.value)+numExtraLines {
	// 		// Can reuse the extra space.
	// 		newGrid = m.value[:len(m.value)+numExtraLines]
	// 	} else {
	// 		// No space left; need a new slice.
	// 		newGrid = make([][]rune, len(m.value)+numExtraLines)
	// 		copy(newGrid, m.value[:m.row+1])
	// 	}
	// 	// Add all the rows that were after the cursor in the original
	// 	// grid at the end of the new grid.
	// 	copy(newGrid[m.row+1+numExtraLines:], m.value[m.row+1:])
	// 	m.value = newGrid
	// 	// Insert all the new lines in the middle.
	// 	for _, l := range lines[1:] {
	// 		m.row++
	// 		m.value[m.row] = l
	// 		m.col = len(l)
	// 	}
	// }
	//
	// // Finally add the tail at the end of the last line inserted.
	// m.value[m.row] = append(m.value[m.row], tail...)
	//
	// m.SetCursor(m.col)
}

// Value returns the value of the text input.
func (m Model) Value() string {
	if m.value == nil {
		return ""
	}

	return m.value.String()
}

// Length returns the number of characters currently in the text input.
func (m *Model) Length() (n int) {
	for _, l := range m.value {
		n += l.RuneLen()
	}
	return
}

// LineCount returns the number of lines that are currently in the text input.
func (m *Model) LineCount() int {
	return len(m.value)
}

// Line returns the line position.
func (m Model) Line() int {
	return m.y
}

// CursorDown moves the cursor down by one line.
func (m *Model) CursorDown() {
	if m.y < len(m.value)-1 {
		m.y++
		line := m.value[m.y]
		if m.x > line.Width(m.width) {
			m.x = line.Width(m.width)
		}
	}
}

// CursorUp moves the cursor up by one line.
func (m *Model) CursorUp() {
	if m.y > 0 {
		m.y--
		line := m.value[m.y]
		if m.x > line.Width(m.width) {
			m.x = line.Width(m.width)
		}
	}
}

// SetCursor moves the cursor to the given position. If the position is
// out of bounds the cursor will be moved to the start or end accordingly.
func (m *Model) SetCursor(col int) {
	m.x = clamp(col, 0, m.value[m.y].Width(m.width))
}

// CursorStart moves the cursor to the start of the input field.
func (m *Model) CursorStart() {
	m.SetCursor(0)
}

// CursorEnd moves the cursor to the end of the input field.
func (m *Model) CursorEnd() {
	m.SetCursor(m.value[m.y].Width(m.width))
}

// Focused returns the focus state on the model.
func (m Model) Focused() bool {
	return m.focus
}

// Focus sets the focus state on the model. When the model is in focus it can
// receive keyboard input and the cursor will be hidden.
func (m *Model) Focus() tea.Cmd {
	m.focus = true
	m.activeStyle = &m.Styles.Focused
	return m.Cursor.Focus()
}

// Blur removes the focus state on the model. When the model is blurred it can
// not receive keyboard input and the cursor will be hidden.
func (m *Model) Blur() {
	m.focus = false
	m.activeStyle = &m.Styles.Blurred
	m.Cursor.Blur()
}

// Reset sets the input to its default state with no input.
func (m *Model) Reset() {
	startCap := m.MaxHeight
	if startCap <= 0 {
		startCap = defaultMaxHeight
	}
	m.value = make([]Line, minHeight, startCap)
	m.x = 0
	m.y = 0
	m.viewport.GotoTop()
	m.SetCursor(0)
}

// deleteBeforeCursor deletes all text before the cursor. Returns whether or
// not the cursor blink should be reset.
func (m *Model) deleteBeforeCursor() {
	// m.value[m.y] = m.value[m.y][m.col:]
	line := m.value[m.y]
	line.Delete(0, m.x)
	m.value[m.y] = line
	m.SetCursor(0)
}

// deleteAfterCursor deletes all text after the cursor. Returns whether or not
// the cursor blink should be reset. If input is masked delete everything after
// the cursor so as not to reveal word breaks in the masked input.
func (m *Model) deleteAfterCursor() {
	// m.value[m.y] = m.value[m.y][:m.col]
	line := m.value[m.y]
	line.Delete(m.x, line.Width(m.width))
	m.value[m.y] = line
	m.SetCursor(len(line))
}

// transposeLeft exchanges the runes at the cursor and immediately
// before. No-op if the cursor is at the beginning of the line.  If
// the cursor is not at the end of the line yet, moves the cursor to
// the right.
func (m *Model) transposeLeft() {
	if m.x == 0 || len(m.value[m.y]) < 2 {
		return
	}
	if m.x >= len(m.value[m.y]) {
		m.SetCursor(m.x - 1)
	}
	m.value[m.y][m.x-1], m.value[m.y][m.x] = m.value[m.y][m.x], m.value[m.y][m.x-1]
	if m.x < len(m.value[m.y]) {
		m.SetCursor(m.x + 1)
	}
}

// deleteWordLeft deletes the word left to the cursor.
func (m *Model) deleteWordLeft() {
	if m.x == 0 || len(m.value[m.y]) == 0 {
		return
	}

	x := m.x - 1
	for x >= 0 {
		c, _ := m.value[m.y].At(x)
		if strings.Trim(c.Content, " ") == "" {
			break
		}
		x -= c.Width
	}

	if x < 0 {
		x = 0
	}

	m.value[m.y].Delete(x, m.x)
	m.x = x
}

// deleteWordRight deletes the word right to the cursor.
func (m *Model) deleteWordRight() {
	if m.x >= len(m.value[m.y]) || len(m.value[m.y]) == 0 {
		return
	}

	x := m.x
	for x < m.value[m.y].Width(m.width) {
		// Ignore series of whitespace after cursor.
		c, _ := m.value[m.y].At(x)
		if strings.Trim(c.Content, " ") != "" {
			break
		}
		x += c.Width
	}

	for x < m.value[m.y].Width(m.width) {
		c, _ := m.value[m.y].At(x)
		if strings.Trim(c.Content, " ") == "" {
			break
		}
		x += c.Width
	}

	if x > m.value[m.y].Width(m.width) {
		x = m.value[m.y].Width(m.width)
	}

	m.value[m.y].Delete(m.x, x)
}

// characterRight moves the cursor one character to the right.
func (m *Model) characterRight() {
	if line := m.value[m.y]; m.x < line.Width(m.width) {
		c, _ := line.At(m.x)
		m.SetCursor(m.x + c.Width)
	} else if m.y < len(m.value)-1 {
		m.y++
		m.CursorStart()
	}
}

// characterLeft moves the cursor one character to the left.
// If insideLine is set, the cursor is moved to the last
// character in the previous line, instead of one past that.
func (m *Model) characterLeft(insideLine bool) {
	if m.x == 0 && m.y != 0 {
		m.y--
		m.CursorEnd()
		if !insideLine {
			return
		}
	}
	if line := m.value[m.y]; m.x > 0 {
		c, _ := line.At(m.x - 1)
		m.SetCursor(m.x - c.Width)
	}
}

// wordLeft moves the cursor one word to the left. If input is masked, move
// input to the start so as not to reveal word breaks in the masked input.
func (m *Model) wordLeft() {
	for m.x > 0 {
		m.characterLeft(true /* insideLine */)
		c, _ := m.value[m.y].At(m.x)
		if m.x > 0 && strings.Trim(c.Content, " ") != "" {
			break
		}
	}

	for m.x > 0 {
		c, _ := m.value[m.y].At(m.x - 1)
		if strings.Trim(c.Content, " ") == "" {
			break
		}
		m.SetCursor(m.x - c.Width)
	}

	// for {
	// 	m.characterLeft(true /* insideLine */)
	// 	if m.x < len(m.value[m.y]) && !unicode.IsSpace(m.value[m.y][m.x]) {
	// 		break
	// 	}
	// }
	//
	// for m.x > 0 {
	// 	if unicode.IsSpace(m.value[m.y][m.x-1]) {
	// 		break
	// 	}
	// 	m.SetCursor(m.x - 1)
	// }
}

// wordRight moves the cursor one word to the right. Returns whether or not the
// cursor blink should be reset. If the input is masked, move input to the end
// so as not to reveal word breaks in the masked input.
func (m *Model) wordRight() {
	m.doWordRight(func(_ int, c Cell) Cell { return c })
}

func (m *Model) doWordRight(fn func(i int, c Cell) Cell) {
	// Skip spaces forward.
	c, _ := m.value[m.y].At(m.x)
	for m.x >= m.value[m.y].Width(m.width) || strings.Trim(c.Content, " ") == "" {
		if m.y == len(m.value)-1 && m.x == m.value[m.y].Width(m.width) {
			// End of text.
			break
		}
		m.characterRight()
		c, _ = m.value[m.y].At(m.x)
	}

	cellIdx := 0
	for m.x < m.value[m.y].Width(m.width) {
		c, i := m.value[m.y].At(m.x)
		if strings.Trim(c.Content, " ") == "" {
			m.SetCursor(m.x + c.Width)
			break
		}
		c = fn(cellIdx, c)
		m.value[m.y][i] = c
		m.SetCursor(m.x + c.Width)
		cellIdx++
	}
}

// uppercaseRight changes the word to the right to uppercase.
func (m *Model) uppercaseRight() {
	m.doWordRight(func(_ int, c Cell) Cell {
		c.Content = strings.ToUpper(c.Content)
		return c
	})
}

// lowercaseRight changes the word to the right to lowercase.
func (m *Model) lowercaseRight() {
	m.doWordRight(func(_ int, c Cell) Cell {
		c.Content = strings.ToLower(c.Content)
		return c
	})
}

// capitalizeRight changes the word to the right to title case.
func (m *Model) capitalizeRight() {
	m.doWordRight(func(i int, c Cell) Cell {
		if i == 0 {
			c.Content = strings.ToUpper(c.Content)
		}
		return c
	})
}

// LineInfo returns the number of characters from the start of the
// (soft-wrapped) line and the (soft-wrapped) line width.
func (m Model) LineInfo() LineInfo {
	// grid := m.memoizedWrap(m.value[m.y], m.width)
	grid := m.value

	// Find out which line we are currently on. This can be determined by the
	// m.col and counting the number of runes that we need to skip.
	var counter int
	for i, line := range grid {
		// We've found the line that we are on
		if counter+len(line) == m.x && i+1 < len(grid) {
			// We wrap around to the next line if we are at the end of the
			// previous line so that we can be at the very beginning of the row
			return LineInfo{
				CharOffset:   0,
				ColumnOffset: 0,
				Height:       len(grid),
				RowOffset:    i + 1,
				StartColumn:  m.x,
				Width:        len(grid[i+1]),
				CharWidth:    line.Width(m.width),
			}
		}

		if counter+len(line) >= m.x {
			return LineInfo{
				CharOffset:   uniseg.StringWidth(line.String()[:max(0, m.x-counter)]),
				ColumnOffset: m.x - counter,
				Height:       len(grid),
				RowOffset:    i,
				StartColumn:  counter,
				Width:        len(line),
				CharWidth:    line.Width(m.width),
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
	max := min + m.viewport.Height() - 1

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

// moveToBegin moves the cursor to the beginning of the input.
func (m *Model) moveToBegin() {
	m.y = 0
	m.SetCursor(0)
}

// moveToEnd moves the cursor to the end of the input.
func (m *Model) moveToEnd() {
	m.y = len(m.value) - 1
	m.SetCursor(len(m.value[m.y]))
}

// SetWidth sets the width of the textarea to fit exactly within the given width.
// This means that the textarea will account for the width of the prompt and
// whether or not line numbers are being shown.
//
// Ensure that SetWidth is called after setting the Prompt and ShowLineNumbers,
// It is important that the width of the textarea be exactly the given width
// and no more.
func (m *Model) SetWidth(w int) {
	// Update prompt width only if there is no prompt function as SetPromptFunc
	// updates the prompt width when it is called.
	if m.promptFunc == nil {
		m.promptWidth = uniseg.StringWidth(m.Prompt)
	}

	// Add base style borders and padding to reserved outer width.
	reservedOuter := m.activeStyle.Base.GetHorizontalFrameSize()

	// Add prompt width to reserved inner width.
	reservedInner := m.promptWidth

	// Add line number width to reserved inner width.
	if m.ShowLineNumbers {
		const lnWidth = 4 // Up to 3 digits for line number plus 1 margin.
		reservedInner += lnWidth
	}

	// Input width must be at least one more than the reserved inner and outer
	// width. This gives us a minimum input width of 1.
	minWidth := reservedInner + reservedOuter + 1
	inputWidth := max(w, minWidth)

	// Input width must be no more than maximum width.
	if m.MaxWidth > 0 {
		inputWidth = min(inputWidth, m.MaxWidth)
	}

	// Since the width of the viewport and input area is dependent on the width of
	// borders, prompt and line numbers, we need to calculate it by subtracting
	// the reserved width from them.

	m.viewport.SetWidth(inputWidth - reservedOuter)
	m.width = inputWidth - reservedOuter - reservedInner
}

// SetPromptFunc supersedes the Prompt field and sets a dynamic prompt
// instead.
// If the function returns a prompt that is shorter than the
// specified promptWidth, it will be padded to the left.
// If it returns a prompt that is longer, display artifacts
// may occur; the caller is responsible for computing an adequate
// promptWidth.
func (m *Model) SetPromptFunc(promptWidth int, fn func(lineIdx int) string) {
	m.promptFunc = fn
	m.promptWidth = promptWidth
}

// Height returns the current height of the textarea.
func (m Model) Height() int {
	return m.height
}

// SetHeight sets the height of the textarea.
func (m *Model) SetHeight(h int) {
	if m.MaxHeight > 0 {
		m.height = clamp(h, minHeight, m.MaxHeight)
		m.viewport.SetHeight(clamp(h, minHeight, m.MaxHeight))
	} else {
		m.height = max(h, minHeight)
		m.viewport.SetHeight(max(h, minHeight))
	}
}

// Update is the Bubble Tea update loop.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focus {
		m.Cursor.Blur()
		return m, nil
	}

	var cmds []tea.Cmd

	if m.y < len(m.value) && m.value[m.y] == nil {
		m.value[m.y] = make(Line, 0)
	}

	// if m.MaxHeight > 0 && m.MaxHeight != m.cache.Capacity() {
	// 	m.cache = memoization.NewMemoCache[line, [][]rune](m.MaxHeight)
	// }

	switch msg := msg.(type) {
	case tea.PasteMsg:
		m.insertRunesFromUserInput(string(msg))
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.KeyMap.DeleteAfterCursor):
			m.x = clamp(m.x, 0, m.value[m.y].Width(m.width))
			if m.x >= len(m.value[m.y]) {
				m.mergeLineBelow(m.y)
				break
			}
			m.deleteAfterCursor()
		case key.Matches(msg, m.KeyMap.DeleteBeforeCursor):
			m.x = clamp(m.x, 0, m.value[m.y].Width(m.width))
			if m.x <= 0 {
				m.mergeLineAbove(m.y)
				break
			}
			m.deleteBeforeCursor()
		case key.Matches(msg, m.KeyMap.DeleteCharacterBackward):
			m.x = clamp(m.x, 0, m.value[m.y].Width(m.width))
			if m.x <= 0 {
				m.mergeLineAbove(m.y)
				break
			}
			if len(m.value[m.y]) > 0 {
				// m.value[m.y] = append(m.value[m.y][:max(0, m.col-1)], m.value[m.y][m.col:]...)
				c, _ := m.value[m.y].At(m.x - 1)
				m.value[m.y].Delete(m.x-c.Width, m.x)
				if m.x > 0 {
					m.SetCursor(m.x - c.Width)
				}
			}
		case key.Matches(msg, m.KeyMap.DeleteCharacterForward):
			if len(m.value[m.y]) > 0 && m.x < len(m.value[m.y]) {
				// m.value[m.y] = append(m.value[m.y][:m.col], m.value[m.y][m.col+1:]...)
				m.value[m.y].Delete(m.x, m.x+1)
			}
			if m.x >= len(m.value[m.y]) {
				m.mergeLineBelow(m.y)
				break
			}
		case key.Matches(msg, m.KeyMap.DeleteWordBackward):
			if m.x <= 0 {
				m.mergeLineAbove(m.y)
				break
			}
			m.deleteWordLeft()
		case key.Matches(msg, m.KeyMap.DeleteWordForward):
			m.x = clamp(m.x, 0, len(m.value[m.y]))
			if m.x >= len(m.value[m.y]) {
				m.mergeLineBelow(m.y)
				break
			}
			m.deleteWordRight()
		case key.Matches(msg, m.KeyMap.InsertNewline):
			if m.MaxHeight > 0 && m.value.Len() >= m.MaxHeight {
				return m, nil
			}
			m.x = clamp(m.x, 0, m.value[m.y].Width(m.width))
			m.splitLine(m.y, m.x)
		case key.Matches(msg, m.KeyMap.LineEnd):
			m.CursorEnd()
		case key.Matches(msg, m.KeyMap.LineStart):
			m.CursorStart()
		case key.Matches(msg, m.KeyMap.CharacterForward):
			m.characterRight()
		case key.Matches(msg, m.KeyMap.LineNext):
			m.CursorDown()
		case key.Matches(msg, m.KeyMap.WordForward):
			m.wordRight()
		case key.Matches(msg, m.KeyMap.Paste):
			return m, Paste
		case key.Matches(msg, m.KeyMap.CharacterBackward):
			m.characterLeft(false /* insideLine */)
		case key.Matches(msg, m.KeyMap.LinePrevious):
			m.CursorUp()
		case key.Matches(msg, m.KeyMap.WordBackward):
			m.wordLeft()
		case key.Matches(msg, m.KeyMap.InputBegin):
			m.moveToBegin()
		case key.Matches(msg, m.KeyMap.InputEnd):
			m.moveToEnd()
		case key.Matches(msg, m.KeyMap.LowercaseWordForward):
			m.lowercaseRight()
		case key.Matches(msg, m.KeyMap.UppercaseWordForward):
			m.uppercaseRight()
		case key.Matches(msg, m.KeyMap.CapitalizeWordForward):
			m.capitalizeRight()
		case key.Matches(msg, m.KeyMap.TransposeCharacterBackward):
			m.transposeLeft()

		default:
			m.insertRunesFromUserInput(msg.Text)
		}

	case pasteMsg:
		m.insertRunesFromUserInput(string(msg))

	case pasteErrMsg:
		m.Err = msg
	}

	vp, cmd := m.viewport.Update(msg)
	m.viewport = &vp
	cmds = append(cmds, cmd)

	newCol, newRow := m.cursor()
	cmds = append(cmds, tea.SetCursorPosition(m.offsetX+newCol, m.offsetY+newRow))
	m.repositionView()

	return m, tea.Batch(cmds...)
}

// View renders the text area in its current state.
func (m Model) View() string {
	if m.value.Len() == 0 && m.y == 0 && m.x == 0 && m.Placeholder != "" {
		return m.placeholderView()
	}

	var (
		b     bytes.Buffer
		style lipgloss.Style
	)
	for j, line := range m.value {
		if m.y == j {
			style = m.activeStyle.computedCursorLine()
		} else {
			style = m.activeStyle.computedText()
		}

		// TODO: Support tabs.
		var lineWidth int
		var lineStr string
		for _, cell := range line {
			if lineWidth >= m.width {
				lineStr += "\n"
				lineWidth -= m.width
			}

			lineWidth += cell.Width
			lineStr += cell.Content
		}

		if lineWidth == m.width {
			lineStr += "\n"
			lineWidth = 0
		}

		if lineWidth < m.width {
			lineStr += strings.Repeat(" ", m.width-lineWidth)
		}

		lineStrs := strings.Split(lineStr, "\n")

		for i, l := range lineStrs {
			prompt := m.getPromptString(j)
			prompt = m.activeStyle.computedPrompt().Render(prompt)
			b.WriteString(style.Render(prompt))
			if m.ShowLineNumbers {
				if i == 0 {
					if m.y == j {
						b.WriteString(style.Render(m.activeStyle.computedCursorLineNumber().Render(m.formatLineNumber(j + 1))))
					} else {
						b.WriteString(style.Render(m.activeStyle.computedLineNumber().Render(m.formatLineNumber(j + 1))))
					}
				} else {
					if m.y == j {
						b.WriteString(style.Render(m.activeStyle.computedCursorLineNumber().Render(m.formatLineNumber(" "))))
					} else {
						b.WriteString(style.Render(m.activeStyle.computedLineNumber().Render(m.formatLineNumber(" "))))
					}
				}
			}

			b.WriteString(style.Render(l))
			if i < len(lineStrs)-1 {
				b.WriteRune('\n')
			}
		}
	}

	m.viewport.SetContent(b.String())

	return m.activeStyle.Base.Render(m.viewport.View())

	var (
		s strings.Builder
		// style            lipgloss.Style
		newLines         int
		widestLineNumber int
		lineInfo         = m.LineInfo()
	)

	displayLine := 0
	for l, line := range m.value {
		// wrappedLines := m.memoizedWrap(line, m.width)
		wrappedLines := m.value

		if m.y == l {
			style = m.activeStyle.computedCursorLine()
		} else {
			style = m.activeStyle.computedText()
		}

		for wl, wrappedLine := range wrappedLines {
			prompt := m.getPromptString(displayLine)
			prompt = m.activeStyle.computedPrompt().Render(prompt)
			s.WriteString(style.Render(prompt))
			displayLine++

			var ln string
			if m.ShowLineNumbers { //nolint:nestif
				if wl == 0 {
					if m.y == l {
						ln = style.Render(m.activeStyle.computedCursorLineNumber().Render(m.formatLineNumber(l + 1)))
						s.WriteString(ln)
					} else {
						ln = style.Render(m.activeStyle.computedLineNumber().Render(m.formatLineNumber(l + 1)))
						s.WriteString(ln)
					}
				} else {
					if m.y == l {
						ln = style.Render(m.activeStyle.computedCursorLineNumber().Render(m.formatLineNumber(" ")))
						s.WriteString(ln)
					} else {
						ln = style.Render(m.activeStyle.computedLineNumber().Render(m.formatLineNumber(" ")))
						s.WriteString(ln)
					}
				}
			}

			// Note the widest line number for padding purposes later.
			lnw := lipgloss.Width(ln)
			if lnw > widestLineNumber {
				widestLineNumber = lnw
			}

			strwidth := uniseg.StringWidth(wrappedLine.String())
			padding := m.width - strwidth
			// If the trailing space causes the line to be wider than the
			// width, we should not draw it to the screen since it will result
			// in an extra space at the end of the line which can look off when
			// the cursor line is showing.
			if strwidth > m.width {
				// The character causing the line to be wider than the width is
				// guaranteed to be a space since any other character would
				// have been wrapped.
				wrappedLine.TrimSuffix(" ")
				padding -= m.width - strwidth
			}
			if m.y == l && lineInfo.RowOffset == wl && m.Cursor.Mode() != cursor.CursorHide {
				s.WriteString(style.Render(wrappedLine[:lineInfo.ColumnOffset].String()))
				if m.x >= len(line) && lineInfo.CharOffset >= m.width {
					m.Cursor.SetChar(" ")
					s.WriteString(m.Cursor.View())
				} else {
					m.Cursor.SetChar(wrappedLine[lineInfo.ColumnOffset].Content)
					s.WriteString(style.Render(m.Cursor.View()))
					s.WriteString(style.Render(wrappedLine[lineInfo.ColumnOffset+1:].String()))
				}
			} else {
				s.WriteString(style.Render(wrappedLine.String()))
			}
			s.WriteString(style.Render(strings.Repeat(" ", max(0, padding))))
			s.WriteRune('\n')
			newLines++
		}
	}

	// Always show at least `m.Height` lines at all times.
	// To do this we can simply pad out a few extra new lines in the view.
	for i := 0; i < m.height; i++ {
		prompt := m.getPromptString(displayLine)
		prompt = m.activeStyle.computedPrompt().Render(prompt)
		s.WriteString(prompt)
		displayLine++

		// Write end of buffer content
		leftGutter := string(m.EndOfBufferCharacter)
		rightGapWidth := m.Width() - lipgloss.Width(leftGutter) + widestLineNumber
		rightGap := strings.Repeat(" ", max(0, rightGapWidth))
		s.WriteString(m.activeStyle.computedEndOfBuffer().Render(leftGutter + rightGap))
		s.WriteRune('\n')
	}

	m.viewport.SetContent(s.String())
	return m.activeStyle.Base.Render(m.viewport.View())
}

// formatLineNumber formats the line number for display dynamically based on
// the maximum number of lines.
func (m Model) formatLineNumber(x any) string {
	// XXX: ultimately we should use a max buffer height, which has yet to be
	// implemented.
	digits := len(strconv.Itoa(m.MaxHeight))
	return fmt.Sprintf(" %*v ", digits, x)
}

func (m Model) getPromptString(displayLine int) (prompt string) {
	prompt = m.Prompt
	if m.promptFunc == nil {
		return prompt
	}
	prompt = m.promptFunc(displayLine)
	pl := uniseg.StringWidth(prompt)
	if pl < m.promptWidth {
		prompt = fmt.Sprintf("%*s%s", m.promptWidth-pl, "", prompt)
	}
	return prompt
}

// placeholderView returns the prompt and placeholder view, if any.
func (m Model) placeholderView() string {
	var (
		s     strings.Builder
		p     = m.Placeholder
		style = m.activeStyle.computedPlaceholder()
	)

	// word wrap lines
	pwordwrap := ansi.Wordwrap(p, m.width, "")
	// wrap lines (handles lines that could not be word wrapped)
	pwrap := ansi.Hardwrap(pwordwrap, m.width, true)
	// split string by new lines
	plines := strings.Split(strings.TrimSpace(pwrap), "\n")

	for i := 0; i < m.height; i++ {
		lineStyle := m.activeStyle.computedPlaceholder()
		lineNumberStyle := m.activeStyle.computedLineNumber()
		if len(plines) > i {
			lineStyle = m.activeStyle.computedCursorLine()
			lineNumberStyle = m.activeStyle.computedCursorLineNumber()
		}

		// render prompt
		prompt := m.getPromptString(i)
		prompt = m.activeStyle.computedPrompt().Render(prompt)
		s.WriteString(lineStyle.Render(prompt))

		// when show line numbers enabled:
		// - render line number for only the cursor line
		// - indent other placeholder lines
		// this is consistent with vim with line numbers enabled
		if m.ShowLineNumbers {
			var ln string

			switch {
			case i == 0:
				ln = strconv.Itoa(i + 1)
				fallthrough
			case len(plines) > i:
				s.WriteString(lineStyle.Render(lineNumberStyle.Render(m.formatLineNumber(ln))))
			default:
			}
		}

		switch {
		// first line
		case i == 0:
			// first character of first line as cursor with character
			m.Cursor.TextStyle = m.activeStyle.computedPlaceholder()
			m.Cursor.SetChar(string(plines[0][0]))
			s.WriteString(lineStyle.Render(m.Cursor.View()))

			// the rest of the first line
			s.WriteString(lineStyle.Render(style.Render(plines[0][1:] + strings.Repeat(" ", max(0, m.width-uniseg.StringWidth(plines[0]))))))
		// remaining lines
		case len(plines) > i:
			// current line placeholder text
			if len(plines) > i {
				s.WriteString(lineStyle.Render(style.Render(plines[i] + strings.Repeat(" ", max(0, m.width-uniseg.StringWidth(plines[i]))))))
			}
		default:
			// end of line buffer character
			eob := m.activeStyle.computedEndOfBuffer().Render(string(m.EndOfBufferCharacter))
			s.WriteString(eob)
		}

		// terminate with new line
		s.WriteRune('\n')
	}

	m.viewport.SetContent(s.String())
	return m.activeStyle.Base.Render(m.viewport.View())
}

// Blink returns the blink command for the cursor.
func Blink() tea.Msg {
	return cursor.Blink()
}

// func (m Model) memoizedWrap(runes []rune, width int) [][]rune {
// 	input := line{runes: runes, width: width}
// 	if v, ok := m.cache.Get(input); ok {
// 		return v
// 	}
// 	v := wrap(runes, width)
// 	m.cache.Set(input, v)
// 	return v
// }

// cursorLineNumber returns the line number that the cursor is on.
// This accounts for soft wrapped lines.
// TODO: remove
func (m Model) cursorLineNumber() int {
	row := m.y
	for _, line := range m.value {
		row += line.Width(m.width) / m.width
	}
	// for i := 0; i < m.y; i++ {
	// 	// Calculate the number of lines that the current line will be split
	// 	// into.
	// 	// line += len(m.memoizedWrap(m.value[i], m.width))
	// 	row += m.value[i].Width() / m.width
	// }
	// line += m.LineInfo().RowOffset
	return row
}

// cursor returns the cursor position. This accounts for soft wrapped lines.
func (m Model) cursor() (x, y int) {
	col := m.x
	row := m.y
	for y, line := range m.value {
		offset := line.Width(m.width) / m.width
		if m.y == y {
			col = col % m.width
		}
		row += offset
	}
	return col, row - m.viewport.YOffset
}

// mergeLineBelow merges the current line the cursor is on with the line below.
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

	m.x = len(m.value[row-1])
	m.y = m.y - 1

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
	head, tail := m.value[row].Split(col)
	m.value[row] = head
	m.value.Insert(row+1, tail)
	m.x = 0
	if m.y >= m.viewport.Height() {
		m.viewport.YOffset++
	}
	m.y++
}

// SetOffset sets the offset of the viewport.
func (m *Model) SetOffset(x, y int) {
	m.offsetX, m.offsetY = x, y
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

		if spaces > 0 { //nolint:nestif
			if uniseg.StringWidth(string(lines[row]))+uniseg.StringWidth(string(word))+spaces > width {
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
			if uniseg.StringWidth(string(word))+lastCharLen > width {
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

	if uniseg.StringWidth(string(lines[row]))+uniseg.StringWidth(string(word))+spaces >= width {
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
