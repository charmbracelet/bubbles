package textinput

import (
	"strings"
	"time"
	"unicode"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	rw "github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"
)

const defaultBlinkSpeed = time.Millisecond * 600

const (
	// EchoNormal displays text as is. This is the default behavior.
	EchoNormal EchoMode = iota

	// EchoPassword displays the EchoCharacter mask instead of actual
	// characters.  This is commonly used for password fields.
	EchoPassword

	// EchoNone displays nothing as characters are entered. This is commonly
	// seen for password fields on the command line.
	EchoNone

	// EchoOnEdit
)

// EchoMode sets the input behavior of the text input field.
type EchoMode int

var (
	// color is a helper for returning colors.
	color func(s string) termenv.Color = termenv.ColorProfile().Color
)

// Model is the Bubble Tea model for this text input element.
type Model struct {
	Err error

	Prompt      string
	Placeholder string

	Cursor     string
	BlinkSpeed time.Duration

	TextColor        string
	BackgroundColor  string
	PlaceholderColor string
	CursorColor      string

	EchoMode      EchoMode
	EchoCharacter rune

	// CharLimit is the maximum amount of characters this input element will
	// accept. If 0 or less, there's no limit.
	CharLimit int

	// Width is the maximum number of characters that can be displayed at once.
	// It essentially treats the text field like a horizontally scrolling
	// viewport. If 0 or less this setting is ignored.
	Width int

	// Underlying text value.
	value []rune

	// Focus indicates whether user input focus should be on this input
	// component. When false, don't blink and ignore keyboard input.
	focus bool

	// Cursor blink state.
	blink bool

	// Cursor position.
	pos int

	// Used to emulate a viewport when width is set and the content is
	// overflowing.
	offset      int
	offsetRight int
}

// SetValue sets the value of the text input.
func (m *Model) SetValue(s string) {
	runes := []rune(s)
	if m.CharLimit > 0 && len(runes) > m.CharLimit {
		m.value = runes[:m.CharLimit]
	} else {
		m.value = runes
	}
	if m.pos > len(m.value) {
		m.pos = len(m.value)
	}
	m.handleOverflow()
}

// Value returns the value of the text input.
func (m Model) Value() string {
	return string(m.value)
}

// SetCursor start moves the cursor to the given position. If the position is
// out of bounds the cursor will be moved to the start or end accordingly.
func (m *Model) SetCursor(pos int) {
	m.pos = clamp(pos, 0, len(m.value))
	m.handleOverflow()
}

// CursorStart moves the cursor to the start of the field.
func (m *Model) CursorStart() {
	m.pos = 0
	m.handleOverflow()
}

// CursorEnd moves the cursor to the end of the field.
func (m *Model) CursorEnd() {
	m.pos = len(m.value)
	m.handleOverflow()
}

// Focused returns the focus state on the model.
func (m Model) Focused() bool {
	return m.focus
}

// Focus sets the focus state on the model.
func (m *Model) Focus() {
	m.focus = true
	m.blink = false
}

// Blur removes the focus state on the model.
func (m *Model) Blur() {
	m.focus = false
	m.blink = true
}

// Reset sets the input to its default state with no input.
func (m *Model) Reset() {
	m.value = nil
	m.pos = 0
	m.blink = false
}

