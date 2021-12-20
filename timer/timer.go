// Package timer provides a simple timeout component.
package timer

import (
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	lastID int
	idMtx  sync.Mutex
)

func nextID() int {
	idMtx.Lock()
	defer idMtx.Unlock()
	lastID++
	return lastID
}

// TickMsg is a message that is sent on every timer tick.
type TickMsg struct {
	// ID is the identifier of the stopwatch that send the message. This makes
	// it possible to determine which timer a tick belongs to when there
	// are multiple timers running.
	//
	// Note, however, that a timer will reject ticks from other stopwatches, so
	// it's safe to flow all TickMsgs through all timers and hvae them still
	// behave appropriately.
	ID int

	// Timeout returns whether or not this tick is a timeout tick. You can
	// alternatively listen for TimeoutMsg.
	Timeout bool
}

// TimeoutMsg is a message that is sent once when the timer times out.
//
// It's a convenience message sent alongside a TickMsg with the Timeout value
// set to true.
type TimeoutMsg struct {
	ID int
}

// Model of the timer component.
type Model struct {
	// How long until the timer expires.
	Timeout time.Duration

	// How long to wait before every tick. Defaults to 1 second.
	Interval time.Duration

	id      int
	running bool
}

// NewWithInterval creates a new timer with the given timeout and tick interval.
func NewWithInterval(timeout, interval time.Duration) Model {
	return Model{
		Timeout:  timeout,
		Interval: interval,
		running:  true,
	}
}

// New creates a new timer with the given timeout and default 1s interval.
func New(timeout time.Duration) Model {
	return NewWithInterval(timeout, time.Second)
}

// ID returns
func (m Model) ID() int {
	return m.id
}

// Running returns whether or not the timer is running. If the timer has timed
// out this will always return false.
func (m Model) Running() bool {
	if m.Timedout() || !m.running {
		return false
	}
	return true
}

// Timedout returns whether or not the timer has timed out.
func (m Model) Timedout() bool {
	return m.Timeout <= 0
}

// Init starts the timer.
func (m Model) Init() tea.Cmd {
	return tick(m.id, m.Interval, m.Timedout())
}

// Update handles the timer tick.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		if !m.Running() || (msg.ID != 0 && msg.ID != m.id) {
			break
		}

		m.Timeout -= m.Interval
		tickCmd := tick(m.id, m.Interval, m.Timedout())

		if m.Timedout() {
			return m, tea.Batch(tickCmd, timeout(m.id))
		}
		return m, tickCmd
	}

	return m, nil
}

// View of the timer component.
func (m Model) View() string {
	return m.Timeout.String()
}

// Start resumes the timer. Has no effect if the timer has timed out.
func (m *Model) Start() tea.Cmd {
	m.running = true
	if m.Timedout() {
		return nil
	}
	return tick(m.id, m.Interval, m.Timedout())
}

// Stop pauses the timer. Has no effect if the timer has timed out.
func (m *Model) Stop() tea.Cmd {
	m.running = false
	return func() tea.Msg {
		return nil
	}
}

// Toggle stops the timer if it's running and starts it if it's stopped.
func (m *Model) Toggle() tea.Cmd {
	if m.Running() {
		return m.Stop()
	}
	return m.Start()
}

func tick(id int, d time.Duration, timedout bool) tea.Cmd {
	return tea.Tick(d, func(_ time.Time) tea.Msg {
		return TickMsg{ID: id, Timeout: timedout}
	})
}

func timeout(id int) tea.Cmd {
	return func() tea.Msg {
		return TimeoutMsg{ID: id}
	}
}
