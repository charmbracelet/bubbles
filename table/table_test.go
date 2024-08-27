package table

import (
	"reflect"
	"testing"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
)

var testCols = []Column{
	{Title: "col1", Width: 10},
	{Title: "col2", Width: 10},
	{Title: "col3", Width: 10},
}

func TestNew(t *testing.T) {
	tests := map[string]struct {
		opts []Option
		want Model
	}{
		"Default": {
			want: Model{
				// Default fields
				cursor:   0,
				viewport: viewport.New(0, 20),
				KeyMap:   DefaultKeyMap(),
				Help:     help.New(),
				styles:   DefaultStyles(),
			},
		},
		"WithColumns": {
			opts: []Option{
				WithColumns([]Column{
					{Title: "Foo", Width: 1},
					{Title: "Bar", Width: 2},
				}),
			},
			want: Model{
				// Default fields
				cursor:   0,
				viewport: viewport.New(0, 20),
				KeyMap:   DefaultKeyMap(),
				Help:     help.New(),
				styles:   DefaultStyles(),

				// Modified fields
				cols: []Column{
					{Title: "Foo", Width: 1},
					{Title: "Bar", Width: 2},
				},
			},
		},
		"WithCols; WithRows": {
			opts: []Option{
				WithColumns([]Column{
					{Title: "Foo", Width: 1},
					{Title: "Bar", Width: 2},
				}),
				WithRows([]Row{
					{"1", "Foo"},
					{"2", "Bar"},
				}),
			},
			want: Model{
				// Default fields
				cursor:   0,
				viewport: viewport.New(0, 20),
				KeyMap:   DefaultKeyMap(),
				Help:     help.New(),
				styles:   DefaultStyles(),

				// Modified fields
				cols: []Column{
					{Title: "Foo", Width: 1},
					{Title: "Bar", Width: 2},
				},
				rows: []Row{
					{"1", "Foo"},
					{"2", "Bar"},
				},
			},
		},
		"WithHeight": {
			opts: []Option{
				WithHeight(10),
			},
			want: Model{
				// Default fields
				cursor: 0,
				KeyMap: DefaultKeyMap(),
				Help:   help.New(),
				styles: DefaultStyles(),

				// Modified fields
				// Viewport height is 1 less than the provided height when no header is present since lipgloss.Height adds 1
				viewport: viewport.New(0, 9),
			},
		},
		"WithWidth": {
			opts: []Option{
				WithWidth(10),
			},
			want: Model{
				// Default fields
				cursor: 0,
				KeyMap: DefaultKeyMap(),
				Help:   help.New(),
				styles: DefaultStyles(),

				// Modified fields
				// Viewport height is 1 less than the provided height when no header is present since lipgloss.Height adds 1
				viewport: viewport.New(10, 20),
			},
		},
		"WithFocused": {
			opts: []Option{
				WithFocused(true),
			},
			want: Model{
				// Default fields
				cursor:   0,
				viewport: viewport.New(0, 20),
				KeyMap:   DefaultKeyMap(),
				Help:     help.New(),
				styles:   DefaultStyles(),

				// Modified fields
				focus: true,
			},
		},
		"WithStyles": {
			opts: []Option{
				WithStyles(Styles{}),
			},
			want: Model{
				// Default fields
				cursor:   0,
				viewport: viewport.New(0, 20),
				KeyMap:   DefaultKeyMap(),
				Help:     help.New(),

				// Modified fields
				styles: Styles{},
			},
		},
		"WithKeyMap": {
			opts: []Option{
				WithKeyMap(KeyMap{}),
			},
			want: Model{
				// Default fields
				cursor:   0,
				viewport: viewport.New(0, 20),
				Help:     help.New(),
				styles:   DefaultStyles(),

				// Modified fields
				KeyMap: KeyMap{},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.want.UpdateViewport()

			got := New(tc.opts...)

			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("\n\nwant %v\n\ngot %v", tc.want, got)
			}
		})
	}
}

func TestModel_FromValues(t *testing.T) {
	input := "foo1,bar1\nfoo2,bar2\nfoo3,bar3"
	table := New(WithColumns([]Column{{Title: "Foo"}, {Title: "Bar"}}))
	table.FromValues(input, ",")

	if len(table.rows) != 3 {
		t.Fatalf("expect table to have 3 rows but it has %d", len(table.rows))
	}

	expect := []Row{
		{"foo1", "bar1"},
		{"foo2", "bar2"},
		{"foo3", "bar3"},
	}
	if !reflect.DeepEqual(table.rows, expect) {
		t.Fatalf("\n\nwant %v\n\ngot %v", expect, table.rows)
	}
}

