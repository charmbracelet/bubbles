package table

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/lipgloss/v2/table"
	"github.com/charmbracelet/x/exp/golden"
)

// Reusable inputs

var niceMargins = lipgloss.NewStyle().Padding(0, 1)
var headers = []string{"Rank", "City", "Country", "Population"}
var rows = [][]string{
	{"1", "Tokyo", "Japan", "37,274,000"},
	{"2", "Delhi", "India", "32,065,760"},
	{"3", "Shanghai", "China", "28,516,904"},
	{"4", "Dhaka", "Bangladesh", "22,478,116"},
	{"5", "São Paulo", "Brazil", "22,429,800"},
	{"6", "Mexico City", "Mexico", "22,085,140"},
	{"7", "Cairo", "Egypt", "21,750,020"},
	{"8", "Beijing", "China", "21,333,332"},
	{"9", "Mumbai", "India", "20,961,472"},
	{"10", "Osaka", "Japan", "19,059,856"},
	{"11", "Chongqing", "China", "16,874,740"},
	{"12", "Karachi", "Pakistan", "16,839,950"},
	{"13", "Istanbul", "Turkey", "15,636,243"},
	{"14", "Kinshasa", "DR Congo", "15,628,085"},
	{"15", "Lagos", "Nigeria", "15,387,639"},
	{"16", "Buenos Aires", "Argentina", "15,369,919"},
}

// Tests

func TestNew(t *testing.T) {
	headers := []string{"Rank", "City", "Country", "Population"}
	rows := [][]string{
		{"1", "Tokyo", "Japan", "37,274,000"},
		{"2", "Delhi", "India", "32,065,760"},
		{"3", "Shanghai", "China", "28,516,904"},
		{"4", "Dhaka", "Bangladesh", "22,478,116"},
		{"5", "São Paulo", "Brazil", "22,429,800"},
		{"6", "Mexico City", "Mexico", "22,085,140"},
		{"7", "Cairo", "Egypt", "21,750,020"},
		{"8", "Beijing", "China", "21,333,332"},
		{"9", "Mumbai", "India", "20,961,472"},
		{"10", "Osaka", "Japan", "19,059,856"},
	}
	t.Run("new with options", func(t *testing.T) {
		tb := New(
			WithHeaders(headers...),
			WithRows(rows...),
			WithHeight(10),
		)
		tb.View()
	})
	t.Run("new, no options", func(t *testing.T) {
		tb := New().SetHeaders(headers...).SetRows(rows...)
		tb.View()
	})
}

func TestFromValues(t *testing.T) {
	input := "foo1,bar1\nfoo2,bar2\nfoo3,bar3"
	table := New(WithHeaders("Foo", "Bar"))
	table.FromValues(input, ",")

	if len(table.rows) != 3 {
		t.Fatalf("expect table to have 3 rows but it has %d", len(table.rows))
	}

	expect := [][]string{
		{"foo1", "bar1"},
		{"foo2", "bar2"},
		{"foo3", "bar3"},
	}
	if !reflect.DeepEqual(table.rows, expect) {
		t.Fatal("table rows is not equals to the input")
	}
}

func TestFromValuesWithTabSeparator(t *testing.T) {
	input := "foo1.\tbar1\nfoo,bar,baz\tbar,2"
	table := New(WithHeaders("Foo", "Bar"))
	table.FromValues(input, "\t")

	if len(table.rows) != 2 {
		t.Fatalf("expect table to have 2 rows but it has %d", len(table.rows))
	}

	expect := [][]string{
		{"foo1.", "bar1"},
		{"foo,bar,baz", "bar,2"},
	}
	if !reflect.DeepEqual(table.rows, expect) {
		t.Fatal("table rows is not equals to the input")
	}
}

func TestTableAlignment(t *testing.T) {
	headers := []string{
		"Name",
		"Country of Origin",
		"Dunk-able",
	}
	rows := [][]string{
		{"Chocolate Digestives", "UK", "Yes"},
		{"Tim Tams", "Australia", "No"},
		{"Hobnobs", "UK", "Yes"},
	}
	t.Run("No border", func(t *testing.T) {
		biscuits := New(
			WithHeaders(headers...),
			WithRows(rows...),
		).
			// Remove default border.
			SetBorder(false).
			// Remove default border under header.
			BorderHeader(false).
			// Strip styles.
			SetStyleFunc(func(_, _ int) lipgloss.Style {
				return niceMargins
			})
		golden.RequireEqual(t, []byte(biscuits.View()))
	})
	t.Run("With border", func(t *testing.T) {
		// TODO how do we style header border?
		//		s.Header = s.Header.
		//			BorderStyle(lipgloss.NormalBorder()).
		//			BorderForeground(lipgloss.Color("240")).
		//			Bold(false)

		biscuits := New(
			WithHeaders(headers...),
			WithRows(rows...),
		).
			// Strip styles
			SetStyleFunc(func(_, _ int) lipgloss.Style {
				return niceMargins
			})
		golden.RequireEqual(t, []byte(biscuits.View()))
	})
}

// Test Styles

func TestOverwriteStyles(t *testing.T) {
	tests := []struct {
		name   string
		styles Styles
	}{

		{"clear styles", Styles{
			Selected: lipgloss.NewStyle(),
			Header:   lipgloss.NewStyle(),
			Cell:     lipgloss.NewStyle(),
		}},
		{"new styles", Styles{
			Selected: niceMargins.Foreground(lipgloss.Color("68")),
			Header:   niceMargins,
			Cell:     niceMargins,
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tb := New(
				WithHeaders(headers...),
				WithRows(rows...),
				WithFocused(true),
				WithHeight(10),
			)
			tb.OverwriteStyles(tc.styles)
			golden.RequireEqual(t, []byte(tb.View()))
		})
	}
}

