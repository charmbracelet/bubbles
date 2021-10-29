/*
Calendar component

Mo Tu We Th Fr Sa Su
             1  2  3
 4  5  6  7  8  9 10
11 12 13 14 15 16 17
18 19 20 21 22 23 24
25 26 27 28 29 30 31
*/

package calendar

import (
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model stores component.
type Model struct {
	CurrentDate time.Time
	Styles      Styles
	Weekdays    []Weekday
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
			CurrentDate:    lipgloss.NewStyle().ColorWhitespace(false).Width(3).Align(lipgloss.Center).Background(lipgloss.Color("#7571F9")),
			Date:           lipgloss.NewStyle().ColorWhitespace(false).Width(3).Align(lipgloss.Center),
			WeekdaysHeader: lipgloss.NewStyle().ColorWhitespace(false).Width(3).Align(lipgloss.Left).Background(lipgloss.Color("#F25D94")),
		},
		Weekdays: EnglishWeekdays,
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
	firstofmonth := time.Date(m.CurrentDate.Year(), m.CurrentDate.Month(), 1, 0, 0, 0, 0, m.CurrentDate.Location())
	lastofmonth := firstofmonth.AddDate(0, 1, -1)

	s := ""
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
		if weekday.Name == firstofmonth.Weekday().String() {
			monthStartingDayOffset = i
			break
		}
	}
	s += strings.Repeat("   ", monthStartingDayOffset)

	// Render calendar
	// Current selected day is highlighted
	for i := 1; i <= lastofmonth.Day(); i++ {
		if i == m.CurrentDate.Day() {
			s += m.Styles.CurrentDate.Render(strconv.Itoa(i))
		} else {
			s += m.Styles.Date.Render(strconv.Itoa(i))
		}

		// Line return on week end, except when the last day in the month ends on the last weekday
		if (i+monthStartingDayOffset)%7 == 0 && i != lastofmonth.Day() {
			s += "\n"
		}
	}
	s += "\n"
	return s
}
