// package router provides a simple router for bubbletea applications
// The router is based on a history stack (similar to browser history)
// and allows for navigating between screens both programmatically and using keybindings too.
package router

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Screen represents a model with a particular binding to trigger it
type Screen struct {
	model       tea.Model
	binding     key.Binding
	initialized bool
}

// Model stores different screens and the currently focused screens
type Model struct {
	screens     map[string]Screen
	initialPath string
	history     []string
}

// New creates a new empty Model
func New(initialPath string) Model {
	return Model{initialPath: initialPath}
}

// NewWithScreens creates a new model with an array of Screen
func NewWithScreens(screens map[string]Screen, initialPath string) Model {
	history := make([]string, 0)
	history = append(history, initialPath)
	return Model{
		screens,
		initialPath,
		history,
	}
}

func (m Model) current() string {
	return m.history[len(m.history)-1]
}

func (m Model) updateCurrent(msg tea.Msg) (Model, tea.Cmd) {
	if _, ok := m.screens[m.current()]; !ok {
		return m, nil
	}
	var cmd tea.Cmd
	screen := m.screens[m.current()]
	screen.model, cmd = m.screens[m.current()].model.Update(msg)
	m.screens[m.current()] = screen
	return m, cmd
}

// AddScreen adds a new screen to the router
func (m *Model) AddScreen(model tea.Model, path string, binding key.Binding) {
	initialized := false
	m.screens[path] = Screen{model, binding, initialized}
}

// Navigates to a screen by path (replacing top of history)
func (m *Model) NavigateTo(path string) tea.Cmd {
	m.history[len(m.history)-1] = path
	return m.initCurrent()
}

// Navigates to a screen by path (pushing an element to history)
func (m *Model) Push(path string) tea.Cmd {
	m.history = append(m.history, path)
	return m.initCurrent()
}

func (m *Model) Pop() tea.Cmd {
	if len(m.history) == 1 {
		return nil
	}
	m.history = m.history[:len(m.history)-1]
	return m.initCurrent()
}

// initCurrent sets the current screen to the given integer and initializes the screen if not already initailized
func (m *Model) initCurrent() tea.Cmd {
	if _, ok := m.screens[m.current()]; !ok {
		return nil
	}
	if !m.screens[m.current()].initialized {
		currentScreen := m.screens[m.current()]
		currentScreen.initialized = true
		m.screens[m.current()] = currentScreen
		return m.screens[m.current()].model.Init()
	}
	return nil
}

// Init implements tea.Model
func (m *Model) Init() tea.Cmd {
	return m.NavigateTo(m.initialPath)
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		for path, screen := range m.screens {
			if key.Matches(msg, screen.binding) {
				cmd := m.NavigateTo(path)
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
	return m.screens[m.current()].model.View()
}
