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

// String returns a the cursor mode in a human-readable format. This method is
// provisional and for informational purposes only.
func (c CursorMode) String() string {
	return [...]string{
		"blink",
		"static",
		"hidden",
	}[c]
}

// Model is the Bubble Tea model for this text input element.
type Model struct {
	Err error

	// General settings.
	Prompt        string
	Placeholder   string
	BlinkSpeed    time.Duration
	EchoMode      EchoMode
	EchoCharacter rune

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
	CursorLineStyle  lipgloss.Style
	LineNumberStyle  lipgloss.Style

	// CharLimit is the maximum number of characters this input element will
	// accept. If 0 or less, there's no limit.
	CharLimit int

	// Width is the maximum number of characters that can be displayed at once.
	// It essentially treats the text field like a horizontally scrolling
	// viewport. If 0 or less this setting is ignored.
	Width int

	// LineLimit is the maximum number of lines this input element will accept.
	// If 0 or less, there's no limit.
	LineLimit int

	// Height is the maximum number of lines that can be displayed at once.
	// It essentially treats the text field like a vertically scrolling viewport
	// if there are more lines that permitted height.
	Height int

	// The ID of this Model as it relates to other textinput Models.
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

	// Used to manage cursor blink
	blinkCtx *blinkCtx

	// cursorMode determines the behavior of the cursor
	cursorMode CursorMode

	// lineNumberFormat is the format string used to display line numbers.
	lineNumberFormat string

	// Viewport is the vertically-scrollable Viewport of the multi-line text
	// input.
	Viewport *viewport.Model
}

// NewModel creates a new model with default settings.
func New() Model {
	vp := viewport.New(0, 0)
	vp.KeyMap = viewport.KeyMap{}

	return Model{
		Prompt:           "│ ",
		BlinkSpeed:       defaultBlinkSpeed,
		EchoCharacter:    '*',
		CharLimit:        0,
		PlaceholderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		LineNumberStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		LineLimit:        1,
		Height:           1,

		id:               nextID(),
		value:            nil,
		focus:            false,
		blink:            true,
		col:              0,
		row:              0,
		cursorMode:       CursorBlink,
		lineNumberFormat: "%3d ",

		blinkCtx: &blinkCtx{
			ctx: context.Background(),
		},

		Viewport: &vp,
	}
}

// SetValue sets the value of the text input.
func (m *Model) SetValue(s string) {
	runes := []rune(s)
	if m.CharLimit > 0 && len(runes) > m.CharLimit {
		m.value[m.row] = runes[:m.CharLimit]
	} else {
		m.value[m.row] = runes
	}
	if m.col == 0 || m.col > len(m.value[m.row]) {
		m.setCursor(len(m.value[m.row]))
	}
	m.handleOverflow()
}

