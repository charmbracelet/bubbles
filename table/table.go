package table

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Column struct {
	Title string
	Width int
}

type Row []string

type Model struct {
	// Public API
	// Key mappings for navigating the list.
	KeyMap KeyMap
	Styles Styles

	// Private API
	cols   []Column
	rows   []Row
	width  int
	height int
	cursor int

	viewport viewport.Model
}

func New(cols []Column, rows []Row, w, h int) Model {
	vp := viewport.New(w, max(h-1, 0))
	return Model{
		cols:     cols,
		rows:     rows,
		width:    w,
		height:   h,
		cursor:   0,
		viewport: vp,

		KeyMap: DefaultKeyMap(),
		Styles: DefaultStyles(),
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.MoveUp):
			m.MoveUp()
		case key.Matches(msg, m.KeyMap.MoveDown):
			m.MoveDown()
		case key.Matches(msg, m.KeyMap.GotoTop):
			m.GotoTop()
		case key.Matches(msg, m.KeyMap.GotoBottom):
			m.GotoBottom()
		case key.Matches(msg, m.KeyMap.Quit):
			return m, tea.Quit
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	hCols := m.renderHeaderCols()
	m.SyncViewportContent()

	header := m.Styles.HeaderStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top, hCols...),
	)
	body := m.Styles.RowStyle.Render(m.viewport.View())
	return lipgloss.JoinVertical(lipgloss.Left, header, body)
}

func (m *Model) SyncViewportContent() {
	hCols := m.renderHeaderCols()
	renderedRows := make([]string, 0, len(m.rows))
	for i := range m.rows {
		renderedRows = append(renderedRows, m.renderRow(i, hCols))
	}

	m.viewport.SetContent(
		lipgloss.JoinVertical(lipgloss.Left, renderedRows...),
	)
}

// Cursor returns the index of the selected row.
func (m Model) Cursor() int {
	return m.cursor
}

// SelectedRow returns the selected row.
// You can cast it to your own implementation.
func (m Model) SelectedRow() Row {
	return m.rows[m.cursor]
}

func (m *Model) SetRows(r []Row) {
	m.rows = r
	m.SyncViewportContent()
}

// CursorIsAtTop of the table.
func (m Model) CursorIsAtTop() bool {
	return m.cursor == 0
}

// CursorIsAtBottom of the table.
func (m Model) CursorIsAtBottom() bool {
	return m.cursor == len(m.rows)-1
}

// MoveUp moves the selection to the previous row.
// It can not go above the first row.
func (m *Model) MoveUp() {
	if m.CursorIsAtTop() {
		return
	}

	m.cursor--
	m.SyncViewportContent()

	if m.cursor < m.viewport.YOffset {
		m.viewport.LineUp(1)
	}
}

// MoveDown moves the selection to the next row.
// It can not go below the last row.
func (m *Model) MoveDown() {
	if m.CursorIsAtBottom() {
		return
	}

	m.cursor++
	m.SyncViewportContent()

	if m.cursor > (m.viewport.YOffset + (m.viewport.Height - 1)) {
		m.viewport.LineDown(1)
	}
}

//  moves the selection to the first row.
func (m *Model) GotoTop() {
	if m.CursorIsAtTop() {
		return
	}

	m.cursor = 0
	m.SyncViewportContent()
	m.viewport.GotoTop()
}

// GoBottom moves the selection to the last row.
func (m *Model) GotoBottom() {
	if m.CursorIsAtBottom() {
		return
	}

	m.cursor = len(m.rows) - 1
	m.SyncViewportContent()
	m.viewport.GotoBottom()
}

func (m Model) renderHeaderCols() []string {
	var hCols []string
	for _, c := range m.cols {
		hCols = append(hCols, m.Styles.TitleCellStyle.
			Copy().
			Width(c.Width).
			Render(c.Title))
	}

	return hCols
}

func (m *Model) renderRow(rowId int, headerColumns []string) string {
	var style lipgloss.Style
	if m.Cursor() == rowId {
		style = m.Styles.SelectedCellStyle
	} else {
		style = m.Styles.CellStyle
	}

	renderedColumns := make([]string, len(m.cols))
	for i, value := range m.rows[rowId] {
		colWidth := lipgloss.Width(headerColumns[i])
		col := style.Copy().Width(colWidth).MaxWidth(colWidth).Render(value)
		renderedColumns = append(renderedColumns, col)
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, renderedColumns...)
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}
