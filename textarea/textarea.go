package textarea

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	rw "github.com/mattn/go-runewidth"
)

const defaultBlinkSpeed = time.Millisecond * 530

const (
	minHeight        = 1
	minWidth         = 2
	defaultHeight    = 6
	defaultWidth     = 40
	defaultCharLimit = 400
	maxHeight        = 30
	maxWidth         = 500
)

// Internal ID management for text inputs. Necessary for blink integrity when
// multiple text inputs are involved.
var (
	lastID int
	idMtx  sync.Mutex
)

// Return the next ID we should use on the Model.
func nextID() int {
	idMtx.Lock()
	defer idMtx.Unlock()
	lastID++
	return lastID
}

// initialBlinkMsg initializes cursor blinking.
type initialBlinkMsg struct{}

// blinkMsg signals that the cursor should blink. It contains metadata that
// allows us to tell if the blink message is the one we're expecting.
type blinkMsg struct {
	id  int
	tag int
}

// blinkCanceled is sent when a blink operation is canceled.
type blinkCanceled struct{}

// Internal messages for clipboard operations.
type pasteMsg string
type pasteErrMsg struct{ error }

// EchoMode sets the input behavior of the text input field.
type EchoMode int

const (
	// EchoNormal displays text as is. This is the default behavior.
	EchoNormal EchoMode = iota

	// EchoPassword displays the EchoCharacter mask instead of actual
	// characters.  This is commonly used for password fields.
	EchoPassword

	// EchoNone displays nothing as characters are entered. This is commonly
	// seen for password fields on the command line.
	EchoNone

	// EchoOnEdit.
)

// blinkCtx manages cursor blinking.
type blinkCtx struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// CursorMode describes the behavior of the cursor.
type CursorMode int

// Available cursor modes.
const (
	CursorBlink CursorMode = iota
	CursorStatic
	CursorHide
)

// String returns the cursor mode in a human-readable format. This method is
// provisional and for informational purposes only.
func (c CursorMode) String() string {
	return [...]string{
		"blink",
		"static",
		"hidden",
	}[c]
}

