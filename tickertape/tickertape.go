package tickertape

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the state of the ticker tape.
type Model struct {
	Text        string // The Text to be displayed in the ticker tape.
	Position    int    // The current Position of the ticker tape.
	TickerWidth int    // The TickerWidth of the ticker tape display.
}

// tickMsg is a message used to trigger the ticker tape update.
type tickMsg struct{}

// Init initializes the ticker tape model and starts the ticking process.
func (m *Model) Init() tea.Cmd {
	return m.tick()
}

// tick returns a command that sends a tickMsg after a specified duration.
func (m *Model) tick() tea.Cmd {
	return tea.Tick(time.Millisecond*150, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

// Update handles incoming messages and updates the ticker tape model accordingly.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.TickerWidth = msg.Width // Update the TickerWidth of the ticker tape display.
		return m, nil
	case tickMsg:
		m.Position = (m.Position + 1) % len(m.Text) // Update the Position of the ticker tape.
		return m, m.tick()                          // Schedule the next tick.
	}
	return m, nil
}

// View renders the ticker tape view.
func (m *Model) View() string {
	ticker := m.Text[m.Position:] + m.Text[:m.Position]

	// Get the actual displayable TickerWidth
	displayWidth := m.TickerWidth
	if displayWidth < len(ticker) {
		ticker = ticker[:displayWidth]
	}

	return ticker
}

// UpdateText updates the Text of the ticker tape.
func (m *Model) UpdateText(newText string) {
	m.Text = newText
	m.Position = 0 // Reset Position to start.
}

// UpdateWidth updates the TickerWidth of the ticker tape.
func (m *Model) UpdateWidth(newWidth int) {
	m.TickerWidth = newWidth
}
