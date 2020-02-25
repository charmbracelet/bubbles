package input

import (
	"time"

	"github.com/charmbracelet/tea"

	"github.com/muesli/termenv"
)

var (
	// Helper for returning colors
	color func(s string) termenv.Color = termenv.ColorProfile().Color
)

type Model struct {
	Prompt           string
	Value            string
	Cursor           string
	BlinkSpeed       time.Duration
	Placeholder      string
	TextColor        string
	PlaceholderColor string
	CursorColor      string

	// Focus indicates whether user input focus should be on this input
	// component. When false, don't blink and ignore keyboard input.
	focus bool

	blink bool
	pos   int
}

// Focused returns the focus state on the model
func (m Model) Focused() bool {
	if m.focus {
		return true
	}
	return false
}

// Focus sets the focus state on the model
func (m *Model) Focus() {
	m.focus = true
	m.blink = false
}

// Blur removes the focus state on the model
func (m *Model) Blur() {
	m.focus = false
	m.blink = true
}

// colorText colorizes a given string according to the TextColor value of the
// model
func (m *Model) colorText(s string) string {
	return termenv.
		String(s).
		Foreground(color(m.TextColor)).
		String()
}

// colorPlaceholder colorizes a given string according to the TextColor value
// of the model
func (m *Model) colorPlaceholder(s string) string {
	return termenv.
		String(s).
		Foreground(color(m.PlaceholderColor)).
		String()
}

type CursorBlinkMsg struct{}

func DefaultModel() Model {
	return Model{
		Prompt:           "> ",
		Value:            "",
		BlinkSpeed:       time.Millisecond * 600,
		Placeholder:      "",
		TextColor:        "",
		PlaceholderColor: "240",
		CursorColor:      "",

		focus: false,
		blink: true,
		pos:   0,
	}
}

func Update(msg tea.Msg, m Model) (Model, tea.Cmd) {
	if !m.focus {
		m.blink = true
		return m, nil
	}

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyBackspace:
			fallthrough
		case tea.KeyDelete:
			if len(m.Value) > 0 {
				m.Value = m.Value[:m.pos-1] + m.Value[m.pos:]
				m.pos--
			}
			return m, nil
		case tea.KeyLeft:
			if m.pos > 0 {
				m.pos--
			}
			return m, nil
		case tea.KeyRight:
			if m.pos < len(m.Value) {
				m.pos++
			}
			return m, nil
		case tea.KeyCtrlF: // ^F, forward one character
			fallthrough
		case tea.KeyCtrlB: // ^B, back one charcter
			fallthrough
		case tea.KeyCtrlA: // ^A, beginning
			m.pos = 0
			return m, nil
		case tea.KeyCtrlD: // ^D, delete char under cursor
			if len(m.Value) > 0 && m.pos < len(m.Value) {
				m.Value = m.Value[:m.pos] + m.Value[m.pos+1:]
			}
			return m, nil
		case tea.KeyCtrlE: // ^E, end
			m.pos = len(m.Value)
			return m, nil
		case tea.KeyCtrlK: // ^K, kill text after cursor
			m.Value = m.Value[:m.pos]
			m.pos = len(m.Value)
			return m, nil
		case tea.KeyCtrlU: // ^U, kill text before cursor
			m.Value = m.Value[m.pos:]
			m.pos = 0
			return m, nil
		case tea.KeyRune:
			m.Value = m.Value[:m.pos] + string(msg.Rune) + m.Value[m.pos:]
			m.pos++
			return m, nil
		default:
			return m, nil
		}

	case CursorBlinkMsg:
		m.blink = !m.blink
		return m, nil

	default:
		return m, nil
	}
}

func View(model tea.Model) string {
	m, _ := model.(Model)

	// Placeholder text
	if m.Value == "" && m.Placeholder != "" {
		return placeholderView(m)
	}

	v := m.colorText(m.Value[:m.pos])

	if m.pos < len(m.Value) {
		v += cursorView(string(m.Value[m.pos]), m)
		v += m.colorText(m.Value[m.pos+1:])
	} else {
		v += cursorView(" ", m)
	}
	return m.Prompt + v
}

func placeholderView(m Model) string {
	var (
		v string
		p = m.Placeholder
	)

	// Cursor
	if m.blink && m.PlaceholderColor != "" {
		v += cursorView(
			m.colorPlaceholder(p[:1]),
			m,
		)
	} else {
		v += cursorView(p[:1], m)
	}

	// The rest of the palceholder text
	v += m.colorPlaceholder(p[1:])

	return m.Prompt + v
}

// Style the cursor
func cursorView(s string, m Model) string {
	if m.blink {
		return s
	} else {
		return termenv.String(s).
			Foreground(color(m.CursorColor)).
			Reverse().
			String()
	}
}

// Subscription
func Blink(model tea.Model) tea.Msg {
	m, ok := model.(Model)
	if !ok {
		return tea.NewErrMsg("could not assert given model to the model we expected; make sure you're passing as input model")
	}
	time.Sleep(m.BlinkSpeed)
	return CursorBlinkMsg{}
}
