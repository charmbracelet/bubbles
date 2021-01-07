package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"strings"
)

// DISCLAIMER: This is not a template but a example.
// This code is NOT performant or good for any other purpose except to show the possibility's of the list bubble.

func main() {
	allNodes := []fmt.Stringer{
		node{
			parentIDs: []int{7},
			value:     "no children here"},
		node{
			parentIDs: []int{1},
			value:     "use '+' to unfold a node"},
		node{
			parentIDs: []int{1, 4},
			value:     "use '-' to hide all children of this node"},
		node{
			parentIDs: []int{1, 8},
			value:     "use 'up' and 'down' to move around"},
		node{
			parentIDs: []int{1, 4, 5},
			value:     "grand child\nwith a line break"},
		node{
			parentIDs: []int{3},
			value:     "parent with no grand children"},
		node{
			parentIDs: []int{3, 2},
			value:     "hänsel"},
		node{
			parentIDs: []int{3, 6},
			value:     "gretel"},
	}
	var visNodes []fmt.Stringer
	for i, v := range allNodes {
		n, ok := v.(node)
		if !ok {
			continue
		}
		n.vis = true
		visNodes = append(visNodes, n)
		allNodes[i] = n
	}
	m := model{allNodes: allNodes}
	m.visible = list.NewModel()
	m.visible.SetLess(less)
	m.visible.SetEquals(equals)
	m.visible.AddItems(visNodes)
	m.startCmd = func() tea.Msg { return startMsg{} }

	m.visible.PrefixGen = NewPrefixer()

	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}

type model struct {
	visible  list.Model
	allNodes []fmt.Stringer
	startCmd tea.Cmd
}

type node struct {
	parentIDs []int
	value     string
	vis       bool
}

type startMsg struct{}

func (n node) String() string {
	return n.value
}

func (n node) GetID() (int, error) {
	lenID := len(n.parentIDs)
	if lenID == 0 {
		return 0, fmt.Errorf("no id set")
	}
	return n.parentIDs[lenID-1], nil
}

func (m model) Init() tea.Cmd { return m.startCmd }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "+":
			v, cmd := m.visible.GetCursorItem()
			if cmd != nil {
				msg := cmd()
				if _, ok := msg.(error); ok {
					return nil, cmd
				}
			}

			parent, ok := v.(node)
			if !ok {
				return m, nil
			}
			var newNodes []fmt.Stringer
			for i, v := range m.allNodes {
				n, ok := v.(node)
				parLen := len(parent.parentIDs)
				if !ok || len(n.parentIDs) <= parLen {
					continue
				}
				if len(n.parentIDs) == parLen+1 && n.parentIDs[parLen-1] == parent.parentIDs[parLen-1] && !n.vis {
					n.vis = true
					newNodes = append(newNodes, n)
					m.allNodes[i] = n
				}
			}
			cmd = m.visible.AddItems(newNodes)
			m.visible.Sort()
			return m, cmd

		case "-":
			v, cmd := m.visible.GetCursorItem()
			if cmd != nil {
				msg := cmd()
				if _, ok := msg.(error); ok {
					return m, cmd
				}
			}
			parent, ok := v.(node)
			if !ok {
				return m, nil
			}

			// TODO NOTE this is not performant:  a round O(1/2(n*m)) (average)
			// whereby 'n' m.allNodes and 'm' all m.visible.GetAllItems()
			for i, v := range m.allNodes {
				n, ok := v.(node)
				parLen := len(parent.parentIDs)
				if !ok || len(n.parentIDs) <= parLen {
					continue
				}
				if n.vis && len(n.parentIDs) > parLen && n.parentIDs[parLen-1] == parent.parentIDs[parLen-1] {
					index, err := m.visible.GetIndex(n)
					if err != nil {
						continue
					}
					m.visible.RemoveIndex(index)
					n.vis = false
					m.allNodes[i] = n
				}

			}
			return m, cmd
		case "s":
			// dont return the command issued by the Sort command, possible endless sort loop!
			// Because we are sorting when receiving a ListChange Msg and they get issued by the Sort method -> loop.
			_ = m.visible.Sort()
			return m, nil
		default:
			newList, cmd := m.visible.Update(msg)
			newVis, ok := newList.(list.Model)
			if ok {
				m.visible = newVis
			}
			return m, cmd
		}
	case list.ListChange:
		// dont return the command issued by the Sort command, endless sort loop!
		// Because we are sorting here when receiving a ListChange Msg and they get issued by the Sort method -> loop.
		_ = m.visible.Sort()
		return m, nil
	case startMsg:
		_ = m.visible.Sort()
		_, _ = m.visible.SetCursor(0)
		return m, nil
	default:
		newList, cmd := m.visible.Update(msg)
		newVis, ok := newList.(list.Model)
		if ok {
			m.visible = newVis
		}
		return m, cmd
	}
}
func (m model) View() string {
	lines, err := m.visible.Lines()
	if err != nil {
		return err.Error()
	}
	return strings.Join(lines, "\n")
}
func (m model) Lines() ([]string, error) {
	return m.visible.Lines()
}

func less(a, b fmt.Stringer) bool {
	first, ok1 := a.(node)
	if !ok1 {
		panic("cant sort something else than nodes")
	}
	second, ok2 := b.(node)
	if !ok2 {
		panic("cant sort something else than nodes")
	}
	firLen, secLen := len(first.parentIDs), len(second.parentIDs)
	shorter := firLen
	if secLen < shorter {
		shorter = secLen
	}
	for c := 0; c < shorter; c++ {
		if first.parentIDs[c] > second.parentIDs[c] {
			return false
		}
		if first.parentIDs[c] < second.parentIDs[c] {
			return true
		}
	}

	return firLen <= secLen
}
func equals(a, b fmt.Stringer) bool {
	first, ok1 := a.(node)
	second, ok2 := b.(node)
	if !ok1 || !ok2 {
		return false
	}
	firLen, secLen := len(first.parentIDs), len(second.parentIDs)
	return firLen == secLen && first.parentIDs[firLen-1] == second.parentIDs[secLen-1]
}
