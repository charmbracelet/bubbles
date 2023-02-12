package router

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Screen represents a model with a particular binding to trigger it
type Screen struct {
	model   tea.Model
	binding key.Binding
}

// Model stores different screens and the currently focused screens
type Model struct {
	screens []Screen
	current int
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
	m.screens = append(m.screens, Screen{model, binding})
}

// Init implements tea.Model
func (Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		for i, screen := range m.screens {
			if key.Matches(msg, screen.binding) {
				m.current = i
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
