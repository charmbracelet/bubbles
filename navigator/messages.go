package navigator

import (
	tea "github.com/charmbracelet/bubbletea"
)

type PopNavigationMsg struct{}
type PushNavigationMsg tea.Model

func PushCmd(m tea.Model) tea.Cmd {
	return func() tea.Msg {
		return PushNavigationMsg(m)
	}
}

func PopCmd() tea.Cmd {
	return func() tea.Msg {
		return PopNavigationMsg{}
	}
}
