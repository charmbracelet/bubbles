package main

import (
	"bytes"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"os"
	"strconv"
)

type model struct {
	ready     bool
	list      list.Model
	finished  bool
	endResult chan<- string
	jump      string
	lastViews []string

	requestID chan<- struct{}
	resultID  <-chan int
}

// newModel returns a new example Model and starts the goroutine go generate the unique id
func newModel() *model {
	m := model{}

	req := make(chan struct{})
	res := make(chan int)

	m.requestID = req
	m.resultID = res

	go func(requ <-chan struct{}, send chan<- int) {
		for c := 0; true; c++ {
			_ = <-requ
			send <- c
		}
	}(req, res)

	l := list.NewModel()
	l.SuffixGen = list.NewSuffixer()

	l.SetEquals(func(first, second fmt.Stringer) bool {
		f := first.(stringItem)
		s := second.(stringItem)
		return f.id == s.id
	})
	l.SetLess(func(first, second fmt.Stringer) bool {
		f := first.(stringItem)
		s := second.(stringItem)
		return f.id < s.id
	})

	m.list = l

	return &m
}

// GetID returns a new for this list unique id
func (m *model) GetID() (int, error) {
	if m.requestID == nil || m.resultID == nil {
		return 0, fmt.Errorf("no ID generator running")
	}
	var e struct{}
	m.requestID <- e
	return <-m.resultID, nil
}

func (m *model) AddStrings(items []string) error {
	newList := make([]fmt.Stringer, 0, len(items))
	for _, i := range items {
		id, e := m.GetID()
		if e != nil {
			return e
		}
		newList = append(newList, stringItem{value: i, id: id})
	}
	m.list.AddItems(newList)
	return nil
}

func (m *model) SetStyle(index int, style termenv.Style) error {
	updater := func(toUp fmt.Stringer) fmt.Stringer {
		i := toUp.(stringItem)
		i.style = style
		return i
	}
	return m.list.UpdateItem(index, updater)
}

type stringItem struct {
	value string
	id    int
	style termenv.Style
}

func (s stringItem) String() string {
	return s.style.Styled(string(s.value))
}

func main() {
	m := newModel()
	itemList := []string{
		"Welcome to the bubbles-list example!",
		"",
		"Use 'q' or 'ctrl-c' to quit!",
		"The list can handel linebreaks,\nand has wordwrap enabled if the line gets to long.",
		"",
		"Movements:",
		"You can move the highlighted index up and down with the arrow keys or 'k' and 'j'.",
		"Move to the top with 't' and to the bottom with 'b'.",
		"",
		"Order:",
		"Use '+' or '-' to move the item under the curser up and down.",
		"Sort the entrys with 's' depending on a costum less function, here the orginal order.",
		"To toggle between only absolute item numbers and relativ numbers use the 'r' key.",
		"",
		"Select:",
		"To select a item use 'm'\nor to select all items 'M'.",
		"To unselect a item use 'u'\nor to unselect all items 'U'.",
		"You can toggle the select state of the current item with the space key.",
		"The key 'v' inverts the selected state of all items.",
		"",
		"Ones you hit 'enter', the selected lines will be printed to StdOut and the program exits.",
		"When you print the items there will be a loss of information,\nsince one can not say what was a line break within an item or what is a new item",
		"",
		"Here are some more items for you to test the scrolling\nand the cursor-Offset which defaults to 5 lines from the screen border.",
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
		"Be aware that items with more linebreaks than the screen height minus twice the scroll offset cause some display problems to the hole list, but the cursor will be on the right item, even if the cursor jumps relative to the screen.\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n",
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
		"If you want to jump directly to me type '5' and than 'b',\nbecause i am the fifth item (not lines) from the bottom.", "", "", "",
		"Hey i am the last item :) you can move directly to me with the 'b' key, which stands for bottom",
	}

	m.AddStrings(itemList)

	m.SetStyle(0, termenv.Style{}.Foreground(termenv.ColorProfile().Color("#ffff00")))

	endResult := make(chan string, 1)
	m.endResult = endResult

	p := tea.NewProgram(m)

	// Use the full size of the terminal in its "alternate screen buffer"
	fullScreen := true // change to true if you want fullscreen

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
	// allways print a newline even on empty string result
	fmt.Println(res)
}

func (m model) Init() tea.Cmd {
	return nil
}

