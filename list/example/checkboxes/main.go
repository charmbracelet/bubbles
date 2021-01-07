package main

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// DISCLAIMER: This is not a template but a example.
// This code is NOT performant or good for any other purpose except to show the possibility's of the list bubble.

func main() {
	m := model{}
	m.vis = list.NewModel()
	m.vis.PrefixGen = NewPrefixer()
	m.head = "My TODO list!\n============="
	m.AddItems([]string{
		"buying eggs",
		"take the trash out",
		"get a hair cut",
		"be nice\nto the neighbours",
		"get milk",
	})
	m.tail = "============================================\nuse ' ' to change the done state of a item\nuse 'q' or 'ctrl+c' to exit"
	p := tea.NewProgram(m)
	p.Start()

}

type item struct {
	selected bool
	content  string
	id       int
}

func (m item) String() string {
	return m.content
}

type model struct {
	vis   list.Model
	jump  string
	ready bool
	head  string
	tail  string
}

func (m model) Init() tea.Cmd {
	return nil
}

// update recives messages and the model and changes the model accordingly to the messages
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.vis.PrefixGen == nil {
		// use default
		m.vis.PrefixGen = NewPrefixer()
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:

		// Ctrl+c exits
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		keyString := msg.String()
		switch keyString {
		case "q":
			return m, tea.Quit
		case "1", "2", "3", "4", "5", "6", "7", "8", "9", "0":
			m.jump += keyString
			return m, nil
		case "down", "j":
			m.vis.MoveCursor(m.popJump(1))
			return m, nil
		case "up", "k":
			m.vis.MoveCursor(-m.popJump(1))
			return m, nil
		case "r":
			d, ok := m.vis.PrefixGen.(*SelectPrefixer)
			if ok {
				d.NumberRelative = !d.NumberRelative
			}
			return m, nil
		case "J":
			m.vis.MoveItem(m.popJump(1))
			return m, nil
		case "K":
			m.vis.MoveItem(-m.popJump(1))
			return m, nil
		case "t", "home":
			j := m.popJump(0)
			if j > 0 {
				j--
			}
			m.vis.Top()
			m.vis.MoveCursor(j)
			return m, nil
		case "b", "end":
			j := m.popJump(0)
			if j > 0 {
				j--
			}
			m.vis.Bottom()
			m.vis.MoveCursor(-j)
			return m, nil
		case "w":
			m.vis.Wrap = m.popJump(0)
			return m, nil
		case "s":
			less := func(a, b fmt.Stringer) bool { return a.String() < b.String() }
			m.vis.SetLess(less)
			m.vis.Sort()
			return m, nil
		case "o":
			less := func(a, b fmt.Stringer) bool {
				d, _ := a.(item)
				e, _ := b.(item)
				return d.id < e.id
			}
			m.vis.SetLess(less)
			m.vis.Sort()
			return m, nil
		case " ":
			updater := func(a fmt.Stringer) (fmt.Stringer, tea.Cmd) {
				i, ok := a.(item)
				if !ok {
					return a, nil
				}
				i.selected = !i.selected
				return i, nil
			}
			i, _ := m.vis.GetCursorIndex()
			cmd := m.vis.UpdateItem(i, updater)
			return m, cmd
		default:
			// resets jump buffer to prevent confusion
			m.jump = ""

			// pipe all other commands to the update from the vis
			l, newMsg := m.vis.Update(msg)
			vis, _ := l.(list.Model)
			m.vis = vis
			return m, newMsg
		}

	case tea.WindowSizeMsg:

		width := msg.Width
		height := msg.Height
		m.vis.Screen.Width = width
		m.vis.Screen.Height = height

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
		// pipe all other commands to the update from the vis
		l, newMsg := m.vis.Update(msg)
		vis, _ := l.(list.Model)
		m.vis = vis
		return m, newMsg
	}
}

func (m model) View() string {
	return fmt.Sprintf("%s\n%s\n%s", m.head, m.vis.View(), m.tail)
}
func (m *model) AddItems(toAdd []string) {
	strList := make([]fmt.Stringer, len(toAdd))
	for i, str := range toAdd {
		strList[i] = item{content: str}
	}
	m.vis.AddItems(strList)
}

// popJump takes default vaule and returns the integer value of the jump string
// if its empty or fails the default is returned.
func (m *model) popJump(dft int) int {
	if m.jump == "" {
		return dft
	}
	j, err := strconv.Atoi(m.jump)
	if err != nil {
		return dft
	}
	m.jump = ""
	return j
}