// Paste pastes the contents of the clipboard into the text area (if supported).
func (m *Model) Paste() {
	pasteString, err := clipboard.ReadAll()
	if err != nil {
		m.Err = err
	}
	paste := []rune(pasteString)

	availSpace := m.CharLimit - len(m.value)

	// If the char limit's been reached cancel
	if m.CharLimit > 0 && availSpace <= 0 {
		return
	}

	// If there's not enough space to paste the whole thing cut the pasted
	// runes down so they'll fit
	if availSpace < len(paste) {
		paste = paste[:len(paste)-availSpace]
	}

	// Stuff before and after the cursor
	head := m.value[:m.pos]
	tailSrc := m.value[m.pos:]
	tail := make([]rune, len(tailSrc))
	copy(tail, tailSrc)

	// Insert pasted runes
	for _, r := range paste {
		head = append(head, r)
		availSpace--
		m.pos++
		if m.CharLimit > 0 && availSpace <= 0 {
			break
		}
	}

	// Put it all back together
	m.value = append(head, tail...)
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

// colorText colorizes a given string according to the TextColor value of the
// model.
func (m *Model) colorText(s string) string {
	return termenv.
		String(s).
		Foreground(color(m.TextColor)).
		Background(color(m.BackgroundColor)).
		String()
}

// colorPlaceholder colorizes a given string according to the TextColor value
// of the model.
func (m *Model) colorPlaceholder(s string) string {
	return termenv.
		String(s).
		Foreground(color(m.PlaceholderColor)).
		Background(color(m.BackgroundColor)).
		String()
}

func (m *Model) wordLeft() {
	if m.pos == 0 || len(m.value) == 0 {
		return
	}

	i := m.pos - 1

	for i >= 0 {
		if unicode.IsSpace(m.value[i]) {
			m.pos--
			i--
		} else {
			break
		}
	}

	for i >= 0 {
		if !unicode.IsSpace(m.value[i]) {
			m.pos--
			i--
		} else {
			break
		}
	}
}

func (m *Model) wordRight() {
	if m.pos >= len(m.value) || len(m.value) == 0 {
		return
	}

	i := m.pos

	for i < len(m.value) {
		if unicode.IsSpace(m.value[i]) {
			m.pos++
			i++
		} else {
			break
		}
	}

	for i < len(m.value) {
		if !unicode.IsSpace(m.value[i]) {
			m.pos++
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

// BlinkMsg is sent when the cursor should alternate it's blinking state.
type BlinkMsg struct{}

// NewModel creates a new model with default settings.
func NewModel() Model {
	return Model{
		Prompt:      "> ",
		Placeholder: "",

		BlinkSpeed: defaultBlinkSpeed,

		TextColor:        "",
		PlaceholderColor: "240",
		CursorColor:      "",

		EchoCharacter: '*',

		CharLimit: 0,

		value: nil,
		focus: false,
		blink: true,
		pos:   0,
	}
}

// Update is the Tea update loop.
func Update(msg tea.Msg, m Model) (Model, tea.Cmd) {
	if !m.focus {
		m.blink = true
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyBackspace: // delete character before cursor
			if len(m.value) > 0 {
				m.value = append(m.value[:max(0, m.pos-1)], m.value[m.pos:]...)
				if m.pos > 0 {
					m.pos--
				}
			}
		case tea.KeyLeft:
			if msg.Alt { // alt+left arrow, back one word
				m.wordLeft()
				break
			}
			if m.pos > 0 {
				m.pos--
			}
		case tea.KeyRight:
			if msg.Alt { // alt+right arrow, forward one word
				m.wordRight()
				break
			}
			if m.pos < len(m.value) {
				m.pos++
			}
		case tea.KeyCtrlF: // ^F, forward one character
			fallthrough
		case tea.KeyCtrlB: // ^B, back one charcter
			fallthrough
		case tea.KeyHome, tea.KeyCtrlA: // ^A, go to beginning
			m.CursorStart()
		case tea.KeyDelete, tea.KeyCtrlD: // ^D, delete char under cursor
			if len(m.value) > 0 && m.pos < len(m.value) {
				m.value = append(m.value[:m.pos], m.value[m.pos+1:]...)
			}
		case tea.KeyCtrlE, tea.KeyEnd: // ^E, go to end
			m.CursorEnd()
		case tea.KeyCtrlK: // ^K, kill text after cursor
			m.value = m.value[:m.pos]
			m.pos = len(m.value)
		case tea.KeyCtrlU: // ^U, kill text before cursor
			m.value = m.value[m.pos:]
			m.pos = 0
			m.offset = 0
		case tea.KeyCtrlV: // ^V paste
			m.Paste()
		case tea.KeyRune: // input a regular character

			if msg.Alt {
				if msg.Rune == 'b' { // alt+b, back one word
					m.wordLeft()
					break
				}
				if msg.Rune == 'f' { // alt+f, forward one word
					m.wordRight()
					break
				}
			}

			// Input a regular character
			if m.CharLimit <= 0 || len(m.value) < m.CharLimit {
				m.value = append(m.value[:m.pos], append([]rune{msg.Rune}, m.value[m.pos:]...)...)
				m.pos++
			}
		}

	case BlinkMsg:
		m.blink = !m.blink
		return m, Blink(m)
	}

	m.handleOverflow()

	return m, nil
}

// View renders the textinput in its current state.
func View(m Model) string {
	// Placeholder text
	if len(m.value) == 0 && m.Placeholder != "" {
		return placeholderView(m)
	}

	value := m.value[m.offset:m.offsetRight]
	pos := max(0, m.pos-m.offset)
	v := m.colorText(m.echoTransform(string(value[:pos])))

	if pos < len(value) {
		v += cursorView(m.echoTransform(string(value[pos])), m)  // cursor and text under it
		v += m.colorText(m.echoTransform(string(value[pos+1:]))) // text after cursor
	} else {
		v += cursorView(" ", m)
	}

	// If a max width and background color were set fill the empty spaces with
	// the background color.
	valWidth := rw.StringWidth(string(value))
	if m.Width > 0 && len(m.BackgroundColor) > 0 && valWidth <= m.Width {
		padding := max(0, m.Width-valWidth)
		if valWidth+padding <= m.Width && pos < len(value) {
			padding++
		}
		v += strings.Repeat(
			termenv.String(" ").Background(color(m.BackgroundColor)).String(),
			padding,
		)
	}

	return m.Prompt + v
}

// placeholderView.
func placeholderView(m Model) string {
	var (
		v string
		p = m.Placeholder
	)

	// Cursor
	if m.blink && m.PlaceholderColor != "" {
		v += cursorView(m.colorPlaceholder(p[:1]), m)
	} else {
		v += cursorView(p[:1], m)
	}

	// The rest of the placeholder text
	v += m.colorPlaceholder(p[1:])

	return m.Prompt + v
}

// cursorView styles the cursor.
func cursorView(s string, m Model) string {
	if m.blink {
		if m.TextColor != "" || m.BackgroundColor != "" {
			return termenv.String(s).
				Foreground(color(m.TextColor)).
				Background(color(m.BackgroundColor)).
				String()
		}
		return s
	}
	return termenv.String(s).
		Foreground(color(m.CursorColor)).
		Background(color(m.BackgroundColor)).
		Reverse().
		String()
}

// Blink is a command used to time the cursor blinking.
func Blink(model Model) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(model.BlinkSpeed)
		return BlinkMsg{}
	}
}

func clamp(v, low, high int) int {
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
