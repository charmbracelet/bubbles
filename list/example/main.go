package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
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
	updater := func(toUp fmt.Stringer) (fmt.Stringer, tea.Cmd) {
		i := toUp.(stringItem)
		i.style = style
		return i, nil
	}
	_, err := m.list.UpdateItem(index, updater)
	return err
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
		"The list can handel linebreaks,\nand has wordwrap enabled if the line gets to long.",
		"",
		"Movements:",
		"You can move the highlighted index up and down with the arrow keys or 'k' and 'j'.",
		"Move to the top with 't' and to the bottom with 'b'.",
		"All keys that change the cursor position can be preceded with the press of numbers and change the movemet to that amount.\nI.e.: the key press order '1','2' and 't' moves the cursor to the twelfth item from the top.",
		"",
		"Order:",
		"Use 'K' or 'J' to move the item under the curser up and down.",
		"Sort the entrys with 's' depending on a costum less function, in this case string sorting.",
		"Or bring them back into the original order with the 'o' key.",
		"",
		"Settings:",
		"To toggle between only absolute item numbers and relativ numbers use the 'r' key.",
		"To toggle between showing the wrapped lines of a item use the 'w' key.",
		"",
		"Select:",
		"To select a item use 'm'\nor to select all items 'M'.",
		"To unselect a item use 'u'\nor to unselect all items 'U'.",
		"You can toggle the select state of the current item with the space key.",
		"The key 'v' inverts the selected state of all items.",
		"",
		"Edit:",
		"With the key 'e' you can edit the string of the current item.",
		"There you can make changes to the string and apply them with 'enter' or discard them with 'escape'",
		"",
		"Here are some more items for you to test the scrolling\nand the cursor-Offset which defaults to 5 lines from the screen border.",
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
		//"Be aware, that items with more linebreaks than the screen height minus twice the scroll offset, cause some display problems to the hole list, but the cursor will be on the right item, even if the cursor jumps relative to the screen.\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\nBut you can avoid this ,by toggeling wrap to be off, with the 'w' key.",
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
		"If you want to jump directly to me type '5' and than 'b',\nbecause i am the fifth item (not line) from the bottom.", "", "", "",
		"Hey, i am the last item :) you can move directly to me with the 'b' key, which stands for bottom.",
	}

	m.AddStrings(itemList)

	m.SetStyle(0, termenv.Style{}.Foreground(termenv.ColorProfile().Color("#ffff00")))

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
		cmd, _ := m.list.UpdateItem(i, updater)
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
				m.list.UpdateAllItems(updater)

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
				m.list.UpdateAllItems(updater)

			}
			m.edit = false
			return m, nil

		case "c":
			m.list.Move(1)
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

			//		case "t":
			//			m.lastViews = append(m.lastViews, m.View())
			//			return m, nil
			//		case "T":
			//			f, _ := os.OpenFile("test_cases.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			//			f.WriteString(strings.Join(m.lastViews, "\n##########################\n"))
			//			return m, tea.Quit
		case "w":
			m.list.Wrap = !m.list.Wrap
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
