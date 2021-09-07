// Package stopwatch provides a simple stopwatch component.
package stopwatch

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TickMsg is a message that is sent on every timer tick.
type TickMsg struct{}

type startStopMsg struct {
	running bool
}

type resetMsg struct{}

// Model of the timer component.
type Model struct {
	d time.Duration

	running bool

	// How long to wait before every tick. Defaults to 1 second.
	Interval time.Duration
}

// NewWithInterval creates a new stopwatch with the given timeout and tick interval.
func NewWithInterval(interval time.Duration) Model {
	return Model{
		Interval: interval,
	}
}

// New creates a new stopwatch with 1s interval.
func New() Model {
	return NewWithInterval(time.Second)
}

// Init starts the stopwatch..
func (m Model) Init() tea.Cmd {
	return m.Start()
}

// Start starts the stopwatch.
func (m Model) Start() tea.Cmd {
	return tea.Batch(func() tea.Msg {
		return startStopMsg{true}
	}, tick(m.Interval))
}

// Stop stops the stopwatch.
func (m Model) Stop() tea.Cmd {
	return func() tea.Msg {
		return startStopMsg{false}
	}
}

// Toggle stops the stopwatch if it is running and starts it if it is stopped.
func (m Model) Toggle() tea.Cmd {
	if m.Running() {
		return m.Stop()
	}
	return m.Start()
}

// Reset restes the stopwatch to 0.
func (m Model) Reset() tea.Cmd {
	return func() tea.Msg {
		return resetMsg{}
	}
}

// Running returns true if the stopwatch is running or false if it is stopped.
func (m Model) Running() bool {
	return m.running
}

// Update handles the timer tick.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case startStopMsg:
		m.running = msg.running
	case resetMsg:
		m.d = 0
	case TickMsg:
		if !m.running {
			break
		}
		m.d += m.Interval
		return m, tick(m.Interval)
	}

	return m, nil
}

// Elapsed returns the time elapsed.
func (m Model) Elapsed() time.Duration {
	return m.d
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
