// Package timer provides a simple timeout component that ticks every second.
package timer

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg struct{}

// Model of the timer component.
type Model struct {
	// How long until the timer expires.
	Timeout time.Duration

	// What to do when the timer expires.
	OnTimeout func() tea.Cmd
}

// Init starts the timer.
func (m Model) Init() tea.Cmd {
	return m.tick()
}

// Update handles the timer tick.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tickMsg:
		m.Timeout -= time.Second
		if m.Timeout <= 0 {
			if m.OnTimeout != nil {
				return m, m.OnTimeout()
			}
			return m, nil
		}
		return m, m.tick()
	}

	return m, nil
}

// View of the timer component.
func (m Model) View() string {
	return m.Timeout.String()
}

func (m Model) tick() tea.Cmd {
	return tea.Tick(time.Second, func(_ time.Time) tea.Msg {
		return tickMsg{}
	})
}
