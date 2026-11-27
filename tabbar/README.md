# TabBar Component for Bubble Tea

A simple, customizable tab bar component for the [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework.

![TabBar Demo](https://github.com/charmbracelet/bubbles/raw/master/tabbar/demo.gif)

## Features

- Horizontal tabs with customizable styling
- Active tab highlighting
- Keyboard navigation between tabs
- Adjustable width
- Simple API integration with Bubble Tea applications

## Installation

```bash
go get github.com/charmbracelet/bubbles
```

## Usage

```go
import "github.com/charmbracelet/bubbles/tabbar"
```

### Basic Example

```go
// Create tabs
tabs := []string{"Home", "Projects", "Settings", "About"}

// Initialize the tab bar with Home tab active (index 0)
bar := tabbar.New(tabs, 0)

// In your model's Update method:
switch msg := msg.(type) {
case tea.KeyMsg:
    switch msg.String() {
    case "right", "tab":
        return m, m.tabBar.Next()
    case "left", "shift+tab":
        return m, m.tabBar.Prev()
    }
case tabbar.TabChangeMsg:
    // Handle tab change if needed
}

// In your model's View method:
func (m model) View() string {
    return m.tabBar.View()
}
```

### Customization

You can customize the appearance of the tab bar:

```go
// Customize colors
bar.ActiveBorderColor = lipgloss.Color("#ff0000")    // Red for active tab
bar.InactiveBorderColor = lipgloss.Color("#333333")  // Dark gray for inactive tabs

// Set width to fit your layout
bar.SetWidth(100)
```

## API

### Functions

- `New(tabs []string, activeIndex int) TabBar` - Create a new tab bar
- `(t TabBar) ActiveTab() int` - Get the active tab index
- `(t *TabBar) Next() tea.Cmd` - Move to next tab
- `(t *TabBar) Prev() tea.Cmd` - Move to previous tab
- `(t *TabBar) Activate(index int) tea.Cmd` - Activate tab at specific index
- `(t *TabBar) SetWidth(width int)` - Set width of the tab bar
- `(t TabBar) View() string` - Render the tab bar

### Message Types

- `TabChangeMsg` - Sent when the active tab changes

## Example

See the [example](./example) directory for a complete working example.

## License

[MIT](https://github.com/charmbracelet/bubbles/blob/master/LICENSE)

## Credits

This component was contributed to the Bubble Tea ecosystem by the CloudWorkstation project.