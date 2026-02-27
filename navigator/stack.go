package navigator

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Stack []tea.Model

func (s *Stack) Push(v tea.Model) {
	*s = append(*s, v)
}

func (s *Stack) Pop() tea.Model {
	if len(*s) == 0 {
		return nil
	}
	top := len(*s) - 1
	v := (*s)[top]
	*s = (*s)[:top]
	return v
}

func (s *Stack) Top() tea.Model {
	if len(*s) == 0 {
		return nil
	}
	return (*s)[len(*s)-1]
}
