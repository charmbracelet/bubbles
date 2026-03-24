package table

import (
	"reflect"
	"strings"
	"testing"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
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
				cursor: 0,
				viewport: viewport.New(
					viewport.WithWidth(0),
					viewport.WithHeight(20),
				),
				KeyMap: DefaultKeyMap(),
				Help:   help.New(),
				styles: DefaultStyles(),
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
				cursor: 0,
				viewport: viewport.New(
					viewport.WithWidth(0),
					viewport.WithHeight(20),
				),
				KeyMap: DefaultKeyMap(),
				Help:   help.New(),
				styles: DefaultStyles(),

				// Modified fields
				cols: []Column{
					{Title: "Foo", Width: 1},
					{Title: "Bar", Width: 2},
				},
			},
		},
		"WithColumns; WithRows": {
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
				cursor: 0,
				viewport: viewport.New(
					viewport.WithWidth(0),
					viewport.WithHeight(20),
				),
				KeyMap: DefaultKeyMap(),
				Help:   help.New(),
				styles: DefaultStyles(),

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
				viewport: viewport.New(
					viewport.WithWidth(0),
					viewport.WithHeight(9),
				),
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
				viewport: viewport.New(
					viewport.WithWidth(10),
					viewport.WithHeight(20),
				),
			},
		},
		"WithFocused": {
			opts: []Option{
				WithFocused(true),
			},
			want: Model{
				// Default fields
				cursor: 0,
				viewport: viewport.New(
					viewport.WithWidth(0),
					viewport.WithHeight(20),
				),
				KeyMap: DefaultKeyMap(),
				Help:   help.New(),
				styles: DefaultStyles(),

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
				cursor: 0,
				viewport: viewport.New(
					viewport.WithWidth(0),
					viewport.WithHeight(20),
				),
				KeyMap: DefaultKeyMap(),
				Help:   help.New(),

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
				cursor: 0,
				viewport: viewport.New(
					viewport.WithWidth(0),
					viewport.WithHeight(20),
				),
				Help:   help.New(),
				styles: DefaultStyles(),

				// Modified fields
				KeyMap: KeyMap{},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.want.UpdateViewport()

			got := New(tc.opts...)

			// NOTE(@andreynering): Funcs have different references, so we need
			// to clear them out to compare the structs.
			tc.want.viewport.LeftGutterFunc = nil
			got.viewport.LeftGutterFunc = nil

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

func TestModel_RenderRow_AnsiWidth(t *testing.T) {
	value := "\x1b[31mABCDEFGH\x1b[0m"
	table := &Model{
		rows:   []Row{{value}},
		cols:   []Column{{Title: "col1", Width: 8}},
		styles: Styles{Cell: lipgloss.NewStyle()},
	}

	got := ansi.Strip(table.renderRow(0))
	want := "ABCDEFGH"
	if got != want {
		t.Fatalf("\n\nWant: \n%s\n\nGot:  \n%s\n", want, got)
	}
}

func TestTableAlignment(t *testing.T) {
	t.Run("No border", func(t *testing.T) {
		biscuits := New(
			WithWidth(59),
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
		got := ansiStrip(biscuits.View())
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
			WithWidth(59),
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
		got := ansiStrip(baseStyle.Render(biscuits.View()))
		golden.RequireEqual(t, []byte(got))
	})
}

func ansiStrip(s string) string {
	// Replace all \r\n with \n
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return ansi.Strip(s)
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

func TestModel_View(t *testing.T) {
	tests := map[string]struct {
		modelFunc func() Model
		skip      bool
	}{
		"Empty": {
			modelFunc: func() Model {
				return New(
					WithWidth(60),
					WithHeight(21),
				)
			},
		},
		"Single row and column": {
			modelFunc: func() Model {
				return New(
					WithWidth(27),
					WithHeight(21),
					WithColumns([]Column{
						{Title: "Name", Width: 25},
					}),
					WithRows([]Row{
						{"Chocolate Digestives"},
					}),
				)
			},
		},
		"Multiple rows and columns": {
			modelFunc: func() Model {
				return New(
					WithWidth(59),
					WithHeight(21),
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
			},
		},
		// TODO(fix): since the table height is tied to the viewport height, adding vertical padding to the headers' height directly increases the table height.
		"Extra padding": {
			modelFunc: func() Model {
				s := DefaultStyles()
				s.Header = lipgloss.NewStyle().Padding(2, 2)
				s.Cell = lipgloss.NewStyle().Padding(2, 2)

				return New(
					WithWidth(60),
					WithHeight(10),
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
			},
		},
		"No padding": {
			modelFunc: func() Model {
				s := DefaultStyles()
				s.Header = lipgloss.NewStyle()
				s.Cell = lipgloss.NewStyle()

				return New(
					WithWidth(53),
					WithHeight(10),
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
			},
		},
		// TODO(?): the total height is modified with bordered headers, however not with bordered cells. Is this expected/desired?
		"Bordered headers": {
			modelFunc: func() Model {
				return New(
					WithWidth(59),
					WithHeight(23),
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
					WithStyles(Styles{
						Header: lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()),
					}),
				)
			},
		},
		// TODO(fix): Headers are not horizontally aligned with cells due to the border adding width to the cells.
		"Bordered cells": {
			modelFunc: func() Model {
				return New(
					WithWidth(59),
					WithHeight(21),
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
					WithStyles(Styles{
						Cell: lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()),
					}),
				)
			},
		},
		"Height greater than rows": {
			modelFunc: func() Model {
				return New(
					WithWidth(59),
					WithHeight(6),
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
			},
		},
		"Height less than rows": {
			modelFunc: func() Model {
				return New(
					WithWidth(59),
					WithHeight(2),
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
			},
		},
		// TODO(fix): spaces are added to the right of the viewport to fill the width, but the headers end as though they are not aware of the width.
		"Width greater than columns": {
			modelFunc: func() Model {
				return New(
					WithWidth(80),
					WithHeight(21),
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
			},
		},
		// TODO(fix): Setting the table width does not affect the total headers' width. Cells are wrapped.
		// 	Headers are not affected. Truncation/resizing should match lipgloss.table functionality.
		"Width less than columns": {
			modelFunc: func() Model {
				return New(
					WithWidth(30),
					WithHeight(15),
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
			},
			skip: true,
		},
		"Modified viewport height": {
			modelFunc: func() Model {
				m := New(
					WithWidth(59),
					WithHeight(15),
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

				m.viewport.SetHeight(2)

				return m
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.skip {
				t.Skip()
			}

			table := tc.modelFunc()

			got := ansi.Strip(table.View())

			golden.RequireEqual(t, []byte(got))
		})
	}
}

// TODO: Fix table to make this test will pass.
func TestModel_View_CenteredInABox(t *testing.T) {
	t.Skip()

	boxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		Align(lipgloss.Center)

	table := New(
		WithHeight(6),
		WithWidth(80),
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

	tableView := ansi.Strip(table.View())
	got := boxStyle.Render(tableView)

	golden.RequireEqual(t, []byte(got))
}
