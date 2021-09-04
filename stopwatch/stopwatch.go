// Package stopwatch provides a simple stopwatch component.
package stopwatch

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TickMsg is a message that is sent on every timer tick.
type TickMsg struct{}

// StopMsg is a message that can be send to stop the watch.
type StopMsg struct{}

// Model of the timer component.
type Model struct {
	d time.Duration

	// How long to wait before every tick. Defaults to 1 second.
	TickEvery time.Duration
}

// NewWithInterval creates a new timer with the given timeout and tick interval.
func NewWithInterval(interval time.Duration) Model {
	return Model{
		TickEvery: interval,
	}
}

// New creates a new timer with the given timeout and default 1s interval.
func New(timeout time.Duration) Model {
	return NewWithInterval(time.Second)
}

// Init starts the timer.
func (m Model) Init() tea.Cmd {
	return tick(m.TickEvery)
}

// Update handles the timer tick.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg.(type) {
	case TickMsg:
		m.d += m.TickEvery
		return m, tick(m.TickEvery)
	case StopMsg:
		return m, nil
	}

	return m, nil
}

// View of the timer component.
func (m Model) View() string {
	return m.d.String()
}

func tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(_ time.Time) tea.Msg {
		return TickMsg{}
	})
}
