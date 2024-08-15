package table

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
)

func TestFromValues(t *testing.T) {
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
	if !deepEqual(table.rows, expect) {
		t.Fatal("table rows is not equals to the input")
	}
}

func TestFromValuesWithTabSeparator(t *testing.T) {
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
	if !deepEqual(table.rows, expect) {
		t.Fatal("table rows is not equals to the input")
	}
}

func deepEqual(a, b []Row) bool {
	if len(a) != len(b) {
		return false
	}
	for i, r := range a {
		for j, f := range r {
			if f != b[i][j] {
				return false
			}
		}
	}
	return true
}

var cols = []Column{
	{Title: "col1", Width: 10},
	{Title: "col2", Width: 10},
	{Title: "col3", Width: 10},
}

func TestRenderRow(t *testing.T) {
	tests := []struct {
		name     string
		table    *Model
		expected string
	}{
		{
			name: "simple row",
			table: &Model{
				rows:   []Row{{"Foooooo", "Baaaaar", "Baaaaaz"}},
				cols:   cols,
				styles: Styles{Cell: lipgloss.NewStyle()},
			},
			expected: "Foooooo   Baaaaar   Baaaaaz   ",
		},
		{
			name: "simple row with truncations",
			table: &Model{
				rows:   []Row{{"Foooooooooo", "Baaaaaaaaar", "Quuuuuuuuux"}},
				cols:   cols,
				styles: Styles{Cell: lipgloss.NewStyle()},
			},
			expected: "Foooooooo…Baaaaaaaa…Quuuuuuuu…",
		},
		{
			name: "simple row avoiding truncations",
			table: &Model{
				rows:   []Row{{"Fooooooooo", "Baaaaaaaar", "Quuuuuuuux"}},
				cols:   cols,
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

func TestRenderRowStyleFunc(t *testing.T) {
	tests := []struct {
		name     string
		table    *Model
		expected string
	}{
		{
			name: "simple row",
			table: &Model{
				rows: []Row{{"Foooooo", "Baaaaar", "Baaaaaz"}},
				cols: cols,
				styleFunc: func(row, col int, value string) lipgloss.Style {
					if strings.HasSuffix(value, "z") {
						return lipgloss.NewStyle().Transform(strings.ToLower)
					}
					return lipgloss.NewStyle().Transform(strings.ToUpper)
				},
			},
			expected: "FOOOOOO   BAAAAAR   baaaaaz   ",
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

func TestStyleFunc(t *testing.T) {
	t.Run("StyleFunc returns an empty style", func(t *testing.T) {
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
			WithStyleFunc(func(row, col int, value string) lipgloss.Style {
				return lipgloss.NewStyle()
			}),
		)

		got := ansi.Strip(biscuits.View())
		golden.RequireEqual(t, []byte(got))
	})

	t.Run("Use WithStyles and StyleFunc", func(t *testing.T) {
		columns := []Column{
			{Title: "Rank", Width: 4},
			{Title: "City", Width: 10},
			{Title: "Country", Width: 10},
			{Title: "Population", Width: 10},
		}

		rows := []Row{
			{"1", "Tokyo", "Japan", "37,274,000"},
			{"2", "Delhi", "India", "32,065,760"},
			{"3", "Shanghai", "China", "28,516,904"},
			{"4", "Dhaka", "Bangladesh", "22,478,116"},
			{"5", "São Paulo", "Brazil", "22,429,800"},
			{"6", "Mexico City", "Mexico", "22,085,140"},
			{"7", "Cairo", "Egypt", "21,750,020"},
			{"8", "Beijing", "China", "21,333,332"},
		}

		s := DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false).
			PaddingRight(1)
		s.Selected = s.Selected.
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false)
		s.Cell = lipgloss.NewStyle().PaddingRight(1)

		table := New(
			WithColumns(columns),
			WithRows(rows),
			WithFocused(true),
			WithHeight(7),
			WithStyles(s),
			WithStyleFunc(func(row, col int, value string) lipgloss.Style {
				if row == 5 {
					if col == 1 {
						return s.Cell.Background(lipgloss.Color("#006341"))
					} else if col == 2 {
						return s.Cell.Background(lipgloss.Color("#FFFFFF"))
					} else if col == 3 {
						return s.Cell.Background(lipgloss.Color("#C8102E"))
					} else {
						return s.Cell
					}
				}

				return s.Cell
			}),
		)

		got := ansi.Strip(table.View())
		// got := ansi.Strip(table.View()) if you want to remove colors.
		golden.RequireEqual(t, []byte(got))
	})

	t.Run("Use WithStyles after StyleFunc", func(t *testing.T) {
		columns := []Column{
			{Title: "Rank", Width: 4},
			{Title: "City", Width: 10},
			{Title: "Country", Width: 10},
			{Title: "Population", Width: 10},
		}

		rows := []Row{
			{"1", "Tokyo", "Japan", "37,274,000"},
			{"2", "Delhi", "India", "32,065,760"},
			{"3", "Shanghai", "China", "28,516,904"},
			{"4", "Dhaka", "Bangladesh", "22,478,116"},
			{"5", "São Paulo", "Brazil", "22,429,800"},
			{"6", "Mexico City", "Mexico", "22,085,140"},
			{"7", "Cairo", "Egypt", "21,750,020"},
			{"8", "Beijing", "China", "21,333,332"},
		}

		s := DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false).
			PaddingRight(1)
		s.Selected = s.Selected.
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false)
		s.Cell = lipgloss.NewStyle().PaddingRight(1)

		table := New(
			WithColumns(columns),
			WithRows(rows),
			WithFocused(true),
			WithHeight(7),
			WithStyleFunc(func(row, col int, value string) lipgloss.Style {
				if row == 5 {
					if col == 1 {
						return s.Cell.Background(lipgloss.Color("#006341"))
					} else if col == 2 {
						return s.Cell.Background(lipgloss.Color("#FFFFFF"))
					} else if col == 3 {
						return s.Cell.Background(lipgloss.Color("#C8102E"))
					} else {
						return s.Cell
					}
				}

				return s.Cell
			}),
			WithStyles(s),
		)

		// let's remove colors from the tests since the ansi sequences can be tricky to debug
		got := ansi.Strip(table.View())
		golden.RequireEqual(t, []byte(got))
	})
}
