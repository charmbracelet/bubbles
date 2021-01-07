package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	//"log"
	"os"
	"strconv"
)

type model struct {
	ready     bool
	list      list.Model
	finished  bool
	edit      bool
	jump      string
	lastViews []string

	// Channels to create unique ids for all added/new items
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

	// only used if one wants to get the Index of a item.
	l.SetEquals(func(first, second fmt.Stringer) bool {
		f := first.(stringItem)
		s := second.(stringItem)
		return f.id == s.id
	})
	// used for custom sorting, if not set string comparison will be used.
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
	updater := func(toUp fmt.Stringer) (fmt.Stringer, tea.Cmd) {
		i := toUp.(stringItem)
		i.style = style
		return i, nil
	}
	_, err := m.list.ValidIndex(index)
	if err != nil {
		return err
	}
	m.list.UpdateItem(index, updater)
	return nil
}

type stringItem struct {
	value string
	id    int
	edit  bool
	style termenv.Style
	input textinput.Model
}

func (s stringItem) String() string {
	if s.edit {
		// prepend with ansi-escape sequence to end all hightlighting to not interfere with the textinput-hightlighting
		return "\x1b[0m" + s.input.View()
	}
	return s.style.Styled(string(s.value))
}

func main() {

	m := newModel()
	itemList := []string{
		"Welcome to the bubbles-list example!",
		"",
		"Use 'q' or 'ctrl-c' to quit!",
		"This example serves the purpose to show what one can do with this list, not what is this list.",
		"A Item of the list is one struct value that was added to this list, that satisfies the fmt.Stringer interface (has a String() string methode),\nsince a string can have linebreaks so does a item of this list.",
		"",
		"Movements:",
		"You can move the highlighted index up and down with the arrow keys or 'k' and 'j'.",
		"Move to the top with 't' and to the bottom with 'b'.",
		"All keys that change the cursor position can be preceded with the press of numbers and change the movement to that amount.\nI.e.: the key press order '1','2' and 't' moves the cursor to the twelfth item from the top.",
		"If you know on which index you want to be, type the number and confirm with 'enter'.",
		"",
		"Order:",
		"Use 'K' or 'J' to move the item under the curser up and down.",
		"Sort the entries with 's' depending on a custom sort (less) function, in this case string sorting.",
		"Or bring them back into the original order with the 'o' key.",
		"",
		"Settings:",
		"To toggle between only absolute item numbers and relative numbers use the 'r' key.",
		"To limit the amount of lines displayed per item, type the limit and press 'w' or just press 'w' without any numbers to unlimit again.",
		"",
		"Edit:",
		"With the key 'e' you can edit the string of the current item. Which shows that you can embed other bubbles into the list items.",
		"There you can make changes to the string and apply them with 'enter' or discard them with 'escape'",
		"While you can add new empty Items with the 'a' key.",
		"You can permanently delete an item, with the key 'd'.",
		"",
		"Here are some more items for you to test the scrolling\nand the cursor offset, which defaults to 5 lines relative to the screen border.",
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
		"Multi-line items are not a problem, either.\nBut you may have the problem that you cant see all of the lines of the items, because movement is by design only possible by item and not by line. But since it is possible to embed other bubble-widgets, you could embed a paginator to overcome this problem.\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\nCan you see me?\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n",
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
		"If you want to jump directly to me type '5' and than 'b',\nbecause I am the fifth item (not line) from the bottom.", "", "", "",
		"Hey, i am the last item :) you can move to me directly with the 'b' key, which stands for bottom.",
	}

	m.AddStrings(itemList)

	m.SetStyle(0, termenv.Style{}.Foreground(termenv.ColorProfile().Color("#ffff00")))

	p := tea.NewProgram(m)

	// Use the full size of the terminal in its "alternate screen buffer"
	fullScreen := true // change to false if you dont want fullscreen

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

	// if there is a item to be edit, pass the massage to the Update methode of the item.
	if k, ok := msg.(tea.KeyMsg); m.edit && ok && k.Type != tea.KeyEscape && k.Type != tea.KeyEnter {
		updater := func(toUp fmt.Stringer) (fmt.Stringer, tea.Cmd) {
			item, _ := toUp.(stringItem)
			if !item.edit {
				return item, nil
			}
			newInput, cmd := item.input.Update(msg)
			item.input = newInput
			return item, cmd
		}
		i, _ := m.list.GetCursorIndex()
		cmd := m.list.UpdateItem(i, updater)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEscape {
			if m.edit {
				// make sure that all items edit-fields are false and discard the change
				updater := func(toUp fmt.Stringer) (fmt.Stringer, tea.Cmd) {
					item, _ := toUp.(stringItem)

					item.edit = false
					return item, nil
				}
				for c := 0; c < m.list.Len(); c++ {
					m.list.UpdateItem(c, updater)
				}

			}
			m.edit = false
			return m, nil
		}

		// Ctrl+c exits
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		keyString := msg.String()
		switch keyString {
		case "e":
			m.edit = true
			i, _ := m.list.GetCursorIndex()

			updater := func(toUp fmt.Stringer) (fmt.Stringer, tea.Cmd) {
				item, _ := toUp.(stringItem)
				item.input = textinput.NewModel()
				item.input.Prompt = ""
				item.input.SetValue(item.value)
				item.input.Focus()
				item.edit = true

				j, _ := strconv.Atoi(m.jump)
				item.input.SetCursor(j)
				m.jump = ""
				return item, nil
			}
			m.list.UpdateItem(i, updater)
			return m, nil

		case "enter":
			if m.edit {
				// Update the value and make sure that all items edit-fields are false
				updater := func(toUp fmt.Stringer) (fmt.Stringer, tea.Cmd) {
					item, _ := toUp.(stringItem)
					if item.edit {
						item.value = item.input.Value()
					}

					item.edit = false
					return item, nil
				}
				for c := 0; c < m.list.Len(); c++ {
					m.list.UpdateItem(c, updater)
				}
				m.edit = false
				return m, nil

			}

			j := 0
			if m.jump != "" {
				j, _ = strconv.Atoi(m.jump)
				m.jump = ""
				m.list.SetCursor(j - 1)
			}
			return m, nil

		case "q":
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
			m.list.MoveCursor(j)
			return m, nil
		case "up", "k":
			j := 1
			if m.jump != "" {
				j, _ = strconv.Atoi(m.jump)
				m.jump = ""
			}
			m.list.MoveCursor(-j)
			return m, nil
		case "r":
			d, ok := m.list.PrefixGen.(*list.DefaultPrefixer)
			if ok {
				d.NumberRelative = !d.NumberRelative
			}
			return m, nil
		case "J":
			j := 1
			if m.jump != "" {
				j, _ = strconv.Atoi(m.jump)
				m.jump = ""
			}
			m.list.MoveItem(j)
			return m, nil
		case "K":
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
			m.list.MoveCursor(j)
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
			m.list.MoveCursor(-j)
			return m, nil

		case "w":
			if m.jump != "" {
				j, _ := strconv.Atoi(m.jump)
				m.jump = ""
				m.list.Wrap = j
				return m, nil
			}
			m.list.Wrap = 0
			return m, nil
		case "s":
			less := func(a, b fmt.Stringer) bool { return a.String() < b.String() }
			m.list.SetLess(less)
			m.list.Sort()
			return m, nil
		case "o":
			less := func(a, b fmt.Stringer) bool {
				d, _ := a.(stringItem)
				e, _ := b.(stringItem)
				return d.id < e.id
			}
			m.list.SetLess(less)
			m.list.Sort()
			return m, nil
		case "a":
			j := 1
			if m.jump != "" {
				j, _ = strconv.Atoi(m.jump)
				m.jump = ""
			}
			m.AddStrings(make([]string, j))
			return m, nil
		case "d":
			j := 1
			if m.jump != "" {
				j, _ = strconv.Atoi(m.jump)
				m.jump = ""
			}
			var ok bool
			var i int
			for c := 0; c < j && !ok; c++ {
				i, _ = m.list.GetCursorIndex()
				_, cmd := m.list.RemoveIndex(i)
				_, ok = cmd().(error)
			}
			return m, nil

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
