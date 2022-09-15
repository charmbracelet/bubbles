package textinput

import (
	"strings"
	"time"
	"unicode"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/internal/cursor"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	rw "github.com/mattn/go-runewidth"
)

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

// ValidateFunc is a function that returns an error if the input is invalid.
type ValidateFunc func(string) error

// Model is the Bubble Tea model for this text input element.
type Model struct {
	Err error

	// General settings.
	Prompt        string
	Placeholder   string
	EchoMode      EchoMode
	EchoCharacter rune
	Cursor        cursor.Model

	// Deprecated: use cursor.BlinkSpeed instead.
	// This is unused and will be removed in the future.
	BlinkSpeed time.Duration

	// Styles. These will be applied as inline styles.
	//
	// For an introduction to styling with Lip Gloss see:
	// https://github.com/charmbracelet/lipgloss
	PromptStyle      lipgloss.Style
	TextStyle        lipgloss.Style
	BackgroundStyle  lipgloss.Style
	PlaceholderStyle lipgloss.Style
	CursorStyle      lipgloss.Style

	// CharLimit is the maximum amount of characters this input element will
	// accept. If 0 or less, there's no limit.
	CharLimit int

	// Width is the maximum number of characters that can be displayed at once.
	// It essentially treats the text field like a horizontally scrolling
	// viewport. If 0 or less this setting is ignored.
	Width int

	// Underlying text value.
	value []rune

	// focus indicates whether user input focus should be on this input
	// component. When false, ignore keyboard input and hide the cursor.
	focus bool

	// Cursor position.
	pos int

	// Used to emulate a viewport when width is set and the content is
	// overflowing.
	offset      int
	offsetRight int

	// Validate is a function that checks whether or not the text within the
	// input is valid. If it is not valid, the `Err` field will be set to the
	// error returned by the function. If the function is not defined, all
	// input is considered valid.
	Validate ValidateFunc
}

// New creates a new model with default settings.
func New() Model {
	return Model{
		Prompt:           "> ",
		EchoCharacter:    '*',
		CharLimit:        0,
		PlaceholderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Cursor:           cursor.New(),

		value: nil,
		focus: false,
		pos:   0,
	}
}

// NewModel creates a new model with default settings.
//
// Deprecated. Use New instead.
var NewModel = New

// SetValue sets the value of the text input.
func (m *Model) SetValue(s string) {
	if m.Validate != nil {
		if err := m.Validate(s); err != nil {
			m.Err = err
			return
		}
	}

	m.Err = nil

	runes := []rune(s)
	if m.CharLimit > 0 && len(runes) > m.CharLimit {
		m.value = runes[:m.CharLimit]
	} else {
		m.value = runes
	}
	if m.pos > len(m.value) {
		m.SetCursor(len(m.value))
	}
	m.handleOverflow()
}

// Value returns the value of the text input.
func (m Model) Value() string {
	return string(m.value)
}

// Position returns the cursor position.
func (m Model) Position() int {
	return m.pos
}

// SetCursor moves the cursor to the given position. If the position is
// out of bounds the cursor will be moved to the start or end accordingly.
func (m *Model) SetCursor(pos int) {
	m.pos = clamp(pos, 0, len(m.value))
	m.handleOverflow()
}

// CursorStart moves the cursor to the start of the input field.
func (m *Model) CursorStart() {
	m.SetCursor(0)
}

// CursorEnd moves the cursor to the end of the input field.
func (m *Model) CursorEnd() {
	m.SetCursor(len(m.value))
}

// Focused returns the focus state on the model.
func (m Model) Focused() bool {
	return m.focus
}

// Focus sets the focus state on the model. When the model is in focus it can
// receive keyboard input and the cursor will be shown.
func (m *Model) Focus() tea.Cmd {
	m.focus = true
	return m.Cursor.Focus()
}

// Blur removes the focus state on the model.  When the model is blurred it can
// not receive keyboard input and the cursor will be hidden.
func (m *Model) Blur() {
	m.focus = false
	m.Cursor.Blur()
}

