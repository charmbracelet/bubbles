package router

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Screen represents a model with a particular binding to trigger it
type Screen struct {
	model   tea.Model
	binding key.Binding
  initialized bool
}

// Model stores different screens and the currently focused screens
type Model struct {
	screens     []Screen
	current     int
}

// New creates a new empty Model
func New() Model {
	return Model{}
}

// NewWithScreens creates a new model with an array of Screen
func NewWithScreens(screens []Screen) Model {
	current := 0
	return Model{
		screens,
		current,
	}
}

func (m Model) updateCurrent(msg tea.Msg) (Model, tea.Cmd) {
	if len(m.screens) <= m.current {
		return m, nil
	}
	var cmd tea.Cmd
	m.screens[m.current].model, cmd = m.screens[m.current].model.Update(msg)
	return m, cmd
}

// AddScreen adds a new screen to the router
func (m *Model) AddScreen(model tea.Model, binding key.Binding) {
  initialized := false
	m.screens = append(m.screens, Screen{model, binding, initialized})
}

// setCurrent sets the current screen to the given integer and initializes the screen if not already initailized
func (m *Model) setCurrent(current int) tea.Cmd {
	if len(m.screens) >= m.current {
		return nil
	}
	m.current = current
	if !m.screens[m.current].initialized {
		m.screens[m.current].initialized = true
		return m.screens[m.current].model.Init()
	}
	return nil
}

// Init implements tea.Model
func (m *Model) Init() tea.Cmd {
	return m.setCurrent(0) // assumes that there is atleast one screen
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		for i, screen := range m.screens {
			if key.Matches(msg, screen.binding) {
				cmd := m.setCurrent(i)
				if cmd != nil {
					return m, cmd
				}
				break
			}
		}
	}

	return m.updateCurrent(msg)
}

// View implements tea.Model
func (m Model) View() string {
	return m.screens[m.current].model.View()
}
