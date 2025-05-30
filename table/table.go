// Package table provides a table component for Bubble Tea applications.
package table

import (
	"reflect"

	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/lipgloss/v2/table"
)

// Model defines a state for the table widget.
type Model struct {
	KeyMap KeyMap
	Help   help.Model

	cursor       int
	focus        bool
	styles       Styles
	useStyleFunc bool

	table *table.Table
}

// KeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type KeyMap struct {
	LineUp       key.Binding
	LineDown     key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	GotoTop      key.Binding
	GotoBottom   key.Binding
}

// ShortHelp implements the KeyMap interface.
func (km KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.LineUp, km.LineDown}
}

// FullHelp implements the KeyMap interface.
func (km KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{km.LineUp, km.LineDown, km.GotoTop, km.GotoBottom},
		{km.PageUp, km.PageDown, km.HalfPageUp, km.HalfPageDown},
	}
}

// DefaultKeyMap returns a default set of keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		LineUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		LineDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("b", "pgup"),
			key.WithHelp("b/pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("f", "pgdown", "space"),
			key.WithHelp("f/pgdn", "page down"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("u", "ctrl+u"),
			key.WithHelp("u", "½ page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("d", "ctrl+d"),
			key.WithHelp("d", "½ page down"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "go to start"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "go to end"),
		),
	}
}

// Styles contains style definitions for this table component. Load default
// styles to your table with [DefaultStyles].
type Styles struct {
	Header   lipgloss.Style
	Cell     lipgloss.Style
	Selected lipgloss.Style
}

// DefaultStyles returns sensible default table styles.
func DefaultStyles() Styles {
	return Styles{
		Header:   lipgloss.NewStyle().Bold(true).Padding(0, 1),
		Cell:     lipgloss.NewStyle().Padding(0, 1),
		Selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Padding(0, 1),
	}
}

// NewFromTemplate lets you create a table [Model] from Lip Gloss'
// [table.Table].
func NewFromTemplate(t *table.Table) *Model {
	return &Model{
		cursor:       0,
		KeyMap:       DefaultKeyMap(),
		Help:         help.New(),
		table:        t,
		useStyleFunc: true,
	}
}

// SetBorder is a shorthand function for setting or unsetting borders on a
// table. The arguments work as follows:
//
// With one argument, the argument is applied to all sides.
//
// With two arguments, the arguments are applied to the vertical and horizontal
// sides, in that order.
//
// With three arguments, the arguments are applied to the top side, the
// horizontal sides, and the bottom side, in that order.
//
// With four arguments, the arguments are applied clockwise starting from the
// top side, followed by the right side, then the bottom, and finally the left.
//
// With five arguments, the arguments are applied clockwise starting from the
// top side, followed by the right side, then the bottom, and finally the left.
// The final value will set the row separator.
//
// With six arguments, the arguments are applied clockwise starting from the
// top side, followed by the right side, then the bottom, and finally the left.
// The final two values will set the row and column separators in that order.
//
// With more than six arguments nothing will be set.
func (m *Model) SetBorder(s ...bool) *Model {
	top, right, bottom, left, rowSeparator, columnSeparator := m.whichSides(s...)
	m.table.
		BorderTop(top).
		BorderRight(right).
		BorderBottom(bottom).
		BorderLeft(left).
		BorderRow(rowSeparator).
		BorderColumn(columnSeparator)
	return m
}

// Border sets the kind of border to use for the table. See [lipgloss.Border].
func (m *Model) Border(border lipgloss.Border) *Model {
	m.table.Border(border)
	return m
}

// BorderStyle sets the style for the table border.
func (m *Model) BorderStyle(style lipgloss.Style) *Model {
	m.table.BorderStyle(style)
	return m
}

// BorderBottom sets the bottom border.
func (m *Model) BorderBottom(v bool) *Model {
	m.table.BorderBottom(v)
	return m
}

