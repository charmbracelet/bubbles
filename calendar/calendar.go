/*
Calendar component

July                  August                September
Mo Tu We Th Fr Sa Su  Mo Tu We Th Fr Sa Su  Mo Tu We Th Fr Sa Su
             1  2  3   1  2  3  4  5  6  7            1  2  3  4
 4  5  6  7  8  9 10   8  9 10 11 12 13 14   5  6  7  8  9 10 11
11 12 13 14 15 16 17  15 16 17 18 19 20 21  12 13 14 15 16 17 18
18 19 20 21 22 23 24  22 23 24 25 26 27 28  19 20 21 22 23 24 25
25 26 27 28 29 30 31  29 30 31              26 27 28 29 30
*/

package calendar

import (
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const columnWidth = 3

// Model stores component.
type Model struct {
	CurrentDate      time.Time
	Styles           Styles
	Weekdays         []Weekday
	NbMonthDisplayed int
}

// Styles stores all component styles.
type Styles struct {
	CurrentDate    lipgloss.Style
	Date           lipgloss.Style
	WeekdaysHeader lipgloss.Style
}

// Weekday represents a full weekday name and its abbreviation.
type Weekday struct {
	Name         string
	Abbreviation string
}

// EnglishWeekdays.
var EnglishWeekdays = []Weekday{
	{Name: "Monday", Abbreviation: "Mo"},
	{Name: "Tuesday", Abbreviation: "Tu"},
	{Name: "Wednesday", Abbreviation: "We"},
	{Name: "Thursday", Abbreviation: "Th"},
	{Name: "Friday", Abbreviation: "Fr"},
	{Name: "Saturday", Abbreviation: "Sa"},
	{Name: "Sunday", Abbreviation: "Su"},
}

// NewModel initializes a calendar component with default values.
func NewModel() Model {
	return Model{
		CurrentDate: time.Now(),
		Styles: Styles{
			CurrentDate:    lipgloss.NewStyle().ColorWhitespace(false).Width(columnWidth).Align(lipgloss.Center).Background(lipgloss.Color("#7571F9")),
			Date:           lipgloss.NewStyle().ColorWhitespace(false).Width(columnWidth).Align(lipgloss.Center),
			WeekdaysHeader: lipgloss.NewStyle().ColorWhitespace(false).Width(columnWidth).Align(lipgloss.Left).Background(lipgloss.Color("#F25D94")),
		},
		Weekdays:         EnglishWeekdays,
		NbMonthDisplayed: 3,
	}
}

// Init method.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update method.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		}
	}

	return m, nil
}

// View method.
func (m Model) View() string {
	// Each month will have their days represented in a string array
	calendarMonthRender := make([][]string, m.NbMonthDisplayed)

	for monthIndex := range calendarMonthRender {
		firstDayOfMonth := time.Date(m.CurrentDate.Year(), m.CurrentDate.Month(), 1, 0, 0, 0, 0, m.CurrentDate.Location())

		// Use an index relative to the current month. 0 is current month, -1 is the month before, +1 is the month after etc.
		// This makes it easy to calculate first and last days for each months in our calendar
		monthRelativePosition := monthIndex - (len(calendarMonthRender) / 2)
		firstDayOfMonth = firstDayOfMonth.AddDate(0, monthRelativePosition, 0)
		lastDayOfMonth := firstDayOfMonth.AddDate(0, 1, -1)

		// Month name heading and days
		// July                  August                September
		// Mo Tu We Th Fr Sa Su  Mo Tu We Th Fr Sa Su  Mo Tu We Th Fr Sa Su
		s := ""
		s += firstDayOfMonth.Month().String()
		s += "\n"

		for i := 0; i < len(m.Weekdays); i++ {
			s += m.Styles.WeekdaysHeader.Render(m.Weekdays[i].Abbreviation)
		}
		s += "\n"

		// Determine 1st day in the month position in the week to complete with padding
		// Dashes represent the required offset padding (offset * weekday header width):
		// Mo Tu We Th Fr Sa Su
		// ------------ 1  2  3
		//  4  5  6  7  8  9 10
		monthStartingDayOffset := 0
		for i, weekday := range m.Weekdays {
			if weekday.Name == firstDayOfMonth.Weekday().String() {
				monthStartingDayOffset = i
				break
			}
		}
		s += strings.Repeat("   ", monthStartingDayOffset)

		for i := 1; i <= lastDayOfMonth.Day(); i++ {
			// Current selected day is highlighted
			if i == m.CurrentDate.Day() && monthRelativePosition == 0 {
				s += m.Styles.CurrentDate.Render(strconv.Itoa(i))
			} else {
				s += m.Styles.Date.Render(strconv.Itoa(i))
			}

			// Add a line return on week end to prepare new line
			// Except when the last day in the month ends on the last weekday
			if (i+monthStartingDayOffset)%7 == 0 && i != lastDayOfMonth.Day() {
				s += "\n"
			}
		}

		s += "\n"

		calendarMonthRender[monthIndex] = strings.Split(s, "\n")
	}

	s := ""
	maxNbLine := 0
	for _, month := range calendarMonthRender {
		if len(month) > maxNbLine {
			maxNbLine = len(month)
		}
	}

	// Render calendar by iterating over each month, line by line
	// For example if we have three months, our first line will look like this:
	//        1  2  3  4  5               1  2  3   1  2  3  4  5  6  7
	// The second line will look like this:
	//  6  7  8  9 10 11 12   4  5  6  7  8  9 10   8  9 10 11 12 13 14

	for i := 0; i < maxNbLine; i++ {
		for monthPosition, month := range calendarMonthRender {
			// Padding between months
			if monthPosition > 0 {
				s += " "
			}
			// Render calendar line
			if i < len(month) {
				s += lipgloss.NewStyle().Width(columnWidth * 7).Render(month[i])
			}
		}
		s += "\n"
	}
	return s
}
