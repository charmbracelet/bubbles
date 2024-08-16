package list

import (
	"fmt"
	"io"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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

func TestIndex(t *testing.T) {
	list := New([]Item{item("bar"), item("baz"), item("foo"), item("bar"), item("panda"), item("bay")}, itemDelegate{}, 10, 15)
	list.cursor = 0

	// Hardcode filtering
	list.filterState = Filtering
	list.FilterInput.SetValue("f")
	msg := filterItems(list)
	val, ok := msg().(FilterMatchesMsg)
	if ok {
		list.filteredItems = filteredItems(val)
	}

	// index updates once the cursor changes
	t.Log("cursor", list.Cursor())
	t.Log("i:", list.Index())
	t.Log("gi:", list.GlobalIndex())

	t.Log("\n", list.View())

	// take 2
	list.ResetFilter()

	t.Log("\n", list.View())
	list.filterState = Filtering
	list.FilterInput.SetValue("ba")
	msg = filterItems(list)
	val, ok = msg().(FilterMatchesMsg)
	if ok {
		list.filteredItems = filteredItems(val)
	}
	// cursor at 0
	t.Log("cursor", list.Cursor())
	t.Log("i:", list.Index())
	t.Log("gi:", list.GlobalIndex())

	list.CursorDown()
	t.Log("moved cursor")

	// cursor at 1
	t.Log("cursor", list.Cursor())
	t.Log("i:", list.Index())
	t.Log("gi:", list.GlobalIndex())

	list.CursorDown()
	t.Log("moved cursor")

	// cursor at 2, gi at 3
	t.Log("cursor", list.Cursor())
	t.Log("i:", list.Index())
	t.Log("gi:", list.GlobalIndex())

	list.CursorDown()
	t.Log("moved cursor")

	// cursor at 2, gi at 5
	t.Log("cursor", list.Cursor())
	t.Log("i:", list.Index())
	t.Log("gi:", list.GlobalIndex())

	t.Log("\n", list.View())
	realIndex := list.GlobalIndex()

	if realIndex != 5 {
		t.Fatalf("expected the item's index to be 5, got %d", realIndex)
	}
}
