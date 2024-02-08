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

// PositionType represents the current position (Date, Month, Year, Hour, or Minute).
type PositionType int

const (
	// Date represents the position type for selecting the date.
	Date PositionType = iota

	// Month represents the position type for selecting the month.
	Month

	// Year represents the position type for selecting the year.
	Year

	// Hour represents the position type for selecting the hour.
	Hour

	// Minute represents the position type for selecting the minute.
	Minute
)

// TimeFormat represents the time format (12-hour or 24-hour).
type TimeFormat int

const (
	// Hour12 represents the 12-hour time format.
	Hour12 TimeFormat = iota

	// Hour24 represents the 24-hour time format.
	Hour24
)

// PickerType represents the selection type (Date, Time, or Both).
type PickerType int

const (
	// DateTime represents the picker type for selecting both date and time.
	DateTime PickerType = iota

	// DateOnly represents the picker type for selecting only the date.
	DateOnly

	// TimeOnly represents the picker type for selecting only the time.
	TimeOnly
)

// Model is the Bubble Tea model for the date input element.
type Model struct {
	Err         error
	Prompt      string
	Date        time.Time
	PromptStyle lipgloss.Style
	TextStyle   lipgloss.Style
	CursorStyle lipgloss.Style
	Pos         PositionType
	TimeFormat  TimeFormat
	PickerType  PickerType
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
		TimeFormat:  Hour12,
		PickerType:  DateTime,
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
				prevDate := m.Date
				m.Date = m.Date.Add(time.Hour) // Increase the Hour by 1
				if prevDate.Day() != m.Date.Day() || prevDate.Month() != m.Date.Month() || prevDate.Year() != m.Date.Year() {
					m.Date = m.Date.AddDate(0, 0, -1) // Decrease the date by 1
				}
			}
			if m.Pos == Minute {
				prevDate := m.Date
				m.Date = m.Date.Add(time.Minute) // Increase the minute by 1
				if prevDate.Day() != m.Date.Day() || prevDate.Month() != m.Date.Month() || prevDate.Year() != m.Date.Year() {
					m.Date = m.Date.AddDate(0, 0, -1) // Decrease the date by 1
				}
			}

		case key.Matches(msg, m.KeyMap.Decrement):
			if m.Pos == Date {
				if m.Date.After(time.Date(0, time.January, 1, 23, 59, 0, 0, time.UTC)) { // Date : 1 JAN 0000 (Avoid negative year)
					m.Date = m.Date.AddDate(0, 0, -1) // Decrease the date by 1
				}
			}
			if m.Pos == Month {
				if m.Date.After(time.Date(0, time.January, 31, 23, 59, 0, 0, time.UTC)) { // Date : 31 JAN 0000 (Avoid negative year)
					m.Date = m.Date.AddDate(0, -1, 0) // Decrease the month by 1
				}
			}
			if m.Pos == Year {
				if m.Date.Year() > 0 {
					m.Date = m.Date.AddDate(-1, 0, 0) // Decrease the year by 1
				}
			}
			if m.Pos == Hour {
				prevDate := m.Date
				m.Date = m.Date.Add(-time.Hour) // Decrease the Hour by 1
				if prevDate.Day() != m.Date.Day() || prevDate.Month() != m.Date.Month() || prevDate.Year() != m.Date.Year() {
					m.Date = m.Date.AddDate(0, 0, 1) // Increase the date by 1
				}
			}
			if m.Pos == Minute {
				prevDate := m.Date
				m.Date = m.Date.Add(-time.Minute) // Decrease the minute by 1
				if prevDate.Day() != m.Date.Day() || prevDate.Month() != m.Date.Month() || prevDate.Year() != m.Date.Year() {
					m.Date = m.Date.AddDate(0, 0, 1) // Increase the date by 1
				}
			}

		case key.Matches(msg, m.KeyMap.Forward):
			lastPos := Minute
			if m.PickerType == DateOnly {
				lastPos = Year
			}
			if m.Pos < lastPos {
				m.Pos++
			}

		case key.Matches(msg, m.KeyMap.Backward):
			firstPos := Date
			if m.PickerType == TimeOnly {
				firstPos = Hour
			}
			if m.Pos > firstPos {
				m.Pos--
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	return m, nil
}

// View renders the date input in its current state.
func (m Model) View() string {
	// Apply styles
	prompt := m.PromptStyle.Render(m.Prompt)

	text := ""
	if m.PickerType == DateTime {
		text += m.dateView()
		text += " | "
		text += m.timeView()
	} else if m.PickerType == DateOnly {
		text += m.dateView()
	} else {
		text += m.timeView()
	}

	return prompt + text
}

func (m Model) dateView() string {
	// Customize styles based on the current position
	var (
		dayStyle   = m.TextStyle
		monthStyle = m.TextStyle
		yearStyle  = m.TextStyle
	)

	if m.Pos == Date {
		dayStyle = m.CursorStyle
	} else if m.Pos == Month {
		monthStyle = m.CursorStyle
	} else if m.Pos == Year {
		yearStyle = m.CursorStyle
	}

	day := m.Date.Day()
	month := m.Date.Month().String()
	year := m.Date.Year()

	// Format the date components
	dayText := fmt.Sprintf("%02d", day)
	yearText := fmt.Sprintf("%04d", year)

	return dayStyle.Render(dayText) + " " + monthStyle.Render(month) + " " + yearStyle.Render(yearText)
}

// formatTime formats the time based on the specified format (12-hour or 24-hour).
func (m Model) timeView() string {
	var (
		hourStyle   = m.TextStyle
		minuteStyle = m.TextStyle
	)

	if m.Pos == Hour {
		hourStyle = m.CursorStyle
	} else if m.Pos == Minute {
		minuteStyle = m.CursorStyle
	}

	s := ""
	if m.TimeFormat == Hour12 {
		s = m.Date.Format("03:04 PM")
		return hourStyle.Render(s[:2]) + ":" + minuteStyle.Render(s[3:5]) + " " + m.TextStyle.Render(s[6:])
	}
	s = m.Date.Format("15:04")

	return hourStyle.Render(s[:2]) + ":" + minuteStyle.Render(s[3:5])
}

// SetValue sets the date value of the input.
func (m *Model) SetValue(date time.Time) {
	m.Date = date
}

// SetValue sets the TimeFormat.
func (m *Model) SetTimeFormat(format TimeFormat) {
	if format < 0 {
		format = 0
	} else if format > 1 {
		format = 1
	}
	m.TimeFormat = format
}

// SetPickerType sets the PickerType.
func (m *Model) SetPickerType(pickerType PickerType) {
	if pickerType < 0 {
		pickerType = 0
	} else if pickerType > 2 {
		pickerType = 2
	}
	m.PickerType = pickerType
	if pickerType == DateTime || pickerType == DateOnly {
		m.Pos = Date
	} else {
		m.Pos = Hour
	}
}

// Value returns the formatted date value as a string.
func (m Model) Value() string {
	if m.PickerType <= DateTime {
		return m.Date.Format("02 January 2006 03:04 PM")
	} else if m.PickerType >= TimeOnly {
		if m.TimeFormat <= 0 {
			return m.Date.Format("03:04 PM")
		} else if m.TimeFormat >= 1 {
			return m.Date.Format("15:04")
		}
	}
	return m.Date.Format("02 January 2006")
}

// bubbletea Init function.
func (m Model) Init() tea.Cmd {
	return nil
}
