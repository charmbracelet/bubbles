package table

import (
	"reflect"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
)

func TestFromValues(t *testing.T) {
	t.Run("Headers", func(t *testing.T) {
		input := "foo1,bar1\nfoo2,bar2\nfoo3,bar3"
		table := New()
		table.SetHeaders("Foo", "Bar")
		table.FromValues(input, ",")

		if len(table.rows) != 3 {
			t.Fatalf("expect table to have 3 rows but it has %d", len(table.rows))
		}

		expect := [][]string{
			{"foo1", "bar1"},
			{"foo2", "bar2"},
			{"foo3", "bar3"},
		}
		if !reflect.DeepEqual(table.rows, expect) {
			t.Fatal("table rows is not equals to the input")
		}
	})
	t.Run("WithColumns", func(t *testing.T) {
		input := "foo1,bar1\nfoo2,bar2\nfoo3,bar3"
		table := New(WithColumns([]Column{{Title: "Foo"}, {Title: "Bar"}}))
		table.FromValues(input, ",")

		if len(table.rows) != 3 {
			t.Fatalf("expect table to have 3 rows but it has %d", len(table.rows))
		}

		expect := [][]string{
			{"foo1", "bar1"},
			{"foo2", "bar2"},
			{"foo3", "bar3"},
		}
		if !reflect.DeepEqual(table.rows, expect) {
			t.Fatal("table rows is not equals to the input")
		}
	})
	t.Run("WithHeaders", func(t *testing.T) {
		input := "foo1,bar1\nfoo2,bar2\nfoo3,bar3"
		table := New(WithHeaders([]string{"Foo", "Bar"}))
		table.FromValues(input, ",")

		if len(table.rows) != 3 {
			t.Fatalf("expect table to have 3 rows but it has %d", len(table.rows))
		}

		expect := [][]string{
			{"foo1", "bar1"},
			{"foo2", "bar2"},
			{"foo3", "bar3"},
		}
		if !reflect.DeepEqual(table.rows, expect) {
			t.Fatal("table rows is not equals to the input")
		}
	})
}

func TestFromValuesWithTabSeparator(t *testing.T) {
	input := "foo1.\tbar1\nfoo,bar,baz\tbar,2"
	table := New(WithColumns([]Column{{Title: "Foo"}, {Title: "Bar"}}))
	table.FromValues(input, "\t")

	if len(table.rows) != 2 {
		t.Fatalf("expect table to have 2 rows but it has %d", len(table.rows))
	}

	expect := [][]string{
		{"foo1.", "bar1"},
		{"foo,bar,baz", "bar,2"},
	}
	if !reflect.DeepEqual(table.rows, expect) {
		t.Fatalf("table rows is not equal to the input. got: %#v, want %#v", table.rows, expect)
	}
}

func TestSetCursor(t *testing.T) {
	/*
	   the range for rows goes from 1 to len(rows) because in the bubble, the
	   first row is the headers, so we're adding 1 to the standard range.
	  **/
	tests := []struct {
		name     string
		cursor   int
		expected int
	}{
		{"cursor exceeds rows", 10, 2},
		{"cursor less than rows", -10, 0},
		{"cursor at zero", 0, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			table := New(
				WithRows([]Row{
					{"Foo"},
					{"Bar"},
					{"Baz"},
				}),
			)
			table.SetCursor(tc.cursor)
			if table.cursor != tc.expected {
				t.Fatalf("wrong cursor value, should be %d, got: %d\n%s", tc.expected, table.cursor, table.View())
			}
		})
		t.Run(tc.name+"/ table with headers", func(t *testing.T) {
			table := New(
				WithColumns([]Column{
					{Title: "Name", Width: 10},
				}),
				WithRows([]Row{
					{"Foo"},
					{"Bar"},
					{"Baz"},
				}),
			)
			table.SetCursor(tc.cursor)
			if table.cursor != tc.expected {
				t.Fatalf("wrong cursor value, should be %d, got: %d\n%s", tc.expected, table.cursor, table.View())
			}
		})
	}
}

func TestTableAlignment(t *testing.T) {
	t.Run("No border", func(t *testing.T) {
		s := DefaultStyles()
		s.BorderHeader = false
		biscuits := New(
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

		// unset borders; hidden border leaves space.
		biscuits.SetBorder(false)
		got := ansi.Strip(biscuits.View())
		golden.RequireEqual(t, []byte(got))
	})
	t.Run("With border", func(t *testing.T) {
		biscuits := New(
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
			WithHeight(10),
			WithStyles(DefaultStyles()),
		)
		got := ansi.Strip(biscuits.View())
		golden.RequireEqual(t, []byte(got))
	})
}

func TestSetStyleFunc(t *testing.T) {
	t.Run("single cell styling", func(t *testing.T) {
		s := DefaultStyles()
		s.BorderHeader = false
		biscuits := New(
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
		biscuits.SetStyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				// TODO this should be exported to be usable outside of the lib.
				return s.Header
			}
			if row == 1 && col == 1 {
				return s.Cell.Bold(true)
			}
			return s.Cell
		})
		golden.RequireEqual(t, []byte(biscuits.View()))
	})
}

func TestWithStyleFunc(t *testing.T) {
	t.Run("single cell styling", func(t *testing.T) {
		// #502
		s := DefaultStyles()
		s.BorderHeader = false
		biscuits := New(
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
			WithStyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return s.Header
				}
				// TODO we should probably make it possible to retrieve Style
				// from the model in case it has been modified from the
				// defaults.
				if row == 1 && col == 1 {
					return s.Cell.Bold(true)
				}
				return s.Cell
			}))
		golden.RequireEqual(t, []byte(biscuits.View()))
	})
	t.Run("cell styling by content", func(t *testing.T) {
		// #502
		rows := []Row{
			{"Chocolate Digestives", "UK", "Yes"},
			{"Tim Tams", "Australia", "No"},
			{"Hobnobs", "UK", "Yes"},
		}
		s := DefaultStyles()
		s.BorderHeader = false
		biscuits := New(
			WithColumns([]Column{
				{Title: "Name", Width: 25},
				{Title: "Country of Origin", Width: 16},
				{Title: "Dunk-able", Width: 12},
			}),
			WithRows(rows),
			WithStyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return s.Header
				}
				// TODO we should probably make it possible to retrieve Style
				// from the model in case it has been modified from the
				// defaults.

				// you need to pre-define the rows for this to be accessible in
				// WithStyleFunc
				if rows[row][col] == "Yes" {
					return s.Cell.Bold(true)
				}
				return s.Cell
			}))
		golden.RequireEqual(t, []byte(biscuits.View()))
	})
	t.Run("change column text alignment", func(t *testing.T) {
		// #399
		s := DefaultStyles()
		s.BorderHeader = false
		biscuits := New(
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
			WithStyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return s.Header
				}
				if col == 1 {
					return s.Cell.Align(lipgloss.Right)
				}
				return s.Cell
			}))
		golden.RequireEqual(t, []byte(biscuits.View()))
	})
}
