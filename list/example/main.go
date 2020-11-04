package main

import (
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
	items := []string{
		"Welcome to the bubbles-list example!",
		"Use 'q' or 'ctrl-c' to quit!",
		"You can move the highlighted index up and down with the keys 'k' and 'j'.",
		"Move to the beginning with 'g' and to the end with 'G'.",
		"Sort the entrys with 's', but be carefull you can't unsort it again.",
		"The list can handel linebreaks,\nand has wordwrap enabled if the line gets to long.",
		"You can select items with the space key which will select the line and mark it as such.",
		"Ones you hit 'enter', the selected lines will be printed to StdOut and the program exits.",
		"When you print the items there will be a loss of information,",
		"since one can not say what was a line break within an item or what is a new item",
	}
	endResult := make(chan string, 1)

	p := tea.NewProgram(initialize(items, endResult), update, view)

	// Use the full size of the terminal in its "alternate screen buffer"
	fullScreen := false // change to true if you want fullscreen

	if fullScreen {
		p.EnterAltScreen()
	}

	if err := p.Start(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
	if fullScreen {
		p.ExitAltScreen()
	}

	res := <-endResult
	if res != "" {
		fmt.Println(res)
	}
}

// initialize sets up the model and returns it to the bubbletea runtime
// as a function result, so it can later be handed over to the update and view functions.
func initialize(lineList []string, endResult chan<- string) func() (tea.Model, tea.Cmd) {
	l := list.NewModel()
	l.AddItems(lineList)

	return func() (tea.Model, tea.Cmd) { return model{list: l, endResult: endResult}, nil }
}

// view waits till the terminal sizes is knowen to the model and than,
// pipes the model to the list View for rendering the list
func view(mdl tea.Model) string {
	m, _ := mdl.(model)
	if !m.ready {
		return "\n  Initalizing...\n\n  Waiting for info about window size."
	}

	listString := list.View(m.list)
	return listString
}

// update recives messages and the model and changes the model accordingly to the messages
func update(msg tea.Msg, mdl tea.Model) (tea.Model, tea.Cmd) {
	m, _ := mdl.(model)

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Ctrl+c exits
		if msg.Type == tea.KeyCtrlC {
			m.endResult <- ""
			return m, tea.Quit
		}
		switch msg.String() {
		case "q":
			m.endResult <- ""
			return m, tea.Quit
		}

		// Enter prints the selected lines to StdOut
		if msg.Type == tea.KeyEnter {
			result := strings.Join(m.list.GetSelected(), "\n")
			m.endResult <- result
			return m, tea.Quit
		}

		// pipe all other commands to the update from the list
		list, newMsg := list.Update(msg, m.list)
		m.list = list

		return m, newMsg

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

	}

	return m, nil
}
