package picker

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	state          State
	showIndicators bool
	canCycle       bool
	displayFunc    DisplayFunc
	keys           KeyMap
}

type State interface {
	GetValue() interface{}
	Next(canCycle bool)
	Prev(canCycle bool)
}

type DisplayFunc func(interface{}) string

func NewModel(state State, opts ...func(*Model)) Model {
	defaultDisplayFunc := func(v interface{}) string {
		return fmt.Sprintf("%v", v)
	}

	m := Model{
		state:          state,
		showIndicators: true,
		canCycle:       false,
		displayFunc:    defaultDisplayFunc,
		keys:           DefaultKeyMap(),
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
			m.state.Next(m.canCycle)
		case key.Matches(msg, m.keys.Prev):
			m.state.Prev(m.canCycle)
		}
	}

	return m, nil
}

func (m Model) View() string {
	var leftIndicator, rightIndicator string
	if m.showIndicators {
		leftIndicator = "<"
		rightIndicator = ">"
	}

	var output string

	output += fmt.Sprintf("%s%s%s", leftIndicator, m.GetDisplayValue(), rightIndicator)

	return output
}

func (m Model) GetValue() interface{} {
	return m.state.GetValue()
}

func (m Model) GetDisplayValue() string {
	return m.displayFunc(m.state.GetValue())
}

// Model Options --------------------

func WithKeys(keys KeyMap) func(*Model) {
	return func(m *Model) {
		m.keys = keys
	}
}

func WithoutIndicators() func(*Model) {
	return func(m *Model) {
		m.showIndicators = false
	}
}

func WithCycles() func(*Model) {
	return func(m *Model) {
		m.canCycle = true
	}
}

func WithDisplayFunc(df DisplayFunc) func(*Model) {
	return func(m *Model) {
		m.displayFunc = df
	}
}
