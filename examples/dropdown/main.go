// Example program demonstrating the dropdown component.
//
// Controls:
//
//   - tab / shift+tab  cycle focus between dropdowns
//   - ↑/k, ↓/j        navigate options (when expanded)
//   - enter / space    open or confirm selection
//   - esc              close without selecting
//   - r                reload options in the first dropdown at runtime
//   - q / ctrl+c       quit
package main

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/dropdown"
	tea "charm.land/bubbletea/v2"
)

const numDropdowns = 3

var charmOpts = []dropdown.Option{
	{Label: "Bubble Tea", Value: "bubbletea"},
	{Label: "Lip Gloss", Value: "lipgloss"},
	{Label: "Bubbles", Value: "bubbles"},
	{Label: "Huh", Value: "huh"},
	{Label: "Wish", Value: "wish"},
}

var langOpts = []dropdown.Option{
	{Label: "Go", Value: "go"},
	{Label: "Rust", Value: "rust"},
	{Label: "Zig", Value: "zig"},
	{Label: "C", Value: "c"},
}

type model struct {
	dropdowns    [numDropdowns]dropdown.Model
	focusIndex   int
	lastSelected string
}

func initialModel() model {
	// Dropdown 0: normal, interactive.
	dd0 := dropdown.New(
		dropdown.WithOptions(charmOpts...),
		dropdown.WithWidth(18),
	)
	dd0.Focus() //nolint:errcheck

	// Dropdown 1: disabled, with a pre-committed selection.
	dd1 := dropdown.New(
		dropdown.WithOptions(charmOpts...),
		dropdown.WithWidth(18),
	)
	// Pre-select "Lip Gloss" (index 1) by running Update with a temporary focus.
	dd1.Focus() //nolint:errcheck
	dd1, _ = dd1.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // open
	dd1, _ = dd1.Update(tea.KeyPressMsg{Code: tea.KeyDown})  // cursor → 1
	dd1, _ = dd1.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // confirm
	dd1.Blur()
	dd1.Disabled = true

	// Dropdown 2: empty options — shows placeholder only, never opens.
	dd2 := dropdown.New(dropdown.WithWidth(18))

	return model{
		dropdowns:  [numDropdowns]dropdown.Model{dd0, dd1, dd2},
		focusIndex: 0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab", "shift+tab":
			step := 1
			if msg.String() == "shift+tab" {
				step = -1
			}
			m.dropdowns[m.focusIndex].Blur()
			m.focusIndex = (m.focusIndex + step + numDropdowns) % numDropdowns
			m.dropdowns[m.focusIndex].Focus() //nolint:errcheck
			return m, nil

		case "r":
			// Reload options at runtime (only when collapsed).
			if !m.dropdowns[0].IsOpen() {
				m.dropdowns[0].SetOptions(langOpts)
				m.lastSelected = "(options reloaded)"
			}
			return m, nil
		}

	case dropdown.SelectMsg:
		m.lastSelected = fmt.Sprintf("selected %q (value: %q, index: %d)",
			msg.Option.Label, msg.Option.Value, msg.Index)
		return m, nil

	case dropdown.CloseMsg:
		m.lastSelected = "(closed without selecting)"
		return m, nil
	}

	var cmd tea.Cmd
	m.dropdowns[m.focusIndex], cmd = m.dropdowns[m.focusIndex].Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	var sb strings.Builder

	sb.WriteString("Dropdown Component Demo\n")
	sb.WriteString("───────────────────────────────────────────────\n\n")

	labels := []string{"Normal (tab to focus)", "Disabled", "Empty"}
	for i, dd := range m.dropdowns {
		sb.WriteString(fmt.Sprintf("  %s\n", labels[i]))
		for _, line := range strings.Split(dd.View(), "\n") {
			sb.WriteString("  " + line + "\n")
		}
		sb.WriteRune('\n')
	}

	sb.WriteString("───────────────────────────────────────────────\n")
	if m.lastSelected != "" {
		sb.WriteString("  " + m.lastSelected + "\n")
	}
	sb.WriteString("\n  tab/shift+tab: focus  r: reload opts  q: quit\n")

	return tea.NewView(sb.String())
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
