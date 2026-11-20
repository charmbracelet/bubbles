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

func TestMoveOnEmptyTable_CursorStaysAtZero(t *testing.T) {
	cols := []Column{{Title: "Name", Width: 10}}

	tests := []struct {
		name string
		move func(*Model)
	}{
		{"MoveDown", func(m *Model) { m.MoveDown(1) }},
		{"MoveUp", func(m *Model) { m.MoveUp(1) }},
		{"GotoTop", func(m *Model) { m.GotoTop() }},
		{"GotoBottom", func(m *Model) { m.GotoBottom() }},
		{"SetCursor", func(m *Model) { m.SetCursor(3) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tbl := New(WithColumns(cols), WithHeight(5))

			tt.move(&tbl)
			if got := tbl.Cursor(); got != 0 {
				t.Fatalf("Cursor() on empty table = %d, want 0", got)
			}

			// Rows that arrive later must start with the first row selected.
			tbl.SetRows([]Row{{"a"}, {"b"}})
			if got := tbl.Cursor(); got != 0 {
				t.Fatalf("Cursor() after SetRows = %d, want 0", got)
			}
			if got := tbl.SelectedRow(); len(got) != 1 || got[0] != "a" {
				t.Fatalf("SelectedRow() = %v, want [a]", got)
			}
		})
	}
}
