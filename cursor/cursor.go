package cursor

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const defaultBlinkSpeed = time.Millisecond * 530

// initialBlinkMsg initializes cursor blinking.
type initialBlinkMsg struct{}

// BlinkMsg signals that the cursor should blink. It contains metadata that
// allows us to tell if the blink message is the one we're expecting.
type BlinkMsg struct {
	id  int
	tag int
}

// blinkCanceled is sent when a blink operation is canceled.
type blinkCanceled struct{}

// blinkCtx manages cursor blinking.
type blinkCtx struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// Mode describes the behavior of the cursor.
type Mode int

// Available cursor modes.
const (
	CursorBlink Mode = iota
	CursorStatic
	CursorHide
)

// String returns the cursor mode in a human-readable format. This method is
// provisional and for informational purposes only.
func (c Mode) String() string {
	return [...]string{
		"blink",
		"static",
		"hidden",
	}[c]
}

// Model is the Bubble Tea model for this cursor element.
type Model struct {
	BlinkSpeed time.Duration
	// Style for styling the cursor block.
	Style lipgloss.Style
	// TextStyle is the style used for the cursor when it is hidden (when blinking).
	// I.e. displaying normal text.
	TextStyle lipgloss.Style

	// char is the character under the cursor
	char string
	// The ID of this Model as it relates to other cursors
	id int
	// focus indicates whether the containing input is focused
	focus bool
	// Cursor Blink state.
	Blink bool
	// Used to manage cursor blink
	blinkCtx *blinkCtx
	// The ID of the blink message we're expecting to receive.
	blinkTag int
	// mode determines the behavior of the cursor
	mode Mode
}

// New creates a new model with default settings.
func New() Model {
	return Model{
		BlinkSpeed: defaultBlinkSpeed,

		Blink: true,
		mode:  CursorBlink,

		blinkCtx: &blinkCtx{
			ctx: context.Background(),
		},
	}
}

// Update updates the cursor.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case initialBlinkMsg:
		// We accept all initialBlinkMsgs generated by the Blink command.

		if m.mode != CursorBlink || !m.focus {
			return m, nil
		}

		cmd := m.BlinkCmd()
		return m, cmd

	case BlinkMsg:
		// We're choosy about whether to accept blinkMsgs so that our cursor
		// only exactly when it should.

		// Is this model blink-able?
		if m.mode != CursorBlink || !m.focus {
			return m, nil
		}

		// Were we expecting this blink message?
		if msg.id != m.id || msg.tag != m.blinkTag {
			return m, nil
		}

		var cmd tea.Cmd
		if m.mode == CursorBlink {
			m.Blink = !m.Blink
			cmd = m.BlinkCmd()
		}
		return m, cmd

	case blinkCanceled: // no-op
		return m, nil
	}
	return m, nil
}

// Mode returns the model's cursor mode. For available cursor modes, see
// type Mode.
func (m Model) Mode() Mode {
	return m.mode
}

// SetMode sets the model's cursor mode. This method returns a command.
//
// For available cursor modes, see type CursorMode.
func (m *Model) SetMode(mode Mode) tea.Cmd {
	// Adjust the mode value if it's less than 0.
	if mode < CursorBlink {
		mode = CursorBlink
	}
	// Adjust the mode value if it's greater than 2.
	if mode > CursorHide {
		mode = CursorHide
	}
	m.mode = mode
	m.Blink = m.mode == CursorHide || !m.focus
	if mode == CursorBlink {
		return Blink
	}
	return nil
}

// BlinkCmd is a command used to manage cursor blinking.
func (m *Model) BlinkCmd() tea.Cmd {
	if m.mode != CursorBlink {
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
			return BlinkMsg{id: m.id, tag: m.blinkTag}
		}
		return blinkCanceled{}
	}
}

// Blink is a command used to initialize cursor blinking.
func Blink() tea.Msg {
	return initialBlinkMsg{}
}

// Focus focuses the cursor to allow it to blink if desired.
func (m *Model) Focus() tea.Cmd {
	m.focus = true
	m.Blink = m.mode == CursorHide // show the cursor unless we've explicitly hidden it

	if m.mode == CursorBlink && m.focus {
		return m.BlinkCmd()
	}
	return nil
}

// Blur blurs the cursor.
func (m *Model) Blur() {
	m.focus = false
	m.Blink = true
}

// SetChar sets the character under the cursor.
func (m *Model) SetChar(char string) {
	m.char = char
}

// View displays the cursor.
func (m Model) View() string {
	if m.Blink {
		return m.TextStyle.Inline(true).Render(m.char)
	}
	return m.Style.Inline(true).Reverse(true).Render(m.char)
}
