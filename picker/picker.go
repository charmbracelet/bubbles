package picker

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	State          State
	ShowIndicators bool
	CanCycle       bool
	DisplayFunc    DisplayFunc
	Keys           KeyMap
}

type State interface {
	GetValue() interface{}
	Next(canCycle bool)
	Prev(canCycle bool)
}

type DisplayFunc func(stateValue interface{}) string

func NewModel(state State, opts ...func(*Model)) Model {
	defaultDisplayFunc := func(v interface{}) string {
		return fmt.Sprintf("%v", v)
	}

	m := Model{
		State:          state,
		ShowIndicators: true,
		CanCycle:       false,
		DisplayFunc:    defaultDisplayFunc,
		Keys:           DefaultKeyMap(),
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
		case key.Matches(msg, m.Keys.Next):
			m.State.Next(m.CanCycle)
		case key.Matches(msg, m.Keys.Prev):
			m.State.Prev(m.CanCycle)
		}
	}

	return m, nil
}

func (m Model) View() string {
	var prevInd, nextInd string
	if m.ShowIndicators {
		prevInd = m.GetPrevIndicator()
		nextInd = m.GetNextIndicator()
	}

	return prevInd + m.GetDisplayValue() + nextInd
}

func (m Model) GetValue() interface{} {
	return m.State.GetValue()
}

func (m Model) GetDisplayValue() string {
	return m.DisplayFunc(m.State.GetValue())
}

func (m Model) GetPrevIndicator() string {
	return "<"
}

func (m Model) GetNextIndicator() string {
	return ">"
}

// Model Options --------------------

func WithKeys(keys KeyMap) func(*Model) {
	return func(m *Model) {
		m.Keys = keys
	}
}

func WithoutIndicators() func(*Model) {
	return func(m *Model) {
		m.ShowIndicators = false
	}
}

func WithCycles() func(*Model) {
	return func(m *Model) {
		m.CanCycle = true
	}
}

func WithDisplayFunc(df DisplayFunc) func(*Model) {
	return func(m *Model) {
		m.DisplayFunc = df
	}
}
