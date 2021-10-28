package calendar

import (
	"fmt"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	weekDayToNumber = map[string]int{
		"Monday": 0, "Tuesday": 1, "Wednesday": 2, "Thursday": 3, "Friday": 4, "Saturday": 5, "Sunday": 6,
	}
	separator = " "
)

type Model struct {
	CurrentDate time.Time
	Styles      map[string]lipgloss.Style
}

func NewModel() Model {
	return Model{
		CurrentDate: time.Now(),
		Styles:      map[string]lipgloss.Style{"current_date": lipgloss.NewStyle().Background(lipgloss.Color("#7571F9"))},
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

	s := fmt.Sprintf("%s %s %s %s %s %s %s\n", "Mo", "Tu", "We", "Th", "Fr", "Sa", "Su")

	offset := weekDayToNumber[firstofmonth.Weekday().String()]
	for i := 0; i < offset; i++ {
		s += fmt.Sprintf("   ")
	}

	for i := 1; i <= lastofmonth.Day(); i++ {
		var char string
		if i == m.CurrentDate.Day() {
			char = m.Styles["current_date"].Render(strconv.Itoa(i))
		} else {
			char = strconv.Itoa(i)
		}

		if i < 10 {
			s += fmt.Sprintf(" %s ", char)
		} else {
			s += fmt.Sprintf("%s ", char)
		}

		if (i+offset)%7 == 0 {
			s += "\n"
		}
	}

	s += "\n"
	return s
}
