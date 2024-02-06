package table

import (
	"fmt"
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

func TestSetCursorAlwaysVisibleHeight1(t *testing.T) {
	input := "foo1,bar1\nfoo2,bar2\nfoo3,bar3\nfoo4,bar4"
	table := New(WithColumns([]Column{{Title: "Foo", Width: 4}, {Title: "Bar", Width: 4}}))
	table.FromValues(input, ",")
	table.SetStyles(Styles{
		Header:   lipgloss.NewStyle(),
		Cell:     lipgloss.NewStyle(),
		Selected: lipgloss.NewStyle(),
	})

	expected := []string{
		"Foo Bar \nfoo1bar1",
		"Foo Bar \nfoo2bar2",
		"Foo Bar \nfoo3bar3",
	}

	table.SetHeight(1)

	for i := range expected {
		var (
			cursor   = i
			expected = expected[i]
		)
		t.Run(fmt.Sprintf("SetCursor(%d) Moving Down", i), func(t *testing.T) {
			table.SetCursor(cursor)
			t.Logf("m.cursor = %d", table.cursor)
			t.Logf("m.start  = %d", table.start)
			t.Logf("m.end    = %d", table.end)
			if table.View() != expected {
				t.Fatalf(`
expected: %q
     got: %q`, expected, table.View())
			}
		})
	}

	expected = []string{
		"Foo Bar \nfoo2bar2",
		"Foo Bar \nfoo1bar1",
	}

	for i := range expected {
		var (
			cursor   = table.cursor - 1
			expected = expected[i]
		)
		t.Run(fmt.Sprintf("SetCursor(%d) Moving Up", cursor), func(t *testing.T) {
			table.SetCursor(cursor)
			t.Logf("m.cursor = %d", table.cursor)
			t.Logf("m.start  = %d", table.start)
			t.Logf("m.end    = %d", table.end)
			if table.View() != expected {
				t.Fatalf(`
expected: %q
     got: %q`, expected, table.View())
			}
		})
	}

	jumps := []struct {
		i int
		s string
	}{
		{3, "Foo Bar \nfoo4bar4"},
		{0, "Foo Bar \nfoo1bar1"},
	}

	for i := range jumps {
		var (
			cursor   = jumps[i].i
			expected = jumps[i].s
		)
		t.Run(fmt.Sprintf("SetCursor(%d->%d)", table.cursor, cursor), func(t *testing.T) {
			table.SetCursor(cursor)
			t.Logf("m.cursor = %d", table.cursor)
			t.Logf("m.start  = %d", table.start)
			t.Logf("m.end    = %d", table.end)
			if table.View() != expected {
				t.Fatalf(`
expected: %q
     got: %q`, expected, table.View())
			}
		})
	}
}

func TestSetCursorAlwaysVisibleHeight2(t *testing.T) {
	input := "foo1,bar1\nfoo2,bar2\nfoo3,bar3\nfoo4,bar4"
	table := New(WithColumns([]Column{{Title: "Foo", Width: 4}, {Title: "Bar", Width: 4}}))
	table.FromValues(input, ",")
	table.SetStyles(Styles{
		Header:   lipgloss.NewStyle(),
		Cell:     lipgloss.NewStyle(),
		Selected: lipgloss.NewStyle(),
	})

	expected := []string{
		"Foo Bar \nfoo1bar1\nfoo2bar2",
		"Foo Bar \nfoo1bar1\nfoo2bar2",
		"Foo Bar \nfoo2bar2\nfoo3bar3",
		"Foo Bar \nfoo3bar3\nfoo4bar4",
	}

	table.SetHeight(2)

	for i := range expected {
		var (
			cursor   = i
			expected = expected[i]
		)
		t.Run(fmt.Sprintf("SetCursor(%d) Moving Down", i), func(t *testing.T) {
			table.SetCursor(cursor)
			t.Logf("m.cursor = %d", table.cursor)
			t.Logf("m.start  = %d", table.start)
			t.Logf("m.end    = %d", table.end)
			if table.View() != expected {
				t.Fatalf(`
expected: %q
     got: %q`, expected, table.View())
			}
		})
	}

	expected = []string{
		"Foo Bar \nfoo2bar2\nfoo3bar3",
		"Foo Bar \nfoo2bar2\nfoo3bar3",
		"Foo Bar \nfoo1bar1\nfoo2bar2",
	}

	for i := range expected {
		var (
			cursor   = table.cursor - 1
			expected = expected[i]
		)
		t.Run(fmt.Sprintf("SetCursor(%d) Moving Up", cursor), func(t *testing.T) {
			table.SetCursor(cursor)
			t.Logf("m.cursor = %d", table.cursor)
			t.Logf("m.start  = %d", table.start)
			t.Logf("m.end    = %d", table.end)
			if table.View() != expected {
				t.Fatalf(`
expected: %q
     got: %q`, expected, table.View())
			}
		})
	}

	jumps := []struct {
		i int
		s string
	}{
		{3, "Foo Bar \nfoo3bar3\nfoo4bar4"},
		{0, "Foo Bar \nfoo1bar1\nfoo2bar2"},
		{2, "Foo Bar \nfoo2bar2\nfoo3bar3"},
		{0, "Foo Bar \nfoo1bar1\nfoo2bar2"},
	}

	for i := range jumps {
		var (
			cursor   = jumps[i].i
			expected = jumps[i].s
		)
		t.Run(fmt.Sprintf("SetCursor(%d->%d)", table.cursor, cursor), func(t *testing.T) {
			table.SetCursor(cursor)
			t.Logf("m.cursor = %d", table.cursor)
			t.Logf("m.start  = %d", table.start)
			t.Logf("m.end    = %d", table.end)
			if table.View() != expected {
				t.Fatalf(`
expected: %q
     got: %q`, expected, table.View())
			}
		})
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
