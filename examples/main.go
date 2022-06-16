package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	table table.Model
	cols  []table.Column
	data  []table.Row
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "a":
			m.appendRow(table.Row{randSeq(10), "25"})
			m.table.SetRows(m.data)
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *model) appendRow(r table.Row) {
	m.data = append(m.data, r)
}

func (m model) View() string {
	return m.table.View()
}

func main() {
	rand.Seed(time.Now().UnixNano())

	m := model{
		cols: []table.Column{
			{Title: "name", Width: 30},
			{Title: "age", Width: 5},
		},
	}

	m.table = table.New(m.cols, m.data, 30, 20)

	if err := tea.NewProgram(m).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
