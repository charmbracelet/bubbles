package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/scrollbar"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func newModel() model {
	// Viewport
	vp := viewport.New(0, 0)

	// Scrollbar
	sb := scrollbar.NewVertical()
	sb.Style = sb.Style.Border(lipgloss.RoundedBorder(), true)

	return model{
		viewport:  vp,
		scrollbar: sb,
	}
}

type model struct {
	content   string
	viewport  viewport.Model
	scrollbar tea.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case " ":
			if m.content != "" {
				m.content += "\n"
			}
			m.content += fmt.Sprintf("%02d: Lorem ipsum dolor sit amet, consectetur adipiscing elit.", lipgloss.Height(m.content)-1)
		}
	case tea.WindowSizeMsg:
		// Update viewport size
		m.viewport.Width = msg.Width - 3
		m.viewport.Height = msg.Height

		// Update scrollbar height
		m.scrollbar, cmd = m.scrollbar.Update(scrollbar.HeightMsg(msg.Height))
		cmds = append(cmds, cmd)
	}

	m.viewport.SetContent(m.content)
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	// Update scrollbar viewport
	m.scrollbar, cmd = m.scrollbar.Update(m.viewport)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.viewport.TotalLineCount() > m.viewport.VisibleLineCount() {
		return lipgloss.JoinHorizontal(lipgloss.Left,
			m.viewport.View(),
			m.scrollbar.View(),
		)
	}

	return m.viewport.View()
}

func main() {
	p := tea.NewProgram(
		newModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}
