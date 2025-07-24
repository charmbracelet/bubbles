package navigator

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	pages Stack
}

func New(splash tea.Model) Model {
	model := Model{
		pages: []tea.Model{splash},
	}

	return model
}

func (m *Model) Init() tea.Cmd {
	return m.sendSeqCmd(m.pageInitCommands(m.pages.Top()))
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case PushNavigationMsg:
		var (
			item = msg
			cmds []tea.Cmd
		)
		// 1. last page
		cmds = append(cmds, m.pageLeaveCommands(m.pages.Top())...)
		// 2. new page
		cmds = append(cmds, m.pageEnterCommands(item)...)

		m.pages.Push(item)
		return m.sendSeqCmd(cmds)
	case PopNavigationMsg:
		var (
			cmds []tea.Cmd
		)

		if 0 == len(m.pages) {
			return nil
		}

		// 1. last page
		cmds = append(cmds, m.pageLeaveCommands(m.pages.Pop())...)
		// 2. curr page
		cmds = append(cmds, m.pageEnterCommands(m.pages.Top())...)

		return m.sendSeqCmd(cmds)
	}

	top := m.pages.Top()
	if top == nil {
		return nil
	}

	// update page
	newTop, cmd := top.Update(msg)
	m.pages[len(m.pages)-1] = newTop
	return cmd
}

func (m *Model) View() string {
	top := m.pages.Top()
	if top == nil {
		return ""
	}

	return top.View()
}
