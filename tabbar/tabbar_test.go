package tabbar

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestNewTabBar(t *testing.T) {
	// Create a new TabBar
	tabs := []string{"Tab 1", "Tab 2", "Tab 3"}
	bar := New(tabs, 0)

	// Check initial state
	if bar.ActiveTabIndex != 0 {
		t.Errorf("Expected ActiveTabIndex to be 0, got %d", bar.ActiveTabIndex)
	}

	if len(bar.Tabs) != len(tabs) {
		t.Errorf("Expected %d tabs, got %d", len(tabs), len(bar.Tabs))
	}

	// Check tab titles
	for i, tab := range bar.Tabs {
		if tab.title != tabs[i] {
			t.Errorf("Expected tab %d to have title %q, got %q", i, tabs[i], tab.title)
		}
	}
}

func TestActiveTab(t *testing.T) {
	// Create a new TabBar with second tab active
	bar := New([]string{"Tab 1", "Tab 2", "Tab 3"}, 1)

	// Check ActiveTab method
	if bar.ActiveTab() != 1 {
		t.Errorf("Expected ActiveTab() to return 1, got %d", bar.ActiveTab())
	}
}

func TestNextPrev(t *testing.T) {
	// Create a new TabBar
	bar := New([]string{"Tab 1", "Tab 2", "Tab 3"}, 0)

	// Test Next method
	var cmd tea.Cmd = bar.Next()
	if bar.ActiveTabIndex != 1 {
		t.Errorf("After Next(), expected ActiveTabIndex to be 1, got %d", bar.ActiveTabIndex)
	}

	// Execute command and check message
	msg := cmd()
	if tabChangeMsg, ok := msg.(TabChangeMsg); !ok || tabChangeMsg.Index != 1 {
		t.Errorf("Expected Next() to return TabChangeMsg with Index 1, got %v", msg)
	}

	// Test Next again (should go to last tab)
	bar.Next()
	if bar.ActiveTabIndex != 2 {
		t.Errorf("After second Next(), expected ActiveTabIndex to be 2, got %d", bar.ActiveTabIndex)
	}

	// Test wrap-around behavior
	bar.Next()
	if bar.ActiveTabIndex != 0 {
		t.Errorf("After third Next(), expected ActiveTabIndex to wrap to 0, got %d", bar.ActiveTabIndex)
	}

	// Test Prev method
	bar.Prev()
	if bar.ActiveTabIndex != 2 {
		t.Errorf("After Prev(), expected ActiveTabIndex to be 2, got %d", bar.ActiveTabIndex)
	}
}

func TestActivate(t *testing.T) {
	// Create a new TabBar
	bar := New([]string{"Tab 1", "Tab 2", "Tab 3"}, 0)

	// Test Activate method with valid index
	var cmd tea.Cmd = bar.Activate(2)
	if bar.ActiveTabIndex != 2 {
		t.Errorf("After Activate(2), expected ActiveTabIndex to be 2, got %d", bar.ActiveTabIndex)
	}

	// Execute command and check message
	msg := cmd()
	if tabChangeMsg, ok := msg.(TabChangeMsg); !ok || tabChangeMsg.Index != 2 {
		t.Errorf("Expected Activate(2) to return TabChangeMsg with Index 2, got %v", msg)
	}

	// Test Activate method with invalid index (too low)
	cmd = bar.Activate(-1)
	if cmd != nil {
		t.Error("Expected Activate(-1) to return nil command")
	}
	if bar.ActiveTabIndex != 2 {
		t.Errorf("After Activate(-1), expected ActiveTabIndex to remain 2, got %d", bar.ActiveTabIndex)
	}

	// Test Activate method with invalid index (too high)
	cmd = bar.Activate(10)
	if cmd != nil {
		t.Error("Expected Activate(10) to return nil command")
	}
	if bar.ActiveTabIndex != 2 {
		t.Errorf("After Activate(10), expected ActiveTabIndex to remain 2, got %d", bar.ActiveTabIndex)
	}
}

func TestSetWidth(t *testing.T) {
	// Create a new TabBar
	bar := New([]string{"Tab 1", "Tab 2"}, 0)
	initialWidth := bar.Width

	// Test SetWidth method
	bar.SetWidth(100)
	if bar.Width != 100 {
		t.Errorf("After SetWidth(100), expected Width to be 100, got %d", bar.Width)
	}

	// Check that initial width was set correctly (default)
	if initialWidth <= 0 {
		t.Errorf("Expected default Width to be positive, got %d", initialWidth)
	}
}

func TestView(t *testing.T) {
	// Create a new TabBar
	bar := New([]string{"Tab 1", "Tab 2"}, 0)
	bar.SetWidth(50)

	// Get rendered view
	view := bar.View()

	// Basic sanity checks
	if len(view) == 0 {
		t.Error("Expected non-empty view")
	}

	// Check that both tab titles appear in the rendered view
	if !strings.Contains(view, "Tab 1") {
		t.Error("Expected view to contain 'Tab 1'")
	}
	if !strings.Contains(view, "Tab 2") {
		t.Error("Expected view to contain 'Tab 2'")
	}

	// Check custom styling
	bar.ActiveBorderColor = lipgloss.Color("#ff0000")
	bar.InactiveBorderColor = lipgloss.Color("#000000")
	
	// Re-render view with custom styling
	view = bar.View()
	if len(view) == 0 {
		t.Error("Expected non-empty view after styling changes")
	}
}