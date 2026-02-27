package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/tabbar"
)

type model struct {
	tabBar      tabbar.TabBar
	content     []string
	width       int
	height      int
}

func initialModel() model {
	tabs := []string{"Home", "Projects", "Settings", "About"}
	
	// Create content for each tab
	content := []string{
		"Welcome to the Home tab!",
		"Here are your projects:\n - Project 1\n - Project 2\n - Project 3",
		"Settings:\n - Theme: Dark\n - Notifications: On\n - Sounds: Off",
		"About:\nThis is a simple example of the tabbar component.",
	}
	
	return model{
		tabBar:  tabbar.New(tabs, 0),
		content: content,
		width:   80,
		height:  24,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "right", "tab":
			return m, m.tabBar.Next()
		case "left", "shift+tab":
			return m, m.tabBar.Prev()
		case "1", "2", "3", "4":
			index := int(msg.Runes[0] - '1')
			if index >= 0 && index < 4 {
				return m, m.tabBar.Activate(index)
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.tabBar.SetWidth(msg.Width)
	case tabbar.TabChangeMsg:
		// No additional action needed, the tabBar state is already updated
	}
	
	return m, nil
}

func (m model) View() string {
	// Tab content
	activeTab := m.tabBar.ActiveTab()
	tabContent := m.content[activeTab]
	
	// Style for the content area
	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#0074D9")).
		Padding(1, 2).
		Width(m.width - 4)
	
	// Help text
	helpText := "\nUse ← and → arrows to switch tabs | 1-4 to jump to tab | q to quit"
	
	// Combine everything
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.tabBar.View(),
		"",
		contentStyle.Render(tabContent),
		helpText,
	)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}