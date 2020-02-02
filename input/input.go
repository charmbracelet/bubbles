package input

import (
	"time"

	"github.com/charmbracelet/tea"

	"github.com/muesli/termenv"
)

type Model struct {
	Prompt           string
	Value            string
	Cursor           string
	BlinkSpeed       time.Duration
	Placeholder      string
	PlaceholderColor string

	blink        bool
	pos          int
	colorProfile termenv.Profile
}

type CursorBlinkMsg struct{}

func DefaultModel() Model {
	return Model{
		Prompt:           "> ",
		Value:            "",
		BlinkSpeed:       time.Millisecond * 600,
		Placeholder:      "",
		PlaceholderColor: "240",

		blink:        false,
		pos:          0,
		colorProfile: termenv.ColorProfile(),
	}
}

func Update(msg tea.Msg, m Model) (Model, tea.Cmd) {
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
		case tea.KeyCtrlF: // ^F, forward one character
			fallthrough
		case tea.KeyLeft:
			if m.pos > 0 {
				m.pos--
			}
			return m, nil
		case tea.KeyCtrlB: // ^B, back one charcter
			fallthrough
		case tea.KeyRight:
			if m.pos < len(m.Value) {
				m.pos++
			}
			return m, nil
		case tea.KeyCtrlA: // ^A, beginning
			m.pos = 0
			return m, nil
		case tea.KeyCtrlD: // ^D, delete char under cursor
			if len(m.Value) > 0 && m.pos < len(m.Value) {
				m.Value = m.Value[:m.pos] + m.Value[m.pos+1:]
			}
			return m, nil
		case tea.KeyCtrlE: // ^E, end
			m.pos = len(m.Value) - 1
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
			m.Value = m.Value[:m.pos] + msg.String() + m.Value[m.pos:]
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
		var v string

		if m.blink {
			v += cursor(
				termenv.String(m.Placeholder[:1]).
					Foreground(m.colorProfile.Color(m.PlaceholderColor)).
					String(),
				m.blink,
			)
		} else {
			v += cursor(m.Placeholder[:1], m.blink)
		}

		v += termenv.String(m.Placeholder[1:]).
			Foreground(m.colorProfile.Color(m.PlaceholderColor)).
			String()
		return m.Prompt + v
	}

	v := m.Value[:m.pos]

	if m.pos < len(m.Value) {
		v += cursor(string(m.Value[m.pos]), m.blink)
		v += m.Value[m.pos+1:]
	} else {
		v += cursor(" ", m.blink)
	}
	return m.Prompt + v
}

func placeholderView(m Model) string {
	var (
		v     string
		p     = m.Placeholder
		c     = m.PlaceholderColor
		color = m.colorProfile.Color
	)

	// Cursor
	if m.blink {
		v += cursor(
			termenv.String(p[:1]).
				Foreground(color(c)).
				String(),
			m.blink,
		)
	} else {
		v += cursor(p[:1], m.blink)
	}

	// The rest of the palceholder text
	v += termenv.String(p[1:]).
		Foreground(color(c)).
		String()

	return m.Prompt + v
}

// Style the cursor
func cursor(s string, blink bool) string {
	if blink {
		return s
	}
	return termenv.String(s).Reverse().String()
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