func TestSetStyles(t *testing.T) {
	tests := []struct {
		name   string
		styles Styles
	}{
		{"empty styles; do nothing", Styles{
			Selected: lipgloss.NewStyle(),
			Header:   lipgloss.NewStyle(),
			Cell:     lipgloss.NewStyle(),
		}},
		{"new styles", Styles{
			Selected: niceMargins.Background(lipgloss.Color("68")),
			Header:   niceMargins,
			Cell:     niceMargins,
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			table := New(
				WithHeaders(headers...),
				WithRows(rows...),
				WithFocused(true),
				WithHeight(10),
			)

			table.SetStyles(tc.styles)
			golden.RequireEqual(t, []byte(table.View()))
		})
	}
}

func TestSetStyleFunc(t *testing.T) {
	t.Run("Clear styles with StyleFunc", func(t *testing.T) {
		tb := New(
			WithHeaders(headers...),
			WithRows(rows...),
			WithFocused(true),
			WithHeight(10),
		)
		tb.SetStyleFunc(table.StyleFunc(func(row, col int) lipgloss.Style {
			if row == tb.Cursor() {
				return niceMargins.Background(lipgloss.Color("68"))
			}
			return niceMargins
		}))
		golden.RequireEqual(t, []byte(tb.View()))
	})
}

func TestSetBorder(t *testing.T) {
	tests := []struct {
		name    string
		borders []bool
	}{
		{"unset all borders", []bool{false}},
		{"set all borders", []bool{true}},
		{"vertical borders only", []bool{true, false}},
		{"no top border", []bool{false, true, true}},
		{"no left border", []bool{true, true, true, false}},
		{"row separator and no right border", []bool{true, false, true, true, true}},
		{"set row and column separators", []bool{false, false, false, false, true, true}},
		{"invalid number of arguments", []bool{true, false, false, false, false, true, true}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tb := New(
				WithHeaders(headers...),
				WithRows(rows...),
				WithFocused(true),
				WithHeight(10),
			).SetBorder(tc.borders...)
			golden.RequireEqual(t, []byte(tb.View()))
		})
	}
}

// Examples

func ExampleOption() {
	var niceMargins = lipgloss.NewStyle().Padding(0, 1)
	var headers = []string{"Rank", "City", "Country", "Population"}
	var rows = [][]string{
		{"1", "Tokyo", "Japan", "37,274,000"},
		{"2", "Delhi", "India", "32,065,760"},
		{"3", "Shanghai", "China", "28,516,904"},
		{"4", "Dhaka", "Bangladesh", "22,478,116"},
		{"5", "São Paulo", "Brazil", "22,429,800"},
		{"6", "Mexico City", "Mexico", "22,085,140"},
		{"7", "Cairo", "Egypt", "21,750,020"},
		{"8", "Beijing", "China", "21,333,332"},
		{"9", "Mumbai", "India", "20,961,472"},
		{"10", "Osaka", "Japan", "19,059,856"},
	}
	t := New(
		WithHeaders(headers...),
		WithRows(rows...),
		WithFocused(true),
		WithHeight(10),
	).OverwriteStyles(Styles{
		Selected: niceMargins,
		Header:   niceMargins,
		Cell:     niceMargins,
	})
	fmt.Println(t.View())
	// Output:
	//┌───────────────────────────────────────────┐
	//│ Rank  City         Country     Population │
	//├───────────────────────────────────────────┤
	//│ 1     Tokyo        Japan       37,274,000 │
	//│ 2     Delhi        India       32,065,760 │
	//│ 3     Shanghai     China       28,516,904 │
	//│ 4     Dhaka        Bangladesh  22,478,116 │
	//│ 5     São Paulo    Brazil      22,429,800 │
	//│ …     …            …           …          │
	//└───────────────────────────────────────────┘
	}

func ExampleModel_SetRows() {
	var niceMargins = lipgloss.NewStyle().Padding(0, 1)
	var headers = []string{"Rank", "City", "Country", "Population"}
	var rows = [][]string{
		{"1", "Tokyo", "Japan", "37,274,000"},
		{"2", "Delhi", "India", "32,065,760"},
		{"3", "Shanghai", "China", "28,516,904"},
		{"4", "Dhaka", "Bangladesh", "22,478,116"},
		{"5", "São Paulo", "Brazil", "22,429,800"},
		{"6", "Mexico City", "Mexico", "22,085,140"},
		{"7", "Cairo", "Egypt", "21,750,020"},
		{"8", "Beijing", "China", "21,333,332"},
		{"9", "Mumbai", "India", "20,961,472"},
		{"10", "Osaka", "Japan", "19,059,856"},
	}
	tb := New().
		SetHeaders(headers...).
		SetRows(rows...).
		SetHeight(10).
		OverwriteStyles(Styles{
			Selected: niceMargins,
			Header:   niceMargins,
			Cell:     niceMargins,
		})
	fmt.Println(tb.View())
	// Output:
	//┌───────────────────────────────────────────┐
	//│ Rank  City         Country     Population │
	//├───────────────────────────────────────────┤
	//│ 1     Tokyo        Japan       37,274,000 │
	//│ 2     Delhi        India       32,065,760 │
	//│ 3     Shanghai     China       28,516,904 │
	//│ 4     Dhaka        Bangladesh  22,478,116 │
	//│ 5     São Paulo    Brazil      22,429,800 │
	//│ …     …            …           …          │
	//└───────────────────────────────────────────┘
}
