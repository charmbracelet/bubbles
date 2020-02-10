package spinner

import (
	"errors"
	"time"

	"github.com/charmbracelet/tea"
)

// Spinner animation frames
var (
	simple = []string{"|", "/", "-", "\\"}
)

// Model contains the state for the spinner. Use NewModel to create new models
// rather than using Model as a struct literal.
type Model struct {
	FPS   int
	frame int
}

var assertionErr = errors.New("could not perform assertion on model to what the spinner expects. are you sure you passed the right value?")

// NewModel returns a model with default values
func NewModel() Model {
	return Model{
		FPS:   9,
		frame: 0,
	}
}

// TickMsg indicates that the timer has ticked and we should render a frame
type TickMsg struct{}

// Update is the Tea update function
func Update(msg tea.Msg, m Model) (Model, tea.Cmd) {
	switch msg.(type) {
	case TickMsg:
		m.frame++
		if m.frame >= len(simple) {
			m.frame = 0
		}
		return m, nil
	default:
		return m, nil
	}
}

// View renders the model's view
func View(model Model) string {
	if model.frame >= len(simple) {
		return "[error]"
	}
	return simple[model.frame]
}

// Sub is the subscription that allows the spinner to spin
func Sub(model tea.Model) tea.Msg {
	m, ok := model.(Model)
	if !ok {
		return assertionErr
	}
	time.Sleep(time.Second / time.Duration(m.FPS))
	return TickMsg{}
}