// Value returns the value of the text input.
func (m Model) Value() string {
	var v string
	for _, l := range m.value {
		v += string(l)
		v += "\n"
	}
	return strings.TrimSpace(v)
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

// Cursor returns the line position.
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

// setCursor moves the cursor to the given position and returns whether or not
// the cursor blink should be reset. If the position is out of bounds the
// cursor will be moved to the start or end accordingly.
func (m *Model) setCursor(col int) bool {
	m.col = clamp(col, 0, len(m.value[m.row]))
	m.handleOverflow()

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
// whether or not the curosr blink should be reset.
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

// CursorMode sets the model's cursor mode. This method returns a command.
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
	m.value = make([][]rune, m.LineLimit)
	m.col = 0
	m.row = 0
	m.Viewport.GotoTop()
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
	m.handleColumnBoundaries()
	resetBlink := m.setCursor(m.col + len(paste))
	return resetBlink
}

// If input is multi-line, the input can scroll vertically,
// otherwise, scroll horizontally.
func (m *Model) handleOverflow() {
	for i := 0; i < len(m.value)-1; i++ {
		l := m.value[i]
		if rw.StringWidth(string(l)) < m.Width {
			// The line is less than the maximum width, so let's move on to the
			// next line.
			continue
		}

		// The line is too long, so let's wrap it
		// Before we do this, we need to find the character that will act as
		// the break point. Since we may have multi-width characters this will
		// not always align with the m.value[row][width-1]
		w := 0
		for j := 0; j < len(l); j++ {
			w += rw.RuneWidth(l[j])
			if w >= m.Width {
				// We've hit the maximum number of characters we can allow on this line
				// Let's work backwards until we find a nice break point
				bp := j
				for bp > m.col+1 && !unicode.IsSpace(l[bp]) {
					bp--
				}
				var overflow []rune = make([]rune, len(l[bp:]))
				copy(overflow, l[bp:])
				m.value[i] = l[:bp]
				m.value[i+1] = concat(overflow, m.value[i+1])
				break
			}
		}
	}
}

// canHandleMoreInput returns whether or not the input can handle `length` more
// characters of input being added.
func (m *Model) canHandleMoreInput(length int) bool {
	if m.CharLimit >= 0 && m.Length() >= m.CharLimit {
		return false
	}

	// Depending on where we are in the multi-line input, we may not be able to insert
	// more characters as they may overflow the input area when wrapping, i.e. may go over the LineLimit.
	// In this case, let's just make sure we can handle the input.

	// We'll need to count the number of characters remaining and the characters we've already inserted
	// starting from the cursor.
	spaceRemaining := ((m.LineLimit - m.row) * (m.Width - 1)) + 1
	spaceUsed := 0
	for i := m.row; i < m.LineLimit; i++ {
		spaceUsed += rw.StringWidth(string(m.value[i]))
	}
	return spaceUsed+length < spaceRemaining
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

// lineDown moves the cursor down by `n` lines.
// Returns whether or not the cursor blink should be reset.
func (m *Model) lineDown(n int) bool {
	if m.row < m.LineLimit-1 {
		m.row++
	}
	m.Viewport.SetYOffset(m.row - m.Height/2)
	return true
}

// lineUp moves the cursor up by `n` lines.
// returns whether or not the cursor blink should be reset.
func (m *Model) lineUp(n int) bool {
	if m.row > 0 {
		m.row--
	}
	m.Viewport.SetYOffset(m.row - m.Height/2)
	return true
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

	if m.value == nil {
		m.value = make([][]rune, m.LineLimit)
		m.Viewport.Height = m.Height
		m.Viewport.Width = m.Width
	}

	var resetBlink bool
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyBackspace: // delete character before cursor
			m.handleColumnBoundaries()

			if msg.Alt {
				resetBlink = m.deleteWordLeft()
			} else {
				// In a multi-line input, if the cursor is at the start of a
				// line, and backspace is pressed move the cursor to the end of
				// the previous line and bring the previous line up.
				if m.col == 0 && m.row > 0 {
					rowIsEmpty := len(m.value[m.row]) == 0

					resetBlink = m.lineUp(1)
					m.cursorEnd()

					// If the current line is full we won't have space to shift
					// all the other lines up, so simply do nothing.
					if !rowIsEmpty && len(m.value[m.row]) >= m.Width {
						break
					}

					m.value[m.row] = append(m.value[m.row], m.value[m.row+1]...)

					// Shift all the lines up by one.
					for i := m.row + 1; i < m.LineLimit-1; i++ {
						m.value[i] = m.value[i+1]
					}
					// Clear the last line
					m.value[m.LineLimit-1] = nil
					break
				}

				if len(m.value[m.row]) > 0 {
					m.value[m.row] = append(m.value[m.row][:max(0, m.col-1)], m.value[m.row][m.col:]...)
					if m.col > 0 {
						resetBlink = m.setCursor(m.col - 1)
					}
				}
			}

		case tea.KeyUp:
			resetBlink = m.lineUp(1)
		case tea.KeyDown:
			resetBlink = m.lineDown(1)
		case tea.KeyEnter:
			m.handleColumnBoundaries()

			lastRow := m.row
			m.lineDown(1)
			currentRow := m.row

			// On a multi-line input, we will need to shift the lines after the
			// cursor line down by one since a new line was inserted.
			if m.LineLimit <= 1 {
				break
			}

			// First, let's ensure that there is enough space to insert a new line.
			// We can do this by ensuring that the last line is empty.
			if len(m.value[m.LineLimit-1]) > 0 {
				break
			}

			// Shift all rows after the current row down by one.
			for i := m.LineLimit - 1; i > currentRow; i-- {
				m.value[i] = make([]rune, len(m.value[i-1]))
				copy(m.value[i], m.value[i-1])
			}

			// Split the current line into two lines.
			s1, s2 := m.value[lastRow][:m.col], m.value[lastRow][m.col:]
			m.value[lastRow], m.value[currentRow] = make([]rune, len(s1)), make([]rune, len(s2))
			copy(m.value[lastRow], s1)
			copy(m.value[currentRow], s2)

			// Reset column only if we've actually changed rows
			if lastRow != currentRow {
				m.col = 0
			}

		case tea.KeyLeft, tea.KeyCtrlB:
			if msg.Alt { // alt+left arrow, back one word
				resetBlink = m.wordLeft()
				break
			}
			if m.col == 0 && m.row != 0 {
				resetBlink = m.lineUp(1)
				m.cursorEnd()
				m.col++
			}
			if m.col > 0 { // left arrow, ^F, back one character
				resetBlink = m.setCursor(m.col - 1)
			}
		case tea.KeyRight, tea.KeyCtrlF:
			if msg.Alt { // alt+right arrow, forward one word
				resetBlink = m.wordRight()
				break
			}
			if m.col >= len(m.value[m.row]) && m.row != m.LineLimit-1 {
				m.lineDown(1)
				m.cursorStart()
				m.col--
			}
			if m.col < len(m.value[m.row]) { // right arrow, ^F, forward one character
				resetBlink = m.setCursor(m.col + 1)
			}
		case tea.KeyCtrlW: // ^W, delete word left of cursor
			m.handleColumnBoundaries()
			resetBlink = m.deleteWordLeft()
		case tea.KeyHome, tea.KeyCtrlA: // ^A, go to beginning
			resetBlink = m.cursorStart()
		case tea.KeyDelete, tea.KeyCtrlD: // ^D, delete char under cursor
			m.handleColumnBoundaries()
			if len(m.value[m.row]) > 0 && m.col < len(m.value[m.row]) {
				m.value[m.row] = append(m.value[m.row][:m.col], m.value[m.row][m.col+1:]...)
			}
		case tea.KeyCtrlE, tea.KeyEnd: // ^E, go to end
			resetBlink = m.cursorEnd()
		case tea.KeyCtrlK: // ^K, kill text after cursor
			m.handleColumnBoundaries()
			resetBlink = m.deleteAfterCursor()
		case tea.KeyCtrlU: // ^U, kill text before cursor
			m.handleColumnBoundaries()
			resetBlink = m.deleteBeforeCursor()
		case tea.KeyCtrlV: // ^V paste
			return m, Paste
		case tea.KeyCtrlN: // ^N next line
			resetBlink = m.lineDown(1)
		case tea.KeyCtrlP: // ^P previous line
			resetBlink = m.lineUp(1)
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

			m.handleColumnBoundaries()

			// We can't allow the user to input if we are already at the maximum width and height.
			lw := rw.StringWidth(string(m.value[m.row]))
			msgw := rw.StringWidth(string(msg.Runes))
			if m.row >= m.LineLimit-1 && lw+msgw >= m.Width {
				break
			}

			// If the cursor is at the end of the line let's move the cursor to
			// the next line
			if rw.StringWidth(string(m.value[m.row][:m.col])) >= m.Width-1 {
				// We've hit the end of the line, let's wrap the word we are
				// currently typing to the next line.
				bp := m.col - 1

				// Words are delimited by spaces
				for bp > 0 && !unicode.IsSpace(m.value[m.row][bp]) {
					bp--
				}

				if bp == 0 {
					// There is no space on the previous line, so let's just
					// split the line at the column
					bp = m.col - 1
				}

				word := string(m.value[m.row][(bp + 1):])
				m.value[m.row] = m.value[m.row][:bp]

				m.lineDown(1)

				m.value[m.row] = concat([]rune(word), m.value[m.row])
				m.col = len(word)
			}

			if len(msg.Runes) > 1 {
				// We are possibly pasting in multiple characters. If this
				// paste contains a new line it can break the input, so strip
				// newlines away.
				for i, r := range msg.Runes {
					if r == '\n' || r == '\r' {
						msg.Runes[i] = ' '
					}
				}
			}

			// Input a regular character
			if m.canHandleMoreInput(msgw) {
				m.value[m.row] = append(m.value[m.row][:m.col], append(msg.Runes, m.value[m.row][m.col:]...)...)
				resetBlink = m.setCursor(m.col + msgw)

				if m.col > m.Width && m.row <= m.LineLimit-1 {
					newLines := m.col / m.Width
					m.row += newLines
					m.col = (m.col % m.Width) + newLines
					// Re-center the viewport
					m.Viewport.SetYOffset(m.row - m.Height/2)
				}
			}
		}

	case initialBlinkMsg:
		// We accept all initialBlinkMsgs genrated by the Blink command.

		if m.cursorMode != CursorBlink || !m.focus {
			return m, nil
		}

		cmd := m.blinkCmd()
		return m, cmd

	case blinkMsg:
		// We're choosy about whether to accept blinkMsgs so that our cursor
		// only exactly when it should.

		// Is this model blinkable?
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

	vp, cmd := m.Viewport.Update(msg)
	m.Viewport = &vp
	cmds = append(cmds, cmd)

	m.handleOverflow()

	if resetBlink {
		m.blink = false
		cmds = append(cmds, m.blinkCmd())
	}
	return m, tea.Batch(cmds...)
}

// View renders the textinput in its current state.
func (m Model) View() string {
	// Placeholder text
	if m.Value() == "" && m.row == 0 && m.Placeholder != "" {
		return m.placeholderView()
	}

	var (
		str             string
		styleText       = m.TextStyle.Inline(true).Render
		styleCursorLine = m.CursorLineStyle.Render
	)

	// Display the value for all it's height
	for i := 0; i < m.LineLimit; i++ {
		var v string
		value := m.value[i]

		// We're at the cursor line now, so display the cursor
		if i == m.row {
			col := min(max(0, m.col), len(value))
			v = styleCursorLine(m.echoTransform(string(value[:col])))
			padding := m.Width - rw.StringWidth(string(value))
			if m.col < len(value) {
				v += styleCursorLine(m.cursorView(m.echoTransform(string(value[m.col])))) // cursor and text under it
				v += styleCursorLine(m.echoTransform(string(value[m.col+1:])))            // text after cursor
			} else {
				v += styleCursorLine(m.cursorView(" "))
				padding--
			}

			// Add padding to fill out the rest of the background
			v += styleCursorLine(strings.Repeat(" ", max(0, padding)))
		} else {
			v = styleText(m.echoTransform(string(value)))
		}

		str += m.PromptStyle.Render(m.Prompt)

		if m.ShowLineNumbers {
			lineNumber := m.LineNumberStyle.Render(fmt.Sprintf(m.lineNumberFormat, i+1))
			if m.row == i {
				str += styleCursorLine(lineNumber)
			} else {
				str += lineNumber
			}
		}
		str += v + "\n"
	}

	m.Viewport.SetContent(str)
	return m.Viewport.View()
}

// placeholderView returns the prompt and placeholder view, if any.
func (m Model) placeholderView() string {
	var (
		v       string
		p       = m.Placeholder
		style   = m.PlaceholderStyle.Inline(true).Render
		clStyle = m.CursorLineStyle.Render
		lnStyle = m.LineNumberStyle.Render
	)

	prompt := m.PromptStyle.Render(m.Prompt)
	v += prompt

	if m.ShowLineNumbers {
		v += clStyle(lnStyle((fmt.Sprintf(m.lineNumberFormat, 1))))
	}

	// Cursor
	if m.blink {
		v += clStyle(m.cursorView(style(p[:1])))
	} else {
		v += clStyle(m.cursorView(p[:1]))
	}

	// The rest of the placeholder text
	v += clStyle(style(p[1:] + strings.Repeat(" ", max(0, m.Width-rw.StringWidth(p)))))

	// The rest of the new lines
	for i := 1; i < m.LineLimit; i++ {
		v += "\n" + prompt

		if m.ShowLineNumbers {
			lineNumber := lnStyle((fmt.Sprintf(m.lineNumberFormat, i+1)))
			if i == 0 {
				v += clStyle(lineNumber)
			} else {
				v += lineNumber
			}
		}
	}

	m.Viewport.SetContent(v)
	return m.Viewport.View()
}

// cursorView styles the cursor.
func (m Model) cursorView(v string) string {
	if m.blink {
		return m.TextStyle.Render(v)
	}
	return m.CursorStyle.Inline(true).Reverse(true).Render(v)
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

func (m *Model) handleColumnBoundaries() {
	// While the user is traversing a multi-line input, the cursor may be past
	// the end of the line. This is not an issue until the user makes a change,
	// in which case we will want to adjust the cursor so that it is within the
	// bounds of the line.
	//
	// We don't want to adjust the cursor if the user is only moving the cursor
	// as it may be disorienting if the user goes from a long line to a short
	// line and then back to a long line, otherwise.
	m.col = clamp(m.col, 0, len(m.value[m.row]))
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

func concat(first []rune, second []rune) []rune {
	n := len(first)
	return append(first[:n:n], second...)
}
