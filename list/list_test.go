package list

import (
	"fmt"
	"io"
	"reflect"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

type item string

func (i item) FilterValue() string { return string(i) }

type itemDelegate struct{}

func (d itemDelegate) Height() int                          { return 1 }
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
	if !slices.Equal(list.VisibleItems(), expected) {
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

func TestRemoveItemRemovesValidItem(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		index int
		want  []Item
	}{
		{"first", 0, []Item{item("b"), item("c")}},
		{"middle", 1, []Item{item("a"), item("c")}},
		{"last", 2, []Item{item("a"), item("b")}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := New([]Item{item("a"), item("b"), item("c")}, itemDelegate{}, 40, 20)
			m.RemoveItem(tt.index)
			if !slices.Equal(m.Items(), tt.want) {
				t.Fatalf("items = %v, want %v", m.Items(), tt.want)
			}
		})
	}
}

func TestRemoveItemPastEndPreservesFilteredState(t *testing.T) {
	t.Parallel()
	m := New([]Item{item("foo"), item("bar"), item("baz")}, itemDelegate{}, 40, 20)
	m.SetFilterText("ba")
	m.Select(1)
	items := slices.Clone(m.Items())
	matches := slices.Clone(m.filteredItems)
	page, cursor := m.Paginator.Page, m.cursor
	for _, index := range []int{len(items), len(items) + 1} {
		m.RemoveItem(index)
		if !slices.Equal(m.Items(), items) || !reflect.DeepEqual(m.filteredItems, matches) {
			t.Fatalf("removing %d changed items or filter matches", index)
		}
		if m.filterState != FilterApplied || m.FilterInput.Value() != "ba" {
			t.Fatalf("removing %d changed the active filter", index)
		}
		if m.Paginator.Page != page || m.cursor != cursor {
			t.Fatalf("removing %d changed the selection", index)
		}
	}
}

func TestRemoveItemInvalidIndicesAreNoOps(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		items  []Item
		filter string
		index  int
	}{
		{"negative", []Item{item("a"), item("b")}, "", -1},
		{"negative while filtered", []Item{item("a"), item("b")}, "b", -1},
		{"negative on empty list", nil, "", -1},
		{"past end with no matches", []Item{item("a"), item("b")}, "missing", 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := New(tt.items, itemDelegate{}, 40, 20)
			if tt.filter != "" {
				m.SetFilterText(tt.filter)
			}
			items := slices.Clone(m.Items())
			matches := slices.Clone(m.filteredItems)
			state := m.filterState
			m.RemoveItem(tt.index)
			if !slices.Equal(m.Items(), items) || !reflect.DeepEqual(m.filteredItems, matches) {
				t.Fatal("invalid removal changed items or filter matches")
			}
			if m.filterState != state || m.FilterInput.Value() != tt.filter {
				t.Fatal("invalid removal changed the active filter")
			}
		})
	}
}
