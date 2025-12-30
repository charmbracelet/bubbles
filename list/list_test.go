package list

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type item string

func (i item) FilterValue() string { return string(i) }

type itemDelegate struct{}

func (d itemDelegate) Height() int                          { return 1 }
func (d itemDelegate) Width() int                           { return 8 }
func (d itemDelegate) Spacing() int                         { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m Model, index int, listItem Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)
	fmt.Fprint(w, m.Styles.TitleBar.Render(str))
}

func TestStatusBarItemName(t *testing.T) {
	list := New([]Item{item("foo"), item("bar")}, itemDelegate{}, 10, 10)
	expected := "2 items"
	if !strings.Contains(list.statusView(), expected) {
		t.Fatalf("Error: expected view to contain %s", expected)
	}

	list.SetItems([]Item{item("foo")})
	expected = "1 item"
	if !strings.Contains(list.statusView(), expected) {
		t.Fatalf("Error: expected view to contain %s", expected)
	}
}

func TestStatusBarWithoutItems(t *testing.T) {
	list := New([]Item{}, itemDelegate{}, 10, 10)

	expected := "No items"
	if !strings.Contains(list.statusView(), expected) {
		t.Fatalf("Error: expected view to contain %s", expected)
	}
}

func TestCustomStatusBarItemName(t *testing.T) {
	list := New([]Item{item("foo"), item("bar")}, itemDelegate{}, 10, 10)
	list.SetStatusBarItemName("connection", "connections")

	expected := "2 connections"
	if !strings.Contains(list.statusView(), expected) {
		t.Fatalf("Error: expected view to contain %s", expected)
	}

	list.SetItems([]Item{item("foo")})
	expected = "1 connection"
	if !strings.Contains(list.statusView(), expected) {
		t.Fatalf("Error: expected view to contain %s", expected)
	}

	list.SetItems([]Item{})
	expected = "No connections"
	if !strings.Contains(list.statusView(), expected) {
		t.Fatalf("Error: expected view to contain %s", expected)
	}
}

func TestSetFilterText(t *testing.T) {
	tc := []Item{item("foo"), item("bar"), item("baz")}

	list := New(tc, itemDelegate{}, 10, 10)
	list.SetFilterText("ba")

	list.SetFilterState(Unfiltered)
	expected := tc
	// TODO: replace with slices.Equal() when project move to go1.18 or later
	if !reflect.DeepEqual(list.VisibleItems(), expected) {
		t.Fatalf("Error: expected view to contain only %s", expected)
	}

	list.SetFilterState(Filtering)
	expected = []Item{item("bar"), item("baz")}
	if !reflect.DeepEqual(list.VisibleItems(), expected) {
		t.Fatalf("Error: expected view to contain only %s", expected)
	}

	list.SetFilterState(FilterApplied)
	if !reflect.DeepEqual(list.VisibleItems(), expected) {
		t.Fatalf("Error: expected view to contain only %s", expected)
	}
}

func TestSetFilterState(t *testing.T) {
	tc := []Item{item("foo"), item("bar"), item("baz")}

	list := New(tc, itemDelegate{}, 10, 10)
	list.SetFilterText("ba")

	list.SetFilterState(Unfiltered)
	expected, notExpected := "up", "clear filter"

	lines := strings.Split(list.View(), "\n")
	footer := lines[len(lines)-1]

	if !strings.Contains(footer, expected) || strings.Contains(footer, notExpected) {
		t.Fatalf("Error: expected view to contain '%s' not '%s'", expected, notExpected)
	}

	list.SetFilterState(Filtering)
	expected, notExpected = "filter", "more"

	lines = strings.Split(list.View(), "\n")
	footer = lines[len(lines)-1]

	if !strings.Contains(footer, expected) || strings.Contains(footer, notExpected) {
		t.Fatalf("Error: expected view to contain '%s' not '%s'", expected, notExpected)
	}

	list.SetFilterState(FilterApplied)
	expected = "clear"

	lines = strings.Split(list.View(), "\n")
	footer = lines[len(lines)-1]

	if !strings.Contains(footer, expected) {
		t.Fatalf("Error: expected view to contain '%s'", expected)
	}
}

