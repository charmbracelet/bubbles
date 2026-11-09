package table

import (
	"testing"
)

func TestSetRows_CursorClampsToZeroOnEmpty(t *testing.T) {
	cols := []Column{{Title: "Name", Width: 10}}
	rows := []Row{{"a"}, {"b"}, {"c"}}
	tbl := New(WithColumns(cols), WithRows(rows), WithHeight(5))

	// Move to last row
	tbl.GotoBottom()
	if tbl.Cursor() != 2 {
		t.Fatalf("expected cursor at 2, got %d", tbl.Cursor())
	}

	// Replace with empty rows
	tbl.SetRows([]Row{})

	got := tbl.Cursor()
	if got < 0 {
		t.Fatalf("SetRows(empty): Cursor() = %d, want >= 0 (got negative cursor on empty table)", got)
	}
}

func TestSetRows_CursorClampsOnShrink(t *testing.T) {
	cols := []Column{{Title: "Name", Width: 10}}
	rows := []Row{{"a"}, {"b"}, {"c"}}
	tbl := New(WithColumns(cols), WithRows(rows), WithHeight(5))

	// Move to last row (index 2)
	tbl.GotoBottom()

	// Shrink to 1 row
	tbl.SetRows([]Row{{"x"}})

	got := tbl.Cursor()
	if got != 0 {
		t.Fatalf("SetRows(fewer rows): Cursor() = %d, want 0", got)
	}
}
