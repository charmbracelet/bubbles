package table

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	columnStyle = lipgloss.NewStyle()
)

type Column struct {
	Title string
	Width int
}

type Row []string

type KeyMap struct {
	// The quit keybinding. This won't be caught when filtering.
	Quit key.Binding
}

type Model struct {
	// Key mappings for navigating the list.
	KeyMap KeyMap

	cols   []Column
	rows   []Row
	width  int
	height int

	rowsViewport viewport.Model
}

func New(cols []Column, rows []Row, w, h int) Model {
	viewport := viewport.New(w, len(rows) + 1)
	return Model{
		cols:         cols,
		rows:         rows,
		width:        w,
		height:       h,
		KeyMap:       DefaultKeyMap(),
		rowsViewport: viewport,
	}
}

func (m *Model) SetRows(r []Row) {
	m.rows = r
}
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.KeyMap.Quit) {
			return m, tea.Quit
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var headerCols []string
	for _, c := range m.cols {
		headerCols = append(headerCols, columnStyle.
			Copy().
			Width(c.Width).
			Render(c.Title))
	}
	header := lipgloss.JoinHorizontal(lipgloss.Top, headerCols...)

	renderedRows := make([]string, 0, len(m.rows))
	for i := range m.rows {
		renderedRows = append(renderedRows, m.renderRow(i, headerCols))
	}

	m.rowsViewport.SetContent(
		lipgloss.JoinVertical(lipgloss.Left, renderedRows...),
	)

	return lipgloss.JoinVertical(lipgloss.Left, header, m.rowsViewport.View())
}

func (m *Model) SyncViewPortContent() {
}

func (m *Model) renderRow(rowId int, headerColumns []string) string {
	style := lipgloss.NewStyle().MaxHeight(1)

	renderedColumns := make([]string, len(m.cols))
	for i, value := range m.rows[rowId] {
		colWidth := lipgloss.Width(headerColumns[i])
		col := style.Copy().Width(colWidth).MaxWidth(colWidth).Render(value)
		renderedColumns = append(renderedColumns, col)
	}

	return style.Copy().Render(
		lipgloss.JoinHorizontal(lipgloss.Left, renderedColumns...),
	)
}

// DefaultKeyMap returns a default set of keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Quitting.
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q", "quit"),
		),
	}
}