func TestModel_FromValues_WithTabSeparator(t *testing.T) {
	input := "foo1.\tbar1\nfoo,bar,baz\tbar,2"
	table := New(WithColumns([]Column{{Title: "Foo"}, {Title: "Bar"}}))
	table.FromValues(input, "\t")

	if len(table.rows) != 2 {
		t.Fatalf("expect table to have 2 rows but it has %d", len(table.rows))
	}

	expect := []Row{
		{"foo1.", "bar1"},
		{"foo,bar,baz", "bar,2"},
	}
	if !reflect.DeepEqual(table.rows, expect) {
		t.Fatalf("\n\nwant %v\n\ngot %v", expect, table.rows)
	}
}

func TestModel_RenderRow(t *testing.T) {
	tests := []struct {
		name     string
		table    *Model
		expected string
	}{
		{
			name: "simple row",
			table: &Model{
				rows:   []Row{{"Foooooo", "Baaaaar", "Baaaaaz"}},
				cols:   testCols,
				styles: Styles{Cell: lipgloss.NewStyle()},
			},
			expected: "Foooooo   Baaaaar   Baaaaaz   ",
		},
		{
			name: "simple row with truncations",
			table: &Model{
				rows:   []Row{{"Foooooooooo", "Baaaaaaaaar", "Quuuuuuuuux"}},
				cols:   testCols,
				styles: Styles{Cell: lipgloss.NewStyle()},
			},
			expected: "Foooooooo…Baaaaaaaa…Quuuuuuuu…",
		},
		{
			name: "simple row avoiding truncations",
			table: &Model{
				rows:   []Row{{"Fooooooooo", "Baaaaaaaar", "Quuuuuuuux"}},
				cols:   testCols,
				styles: Styles{Cell: lipgloss.NewStyle()},
			},
			expected: "FoooooooooBaaaaaaaarQuuuuuuuux",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			row := tc.table.renderRow(0)
			if row != tc.expected {
				t.Fatalf("\n\nWant: \n%s\n\nGot:  \n%s\n", tc.expected, row)
			}
		})
	}
}

func TestTableAlignment(t *testing.T) {
	t.Run("No border", func(t *testing.T) {
		biscuits := New(
			WithHeight(5),
			WithColumns([]Column{
				{Title: "Name", Width: 25},
				{Title: "Country of Origin", Width: 16},
				{Title: "Dunk-able", Width: 12},
			}),
			WithRows([]Row{
				{"Chocolate Digestives", "UK", "Yes"},
				{"Tim Tams", "Australia", "No"},
				{"Hobnobs", "UK", "Yes"},
			}),
		)
		got := ansi.Strip(biscuits.View())
		golden.RequireEqual(t, []byte(got))
	})
	t.Run("With border", func(t *testing.T) {
		baseStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

		s := DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)

		biscuits := New(
			WithHeight(5),
			WithColumns([]Column{
				{Title: "Name", Width: 25},
				{Title: "Country of Origin", Width: 16},
				{Title: "Dunk-able", Width: 12},
			}),
			WithRows([]Row{
				{"Chocolate Digestives", "UK", "Yes"},
				{"Tim Tams", "Australia", "No"},
				{"Hobnobs", "UK", "Yes"},
			}),
			WithStyles(s),
		)
		got := ansi.Strip(baseStyle.Render(biscuits.View()))
		golden.RequireEqual(t, []byte(got))
	})
}

