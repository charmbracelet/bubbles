package navigator

import (
	tea "github.com/charmbracelet/bubbletea"
)

// PageEntering && PageLeaving will call every time
type PageEntering interface {
	OnEntering() tea.Cmd
}
type PageLeaving interface {
	OnLeaving() tea.Cmd
}

// PageDestroy && PageInit only once call
//
//	type PageInit interface {
//		Init() tea.Cmd
//	}

func (m Model) pageInitCommands(p tea.Model) []tea.Cmd {
	if p == nil {
		return nil
	}

	var cmds []tea.Cmd
	cmds = append(cmds, p.Init())
	if e, ok := p.(PageEntering); ok {
		cmds = append(cmds, e.OnEntering())
	}

	return cmds
}

func (m Model) pageLeaveCommands(p tea.Model) []tea.Cmd {
	if p == nil {
		return nil
	}

	var cmds []tea.Cmd
	if e, ok := p.(PageLeaving); ok {
		cmds = append(cmds, e.OnLeaving())
	}

	return cmds
}

func (m Model) pageEnterCommands(p tea.Model) []tea.Cmd {
	if p == nil {
		return nil
	}

	var cmds []tea.Cmd
	if e, ok := p.(PageEntering); ok {
		cmds = append(cmds, e.OnEntering())
	}

	return cmds
}

func (m Model) sendSeqCmd(cmds []tea.Cmd) tea.Cmd {

	filterCMds := make([]tea.Cmd, 0, len(cmds))
	for _, cmd := range cmds {
		if cmd != nil {
			filterCMds = append(filterCMds, cmd)
		}
	}

	if len(filterCMds) == 0 {
		return nil
	}

	return tea.Sequence(filterCMds...)
}
