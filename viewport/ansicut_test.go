package viewport

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

const (
	ANSI_CURSOR_CONTROL = "\033[#C"
	ANSI_ERASE          = "\033[1J"

	ANSI_RESET  = "\033[0m"
	ANSI_HIDDEN = "\033[7m"
)

func bold(s string) string {
	return lipgloss.NewStyle().Bold(true).Render(s)
}

func faint(s string) string {
	return lipgloss.NewStyle().Faint(true).Render(s)
}

func join(strs ...string) string {
	return strings.Join(strs, " ")
}

func TestCut(t *testing.T) {
	type testcase struct {
		offset   int
		expected string
	}

	type testcaseGroup struct {
		desc  string
		input string
		cases []testcase
	}

	oldRenderer := lipgloss.DefaultRenderer()
	defer func() { lipgloss.SetDefaultRenderer(oldRenderer) }()

	renderer := lipgloss.NewRenderer(io.Discard, termenv.WithTTY(true), termenv.WithProfile(termenv.TrueColor))
	lipgloss.SetDefaultRenderer(renderer)

	testcaseGroups := []testcaseGroup{
		// english alphabet word cutting
		// also here we test that we can cut runes
		// with a width of 1
		{
			"English",
			"Word",
			[]testcase{
				{0, ("Word")},
				{1, ("ord")},
				{2, ("rd")},
				{3, ("d")},
				{4, ("")},
				{5, ("")},
			},
		},

		// Chineese alphabet word cutting
		// also here we test that we can cut runes
		// with a width of > 1
		{
			"Chineese",
			"传传传",
			[]testcase{
				{0, ("传传传")},
				{1, ("传传")},
				{2, ("传")},
				{3, ("")},
				{4, ("")},
			},
		},

		{
			"Emoji",
			"🍫🍬🍭",
			[]testcase{
				{0, "🍫🍬🍭"},
				{1, "🍬🍭"},
				{2, "🍭"},
				{3, ""},
				{4, ""},
			},
		},

		// test that indexing works despite control sequences
		// meant for the terminal emulator
		// {
		// 	"Cursor Control Sequence",
		// 	ANSI_CURSOR_CONTROL + "Test",
		// 	[]testcase{
		// 		{0, ANSI_CURSOR_CONTROL + "Test"},
		// 		{1, ANSI_CURSOR_CONTROL + "est"},
		// 		{3, ANSI_CURSOR_CONTROL + "t"},
		// 		{5, ""},
		// 	},
		// },

		// keeping styling even tho we cut into the string
		{
			"Bold",
			bold("Test"),
			[]testcase{
				{0, bold("Test")},
				{1, bold("est")},
				{2, bold("st")},
				{3, bold("t")},
				{4, ""},
				{5, ""},
			},
		},

		{
			"Bold Normal",
			join(bold("Bold"), "Normal"),
			[]testcase{
				{0, join(bold("Bold"), "Normal")},
				{1, join(bold("old"), "Normal")},
				{4, join("", "Normal")},
				{5, join("Normal")},
				{9, join("al")},
				{11, ""},
				{12, ""},
			},
		},

		{
			"Bold Normal Bold",
			join(bold("Bold"), "Normal", bold("Bold")),
			[]testcase{
				// {0, join(bold("Bold"), "Normal", bold("Bold"))},
				{1, join(bold("old"), "Normal", bold("Bold"))},
				{4, join("", "Normal", bold("Bold"))},
				// {5, join("Normal", bold("Bold"))},
				// {10, join("l", bold("Bold"))},
				// {12, join(bold("Bold"))},
				// {15, join(bold("d"))},
				// {16, join("")},
			},
		},

		{
			"Faint Normal Bold",
			join(faint("Faint"), "Normal", bold("Bold")),
			[]testcase{
				// {0, join(faint("Faint"), "Normal", bold("Bold"))},
				// {1, join(faint("aint"), "Normal", bold("Bold"))},
				// {5, join("", "Normal", bold("Bold"))},
				// {6, join("Normal", bold("Bold"))},
				// {11, join("l", bold("Bold"))},
				// {13, join(bold("Bold"))},
				// {16, join(bold("d"))},
				// {17, join("")},
			},
		},

		// {
		// 	"Faint Normal Bold",
		// 	join(faint("Faint"), "Normal", bold("Bold")),
		// 	[]testcase{
		// 		{0, join(faint("Faint"), "Normal", bold("Bold"))},
		// 		{1, join(faint("aint"), "Normal", bold("Bold"))},
		// 		{5, join("", "Normal", bold("Bold"))},
		// 		{6, join("Normal", bold("Bold"))},
		// 		{11, join("l", bold("Bold"))},
		// 		{13, join(bold("Bold"))},
		// 		{16, join(bold("d"))},
		// 		{17, join("")},
		// 	},
		// },
	}

	for _, group := range testcaseGroups {
		for _, testcase := range group.cases {
			t.Run(fmt.Sprintf("%s with offset %d", group.desc, testcase.offset), func(t *testing.T) {
				actual := ansiStringSlice(group.input, testcase.offset)
				if testcase.expected != actual {
					t.Errorf(
						"Expected '%s'[%d:] to equal '%s'(%d) but it is '%s'(%d)\n"+
							"Expected: %v\nActual:   %v\n"+
							"Expected: %v\nActual:   %v\n",
						group.input,
						testcase.offset,
						testcase.expected,
						len(testcase.expected),
						actual,
						len(actual),
						testcase.expected,
						actual,
						[]rune(testcase.expected),
						[]rune(actual),
					)
				}
			})
		}
	}
}
