// Package timer provides a simple timeout component.
package timer

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TickMsg is a message that is sent on every timer tick.
type TickMsg struct{}

// TimeoutMsg is a message that is sent once when the timer times out.
type TimeoutMsg struct{}

// Model of the timer component.
type Model struct {
	// How long until the timer expires.
	Timeout time.Duration

	// How long to wait before every tick. Defaults to 1 second.
	Interval time.Duration
}

// NewWithInterval creates a new timer with the given timeout and tick interval.
func NewWithInterval(timeout, interval time.Duration) Model {
	return Model{
		Timeout:  timeout,
		Interval: interval,
	}
}

// New creates a new timer with the given timeout and default 1s interval.
func New(timeout time.Duration) Model {
	return NewWithInterval(timeout, time.Second)
}

// Init starts the timer.
func (m Model) Init() tea.Cmd {
	return tick(m.Interval)
}

// Update handles the timer tick.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg.(type) {
	case TickMsg:
		m.Timeout -= m.Interval
		if m.Timeout <= 0 {
			return m, func() tea.Msg {
				return TimeoutMsg{}
			}
		}
		return m, tick(m.Interval)
	}

	return m, nil
}

// View of the timer component.
func (m Model) View() string {
	return m.Timeout.String()
}

func tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(_ time.Time) tea.Msg {
		return TickMsg{}
	})
}
