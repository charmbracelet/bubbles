package calendar

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var ()

type Model struct {
	CurrentDate time.Time
	Styles      Styles
	Weekdays    []string
}

type Styles struct {
	CurrentDate lipgloss.Style
}

func NewModel() Model {
	return Model{
		CurrentDate: time.Now(),
		Styles:      Styles{CurrentDate: lipgloss.NewStyle().Background(lipgloss.Color("#7571F9"))},
		Weekdays: []string{
			"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday",
		},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		}
	}

	return m, nil
}

func (m Model) View() string {
	firstofmonth := time.Date(m.CurrentDate.Year(), m.CurrentDate.Month(), 1, 0, 0, 0, 0, m.CurrentDate.Location())
	lastofmonth := firstofmonth.AddDate(0, 1, -1)

	// Header with 2 character prefix of day names
	var s string
	for i := 0; i < len(m.Weekdays); i++ {
		s += m.Weekdays[i][0:2]
		s += " "
	}
	s += "\n"

	var offset int
	for i, weekday := range m.Weekdays {
		if weekday == firstofmonth.Weekday().String() {
			offset = i
			break
		}
	}
	s += strings.Repeat("   ", offset)

	for i := 1; i <= lastofmonth.Day(); i++ {
		var char string
		if i == m.CurrentDate.Day() {
			char = m.Styles.CurrentDate.Render(strconv.Itoa(i))
		} else {
			char = strconv.Itoa(i)
		}

		// Indent days numbers
		if i < 10 {
			s += fmt.Sprintf(" %s ", char)
		} else {
			s += fmt.Sprintf("%s ", char)
		}

		// Line return on week end
		if (i+offset)%7 == 0 {
			s += "\n"
		}
	}

	s += "\n"
	return s
}
