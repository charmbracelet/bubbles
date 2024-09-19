package picker

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	state State

	keys KeyMap
}

type State interface {
	GetValue() interface{}
	GetDisplayValue() string
	Next()
	Prev()
}

func NewModel(state State, opts ...func(*Model)) Model {
	m := Model{
		state: state,
		keys:  DefaultKeyMap(),
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, m.keys.Next):
			m.state.Next()
		case key.Matches(msg, m.keys.Prev):
			m.state.Prev()
		}
	}

	return m, nil
}

func (m Model) View() string {
	var output string

	output += fmt.Sprintf("< %v >", m.state.GetDisplayValue())

	return output
}

// Model Options --------------------

func WithKeys(keys KeyMap) func(*Model) {
	return func(m *Model) {
		m.keys = keys
	}
}
