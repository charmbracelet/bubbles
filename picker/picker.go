package picker

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model is a picker component.
// By default the View method will render it as a horizontal picker.
// Methods are exposed to get the value and indicators separately, allowing you to build your own UI (vertical, 4-dimensional, etc).
type Model struct {
	State          State
	ShowIndicators bool
	CanCycle       bool
	CanJump        bool
	StepSize       int
	DisplayFunc    DisplayFunc
	Keys           *KeyMap
	Styles         Styles
}

type State interface {
	GetValue() interface{}

	NextExists() bool
	PrevExists() bool

	Next(canCycle bool)
	Prev(canCycle bool)
	StepForward(size int)
	StepBackward(size int)
	JumpForward()
	JumpBackward()
}

type DisplayFunc func(stateValue interface{}) string

func New(state State, opts ...func(*Model)) Model {
	defaultDisplayFunc := func(v interface{}) string {
		return fmt.Sprintf("%v", v)
	}

	m := Model{
		State:          state,
		ShowIndicators: true,
		CanCycle:       false,
		CanJump:        false,
		StepSize:       10,
		DisplayFunc:    defaultDisplayFunc,
		Keys:           DefaultKeyMap(),
		Styles:         DefaultStyles(),
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

		case key.Matches(msg, m.Keys.StepForward):
			m.State.StepForward(m.StepSize)

		case key.Matches(msg, m.Keys.StepBackward):
			m.State.StepBackward(m.StepSize)

		case key.Matches(msg, m.Keys.JumpForward):
			if m.CanJump {
				m.State.JumpForward()
			}

		case key.Matches(msg, m.Keys.JumpBackward):
			if m.CanJump {
				m.State.JumpBackward()
			}
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

	value := m.Styles.Selection.Render(
		m.GetDisplayValue(),
	)

	return lipgloss.JoinHorizontal(lipgloss.Center,
		prevInd,
		value,
		nextInd,
	)
}

func (m Model) GetValue() interface{} {
	return m.State.GetValue()
}

func (m Model) GetDisplayValue() string {
	return m.DisplayFunc(m.State.GetValue())
}

func (m Model) GetPrevIndicator() string {
	return getIndicator(m.Styles.Previous, m.State.PrevExists())
}

func (m Model) GetNextIndicator() string {
	return getIndicator(m.Styles.Next, m.State.NextExists())
}

func getIndicator(styles IndicatorStyles, enabled bool) string {
	switch enabled {
	case false:
		return styles.Disabled.Render(styles.Value)
	default:
		return styles.Enabled.Render(styles.Value)
	}
}

// Model Options --------------------

func WithKeys(keys *KeyMap) func(*Model) {
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

func WithStyles(s Styles) func(*Model) {
	return func(m *Model) {
		m.Styles = s
	}
}

func WithJumping() func(*Model) {
	return func(m *Model) {
		m.CanJump = true
	}
}

func WithStepSize(size int) func(*Model) {
	return func(m *Model) {
		m.StepSize = size
	}
}
