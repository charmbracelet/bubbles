package list

import (
	"fmt"
	"io"
	"reflect"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
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

func TestDefaultKeyMapQuit(t *testing.T) {
	quit := DefaultKeyMap().Quit
	if !reflect.DeepEqual(quit.Keys(), []string{"q", "esc"}) {
		t.Fatalf("expected quit keys q and esc, got %v", quit.Keys())
	}
	if help := quit.Help(); help.Key != "q" || help.Desc != "quit" {
		t.Fatalf("expected quit help, got %q, %q", help.Key, help.Desc)
	}
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

func TestRemoveItemWithFilter(t *testing.T) {
	items := []Item{item("foo"), item("bar"), item("baz")}
	l := New(items, itemDelegate{}, 10, 10)
	l.SetFilterText("ba")
	l.SetFilterState(FilterApplied)

	// filtered list contains ["bar", "baz"] at global indices 1 and 2.
	// Remove "foo" (global index 0) — filtered list should be unchanged.
	l.RemoveItem(0)
	visible := l.VisibleItems()
	if !reflect.DeepEqual(visible, []Item{item("bar"), item("baz")}) {
		t.Fatalf("after removing unfiltered item: got %v, want [bar baz]", visible)
	}

	// The global list is now ["bar", "baz"]. Remove "bar" (global index 0).
	// Filtered list should shrink to ["baz"].
	l.RemoveItem(0)
	visible = l.VisibleItems()
	if !reflect.DeepEqual(visible, []Item{item("baz")}) {
		t.Fatalf("after removing filtered item: got %v, want [baz]", visible)
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

func TestFilterMatchedIndexesAreRunePositions(t *testing.T) {
	// sahilm/fuzzy reports matched indexes as byte offsets, but MatchesForItem
	// documents them as rune positions and lipgloss.StyleRunes consumes them as
	// such. DefaultFilter and UnsortedFilter must convert byte offsets to rune
	// offsets so multi-byte titles are highlighted correctly.
	cases := []struct {
		name   string
		term   string
		target string
		want   []int
	}{
		{"ascii", "b", "abc", []int{1}},
		{"multibyte prefix", "b", "日本b", []int{2}},
		{"accented prefix", "x", "café x", []int{5}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, f := range []struct {
				name   string
				filter FilterFunc
			}{
				{"DefaultFilter", DefaultFilter},
				{"UnsortedFilter", UnsortedFilter},
			} {
				ranks := f.filter(tc.term, []string{tc.target})
				if len(ranks) != 1 {
					t.Fatalf("%s: expected 1 rank, got %d", f.name, len(ranks))
				}
				if !reflect.DeepEqual(ranks[0].MatchedIndexes, tc.want) {
					t.Errorf("%s: MatchedIndexes = %v, want %v", f.name, ranks[0].MatchedIndexes, tc.want)
				}
			}
		})
	}
}

func TestTitleViewWithoutStatusMessageDoesNotTruncate(t *testing.T) {
	list := New(nil, itemDelegate{}, 20, 10)
	list.Title = "List Title ABCDEFG"
	list.Styles.Title = list.Styles.Title.Padding(0)
	list.Styles.TitleBar = list.Styles.TitleBar.Padding(0)

	if got := ansi.Strip(list.titleView()); got != list.Title {
		t.Fatalf("expected title %q, got %q", list.Title, got)
	}
}

func TestTitleFitsListWidth(t *testing.T) {
	const width = 20
	for _, tc := range []struct {
		name    string
		padding int
	}{
		{name: "default"},
		{name: "extra padding", padding: 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := New(nil, NewDefaultDelegate(), width, 7)
			m.Title = "A long 日本語 title that exceeds the allotted width"
			m.SetShowHelp(false)
			m.SetShowPagination(false)
			m.SetShowStatusBar(false)
			if tc.padding != 0 {
				m.Styles.TitleBar = m.Styles.TitleBar.Padding(0, tc.padding)
			}

			title, _, _ := strings.Cut(m.View(), "\n")
			if got := ansi.StringWidth(title); got > width {
				t.Errorf("title occupies %d columns, want at most %d: %q", got, width, ansi.Strip(title))
			}
			if !strings.Contains(ansi.Strip(title), "A long") || !strings.Contains(title, ellipsis) {
				t.Errorf("expected a truncated title, got %q", ansi.Strip(title))
			}
		})
	}
}