// Reset sets the input to its default state with no input.
func (m *Model) Reset() {
	m.value = nil
	m.SetCursor(0)
}

// handle a clipboard paste event, if supported.
func (m *Model) handlePaste(v string) {
	paste := []rune(v)

	var availSpace int
	if m.CharLimit > 0 {
		availSpace = m.CharLimit - len(m.value)
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
	head := m.value[:m.pos]
	tailSrc := m.value[m.pos:]
	tail := make([]rune, len(tailSrc))
	copy(tail, tailSrc)

	oldPos := m.pos

	// Insert pasted runes
	for _, r := range paste {
		head = append(head, r)
		m.pos++
		if m.CharLimit > 0 {
			availSpace--
			if availSpace <= 0 {
				break
			}
		}
	}

	// Put it all back together
	value := append(head, tail...)
	m.SetValue(string(value))

	if m.Err != nil {
		m.pos = oldPos
	}

	m.SetCursor(m.pos)
}

// If a max width is defined, perform some logic to treat the visible area
// as a horizontally scrolling viewport.
func (m *Model) handleOverflow() {
	if m.Width <= 0 || rw.StringWidth(string(m.value)) <= m.Width {
		m.offset = 0
		m.offsetRight = len(m.value)
		return
	}

	// Correct right offset if we've deleted characters
	m.offsetRight = min(m.offsetRight, len(m.value))

	if m.pos < m.offset {
		m.offset = m.pos

		w := 0
		i := 0
		runes := m.value[m.offset:]

		for i < len(runes) && w <= m.Width {
			w += rw.RuneWidth(runes[i])
			if w <= m.Width+1 {
				i++
			}
		}

		m.offsetRight = m.offset + i
	} else if m.pos >= m.offsetRight {
		m.offsetRight = m.pos

		w := 0
		runes := m.value[:m.offsetRight]
		i := len(runes) - 1

		for i > 0 && w < m.Width {
			w += rw.RuneWidth(runes[i])
			if w <= m.Width {
				i--
			}
		}

		m.offset = m.offsetRight - (len(runes) - 1 - i)
	}
}

// deleteBeforeCursor deletes all text before the cursor.
func (m *Model) deleteBeforeCursor() {
	m.value = m.value[m.pos:]
	m.offset = 0
	m.SetCursor(0)
}

// deleteAfterCursor deletes all text after the cursor. If input is masked
// delete everything after the cursor so as not to reveal word breaks in the
// masked input.
func (m *Model) deleteAfterCursor() {
	m.value = m.value[:m.pos]
	m.SetCursor(len(m.value))
}

// deleteWordLeft deletes the word left to the cursor.
func (m *Model) deleteWordLeft() {
	if m.pos == 0 || len(m.value) == 0 {
		return
	}

	if m.EchoMode != EchoNormal {
		m.deleteBeforeCursor()
		return
	}

	// Linter note: it's critical that we acquire the initial cursor position
	// here prior to altering it via SetCursor() below. As such, moving this
	// call into the corresponding if clause does not apply here.
	oldPos := m.pos //nolint:ifshort

	m.SetCursor(m.pos - 1)
	for unicode.IsSpace(m.value[m.pos]) {
		if m.pos <= 0 {
			break
		}
		// ignore series of whitespace before cursor
		m.SetCursor(m.pos - 1)
	}

	for m.pos > 0 {
		if !unicode.IsSpace(m.value[m.pos]) {
			m.SetCursor(m.pos - 1)
		} else {
			if m.pos > 0 {
				// keep the previous space
				m.SetCursor(m.pos + 1)
			}
			break
		}
	}

	if oldPos > len(m.value) {
		m.value = m.value[:m.pos]
	} else {
		m.value = append(m.value[:m.pos], m.value[oldPos:]...)
	}
}

// deleteWordRight deletes the word right to the cursor If input is masked
// delete everything after the cursor so as not to reveal word breaks in the
// masked input.
func (m *Model) deleteWordRight() {
	if m.pos >= len(m.value) || len(m.value) == 0 {
		return
	}

	if m.EchoMode != EchoNormal {
		m.deleteAfterCursor()
		return
	}

	oldPos := m.pos
	m.SetCursor(m.pos + 1)
	for unicode.IsSpace(m.value[m.pos]) {
		// ignore series of whitespace after cursor
		m.SetCursor(m.pos + 1)

		if m.pos >= len(m.value) {
			break
		}
	}

	for m.pos < len(m.value) {
		if !unicode.IsSpace(m.value[m.pos]) {
			m.SetCursor(m.pos + 1)
		} else {
			break
		}
	}

	if m.pos > len(m.value) {
		m.value = m.value[:oldPos]
	} else {
		m.value = append(m.value[:oldPos], m.value[m.pos:]...)
	}

	m.SetCursor(oldPos)
}

// wordLeft moves the cursor one word to the left. If input is masked, move
// input to the start so as not to reveal word breaks in the masked input.
func (m *Model) wordLeft() {
	if m.pos == 0 || len(m.value) == 0 {
		return
	}

	if m.EchoMode != EchoNormal {
		m.CursorStart()
		return
	}

	i := m.pos - 1
	for i >= 0 {
		if unicode.IsSpace(m.value[i]) {
			m.SetCursor(m.pos - 1)
			i--
		} else {
			break
		}
	}

	for i >= 0 {
		if !unicode.IsSpace(m.value[i]) {
			m.SetCursor(m.pos - 1)
			i--
		} else {
			break
		}
	}
}

// wordRight moves the cursor one word to the right. If the input is masked,
// move input to the end so as not to reveal word breaks in the masked input.
func (m *Model) wordRight() {
	if m.pos >= len(m.value) || len(m.value) == 0 {
		return
	}

	if m.EchoMode != EchoNormal {
		m.CursorEnd()
		return
	}

	i := m.pos
	for i < len(m.value) {
		if unicode.IsSpace(m.value[i]) {
			m.SetCursor(m.pos + 1)
			i++
		} else {
			break
		}
	}

	for i < len(m.value) {
		if !unicode.IsSpace(m.value[i]) {
			m.SetCursor(m.pos + 1)
			i++
		} else {
			break
		}
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
		return m, nil
	}

	// Let's remember where the position of the cursor currently is so that if
	// the cursor position changes, we can reset the blink.
	oldPos := m.pos //nolint

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyBackspace, tea.KeyCtrlH: // delete character before cursor
			m.Err = nil

			if msg.Alt {
				m.deleteWordLeft()
			} else {
				if len(m.value) > 0 {
					m.value = append(m.value[:max(0, m.pos-1)], m.value[m.pos:]...)
					if m.pos > 0 {
						m.SetCursor(m.pos - 1)
					}
				}
			}
		case tea.KeyLeft, tea.KeyCtrlB:
			if msg.Alt { // alt+left arrow, back one word
				m.wordLeft()
				break
			}
			if m.pos > 0 { // left arrow, ^F, back one character
				m.SetCursor(m.pos - 1)
			}
		case tea.KeyRight, tea.KeyCtrlF:
			if msg.Alt { // alt+right arrow, forward one word
				m.wordRight()
				break
			}
			if m.pos < len(m.value) { // right arrow, ^F, forward one character
				m.SetCursor(m.pos + 1)
			}
		case tea.KeyCtrlW: // ^W, delete word left of cursor
			m.deleteWordLeft()
		case tea.KeyHome, tea.KeyCtrlA: // ^A, go to beginning
			m.CursorStart()
		case tea.KeyDelete, tea.KeyCtrlD: // ^D, delete char under cursor
			if len(m.value) > 0 && m.pos < len(m.value) {
				m.value = append(m.value[:m.pos], m.value[m.pos+1:]...)
			}
		case tea.KeyCtrlE, tea.KeyEnd: // ^E, go to end
			m.CursorEnd()
		case tea.KeyCtrlK: // ^K, kill text after cursor
			m.deleteAfterCursor()
		case tea.KeyCtrlU: // ^U, kill text before cursor
			m.deleteBeforeCursor()
		case tea.KeyCtrlV: // ^V paste
			return m, Paste
		case tea.KeyRunes, tea.KeySpace: // input regular characters
			if msg.Alt && len(msg.Runes) == 1 {
				if msg.Runes[0] == 'd' { // alt+d, delete word right of cursor
					m.deleteWordRight()
					break
				}
				if msg.Runes[0] == 'b' { // alt+b, back one word
					m.wordLeft()
					break
				}
				if msg.Runes[0] == 'f' { // alt+f, forward one word
					m.wordRight()
					break
				}
			}

			// Input a regular character
			if m.CharLimit <= 0 || len(m.value) < m.CharLimit {
				runes := msg.Runes

				value := make([]rune, len(m.value))
				copy(value, m.value)
				value = append(value[:m.pos], append(runes, value[m.pos:]...)...)
				m.SetValue(string(value))
				if m.Err == nil {
					m.SetCursor(m.pos + len(runes))
				}
			}
		}

	case pasteMsg:
		m.handlePaste(string(msg))

	case pasteErrMsg:
		m.Err = msg
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.Cursor, cmd = m.Cursor.Update(msg)
	cmds = append(cmds, cmd)

	if oldPos != m.pos {
		m.Cursor.Blink = false
		cmds = append(cmds, m.Cursor.BlinkCmd())
	}

	m.handleOverflow()
	return m, tea.Batch(cmds...)
}

