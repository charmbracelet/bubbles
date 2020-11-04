package main

import (
	"bufio"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"strings"
)

/*
Reads from StdIn, opens lines as bubbles-list.
When closed print, with space, selected lines to StdOut
*/

type model struct {
	ready     bool
	list      list.Model
	finished  bool
	endResult chan<- string
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	lineList := make([]string, 0, 1000)
	for scanner.Scan() {
		lineList = append(lineList, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading from stdin:", err)
		os.Exit(1)
	}
	endResult := make(chan string, 1)

	p := tea.NewProgram(initialize(lineList, endResult), update, view)

	// Use the full size of the terminal in its "alternate screen buffer"
	p.EnterAltScreen()

	if err := p.Start(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
	p.ExitAltScreen()

	fmt.Println(<-endResult)
}

func initialize(lineList []string, endResult chan<- string) func() (tea.Model, tea.Cmd) {
	l := list.NewModel()
	l.AddItems(lineList)

	return func() (tea.Model, tea.Cmd) { return model{list: l, endResult: endResult}, nil }
}

func view(mdl tea.Model) string {
	m, _ := mdl.(model)
	if !m.ready {
		return "\n  Initalizing..."
	}

	return list.View(m.list)
}

type confirmation struct{}

func update(msg tea.Msg, mdl tea.Model) (tea.Model, tea.Cmd) {
	m, _ := mdl.(model)

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Ctrl+c exits
		if msg.Type == tea.KeyCtrlC {
			result := strings.Join(m.list.GetSelected(), "\n")
			m.endResult <- result
			return m, tea.Quit
		}
		switch msg.String() {
		case "q":
			result := strings.Join(m.list.GetSelected(), "\n")
			m.endResult <- result
			return m, tea.Quit
		case "j":
			m.list.Down()
			return m, nil
		case "k":
			m.list.Up()
			return m, nil
		case " ":
			m.list.ToggleSelect()
			m.list.Down()
			return m, nil
		case "g":
			m.list.Top()
			return m, nil
		case "G":
			m.list.Bottom()
			return m, nil
		case "s":
			m.list.Sort()
			return m, nil
		}

	case tea.WindowSizeMsg:

		m.list.Viewport.Width = msg.Width
		m.list.Viewport.Height = msg.Height

		if !m.ready {
			// Since this program is using the full size of the viewport we need
			// to wait until we've received the window dimensions before we
			// can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.ready = true
		}

		// Because we're using the viewport's default update function (with pager-
		// style navigation) it's important that the viewport's update function:
		//
		// * Recieves messages from the Bubble Tea runtime
		// * Returns commands to the Bubble Tea runtime
		//

		m.list.Viewport, cmd = viewport.Update(msg, m.list.Viewport)

		return m, cmd
	case confirmation:
		if !m.finished {
			return m, nil
		}
	}

	return m, nil
}
