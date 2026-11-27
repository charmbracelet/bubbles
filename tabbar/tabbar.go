// Package tabbar provides a simple tab bar component for Bubble Tea applications.
package tabbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TabChangeMsg is sent when the active tab changes.
type TabChangeMsg struct {
	Index int
}

// Tab represents a tab in the tab bar.
type Tab struct {
	title string
	style lipgloss.Style
}

// TabBar is a simple tab bar component that can be used to switch between different views.
// It renders a horizontal list of tabs with customizable styles and borders.
type TabBar struct {
	Tabs              []Tab
	ActiveTabIndex    int
	ActiveBorderColor lipgloss.Color
	InactiveBorderColor lipgloss.Color
	Width             int
}

// New creates a new tab bar with the given tab titles and active index.
func New(tabs []string, activeIndex int) TabBar {
	// Default colors
	activeBorderColor := lipgloss.Color("#0074D9")   // Blue
	inactiveBorderColor := lipgloss.Color("#AAAAAA") // Light gray
	
	// Create tab objects
	tabItems := make([]Tab, len(tabs))
	for i, title := range tabs {
		tabItems[i] = Tab{
			title: title,
			style: lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(inactiveBorderColor).
				Padding(0, 2),
		}
	}
	
	return TabBar{
		Tabs:              tabItems,
		ActiveTabIndex:    activeIndex,
		ActiveBorderColor: activeBorderColor,
		InactiveBorderColor: inactiveBorderColor,
		Width:             80,
	}
}

// ActiveTab returns the index of the active tab.
func (t TabBar) ActiveTab() int {
	return t.ActiveTabIndex
}

// Next activates the next tab and returns a command that will send a TabChangeMsg.
func (t *TabBar) Next() tea.Cmd {
	if t.ActiveTabIndex < len(t.Tabs)-1 {
		t.ActiveTabIndex++
	} else {
		t.ActiveTabIndex = 0
	}
	return func() tea.Msg {
		return TabChangeMsg{Index: t.ActiveTabIndex}
	}
}

// Prev activates the previous tab and returns a command that will send a TabChangeMsg.
func (t *TabBar) Prev() tea.Cmd {
	if t.ActiveTabIndex > 0 {
		t.ActiveTabIndex--
	} else {
		t.ActiveTabIndex = len(t.Tabs) - 1
	}
	return func() tea.Msg {
		return TabChangeMsg{Index: t.ActiveTabIndex}
	}
}

// Activate activates the tab at the given index and returns a command that will send a TabChangeMsg.
func (t *TabBar) Activate(index int) tea.Cmd {
	if index >= 0 && index < len(t.Tabs) {
		t.ActiveTabIndex = index
		return func() tea.Msg {
			return TabChangeMsg{Index: t.ActiveTabIndex}
		}
	}
	return nil
}

// SetWidth sets the width of the tab bar.
func (t *TabBar) SetWidth(width int) {
	t.Width = width
}

// View renders the tab bar.
func (t TabBar) View() string {
	var renderedTabs []string
	
	// Calculate approximate width for each tab
	tabWidth := (t.Width / len(t.Tabs)) - 4 // Account for borders and spacing
	
	for i, tab := range t.Tabs {
		// Set border color based on active status
		borderColor := t.InactiveBorderColor
		if i == t.ActiveTabIndex {
			borderColor = t.ActiveBorderColor
		}
		
		// Set tab style
		style := tab.style.Copy().
			BorderForeground(borderColor).
			Width(tabWidth)
			
		if i == t.ActiveTabIndex {
			style = style.Bold(true)
		}
		
		// Render tab
		renderedTabs = append(renderedTabs, style.Render(tab.title))
	}
	
	return lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
}