package table

import (
	"strings"
	"testing"
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

func TestRenderCell(t *testing.T) {
	const expected = "rendered"

	styles := DefaultStyles()

	styles.RenderCell = func(model Model, value string, position CellPosition) string {
		switch {
		case position.RowID != 0:
			t.Fatalf("Invalid rowID: %d", position.RowID)
		case position.Column != 0:
			t.Fatalf("Invalid columnID: %d", position.Column)
		case !position.IsRowSelected:
			t.Fatalf("Invalid IsRowSelected: %t", position.IsRowSelected)
		}

		return expected
	}

	table := New(
		WithColumns([]Column{{Title: "Foo", Width: 100}}),
		WithRows([]Row{{"unexpected"}}),
		WithStyles(styles),
	)

	rendered := table.View()

	if !strings.Contains(rendered, expected) {
		t.Fatalf("Expected: %q in \n%s", expected, rendered)
	}
}

func TestCellDefault(t *testing.T) {
	const expected = "rendered"

	table := New(
		WithColumns([]Column{{Title: "Foo", Width: 100}}),
		WithRows([]Row{{expected}}),
	)

	rendered := table.View()

	if !strings.Contains(rendered, expected) {
		t.Fatalf("Expected: %q in \n%s", expected, rendered)
	}
}
