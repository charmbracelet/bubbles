package table

import (
	"fmt"
	"image/color"
	"reflect"
	"testing"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/lipgloss/v2/table"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
)

var (
	niceMargins = lipgloss.NewStyle().Padding(0, 1)
	headers     = []string{"Rank", "City", "Country", "Population"}
	rows        = [][]string{
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
)

// Tests

func TestNewBash(t *testing.T) {
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

func TestModel_FromValues(t *testing.T) {
	table := New(
		WithHeaders("Foo", "Bar"),
		WithRows(
			[]string{"foo1", "bar1"},
			[]string{"foo2", "bar2"},
			[]string{"foo3", "bar3"},
		))

	if len(table.rows) != 3 {
		t.Fatalf("expect table to have 3 rows but it has %d", len(table.rows))
	}

	expect := [][]string{
		{"foo1", "bar1"},
		{"foo2", "bar2"},
		{"foo3", "bar3"},
	}
	if !reflect.DeepEqual(table.rows, expect) {
		t.Fatalf("\n\nwant %v\n\ngot %v", expect, table.rows)
	}
}

func TestModel_FromValues_WithTabSeparator(t *testing.T) {
	table := New(
		WithHeaders("Foo", "Bar"),
		WithRows(
			[]string{"foo1.", "bar1"},
			[]string{"foo,bar,baz", "bar,2"},
		),
	)

	if len(table.rows) != 2 {
		t.Fatalf("expect table to have 2 rows but it has %d", len(table.rows))
	}

	expect := [][]string{
		{"foo1.", "bar1"},
		{"foo,bar,baz", "bar,2"},
	}
	if !reflect.DeepEqual(table.rows, expect) {
		t.Fatalf("\n\nwant %v\n\ngot %v", expect, table.rows)
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
	t.Run("NoBorder", func(t *testing.T) {
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
	t.Run("WithBorder", func(t *testing.T) {
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
		{"ClearStyles", Styles{
			Selected: lipgloss.NewStyle(),
			Header:   lipgloss.NewStyle(),
			Cell:     lipgloss.NewStyle(),
		}},
		{"NewStyles", Styles{
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
		{"EmptyStyles", Styles{
			Selected: lipgloss.NewStyle(),
			Header:   lipgloss.NewStyle(),
			Cell:     lipgloss.NewStyle(),
		}},
		{"NewStyles", Styles{
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
	t.Run("ClearStylesWithStyleFunc", func(t *testing.T) {
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
		{"UnsetAllBorders", []bool{false}},
		{"SetAllBorders", []bool{true}},
		{"VerticalBordersOnly", []bool{true, false}},
		{"NoTopBorder", []bool{false, true, true}},
		{"NoLeftBorder", []bool{true, true, true, false}},
		{"RowSeparatorAndNoRightBorder", []bool{true, false, true, true, true}},
		{"SetRowAndColumnSeparators", []bool{false, false, false, false, true, true}},
		{"InvalidNumberOfArguments", []bool{true, false, false, false, false, true, true}},
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

func TestNewFromTemplate(t *testing.T) {
	// Using Pokemon example from https://github.com/charmbracelet/lipgloss.
	baseStyle := lipgloss.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Foreground(lipgloss.Color("252")).Bold(true)
	selectedStyle := baseStyle.Foreground(lipgloss.Color("#01BE85")).Background(lipgloss.Color("#00432F"))
	typeColors := map[string]color.Color{
		"Bug":      lipgloss.Color("#D7FF87"),
		"Electric": lipgloss.Color("#FDFF90"),
		"Fire":     lipgloss.Color("#FF7698"),
		"Flying":   lipgloss.Color("#FF87D7"),
		"Grass":    lipgloss.Color("#75FBAB"),
		"Ground":   lipgloss.Color("#FF875F"),
		"Normal":   lipgloss.Color("#929292"),
		"Poison":   lipgloss.Color("#7D5AFC"),
		"Water":    lipgloss.Color("#00E2C7"),
	}
	dimTypeColors := map[string]color.Color{
		"Bug":      lipgloss.Color("#97AD64"),
		"Electric": lipgloss.Color("#FCFF5F"),
		"Fire":     lipgloss.Color("#BA5F75"),
		"Flying":   lipgloss.Color("#C97AB2"),
		"Grass":    lipgloss.Color("#59B980"),
		"Ground":   lipgloss.Color("#C77252"),
		"Normal":   lipgloss.Color("#727272"),
		"Poison":   lipgloss.Color("#634BD0"),
		"Water":    lipgloss.Color("#439F8E"),
	}

	pokemonHeaders := []string{"#", "Name", "Type 1", "Type 2", "Japanese", "Official Rom."}
	pokemonData := [][]string{
		{"1", "Bulbasaur", "Grass", "Poison", "フシギダネ", "Fushigidane"},
		{"2", "Ivysaur", "Grass", "Poison", "フシギソウ", "Fushigisou"},
		{"3", "Venusaur", "Grass", "Poison", "フシギバナ", "Fushigibana"},
		{"4", "Charmander", "Fire", "", "ヒトカゲ", "Hitokage"},
		{"5", "Charmeleon", "Fire", "", "リザード", "Lizardo"},
		{"6", "Charizard", "Fire", "Flying", "リザードン", "Lizardon"},
		{"7", "Squirtle", "Water", "", "ゼニガメ", "Zenigame"},
		{"8", "Wartortle", "Water", "", "カメール", "Kameil"},
		{"9", "Blastoise", "Water", "", "カメックス", "Kamex"},
		{"10", "Caterpie", "Bug", "", "キャタピー", "Caterpie"},
		{"11", "Metapod", "Bug", "", "トランセル", "Trancell"},
		{"12", "Butterfree", "Bug", "Flying", "バタフリー", "Butterfree"},
		{"13", "Weedle", "Bug", "Poison", "ビードル", "Beedle"},
		{"14", "Kakuna", "Bug", "Poison", "コクーン", "Cocoon"},
		{"15", "Beedrill", "Bug", "Poison", "スピアー", "Spear"},
		{"16", "Pidgey", "Normal", "Flying", "ポッポ", "Poppo"},
		{"17", "Pidgeotto", "Normal", "Flying", "ピジョン", "Pigeon"},
		{"18", "Pidgeot", "Normal", "Flying", "ピジョット", "Pigeot"},
		{"19", "Rattata", "Normal", "", "コラッタ", "Koratta"},
		{"20", "Raticate", "Normal", "", "ラッタ", "Ratta"},
		{"21", "Spearow", "Normal", "Flying", "オニスズメ", "Onisuzume"},
		{"22", "Fearow", "Normal", "Flying", "オニドリル", "Onidrill"},
		{"23", "Ekans", "Poison", "", "アーボ", "Arbo"},
		{"24", "Arbok", "Poison", "", "アーボック", "Arbok"},
		{"25", "Pikachu", "Electric", "", "ピカチュウ", "Pikachu"},
		{"26", "Raichu", "Electric", "", "ライチュウ", "Raichu"},
		{"27", "Sandshrew", "Ground", "", "サンド", "Sand"},
		{"28", "Sandslash", "Ground", "", "サンドパン", "Sandpan"},
	}

	lipglossTable := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers(pokemonHeaders...).
		Width(80).
		Rows(pokemonData...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}

			if pokemonData[row][1] == "Pikachu" {
				return selectedStyle
			}

			even := row%2 == 0

			switch col {
			case 2, 3: // Type 1 + 2
				c := typeColors
				if even {
					c = dimTypeColors
				}

				color := c[fmt.Sprint(pokemonData[row][col])]
				return baseStyle.Foreground(color)
			}

			if even {
				return baseStyle.Foreground(lipgloss.Color("245"))
			}
			return baseStyle.Foreground(lipgloss.Color("252"))
		})

	bubblesTable := NewFromTemplate(lipglossTable, headers, pokemonData)
	golden.RequireEqual(t, []byte(bubblesTable.View()))
}

func TestOverwriteStylesFromLipgloss(t *testing.T) {
	baseStyle := lipgloss.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Foreground(lipgloss.Color("252")).Bold(true)
	lipglossTable := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		Width(80).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}

			even := row%2 == 0

			if even {
				return baseStyle.Foreground(lipgloss.Color("245"))
			}
			return baseStyle.Foreground(lipgloss.Color("252"))
		})
	bubblesTable := New().SetHeaders(headers...).SetRows(rows...)
	bubblesTable.OverwriteStylesFromLipgloss(lipglossTable)
	golden.RequireEqual(t, []byte(bubblesTable.View()))
}

// Examples

func ExampleOption() {
	niceMargins := lipgloss.NewStyle().Padding(0, 1)
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
	niceMargins := lipgloss.NewStyle().Padding(0, 1)
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

func TestCursorNavigation(t *testing.T) {
	tests := map[string]struct {
		rows   [][]string
		action func(*Model)
		want   int
	}{
		"New": {
			rows: [][]string{
				{"r1"},
				{"r2"},
				{"r3"},
			},
			action: func(_ *Model) {},
			want:   0,
		},
		"MoveDown": {
			rows: [][]string{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.MoveDown(2)
			},
			want: 2,
		},
		"MoveUp": {
			rows: [][]string{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.cursor = 3
				t.MoveUp(2)
			},
			want: 1,
		},
		"GotoBottom": {
			rows: [][]string{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.GotoBottom()
			},
			want: 3,
		},
		"GotoTop": {
			rows: [][]string{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.cursor = 3
				t.GotoTop()
			},
			want: 0,
		},
		"SetCursor": {
			rows: [][]string{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.SetCursor(2)
			},
			want: 2,
		},
		"MoveDown with overflow": {
			rows: [][]string{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.MoveDown(5)
			},
			want: 3,
		},
		"MoveUp with overflow": {
			rows: [][]string{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.cursor = 3
				t.MoveUp(5)
			},
			want: 0,
		},
		"Blur does not stop movement": {
			rows: [][]string{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.Blur()
				t.MoveDown(2)
			},
			want: 2,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			table := New(
				WithHeaders("col1", "col2", "col3"),
				WithRows(tc.rows...),
			)
			tc.action(table)

			if table.Cursor() != tc.want {
				t.Errorf("want %d, got %d", tc.want, table.Cursor())
			}
		})
	}
}

func TestModel_SetRows(t *testing.T) {
	table := New(WithHeaders("col1", "col2", "col3"))

	if len(table.rows) != 0 {
		t.Fatalf("want 0, got %d", len(table.rows))
	}

	table.SetRows([]string{"r1"}, []string{"r2"})

	if len(table.rows) != 2 {
		t.Fatalf("want 2, got %d", len(table.rows))
	}

	want := [][]string{{"r1"}, {"r2"}}
	if !reflect.DeepEqual(table.rows, want) {
		t.Fatalf("\n\nwant %v\n\ngot %v", want, table.rows)
	}
}

func TestModel_SetHeaders(t *testing.T) {
	table := New()

	if len(table.headers) != 0 {
		t.Fatalf("want 0, got %d", len(table.headers))
	}

	table.SetHeaders("Foo", "Bar")

	if len(table.headers) != 2 {
		t.Fatalf("want 2, got %d", len(table.headers))
	}

	want := []string{"Foo", "Bar"}
	if !reflect.DeepEqual(table.headers, want) {
		t.Fatalf("\n\nwant %v\n\ngot %v", want, table.headers)
	}
}

func TestModel_View(t *testing.T) {
	tests := map[string]struct {
		modelFunc func() *Model
		skip      bool
	}{
		"Empty": {
			modelFunc: func() *Model {
				return New()
			},
		},
		"SingleRowAndColumn": {
			modelFunc: func() *Model {
				return New(
					WithHeaders("Name"),
					WithRows(
						[]string{"Chocolate Digestives"},
					),
				)
			},
		},
		"MultipleRowsAndColumns": {
			modelFunc: func() *Model {
				return New(
					WithHeaders("Name", "Country of Origin", "Dunk-able"),
					WithRows(
						[]string{"Chocolate Digestives", "UK", "Yes"},
						[]string{"Tim Tams", "Australia", "No"},
						[]string{"Hobnobs", "UK", "Yes"},
					),
				)
			},
		},
		"ExtraPadding": {
			modelFunc: func() *Model {
				s := DefaultStyles()
				s.Header = lipgloss.NewStyle().Padding(2, 2)
				s.Cell = lipgloss.NewStyle().Padding(2, 2)

				return New(
					WithHeaders("Name", "Country of Origin", "Dunk-able"),
					WithRows(
						[]string{"Chocolate Digestives", "UK", "Yes"},
						[]string{"Tim Tams", "Australia", "No"},
						[]string{"Hobnobs", "UK", "Yes"},
					),
					WithStyles(s),
				)
			},
		},
		"NoPadding": {
			modelFunc: func() *Model {
				s := DefaultStyles()
				s.Header = lipgloss.NewStyle()
				s.Cell = lipgloss.NewStyle()

				return New(
					WithHeight(10),
					WithHeaders("Name", "Country of Origin", "Dunk-able"),
					WithRows(
						[]string{"Chocolate Digestives", "UK", "Yes"},
						[]string{"Tim Tams", "Australia", "No"},
						[]string{"Hobnobs", "UK", "Yes"},
					),
					WithStyles(s),
				)
			},
		},
		"BorderedHeaders": {
			modelFunc: func() *Model {
				return New(
					WithHeaders("Name", "Country of Origin", "Dunk-able"),
					WithRows(
						[]string{"Chocolate Digestives", "UK", "Yes"},
						[]string{"Tim Tams", "Australia", "No"},
						[]string{"Hobnobs", "UK", "Yes"},
					),
					WithStyles(Styles{
						Header: lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()),
					}),
				)
			},
		},
		"BorderedCells": {
			modelFunc: func() *Model {
				return New(
					WithHeaders("Name", "Country of Origin", "Dunk-able"),
					WithRows(
						[]string{"Chocolate Digestives", "UK", "Yes"},
						[]string{"Tim Tams", "Australia", "No"},
						[]string{"Hobnobs", "UK", "Yes"},
					),
					WithStyles(Styles{
						Cell: lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()),
					}),
				)
			},
		},
		// FIXME(@andreynering): Fix in Lip Gloss? Potentially add extra empty lines to the bottom of the table.
		"ManualHeightGreaterThanRows": {
			modelFunc: func() *Model {
				return New(
					WithHeight(15),
					WithHeaders("Name", "Country of Origin", "Dunk-able"),
					WithRows(
						[]string{"Chocolate Digestives", "UK", "Yes"},
						[]string{"Tim Tams", "Australia", "No"},
						[]string{"Hobnobs", "UK", "Yes"},
					),
				)
			},
		},
		"ManualHeightLessThanRows": {
			modelFunc: func() *Model {
				return New(
					WithHeight(2),
					WithHeaders("Name", "Country of Origin", "Dunk-able"),
					WithRows(
						[]string{"Chocolate Digestives", "UK", "Yes"},
						[]string{"Tim Tams", "Australia", "No"},
						[]string{"Hobnobs", "UK", "Yes"},
					),
				)
			},
		},
		"ManualWidthGreaterThanColumns": {
			modelFunc: func() *Model {
				return New(
					WithWidth(80),
					WithHeaders("Name", "Country of Origin", "Dunk-able"),
					WithRows(
						[]string{"Chocolate Digestives", "UK", "Yes"},
						[]string{"Tim Tams", "Australia", "No"},
						[]string{"Hobnobs", "UK", "Yes"},
					),
				)
			},
		},
		"ManualWidthLessThanColumns": {
			modelFunc: func() *Model {
				return New(
					WithWidth(30),
					WithHeaders("Name", "Country of Origin", "Dunk-able"),
					WithRows(
						[]string{"Chocolate Digestives", "UK", "Yes"},
						[]string{"Tim Tams", "Australia", "No"},
						[]string{"Hobnobs", "UK", "Yes"},
					),
				)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.skip {
				t.Skip()
			}

			table := tc.modelFunc()

			golden.RequireEqual(t, []byte(table.View()))
		})
	}
}

// TODO: Fix table to make this test will pass.
func TestModel_View_CenteredInABox(t *testing.T) {
	t.Skip()

	boxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		Align(lipgloss.Center)

	table := New(
		WithHeight(6),
		WithWidth(80),
		WithHeaders("Name", "Country of Origin", "Dunk-able"),
		WithRows(
			[]string{"Chocolate Digestives", "UK", "Yes"},
			[]string{"Tim Tams", "Australia", "No"},
			[]string{"Hobnobs", "UK", "Yes"},
		),
	)

	tableView := ansi.Strip(table.View())
	got := boxStyle.Render(tableView)

	golden.RequireEqual(t, []byte(got))
}
