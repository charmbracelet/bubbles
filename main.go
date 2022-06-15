package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	table   table.Model
	cols []table.Column
	data    []table.Row
}



func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// switch msg := msg.(type) {
	// case tea.WindowSizeMsg:
	// 	m.list.SetWidth(msg.Width)
	// 	return m, nil
	//
	// case tea.KeyMsg:
	// 	switch keypress := msg.String(); keypress {
	// 	case "ctrl+c":
	// 		m.quitting = true
	// 		return m, tea.Quit
	//
	// 	case "enter":
	// 		i, ok := m.list.SelectedItem().(item)
	// 		if ok {
	// 			m.choice = string(i)
	// 		}
	// 		return m, tea.Quit
	// 	}
	// }
	//
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return m.table.View()
}
func main() {
	m := model{
		cols: []table.Column{
			{Title: "name", Width: 30},
			{Title: "age", Width: 5},
		},

		data: []table.Row{
			{"John Doe", "69"},
			{"Jane Doe", "29"},
		},
	}

	m.table = table.New(m.cols, m.data, 40, 50)
	if err := tea.NewProgram(m).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
