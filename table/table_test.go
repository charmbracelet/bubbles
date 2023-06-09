package table

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
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

var (
	modelCols = []Column{
		{Title: "col1", Width: 10},
		{Title: "col2", Width: 10},
		{Title: "col3", Width: 10},
	}
)

func TestRenderRow(t *testing.T) {
	type testcase struct {
		m           *Model
		rowID       int
		expectedRow string
		name        string
	}
	testcases := []testcase{
		{
			m: &Model{
				rows: []Row{
					[]string{"valuea1", "valuea2", "valuea3"},
				},
				cols:   modelCols,
				cursor: -1,
				styles: Styles{
					Cell: lipgloss.NewStyle(),
				},
			},
			rowID:       0,
			expectedRow: "valuea1   valuea2   valuea3   ",
			name:        "simple row",
		},
		{
			m: &Model{
				rows: []Row{
					[]string{"valuea11111", "valuea22222", "valuea33333"},
				},
				cols:   modelCols,
				cursor: -1,
				styles: Styles{
					Cell: lipgloss.NewStyle(),
				},
			},
			rowID:       0,
			expectedRow: "valuea111…valuea222…valuea333…",
			name:        "simple row with truncations",
		},
		{
			m: &Model{
				rows: []Row{
					[]string{"valuea1111", "valuea2222", "valuea3333"},
				},
				cols:   modelCols,
				cursor: -1,
				styles: Styles{
					Cell: lipgloss.NewStyle(),
				},
			},
			rowID:       0,
			expectedRow: "valuea1111valuea2222valuea3333",
			name:        "simple row avoiding truncations",
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			row := testcase.m.renderRow(testcase.rowID)
			if row != testcase.expectedRow {
				t.Fatalf("expected row contents |%s|, got |%s|", testcase.expectedRow, row)
			}
		})
	}
}