// View waits till the terminal sizes is known to the model and than,
// pipes the model to the list View for rendering the list
func (m model) View() string {
	if !m.ready {
		return "\n  Initalizing...\n\n  Waiting for info about window size.\n"
	}

	listString := m.list.View()
	return listString
}

// update recives messages and the model and changes the model accordingly to the messages
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.list.PrefixGen == nil {
		// use default
		m.list.PrefixGen = list.NewPrefixer()
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Ctrl+c exits
		if msg.Type == tea.KeyCtrlC {
			m.endResult <- ""
			return m, tea.Quit
		}
		keyString := msg.String()
		switch keyString {
		case "c":
			m.list.Move(1)
			return m, nil
		case "q":
			m.endResult <- ""
			return m, tea.Quit
		case "1", "2", "3", "4", "5", "6", "7", "8", "9", "0":
			m.jump += keyString
			return m, nil
		case "down", "j":
			j := 1
			if m.jump != "" {
				j, _ = strconv.Atoi(m.jump)
				m.jump = ""
			}
			m.list.Move(j)
			return m, nil
		case "up", "k":
			j := 1
			if m.jump != "" {
				j, _ = strconv.Atoi(m.jump)
				m.jump = ""
			}
			m.list.Move(-j)
			return m, nil
		case "r":
			d, ok := m.list.PrefixGen.(*list.DefaultPrefixer)
			if ok {
				d.NumberRelative = !d.NumberRelative
			}
			return m, nil
		case "m":
			j := 1
			if m.jump != "" {
				j, _ = strconv.Atoi(m.jump)
				m.jump = ""
			}
			m.list.MarkSelected(j, true)
			return m, nil
		case "u":
			j := 1
			if m.jump != "" {
				j, _ = strconv.Atoi(m.jump)
				m.jump = ""
			}
			m.list.MarkSelected(j, false)
			return m, nil
		case " ":
			j := 1
			if m.jump != "" {
				j, _ = strconv.Atoi(m.jump)
				m.jump = ""
			}
			m.list.ToggleSelect(j)
			return m, nil
		case "+":
			j := 1
			if m.jump != "" {
				j, _ = strconv.Atoi(m.jump)
				m.jump = ""
			}
			m.list.MoveItem(j)
			return m, nil
		case "-":
			j := 1
			if m.jump != "" {
				j, _ = strconv.Atoi(m.jump)
				m.jump = ""
			}
			m.list.MoveItem(-j)
			return m, nil
		case "t", "home":
			j := 0
			if m.jump != "" {
				j, _ = strconv.Atoi(m.jump)
				m.jump = ""
			}
			if j > 0 {
				j--
			}
			m.list.Top()
			m.list.Move(j)
			return m, nil
		case "b", "end":
			j := 0
			if m.jump != "" {
				j, _ = strconv.Atoi(m.jump)
				m.jump = ""
			}
			if j > 0 {
				j--
			}
			m.list.Bottom()
			m.list.Move(-j)
			return m, nil

		case "enter":
			// Enter prints the selected lines to StdOut
			var result bytes.Buffer
			for _, item := range m.list.GetSelected() {
				result.WriteString(item.String())
				result.WriteString("\n")
			}
			m.endResult <- result.String()
			return m, tea.Quit
			//		case "t":
			//			m.lastViews = append(m.lastViews, m.View())
			//			return m, nil
			//		case "T":
			//			f, _ := os.OpenFile("test_cases.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			//			f.WriteString(strings.Join(m.lastViews, "\n##########################\n"))
			//			return m, tea.Quit
		default:
			// resets jump buffer to prevent confusion
			m.jump = ""

			// pipe all other commands to the update from the list
			l, newMsg := m.list.Update(msg)
			list, _ := l.(list.Model)
			m.list = list
			return m, newMsg
		}

	case tea.WindowSizeMsg:

		width := msg.Width
		height := msg.Height
		m.list.Screen.Width = width
		m.list.Screen.Height = height

		if !m.ready {
			// Since this program can use the full size of the viewport we need
			// to wait until we've received the window dimensions before we
			// can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.ready = true
		}
		return m, nil

	default:
		// pipe all other commands to the update from the list
		l, newMsg := m.list.Update(msg)
		list, _ := l.(list.Model)
		m.list = list
		return m, newMsg
	}
}
