package table

import "testing"

func TestResizeColumnWidthsExpand(t *testing.T) {
	got := resizeColumnWidths([]int{25, 16, 12}, 80, 2)
	sum := 0
	for _, w := range got {
		sum += w + 2
	}
	if sum != 80 {
		t.Fatalf("total width %d, want 80, widths %v", sum, got)
	}
}

func TestResizeColumnWidthsShrink(t *testing.T) {
	got := resizeColumnWidths([]int{25, 16, 12}, 30, 2)
	sum := 0
	for _, w := range got {
		sum += w + 2
	}
	if sum != 30 {
		t.Fatalf("total width %d, want 30, widths %v", sum, got)
	}
}

func TestModelViewWidthMatchesTableWidth(t *testing.T) {
	m := New(
		WithWidth(80),
		WithColumns([]Column{
			{Title: "Name", Width: 25},
			{Title: "Country", Width: 16},
			{Title: "Dunk", Width: 12},
		}),
		WithRows([]Row{{"foo", "UK", "Yes"}}),
	)

	if got := m.naturalWidth(); got != 59 {
		t.Fatalf("naturalWidth %d want 59", got)
	}
	if w := m.Width(); w != 80 {
		t.Fatalf("Width() %d want 80", w)
	}
}
