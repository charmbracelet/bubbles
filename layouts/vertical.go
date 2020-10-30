package layouts

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Model is the Bubble Tea model for a vertical layout element.
type Model struct {
	Index int
	Items []tea.Model

	// Focus indicates whether user focus should be on this component
	focus bool
}

type FocusItem interface {
	Focus() tea.Model
	Blur() tea.Model
}

// NewModel creates a new model with default settings.
func NewModel() Model {
	return Model{}
}

// Update is the Tea update loop.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focus {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "shift+tab", "up":
			m.Index--
			if m.Index < 0 {
				m.Index = len(m.Items) - 1
			}
			m.updateFocus()

		case "tab", "down":
			m.Index++
			if m.Index >= len(m.Items) {
				m.Index = 0
			}
			m.updateFocus()
		}
	}

	cmd := m.updateItems(msg)
	return m, cmd
}

// View renders the layout in its current state.
func (m Model) View() string {
	var view string

	for _, v := range m.Items {
		if mi, ok := v.(tea.Model); ok {
			view += mi.View() + "\n"
		}
	}

	return view
}

func (m *Model) updateFocus() {
	for i, v := range m.Items {
		if m.Index == i {
			if fi, ok := v.(FocusItem); ok {
				// new focused item
				m.Items[i] = fi.Focus()
			}
		} else {
			if fi, ok := v.(FocusItem); ok {
				m.Items[i] = fi.Blur()
			}
		}
	}
}

// Pass messages and models through to text input components. Only text inputs
// with Focus() set will respond, so it's safe to simply update all of them
// here without any further logic.
func (m *Model) updateItems(msg tea.Msg) tea.Cmd {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	for i, v := range m.Items {
		if mi, ok := v.(tea.Model); ok {
			m.Items[i], cmd = mi.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return tea.Batch(cmds...)
}

// Focused returns the focus state on the model.
func (m Model) Focused() bool {
	return m.focus
}

// Focus sets the focus state on the model.
func (m *Model) Focus() {
	m.focus = true
	m.updateFocus()
}

// Blur removes the focus state on the model.
func (m *Model) Blur() {
	m.focus = false
}
