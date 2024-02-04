package datetimepicker

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// KeyMap is the key bindings for different actions within the datetimepicker.
type KeyMap struct {
	Increment key.Binding
	Decrement key.Binding
	Forward   key.Binding
	Backward  key.Binding
	Quit      key.Binding
}

// DefaultKeyMap is the default set of key bindings for navigating and acting
// upon the datetimepicker.
var DefaultKeyMap = KeyMap{
	Increment: key.NewBinding(key.WithKeys("up")),
	Decrement: key.NewBinding(key.WithKeys("down")),
	Forward:   key.NewBinding(key.WithKeys("right")),
	Backward:  key.NewBinding(key.WithKeys("left")),
	Quit:      key.NewBinding(key.WithKeys("ctrl+c")),
}

// PositionType represents the current position (Date, Month, or Year)
type PositionType int

const (
	Date PositionType = iota
	Month
	Year
	Hour
	Minute
)

// Model is the Bubble Tea model for the date input element.
type Model struct {
	Err         error
	Prompt      string
	Date        time.Time
	Format      string
	PromptStyle lipgloss.Style
	TextStyle   lipgloss.Style
	CursorStyle lipgloss.Style
	Pos         PositionType
	// KeyMap encodes the keybindings.
	KeyMap KeyMap
}

// New creates a new model with default settings.
func New() Model {
	return Model{
		Prompt:      "> ",
		PromptStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("212")),
		TextStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("255")),
		CursorStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("212")),
		Pos:         Date,
		Date:        time.Now(),
		KeyMap:      DefaultKeyMap,
	}
}

// Update is the Bubble Tea update loop.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch {
		// case to exit the program.
		case key.Matches(msg, m.KeyMap.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.KeyMap.Increment):
			if m.Pos == Date {
				m.Date = m.Date.AddDate(0, 0, 1) // Increase the date by 1
			}
			if m.Pos == Month {
				m.Date = m.Date.AddDate(0, 1, 0) // Increase the month by 1
			}
			if m.Pos == Year {
				m.Date = m.Date.AddDate(1, 0, 0) // Increase the year by 1
			}
			if m.Pos == Hour {
				m.Date = m.Date.Add(time.Hour) // Increase the minute by 1
			}
			if m.Pos == Minute {
				m.Date = m.Date.Add(time.Minute) // Increase the minute by 1
			}

		case key.Matches(msg, m.KeyMap.Decrement):
			if m.Pos == Date {
				if m.Date.Year() <= 0 && m.Date.Month() <= time.January && m.Date.Day() <= 1 {
					// Avoid negative year
				} else {
					m.Date = m.Date.AddDate(0, 0, -1) // Decrease the date by 1
				}
			}
			if m.Pos == Month {
				if m.Date.Year() <= 0 && m.Date.Month() <= time.January {
					// Avoid negative year
				} else {
					m.Date = m.Date.AddDate(0, -1, 0) // Decrease the month by 1
				}
			}
			if m.Pos == Year {
				if m.Date.Year() > 0 {
					m.Date = m.Date.AddDate(-1, 0, 0) // Decrease the year by 1
				}
			}
			if m.Pos == Hour {
				m.Date = m.Date.Add(-time.Hour) // Decrease the minute by 1
			}
			if m.Pos == Minute {
				m.Date = m.Date.Add(-time.Minute) // Decrease the minute by 1
			}

		case key.Matches(msg, m.KeyMap.Forward):
			if m.Pos < Minute {
				m.Pos++
			}

		case key.Matches(msg, m.KeyMap.Backward):
			if m.Pos > Date {
				m.Pos--
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	return m, nil
}

// View renders the date input in its current state.
func (m Model) View() string {
	// Customize styles based on the current position
	var (
		dayStyle    = m.TextStyle
		monthStyle  = m.TextStyle
		yearStyle   = m.TextStyle
		hourStyle   = m.TextStyle
		minuteStyle = m.TextStyle
	)

	// Apply styles
	prompt := m.PromptStyle.Render(m.Prompt)

	switch m.Pos {
	case Date:
		dayStyle = m.CursorStyle
	case Month:
		monthStyle = m.CursorStyle
	case Year:
		yearStyle = m.CursorStyle
	case Hour:
		hourStyle = m.CursorStyle
	case Minute:
		minuteStyle = m.CursorStyle
	}

	day := m.Date.Day()
	month := m.Date.Month().String()
	year := m.Date.Year()

	// Format the date components
	dayText := fmt.Sprintf("%02d", day)
	yearText := fmt.Sprintf("%04d", year)
	timeText := m.Date.Format("03:04 PM")

	text := ""
	text += dayStyle.Render(dayText) + " " + monthStyle.Render(month) + " " + yearStyle.Render(yearText)
	text += " | "
	text += hourStyle.Render(timeText[:2]) + ":" + minuteStyle.Render(timeText[3:5]) + " " + m.TextStyle.Render(timeText[6:])
	return prompt + text
}

// SetValue sets the date value of the input.
func (m *Model) SetValue(date time.Time) {
	m.Date = date
}

// Value returns the formatted date value as a string.
func (m Model) Value() string {
	return m.Date.Format("02 January 2006 03:04 PM")
}

// bubbletea Init function
func (m Model) Init() tea.Cmd {
	return nil
}