// BorderTop sets the top border.
func (m *Model) BorderTop(v bool) *Model {
	m.table.BorderTop(v)
	return m
}

// BorderLeft sets the left border.
func (m *Model) BorderLeft(v bool) *Model {
	m.table.BorderLeft(v)
	return m
}

// BorderRight sets the right border.
func (m *Model) BorderRight(v bool) *Model {
	m.table.BorderRight(v)
	return m
}

// BorderColumn sets the column border.
func (m *Model) BorderColumn(v bool) *Model {
	m.table.BorderColumn(v)
	return m
}

// BorderHeader sets the header border.
func (m *Model) BorderHeader(v bool) *Model {
	m.table.BorderHeader(v)
	return m
}

// BorderRow sets the row borders.
func (m *Model) BorderRow(v bool) *Model {
	m.table.BorderRow(v)
	return m
}

// Options

// Option is used to set options in [New]. For example:
//
//	table := New(WithHeaders([]string{"Rank", "City", "Country", "Population"}))
type Option func(*Model)

// WithHeaders sets the table headers. This function is used as an [Option] in
// when creating a table with [New].
func WithHeaders(headers ...string) Option {
	return func(m *Model) {
		m.SetHeaders(headers...)
	}
}

// WithHeight sets the height of the table. The given height will be the total
// table height including borders, margins, and padding. This function is used
// as an [Option] in when creating a table with [New].
func WithHeight(h int) Option {
	return func(m *Model) {
		m.table.Height(h)
	}
}

// WithWidth sets the width of the table. The given width will be the total
// table width including borders, margins, and padding. This function is used as
// an [Option] in when creating a table with [New].
func WithWidth(w int) Option {
	return func(m *Model) {
		m.table.Width(w)
	}
}

// WithRows sets the table rows. This function is used as an [Option] in when
// creating a table with [New].
func WithRows(rows ...[]string) Option {
	return func(m *Model) {
		m.SetRows(rows...)
	}
}

// WithFocused sets the focus state of the table. This function is used as an
// [Option] in when creating a table with [New].
func WithFocused(f bool) Option {
	return func(m *Model) {
		m.focus = f
	}
}

// WithStyles sets the table styles. This function is used as an [Option] in
// when creating a table with [New].
func WithStyles(s Styles) Option {
	return func(m *Model) {
		m.SetStyles(s)
	}
}

// WithStyleFunc sets the table [table.StyleFunc] for conditional styling. This
// function is used as an [Option] in when creating a table with [New].
func WithStyleFunc(s table.StyleFunc) Option {
	return func(m *Model) {
		m.useStyleFunc = true
		m.table.StyleFunc(s)
	}
}

// WithKeyMap sets the [KeyMap]. This function is used as an [Option] in when
// creating a table with [New].
func WithKeyMap(km KeyMap) Option {
	return func(m *Model) {
		m.KeyMap = km
	}
}

// Setters

// SetHeaders sets the table headers.
func (m *Model) SetHeaders(headers ...string) *Model {
	m.table.Headers(headers...)
	return m
}

// SetRows sets the table rows.
func (m *Model) SetRows(rows ...[]string) *Model {
	m.table.Rows(rows...)
	return m
}

// SetCursor sets the cursor position in the table.
func (m *Model) SetCursor(n int) *Model {
	m.cursor = clamp(n, 0, m.RowCount()-1)
	return m
}

// SetHeight sets the width of the table. The given height will be the total
// table height including borders, margins, and padding.
func (m *Model) SetHeight(h int) *Model {
	m.table.Height(h)
	return m
}

// SetWidth sets the width of the table. The given width will be the total
// table width including borders, margins, and padding.
func (m *Model) SetWidth(w int) *Model {
	m.table.Width(w)
	return m
}

// SetYOffset sets the YOffset position in the table.
func (m *Model) SetYOffset(n int) *Model {
	var (
		minimum = 0
		maximum = m.RowCount() - m.table.VisibleRows()
		yOffset = clamp(n, minimum, maximum)
	)
	m.table.YOffset(yOffset)
	return m
}