func TestHorizontalEnabled(t *testing.T) {
	items := make([]Item, 20)
	for i := range items {
		items[i] = item(fmt.Sprintf("item %d", i+1))
	}

	// Create a list with enough height for one row, but enough width for multiple columns
	// Delegate height is 1, spacing 0. So 1 item height + 0 spacing = 1 row height per item.
	// A height of 10 means 10 rows fit vertically.
	// A width of 20, with itemDelegate width 8, spacing 0.
	// 8 width + (0*2) spacing = 8 width per item.
	// 20 width / 8 per item = 2.5 columns, so 2 columns fit.
	list := New(items, itemDelegate{}, 20, 10)

	// Simplify testing
	list.SetShowPagination(false)
	list.SetShowHelp(false)
	list.SetShowStatusBar(false)
	list.SetShowFilter(false)
	list.SetShowTitle(false)

	t.Run("Vertical Layout", func(t *testing.T) {
		list.SetHorizontalEnabled(false)

		// Expect 10 items per page (height 10 / itemHeight 1)
		if list.Paginator.PerPage != 10 {
			t.Errorf("Expected 10 items per page in vertical layout, got %d", list.Paginator.PerPage)
		}
		if list.Paginator.TotalPages != 2 { // 20 items / 10 per page = 2 pages
			t.Errorf("Expected 2 total pages in vertical layout, got %d", list.Paginator.TotalPages)
		}

		// CursorDown should move to the next row (next item)
		list.cursor = 0
		list.CursorDown()
		if list.cursor != 1 {
			t.Errorf("Expected cursor to be 1 after CursorDown in vertical, got %d", list.cursor)
		}

		// CursorUp should move to the previous row (previous item)
		list.CursorUp()
		if list.cursor != 0 {
			t.Errorf("Expected cursor to be 0 after CursorUp in vertical, got %d", list.cursor)
		}

		// CursorLeft/Right should not move the cursor
		list.CursorLeft()
		if list.cursor != 0 {
			t.Errorf("Expected cursor to be 0 after CursorLeft in vertical, got %d", list.cursor)
		}
		list.CursorRight()
		if list.cursor != 0 {
			t.Errorf("Expected cursor to be 0 after CursorRight in vertical, got %d", list.cursor)
		}
	})

	t.Run("Horizontal Layout", func(t *testing.T) {
		list.SetHorizontalEnabled(true)

		// Expected 2 columns per page (width 20 / itemWidth 8)
		// Expected 10 rows per page (height 10 / itemHeight 1)
		// PerPage = 2 columns * 10 rows = 20 items per page
		if list.Paginator.PerPage != 20 {
			t.Errorf("Expected 20 items per page in horizontal layout, got %d", list.Paginator.PerPage)
		}
		if list.Paginator.TotalPages != 1 { // 20 items / 20 per page = 1 page
			t.Errorf("Expected 1 total page in horizontal layout, got %d", list.Paginator.TotalPages)
		}

		// CursorDown should move down by a full column width (2 items)
		list.cursor = 0
		list.CursorDown()
		if list.cursor != 2 { // Moved to the item in the next "row" within the horizontal flow
			t.Errorf("Expected cursor to be 2 after CursorDown in horizontal, got %d", list.cursor)
		}

		// CursorUp should move up by a full column width (2 items)
		list.CursorUp()
		if list.cursor != 0 {
			t.Errorf("Expected cursor to be 0 after CursorUp in horizontal, got %d", list.cursor)
		}

		// CursorRight should move to the next item
		list.CursorRight()
		if list.cursor != 1 {
			t.Errorf("Expected cursor to be 1 after CursorRight in horizontal, got %d", list.cursor)
		}

		// CursorLeft should move to the previous item
		list.CursorLeft()
		if list.cursor != 0 {
			t.Errorf("Expected cursor to be 0 after CursorLeft in horizontal, got %d", list.cursor)
		}

		// Test moving to next page with CursorRight if infinite scrolling is enabled
		list.InfiniteScrolling = true
		list.SetItems(make([]Item, 30)) // More items to create multiple pages horizontally
		list.SetSize(20, 10)            // Recalculate pagination

		// Should still be 2 columns * 10 rows = 20 items per page
		if list.Paginator.PerPage != 20 {
			t.Errorf("Expected 20 items per page for 30 items, got %d", list.Paginator.PerPage)
		}
		if list.Paginator.TotalPages != 2 { // 30 items / 20 per page = 2 pages
			t.Errorf("Expected 2 total pages for 30 items, got %d", list.Paginator.TotalPages)
		}

		list.Paginator.Page = 0
		list.cursor = 19 // Last item on the first page

		list.CursorRight() // Move past the end of the page
		if list.Paginator.Page != 1 || list.cursor != 0 {
			t.Errorf("Expected to move to page 1, cursor 0, but got page %d, cursor %d", list.Paginator.Page, list.cursor)
		}

		list.cursor = 9 // On the second row of the first page
		list.Paginator.Page = 0

		list.CursorDown()
		if list.cursor != 11 { // 9 + 2 (columns)
			t.Errorf("Expected cursor to be 11, got %d", list.cursor)
		}

		// Test moving to previous page with CursorLeft if infinite scrolling is enabled
		list.Paginator.Page = 1
		list.cursor = 0 // First item on the second page

		list.CursorLeft() // Move before the start of the page
		if list.Paginator.Page != 0 || list.cursor != 19 {
			t.Errorf("Expected to move to page 0, cursor 19, but got page %d, cursor %d", list.Paginator.Page, list.cursor)
		}

		list.InfiniteScrolling = false
	})
}
