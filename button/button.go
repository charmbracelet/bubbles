package button

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

var (
	// color is a helper for returning colors.
	color func(s string) termenv.Color = termenv.ColorProfile().Color
)

// Model is the Bubble Tea model for a button element.
type Model struct {
	Err error

	Label   string
	Default bool

	TextColor              string
	BackgroundColor        string
	FocusedTextColor       string
	FocusedBackgroundColor string

	// Focus indicates whether user focus should be on this button component
	focus bool
}

// NewModel creates a new model with default settings.
func NewModel() Model {
	return Model{
		Label: "Button",
	}
}

// Update is the Tea update loop.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focus {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// TODO: implement
		}
	}

	return m, nil
}

// View renders the button in its current state.
func (m Model) View() string {
	margin := m.styled("  ").String()
	label := m.styled(m.Label)
	if m.Default {
		label = label.Underline()
	}

	return margin + label.String() + margin
}

// Focused returns the focus state on the model.
func (m Model) Focused() bool {
	return m.focus
}

// Focus sets the focus state on the model.
func (m *Model) Focus() {
	m.focus = true
}

// Blur removes the focus state on the model.
func (m *Model) Blur() {
	m.focus = false
}

func (m Model) styled(s string) termenv.Style {
	view := termenv.String(s)
	if m.focus {
		view = view.Foreground(color(m.FocusedTextColor)).
			Background(color(m.FocusedBackgroundColor))
	} else {
		view = view.Foreground(color(m.TextColor)).
			Background(color(m.BackgroundColor))
	}

	return view
}