// SetStyles sets the table styles, only applying non-empty [Styles]. Note: using
// [Model.SetStyleFunc] will override styles set in this function.
func (m *Model) SetStyles(s Styles) *Model {
	if !reflect.DeepEqual(s.Selected, lipgloss.Style{}) {
		m.styles.Selected = s.Selected
	}
	if !reflect.DeepEqual(s.Header, lipgloss.Style{}) {
		m.styles.Header = s.Header
	}
	if !reflect.DeepEqual(s.Cell, lipgloss.Style{}) {
		m.styles.Cell = s.Cell
	}
	return m
}

// OverwriteStyles sets the table styles, overwriting all existing styles. Note:
// using [Model.SetStyleFunc] will override styles set in this function.
func (m *Model) OverwriteStyles(s Styles) *Model {
	m.styles = s
	return m
}

// LipglossTable sets the inner [lipgloss.Table].
func (m *Model) LipglossTable(t *table.Table) {
	var (
		previousHeaders = m.table.GetHeaders()
		previousData    = m.table.GetData()
	)
	if len(t.GetHeaders()) == 0 {
		t = t.Headers(previousHeaders...)
	}
	if t.GetData() == nil || t.GetData().Rows() == 0 {
		t = t.Rows(table.DataToMatrix(previousData)...)
	}
	m.table = t
	m.useStyleFunc = true
}

// SetStyleFunc sets the table's custom [table.StyleFunc]. Use this for conditional
// styling e.g. styling a cell by its contents or by index.
func (m *Model) SetStyleFunc(s table.StyleFunc) *Model {
	m.useStyleFunc = true
	m.table.StyleFunc(s)
	return m
}

// Creation

// New creates a new model for the table widget.
func New(opts ...Option) *Model {
	m := Model{
		cursor: 0,
		KeyMap: DefaultKeyMap(),
		Help:   help.New(),
		table:  table.New(),
	}

	m.SetStyles(DefaultStyles())

	// Set border defaults here
	m.Border(lipgloss.NormalBorder())
	m.BorderTop(true)
	m.BorderBottom(true)
	m.BorderLeft(true)
	m.BorderRight(true)
	m.BorderColumn(false)
	m.BorderRow(false)
	m.BorderHeader(true)

	for _, opt := range opts {
		opt(&m)
	}

	return &m
}

// Bubble Tea Methods

// Update is the Bubble Tea [tea.Model] update loop.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focus {
		return m, nil
	}
	table := m.table.String()
	// XXX: make this not hard coded?
	height := lipgloss.Height(table) - 6

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.KeyMap.LineUp):
			m.MoveUp(1)
		case key.Matches(msg, m.KeyMap.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, m.KeyMap.PageUp):
			m.MoveUp(height)
		case key.Matches(msg, m.KeyMap.PageDown):
			m.MoveDown(height)
		case key.Matches(msg, m.KeyMap.HalfPageUp):
			m.MoveUp(height / 2) //nolint:mnd
		case key.Matches(msg, m.KeyMap.HalfPageDown):
			m.MoveDown(height / 2) //nolint:mnd
		case key.Matches(msg, m.KeyMap.GotoTop):
			m.GotoTop()
		case key.Matches(msg, m.KeyMap.GotoBottom):
			m.GotoBottom()
		}
	}

	return m, nil
}

// Focus focuses the table, allowing the user to move around the rows and
// interact.
func (m *Model) Focus() {
	m.focus = true
}

// Blur blurs the table, preventing selection or movement.
func (m *Model) Blur() {
	m.focus = false
}

// View renders the table [Model].
func (m Model) View() string {
	if !m.useStyleFunc {
		// Update the position-sensitive styles as the cursor position may have
		// changed in Update.
		m.table.StyleFunc(func(row, _ int) lipgloss.Style {
			if row == m.cursor {
				return m.styles.Selected
			}
			if row == table.HeaderRow {
				return m.styles.Header
			}
			return m.styles.Cell
		})
	}
	return m.table.String()
}

