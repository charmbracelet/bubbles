package main

import (
	"fmt"
	"log"
	
	"github.com/charmbracelet/bubbles/v2/textarea"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type model struct {
	textarea textarea.Model
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Try text selection with mouse or Shift+arrows!"
	ta.SetValue("Welcome to Bubbles!\n\nThis textarea now supports:\n- Mouse selection (click and drag)\n- Double-click to select words\n- Triple-click to select lines\n- Shift+arrows for keyboard selection\n- Ctrl+A to select all\n- Ctrl+C/X to copy/cut")
	ta.Focus()
	ta.SetWidth(60)
	ta.SetHeight(10)
	
	return model{textarea: ta}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.textarea.SetWidth(msg.Width)
		m.textarea.SetHeight(msg.Height - 4)
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" && !m.textarea.HasSelection() {
			return m, tea.Quit
		}
	}
	
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m model) View() string {
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	
	status := ""
	if m.textarea.HasSelection() {
		selected := m.textarea.GetSelectedText()
		if len(selected) > 30 {
			selected = selected[:27] + "..."
		}
		status = fmt.Sprintf("Selected: %q", selected)
	}
	
	return fmt.Sprintf(
		"%s\n\n%s\n%s",
		m.textarea.View(),
		status,
		help.Render("ctrl+c to quit • mouse/keyboard to select text"),
	)
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseAllMotion(), // Enable mouse support
	)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}