// View renders the textinput in its current state.
func (m Model) View() string {
	// Placeholder text
	if len(m.value) == 0 && m.Placeholder != "" {
		return m.placeholderView()
	}

	styleText := m.TextStyle.Inline(true).Render

	value := m.value[m.offset:m.offsetRight]
	pos := max(0, m.pos-m.offset)
	v := styleText(m.echoTransform(string(value[:pos])))

	if pos < len(value) {
		char := m.echoTransform(string(value[pos]))
		m.Cursor.SetChar(char)
		v += m.Cursor.View()                                   // cursor and text under it
		v += styleText(m.echoTransform(string(value[pos+1:]))) // text after cursor
	} else {
		m.Cursor.SetChar(" ")
		v += m.Cursor.View()
	}

	// If a max width and background color were set fill the empty spaces with
	// the background color.
	valWidth := rw.StringWidth(string(value))
	if m.Width > 0 && valWidth <= m.Width {
		padding := max(0, m.Width-valWidth)
		if valWidth+padding <= m.Width && pos < len(value) {
			padding++
		}
		v += styleText(strings.Repeat(" ", padding))
	}

	return m.PromptStyle.Render(m.Prompt) + v
}

// placeholderView returns the prompt and placeholder view, if any.
func (m Model) placeholderView() string {
	var (
		v     string
		p     = m.Placeholder
		style = m.PlaceholderStyle.Inline(true).Render
	)

	m.Cursor.TextStyle = m.PlaceholderStyle
	m.Cursor.SetChar(p[:1])
	v += m.Cursor.View()

	// The rest of the placeholder text
	v += style(p[1:])

	return m.PromptStyle.Render(m.Prompt) + v
}

// Blink is a command used to initialize cursor blinking.
func Blink() tea.Msg {
	return cursor.Blink()
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

// Deprecated.

// Deprecated: use cursor.Mode.
type CursorMode int

const (
	// Deprecated: use cursor.CursorBlink.
	CursorBlink = CursorMode(cursor.CursorBlink)
	// Deprecated: use cursor.CursorStatic.
	CursorStatic = CursorMode(cursor.CursorStatic)
	// Deprecated: use cursor.CursorHide.
	CursorHide = CursorMode(cursor.CursorHide)
)

func (c CursorMode) String() string {
	return cursor.Mode(c).String()
}

// Deprecated: use cursor.Mode().
func (m Model) CursorMode() CursorMode {
	return CursorMode(m.Cursor.Mode())
}

// Deprecated: use cursor.SetMode().
func (m *Model) SetCursorMode(mode CursorMode) tea.Cmd {
	return m.Cursor.SetMode(cursor.Mode(mode))
}
