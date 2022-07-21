package tabs

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

func New(tabs []Item) *Model {
	m := &Model{}
	m.tabs = tabs
	m.style = NewStyle()

	physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))
	m.SetWidth(physicalWidth)

	return m
}

type Model struct {
	tabs  []Item
	style style
	Index int
	width int
}

func (m *Model) SetWidth(width int) {
	m.width = width
}

func (m *Model) AddTab(title string) {
	m.tabs = append(m.tabs, Item{Title: title})
}

func (m *Model) SetActive(i int) {
	m.Index = i
}
func (m *Model) NextTab() {
	if m.Index+1 >= len(m.tabs) {
		m.SetActive(0)
	} else {
		m.SetActive(m.Index + 1)
	}
}
func (m *Model) PrevTab() {
	if m.Index-1 < 0 {
		m.SetActive(len(m.tabs) - 1)
	} else {
		m.SetActive(m.Index - 1)
	}

}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	doc := strings.Builder{}
	row := lipgloss.JoinHorizontal(lipgloss.Top, m.getTabs()...)
	gap := m.style.GetTabGap().Render(strings.Repeat(" ", max(0, m.width-lipgloss.Width(row)-2)))
	row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)
	doc.WriteString(row)
	return doc.String()
}

func (m Model) getTabs() []string {
	out := make([]string, len(m.tabs))
	for i, t := range m.tabs {
		if i == m.Index {
			t.Active = true
		}
		out = append(out, m.style.Render(t.Title, t.Active))
	}
	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