// SoftLineInfo is a helper for keeping track of line information regarding
// soft-wrapped lines.
type SoftLineInfo struct {
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

// Model is the Bubble Tea model for this text input element.
type Model struct {
	Err error

	// General settings.
	Prompt          string
	Placeholder     string
	BlinkSpeed      time.Duration
	EchoMode        EchoMode
	EchoCharacter   rune
	ShowLineNumbers bool

	// Styles. These will be applied as inline styles.
	//
	// For an introduction to styling with Lip Gloss see:
	// https://github.com/charmbracelet/lipgloss
	PromptStyle      lipgloss.Style
	TextStyle        lipgloss.Style
	BackgroundStyle  lipgloss.Style
	PlaceholderStyle lipgloss.Style
	CursorStyle      lipgloss.Style
	LineNumberStyle  lipgloss.Style
	CursorLineStyle  lipgloss.Style

	// CharLimit is the maximum number of characters this input element will
	// accept. If 0 or less, there's no limit.
	CharLimit int

	// Width is the maximum number of characters that can be displayed at once.
	// It essentially treats the text field like a horizontally scrolling
	// viewport. If 0 or less this setting is ignored.
	Width int

	// Height is the maximum number of lines that can be displayed at once.
	// It essentially treats the text field like a vertically scrolling viewport
	// if there are more lines that permitted height.
	Height int

	// The ID of this Model as it relates to other text area Models.
	id int

	// The ID of the blink message we're expecting to receive.
	blinkTag int

	// Underlying text value.
	value [][]rune

	// focus indicates whether user input focus should be on this input
	// component. When false, ignore keyboard input and hide the cursor.
	focus bool

	// Cursor blink state.
	blink bool

	// Cursor column.
	col int

	// Cursor row.
	row int

	// Last character offset, used to maintain state when the cursor is moved
	// vertically such that we can maintain the same navigating position.
	lastCharOffset int

	// Used to manage cursor blink
	blinkCtx *blinkCtx

	// cursorMode determines the behavior of the cursor
	cursorMode CursorMode

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

	return Model{
		BlinkSpeed:       defaultBlinkSpeed,
		CharLimit:        defaultCharLimit,
		Height:           defaultHeight,
		Width:            defaultWidth,
		EchoCharacter:    '*',
		Prompt:           "│ ",
		PromptStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("230")),
		LineNumberStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		PlaceholderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		ShowLineNumbers:  false,

		id:               nextID(),
		value:            make([][]rune, minHeight, maxWidth),
		focus:            false,
		blink:            true,
		col:              0,
		row:              0,
		cursorMode:       CursorBlink,
		lineNumberFormat: "%2d ",

		blinkCtx: &blinkCtx{
			ctx: context.Background(),
		},

		viewport: &vp,
	}
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

// Cursor returns the cursor row.
func (m Model) Cursor() int {
	return m.col
}

// Line returns the line position.
func (m Model) Line() int {
	return m.row
}

// Blink returns whether or not to draw the cursor.
func (m Model) Blink() bool {
	return m.blink
}

// SetCursor moves the cursor to the given position. If the position is
// out of bounds the cursor will be moved to the start or end accordingly.
func (m *Model) SetCursor(col int) {
	m.setCursor(col)
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

	var offset = 0
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

	var offset = 0
	for offset < charOffset {
		if m.col >= len(m.value[m.row]) || offset >= nli.CharWidth-1 {
			break
		}
		offset += rw.RuneWidth(m.value[m.row][m.col])
		m.col++
	}
}

// setCursor moves the cursor to the given position and returns whether or not
// the cursor blink should be reset. If the position is out of bounds the
// cursor will be moved to the start or end accordingly.
func (m *Model) setCursor(col int) bool {
	m.col = clamp(col, 0, len(m.value[m.row]))

	// Show the cursor unless it's been explicitly hidden
	m.blink = m.cursorMode == CursorHide

	// Reset cursor blink if necessary
	return m.cursorMode == CursorBlink
}

// CursorStart moves the cursor to the start of the input field.
func (m *Model) CursorStart() {
	m.cursorStart()
}

// cursorStart moves the cursor to the start of the input field and returns
// whether or not the cursor blink should be reset.
func (m *Model) cursorStart() bool {
	return m.setCursor(0)
}

// CursorEnd moves the cursor to the end of the input field.
func (m *Model) CursorEnd() {
	m.cursorEnd()
}

// CursorMode returns the model's cursor mode. For available cursor modes, see
// type CursorMode.
func (m Model) CursorMode() CursorMode {
	return m.cursorMode
}

// SetCursorMode sets the model's cursor mode. This method returns a command.
//
// For available cursor modes, see type CursorMode.
func (m *Model) SetCursorMode(mode CursorMode) tea.Cmd {
	m.cursorMode = mode
	m.blink = m.cursorMode == CursorHide || !m.focus
	if mode == CursorBlink {
		return Blink
	}
	return nil
}

// cursorEnd moves the cursor to the end of the input field and returns whether
// the cursor should blink should reset.
func (m *Model) cursorEnd() bool {
	return m.setCursor(len(m.value[m.row]))
}

// Focused returns the focus state on the model.
func (m Model) Focused() bool {
	return m.focus
}

// Focus sets the focus state on the model. When the model is in focus it can
// receive keyboard input and the cursor will be hidden.
func (m *Model) Focus() tea.Cmd {
	m.focus = true
	m.blink = m.cursorMode == CursorHide // show the cursor unless we've explicitly hidden it

	if m.cursorMode == CursorBlink && m.focus {
		return m.blinkCmd()
	}
	return nil
}

// Blur removes the focus state on the model.  When the model is blurred it can
// not receive keyboard input and the cursor will be hidden.
func (m *Model) Blur() {
	m.focus = false
	m.blink = true
}

// Reset sets the input to its default state with no input. Returns whether
// or not the cursor blink should reset.
func (m *Model) Reset() bool {
	m.value = make([][]rune, minHeight, maxWidth)
	m.col = 0
	m.row = 0
	m.viewport.GotoTop()
	return m.setCursor(0)
}

// handle a clipboard paste event, if supported. Returns whether or not the
// cursor blink should reset.
func (m *Model) handlePaste(v string) bool {
	paste := []rune(v)

	var availSpace int
	if m.CharLimit > 0 {
		availSpace = m.CharLimit - m.Length()
	}

	// If the char limit's been reached cancel
	if m.CharLimit > 0 && availSpace <= 0 {
		return false
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
	resetBlink := m.setCursor(m.col + len(paste))
	return resetBlink
}

// deleteBeforeCursor deletes all text before the cursor. Returns whether or
// not the cursor blink should be reset.
func (m *Model) deleteBeforeCursor() bool {
	m.value[m.row] = m.value[m.row][m.col:]
	return m.setCursor(0)
}

// deleteAfterCursor deletes all text after the cursor. Returns whether or not
// the cursor blink should be reset. If input is masked delete everything after
// the cursor so as not to reveal word breaks in the masked input.
func (m *Model) deleteAfterCursor() bool {
	m.value[m.row] = m.value[m.row][:m.col]
	return m.setCursor(len(m.value[m.row]))
}

// deleteWordLeft deletes the word left to the cursor. Returns whether or not
// the cursor blink should be reset.
func (m *Model) deleteWordLeft() bool {
	if m.col == 0 || len(m.value[m.row]) == 0 {
		return false
	}

	if m.EchoMode != EchoNormal {
		return m.deleteBeforeCursor()
	}

	// Linter note: it's critical that we acquire the initial cursor position
	// here prior to altering it via SetCursor() below. As such, moving this
	// call into the corresponding if clause does not apply here.
	oldCol := m.col //nolint:ifshort

	blink := m.setCursor(m.col - 1)
	for unicode.IsSpace(m.value[m.row][m.col]) {
		if m.col <= 0 {
			break
		}
		// ignore series of whitespace before cursor
		blink = m.setCursor(m.col - 1)
	}

	for m.col > 0 {
		if !unicode.IsSpace(m.value[m.row][m.col]) {
			blink = m.setCursor(m.col - 1)
		} else {
			if m.col > 0 {
				// keep the previous space
				blink = m.setCursor(m.col + 1)
			}
			break
		}
	}

	if oldCol > len(m.value[m.row]) {
		m.value[m.row] = m.value[m.row][:m.col]
	} else {
		m.value[m.row] = append(m.value[m.row][:m.col], m.value[m.row][oldCol:]...)
	}

	return blink
}

// deleteWordRight deletes the word right to the cursor. Returns whether or not
// the cursor blink should be reset. If input is masked delete everything after
// the cursor so as not to reveal word breaks in the masked input.
func (m *Model) deleteWordRight() bool {
	if m.col >= len(m.value[m.row]) || len(m.value[m.row]) == 0 {
		return false
	}

	if m.EchoMode != EchoNormal {
		return m.deleteAfterCursor()
	}

	oldCol := m.col
	m.setCursor(m.col + 1)
	for unicode.IsSpace(m.value[m.row][m.col]) {
		// ignore series of whitespace after cursor
		m.setCursor(m.col + 1)

		if m.col >= len(m.value[m.row]) {
			break
		}
	}

	for m.col < len(m.value[m.row]) {
		if !unicode.IsSpace(m.value[m.row][m.col]) {
			m.setCursor(m.col + 1)
		} else {
			break
		}
	}

	if m.col > len(m.value[m.row]) {
		m.value[m.row] = m.value[m.row][:oldCol]
	} else {
		m.value[m.row] = append(m.value[m.row][:oldCol], m.value[m.row][m.col:]...)
	}

	return m.setCursor(oldCol)
}

// wordLeft moves the cursor one word to the left. Returns whether or not the
// cursor blink should be reset. If input is masked, move input to the start
// so as not to reveal word breaks in the masked input.
func (m *Model) wordLeft() bool {
	if m.col == 0 || len(m.value[m.row]) == 0 {
		return false
	}

	if m.EchoMode != EchoNormal {
		return m.cursorStart()
	}

	blink := false
	i := m.col - 1
	for i >= 0 {
		if unicode.IsSpace(m.value[m.row][min(i, len(m.value[m.row])-1)]) {
			blink = m.setCursor(m.col - 1)
			i--
		} else {
			break
		}
	}

	for i >= 0 {
		if !unicode.IsSpace(m.value[m.row][min(i, len(m.value[m.row])-1)]) {
			blink = m.setCursor(m.col - 1)
			i--
		} else {
			break
		}
	}

	return blink
}

// wordRight moves the cursor one word to the right. Returns whether or not the
// cursor blink should be reset. If the input is masked, move input to the end
// so as not to reveal word breaks in the masked input.
func (m *Model) wordRight() bool {
	if m.col >= len(m.value[m.row]) || len(m.value[m.row]) == 0 {
		return false
	}

	if m.EchoMode != EchoNormal {
		return m.cursorEnd()
	}

	blink := false
	i := m.col
	for i < len(m.value[m.row]) {
		if unicode.IsSpace(m.value[m.row][i]) {
			blink = m.setCursor(m.col + 1)
			i++
		} else {
			break
		}
	}

	for i < len(m.value[m.row]) {
		if !unicode.IsSpace(m.value[m.row][i]) {
			blink = m.setCursor(m.col + 1)
			i++
		} else {
			break
		}
	}

	return blink
}

// LineInfo returns the number of characters from the start of the
// (soft-wrapped) line and the (soft-wrapped) line width.
func (m Model) LineInfo() SoftLineInfo {
	grid := wrap(m.value[m.row], m.Width)

	// Find out which line we are currently on. This can be determined by the
	// m.col and counting the number of runes that we need to skip.
	var counter int
	for i, line := range grid {
		// We've found the line that we are on
		if counter+len(line) == m.col && i+1 < len(grid) {
			// We wrap around to the next line if we are at the end of the
			// previous line so that we can be at the very beginning of the row
			return SoftLineInfo{
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
			return SoftLineInfo{
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
	return SoftLineInfo{}
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

func (m Model) echoTransform(v string) string {
	switch m.EchoMode {
	case EchoPassword:
		return strings.Repeat(string(m.EchoCharacter), rw.StringWidth(v))
	case EchoNone:
		return ""

	default:
		return v
	}
}

// Update is the Bubble Tea update loop.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focus {
		m.blink = true
		return m, nil
	}

	var (
		resetBlink      bool
		resetCharOffset = true
	)
	var cmds []tea.Cmd

	if m.Height != m.viewport.Height {
		m.Height = clamp(m.Height, minHeight, maxHeight)
		m.viewport.Height = clamp(m.Height, minHeight, maxHeight)
	}
	if m.Width != m.viewport.Width {
		m.Width = clamp(m.Width, minWidth, maxWidth)
		m.viewport.Width = clamp(m.Width, minWidth, maxWidth)
	}
	if m.value[m.row] == nil {
		m.value[m.row] = make([]rune, 0)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyBackspace: // delete character before cursor
			if msg.Alt {
				resetBlink = m.deleteWordLeft()
			} else {
				m.col = clamp(m.col, 0, len(m.value[m.row]))
				if m.col <= 0 {
					m.mergeLineAbove(m.row)
					resetBlink = true
					break
				}
				if len(m.value[m.row]) > 0 {
					m.value[m.row] = append(m.value[m.row][:max(0, m.col-1)], m.value[m.row][m.col:]...)
					if m.col > 0 {
						resetBlink = m.setCursor(m.col - 1)
					}
				}
			}

		case tea.KeyUp, tea.KeyCtrlP:
			resetBlink = true
			m.CursorUp()
			resetCharOffset = false
		case tea.KeyDown, tea.KeyCtrlN:
			resetBlink = true
			m.CursorDown()
			resetCharOffset = false
		case tea.KeyEnter:
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			m.splitLine(m.row, m.col)
			resetBlink = true
		case tea.KeyLeft, tea.KeyCtrlB:
			if msg.Alt { // alt+left arrow, back one word
				resetBlink = m.wordLeft()
				break
			}
			if m.col == 0 && m.row != 0 {
				m.row--
				m.cursorEnd()
				resetBlink = true
				break
			}
			if m.col > 0 { // left arrow, ^B, back one character
				resetBlink = m.setCursor(m.col - 1)
			}
		case tea.KeyRight, tea.KeyCtrlF:
			if msg.Alt { // alt+right arrow, forward one word
				resetBlink = m.wordRight()
				break
			}
			if m.col < len(m.value[m.row]) { // right arrow, ^F, forward one character
				resetBlink = m.setCursor(m.col + 1)
			} else {
				if m.row < len(m.value)-1 {
					m.row++
					m.cursorStart()
				}
				resetBlink = true
			}
		case tea.KeyCtrlW: // ^W, delete word left of cursor
			if m.col <= 0 {
				m.mergeLineAbove(m.row)
				resetBlink = true
				break
			}
			resetBlink = m.deleteWordLeft()
		case tea.KeyHome, tea.KeyCtrlA: // ^A, go to beginning
			resetBlink = m.cursorStart()
		case tea.KeyDelete, tea.KeyCtrlD: // ^D, delete char under cursor
			if len(m.value[m.row]) > 0 && m.col < len(m.value[m.row]) {
				m.value[m.row] = append(m.value[m.row][:m.col], m.value[m.row][m.col+1:]...)
			}
			if m.col >= len(m.value[m.row]) {
				resetBlink = true
				m.mergeLineBelow(m.row)
				break
			}
		case tea.KeyCtrlE, tea.KeyEnd: // ^E, go to end
			resetBlink = m.cursorEnd()
		case tea.KeyCtrlK: // ^K, kill text after cursor
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			if m.col >= len(m.value[m.row]) {
				resetBlink = true
				m.mergeLineBelow(m.row)
				break
			}
			resetBlink = m.deleteAfterCursor()
		case tea.KeyCtrlU: // ^U, kill text before cursor
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			if m.col <= 0 {
				resetBlink = true
				m.mergeLineAbove(m.row)
				break
			}
			resetBlink = m.deleteBeforeCursor()
		case tea.KeyCtrlV: // ^V paste
			return m, Paste
		case tea.KeyRunes, tea.KeySpace: // input regular characters
			if msg.Alt && len(msg.Runes) == 1 {
				if msg.Runes[0] == 'd' { // alt+d, delete word right of cursor
					resetBlink = m.deleteWordRight()
					break
				}
				if msg.Runes[0] == 'b' { // alt+b, back one word
					resetBlink = m.wordLeft()
					break
				}
				if msg.Runes[0] == 'f' { // alt+f, forward one word
					resetBlink = m.wordRight()
					break
				}
			}

			if rw.StringWidth(m.Value()) >= m.CharLimit {
				break
			}

			m.col = min(m.col, len(m.value[m.row]))
			m.value[m.row] = append(m.value[m.row][:m.col], append(msg.Runes, m.value[m.row][m.col:]...)...)
			resetBlink = m.setCursor(m.col + len(msg.Runes))
		}

	case initialBlinkMsg:
		// We accept all initialBlinkMsgs generated by the Blink command.

		if m.cursorMode != CursorBlink || !m.focus {
			return m, nil
		}

		cmd := m.blinkCmd()
		return m, cmd

	case blinkMsg:
		// We're choosy about whether to accept blinkMsgs so that our cursor
		// only exactly when it should.

		// Is this model blink-able?
		if m.cursorMode != CursorBlink || !m.focus {
			return m, nil
		}

		// Were we expecting this blink message?
		if msg.id != m.id || msg.tag != m.blinkTag {
			return m, nil
		}

		var cmd tea.Cmd
		if m.cursorMode == CursorBlink {
			m.blink = !m.blink
			cmd = m.blinkCmd()
		}
		return m, cmd

	case blinkCanceled: // no-op
		return m, nil

	case pasteMsg:
		resetBlink = m.handlePaste(string(msg))

	case pasteErrMsg:
		m.Err = msg
	}

	vp, cmd := m.viewport.Update(msg)
	m.viewport = &vp
	cmds = append(cmds, cmd)

	if resetCharOffset {
		m.lastCharOffset = 0
	}

	if resetBlink {
		m.blink = false
		cmds = append(cmds, m.blinkCmd())
	}

	m.repositionView()

	return m, tea.Batch(cmds...)
}

// View renders the text area in its current state.
func (m Model) View() string {
	if m.Value() == "" && m.row == 0 && m.col == 0 && m.Placeholder != "" {
		return m.placeholderView()
	}

	var s strings.Builder
	var style lipgloss.Style
	lineInfo := m.LineInfo()

	var newLines int

	for l, line := range m.value {
		wrappedLines := wrap(line, m.Width)

		if m.row == l {
			style = m.CursorLineStyle
		} else {
			style = m.TextStyle
		}

		for wl, wrappedLine := range wrappedLines {
			s.WriteString(m.PromptStyle.Render(style.Render(m.Prompt)))

			if m.ShowLineNumbers {
				if wl == 0 {
					s.WriteString(m.LineNumberStyle.Render(style.Render(fmt.Sprintf(m.lineNumberFormat, l+1))))
				} else {
					s.WriteString(m.LineNumberStyle.Render(style.Render("   ")))
				}
			}

			strwidth := rw.StringWidth(string(wrappedLine))
			padding := m.Width - strwidth
			// If the trailing space causes the line to be wider than the
			// width, we should not draw it to the screen since it will result
			// in an extra space at the end of the line which can look off when
			// the cursor line is showing.
			if strwidth > m.Width {
				// The character causing the line to be wider than the width is
				// guaranteed to be a space since any other character would
				// have been wrapped.
				wrappedLine = []rune(strings.TrimSuffix(string(wrappedLine), " "))
				padding -= m.Width - strwidth
			}
			if m.row == l && lineInfo.RowOffset == wl {
				s.WriteString(style.Render(string(wrappedLine[:lineInfo.ColumnOffset])))
				if m.col >= len(line) && lineInfo.CharOffset >= m.Width {
					s.WriteString(m.cursorView(" "))
				} else {
					s.WriteString(style.Render(m.cursorView(string(wrappedLine[lineInfo.ColumnOffset]))))
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
	for i := 0; i < m.Height; i++ {
		s.WriteString(m.PromptStyle.Render(m.TextStyle.Render(m.Prompt)))

		if m.ShowLineNumbers {
			lineNumber := m.LineNumberStyle.Render((fmt.Sprintf(m.lineNumberFormat, len(m.value)+i+1)))
			s.WriteString(lineNumber)
		}
		s.WriteRune('\n')
	}

	m.viewport.SetContent(s.String())

	return m.viewport.View()
}

// placeholderView returns the prompt and placeholder view, if any.
func (m Model) placeholderView() string {
	var (
		v     string
		p     = m.Placeholder
		style = m.PlaceholderStyle.Inline(true).Render
	)

	prompt := m.PromptStyle.Render(m.Prompt)
	v += m.CursorLineStyle.Render(prompt)

	if m.ShowLineNumbers {
		v += m.CursorLineStyle.Render(m.LineNumberStyle.Render((fmt.Sprintf(m.lineNumberFormat, 1))))
	}

	// Cursor
	if m.blink {
		v += m.CursorLineStyle.Render(m.cursorView(style(p[:1])))
	} else {
		v += m.CursorLineStyle.Render(m.cursorView(p[:1]))
	}

	// The rest of the placeholder text
	v += m.CursorLineStyle.Render(style(p[1:] + strings.Repeat(" ", max(0, m.Width-rw.StringWidth(p)))))

	// The rest of the new lines
	for i := 1; i < m.Height; i++ {
		v += "\n" + prompt

		if m.ShowLineNumbers {
			lineNumber := m.LineNumberStyle.Render((fmt.Sprintf(m.lineNumberFormat, i+1)))
			v += lineNumber
		}
	}

	m.viewport.SetContent(v)
	return m.viewport.View()
}

// cursorView styles the cursor.
func (m Model) cursorView(v string) string {
	if m.blink {
		return m.TextStyle.Render(v)
	}
	return m.CursorStyle.Inline(true).Reverse(true).Render(v)
}

// cursorLineNumber returns the line number that the cursor is on.
// This accounts for soft wrapped lines.
func (m Model) cursorLineNumber() int {
	line := 0
	for i := 0; i < m.row; i++ {
		// Calculate the number of lines that the current line will be split
		// into.
		line += len(wrap(m.value[i], m.Width))
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

// blinkCmd is an internal command used to manage cursor blinking.
func (m *Model) blinkCmd() tea.Cmd {
	if m.cursorMode != CursorBlink {
		return nil
	}

	if m.blinkCtx != nil && m.blinkCtx.cancel != nil {
		m.blinkCtx.cancel()
	}

	ctx, cancel := context.WithTimeout(m.blinkCtx.ctx, m.BlinkSpeed)
	m.blinkCtx.cancel = cancel

	m.blinkTag++

	return func() tea.Msg {
		defer cancel()
		<-ctx.Done()
		if ctx.Err() == context.DeadlineExceeded {
			return blinkMsg{id: m.id, tag: m.blinkTag}
		}
		return blinkCanceled{}
	}
}

// Blink is a command used to initialize cursor blinking.
func Blink() tea.Msg {
	return initialBlinkMsg{}
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
			if rw.StringWidth(string(word)) >= width {
				// The word is longer than the width we're trying to contain it
				// to so we will have to break up this word no matter what.
				// Let's try to fit as much as we can on the current line and
				// put the rest on the next line.
				remainingWidth := width - rw.StringWidth(string(lines[row]))

				// Find the column in the word that corresponds to the
				// remaining width. This is entirely to handle double-width
				// runes. As for single-width runes, splitCol will be the same
				// as the remainingWidth.
				var splitCol int
				var stringWidth int
				for ; splitCol < len(word); splitCol++ {
					stringWidth += rw.RuneWidth(word[splitCol])
					if stringWidth > remainingWidth {
						break
					}
				}

				lines[row] = append(lines[row], word[:splitCol]...)
				row++
				lines = append(lines, []rune{})
				lines[row] = append(lines[row], word[splitCol:]...)
				word = nil
			}
		}
	}

	if rw.StringWidth(string(lines[row]))+rw.StringWidth(string(word))+spaces > width {
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