func TestCellPadding(t *testing.T) {
	tt := map[string]struct {
		tableWidth int
		styles     Styles
	}{
		"With padding": {
			tableWidth: 21,
			styles: Styles{
				Selected: lipgloss.NewStyle(),
				Header:   lipgloss.NewStyle().Padding(0, 1),
				Cell:     lipgloss.NewStyle().Padding(0, 1),
			},
		},
		"Without padding; exact width": {
			tableWidth: 15,
			styles: Styles{
				Selected: lipgloss.NewStyle(),
				Header:   lipgloss.NewStyle(),
				Cell:     lipgloss.NewStyle(),
			},
		},
		// TODO: Adjust the golden file once a desired output has been decided
		// https://github.com/charmbracelet/bubbles/issues/472
		"Without padding; too narrow": {
			tableWidth: 10,
			styles: Styles{
				Selected: lipgloss.NewStyle(),
				Header:   lipgloss.NewStyle(),
				Cell:     lipgloss.NewStyle(),
			},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			table := New(
				WithHeight(4),
				WithWidth(tc.tableWidth),
				WithColumns([]Column{
					{Title: "One", Width: 5},
					{Title: "Two", Width: 5},
					{Title: "Three", Width: 5},
				}),
				WithRows([]Row{
					{"r1c1-", "r1c2-", "r1c3-"},
					{"r2c1-", "r2c2-", "r2c3-"},
					{"r3c1-", "r3c2-", "r3c3-"},
					{"r4c1-", "r4c2-", "r4c3-"},
				}),
				WithStyles(tc.styles),
			)

			got := ansi.Strip(table.View())

			// TODO: Adjust the golden file once this bug has been resolved
			// https://github.com/charmbracelet/bubbles/issues/576
			golden.RequireEqual(t, []byte(got))
		})
	}
}

func TestTableCentering(t *testing.T) {
	t.Run("Centered in a box", func(t *testing.T) {
		boxStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			Align(lipgloss.Center)

		table := New(
			WithHeight(5),
			WithWidth(40),
			WithColumns([]Column{
				{Title: "One", Width: 5},
				{Title: "Two", Width: 5},
				{Title: "Three", Width: 5},
			}),
			WithRows([]Row{
				{"r1c1-", "r1c2-", "r1c3-"},
				{"r2c1-", "r2c2-", "r2c3-"},
				{"r3c1-", "r3c2-", "r3c3-"},
				{"r4c1-", "r4c2-", "r4c3-"},
			}),
		)

		tableView := ansi.Strip(table.View())
		got := boxStyle.Render(tableView)

		golden.RequireEqual(t, []byte(got))
	})
}

func TestCursorNavigation(t *testing.T) {
	tests := map[string]struct {
		rows   []Row
		action func(*Model)
		want   int
	}{
		"New": {
			rows: []Row{
				{"r1"},
				{"r2"},
				{"r3"},
			},
			action: func(_ *Model) {},
			want:   0,
		},
		"MoveDown": {
			rows: []Row{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.MoveDown(2)
			},
			want: 2,
		},
		"MoveUp": {
			rows: []Row{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.cursor = 3
				t.MoveUp(2)
			},
			want: 1,
		},
		"GotoBottom": {
			rows: []Row{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.GotoBottom()
			},
			want: 3,
		},
		"GotoTop": {
			rows: []Row{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.cursor = 3
				t.GotoTop()
			},
			want: 0,
		},
		"SetCursor": {
			rows: []Row{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.SetCursor(2)
			},
			want: 2,
		},
		"MoveDown with overflow": {
			rows: []Row{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.MoveDown(5)
			},
			want: 3,
		},
		"MoveUp with overflow": {
			rows: []Row{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.cursor = 3
				t.MoveUp(5)
			},
			want: 0,
		},
		"Blur does not stop movement": {
			rows: []Row{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.Blur()
				t.MoveDown(2)
			},
			want: 2,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			table := New(WithColumns(testCols), WithRows(tc.rows))
			tc.action(&table)

			if table.Cursor() != tc.want {
				t.Errorf("want %d, got %d", tc.want, table.Cursor())
			}
		})
	}
}

func TestModel_SetRows(t *testing.T) {
	table := New(WithColumns(testCols))

	if len(table.rows) != 0 {
		t.Fatalf("want 0, got %d", len(table.rows))
	}

	table.SetRows([]Row{{"r1"}, {"r2"}})

	if len(table.rows) != 2 {
		t.Fatalf("want 2, got %d", len(table.rows))
	}

	want := []Row{{"r1"}, {"r2"}}
	if !reflect.DeepEqual(table.rows, want) {
		t.Fatalf("\n\nwant %v\n\ngot %v", want, table.rows)
	}
}

func TestModel_SetColumns(t *testing.T) {
	table := New()

	if len(table.cols) != 0 {
		t.Fatalf("want 0, got %d", len(table.cols))
	}

	table.SetColumns([]Column{{Title: "Foo"}, {Title: "Bar"}})

	if len(table.cols) != 2 {
		t.Fatalf("want 2, got %d", len(table.cols))
	}

	want := []Column{{Title: "Foo"}, {Title: "Bar"}}
	if !reflect.DeepEqual(table.cols, want) {
		t.Fatalf("\n\nwant %v\n\ngot %v", want, table.cols)
	}
}
