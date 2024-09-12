package table

import (
	"reflect"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
)

func TestFromValues(t *testing.T) {
	t.Run("Headers", func(t *testing.T) {
		input := "foo1,bar1\nfoo2,bar2\nfoo3,bar3"
		table := New()
		table.Headers("Foo", "Bar")
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

	expect := []Row{
		{"foo1.", "bar1"},
		{"foo,bar,baz", "bar,2"},
	}
	if !reflect.DeepEqual(table.rows, expect) {
		t.Fatal("table rows is not equals to the input")
	}
}

func TestTableAlignment(t *testing.T) {
	t.Run("No border", func(t *testing.T) {
		s := DefaultStyles()
		s.Border = lipgloss.HiddenBorder()
		s.BorderHeader = false
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
		got := ansi.Strip(biscuits.View())
		golden.RequireEqual(t, []byte(got))
	})
	t.Run("With border", func(t *testing.T) {
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
			WithStyles(DefaultStyles()),
		)
		got := ansi.Strip(biscuits.View())
		golden.RequireEqual(t, []byte(got))
	})
}