// HelpView is a helper method for rendering the help menu from the keymap.
// Note that this view is not rendered by default and you must call it
// manually in your application, where applicable.
func (m Model) HelpView() string {
	return m.Help.View(m.KeyMap)
}

// Getters

// Focused returns the focus state of the table.
func (m Model) Focused() bool {
	return m.focus
}

// Rows returns the current rows.
func (m Model) Rows() [][]string {
	return table.DataToMatrix(m.table.GetData())
}

// RowCount returns the number of rows in the table.
func (m Model) RowCount() int {
	return m.table.GetData().Rows()
}

// Headers returns the current headers.
func (m Model) Headers() []string {
	return m.table.GetHeaders()
}

// Cursor returns the index of the selected row.
func (m Model) Cursor() int {
	return m.cursor
}

// SelectedRow returns the selected row. You can cast it to your own
// implementation.
func (m Model) SelectedRow() []string {
	if m.cursor < 0 || m.cursor >= m.RowCount() {
		return nil
	}

	return table.DataToMatrix(m.table.GetData())[m.cursor]
}

// Movement

// MoveUp moves the selection up by any number of rows.
// It can not go above the first row.
func (m *Model) MoveUp(n int) {
	m.SetCursor(m.cursor - n)

	if m.cursor < m.table.FirstVisibleRowIndex() {
		m.SetYOffset(m.table.GetYOffset() - n)
	}
}

// MoveDown moves the selection down by any number of rows.
// It can not go below the last row.
func (m *Model) MoveDown(n int) {
	m.SetCursor(m.cursor + n)

	if m.cursor > m.table.LastVisibleRowIndex() {
		m.SetYOffset(m.table.GetYOffset() + n)
	}
}

// GotoTop moves the selection to the first row.
func (m *Model) GotoTop() {
	m.MoveUp(m.cursor)
}

// GotoBottom moves the selection to the last row.
func (m *Model) GotoBottom() {
	m.MoveDown(m.RowCount())
}

// Helpers

func clamp(v, low, high int) int {
	return min(max(v, low), high)
}

// whichSides is a helper method for setting values on sides of a block based on
// the number of arguments given.
// 0: set all sides to true
// 1: set all sides to given arg
// 2: top -> bottom
// 3: top -> horizontal -> bottom
// 4: top -> right -> bottom -> left
// 5: top -> right -> bottom -> left -> rowSeparator
// 6: top -> right -> bottom -> left -> rowSeparator -> columnSeparator
func (m Model) whichSides(s ...bool) (top, right, bottom, left, rowSeparator, columnSeparator bool) {
	// set the separators to true unless otherwise set.
	rowSeparator = m.table.GetBorderRow()
	columnSeparator = m.table.GetBorderColumn()

	switch len(s) {
	case 1:
		top = s[0]
		right = s[0]
		bottom = s[0]
		left = s[0]
		rowSeparator = s[0]
		columnSeparator = s[0]
	case 2:
		top = s[0]
		right = s[1]
		bottom = s[0]
		left = s[1]
	case 3:
		top = s[0]
		right = s[1]
		bottom = s[2]
		left = s[1]
	case 4:
		top = s[0]
		right = s[1]
		bottom = s[2]
		left = s[3]
	case 5:
		top = s[0]
		right = s[1]
		bottom = s[2]
		left = s[3]
		rowSeparator = s[4]
	case 6:
		top = s[0]
		right = s[1]
		bottom = s[2]
		left = s[3]
		rowSeparator = s[4]
		columnSeparator = s[5]
	default:
		top = m.table.GetBorderTop()
		right = m.table.GetBorderRight()
		bottom = m.table.GetBorderBottom()
		left = m.table.GetBorderLeft()
	}
	return top, right, bottom, left, rowSeparator, columnSeparator
}
