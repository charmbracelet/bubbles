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

	RESET       = "\033[0m"
	ANSI_HIDDEN = "\033[7m"
)

func bold(s string) string {
	return lipgloss.NewStyle().Bold(true).Render(s)
}

func faint(s string) string {
	return lipgloss.NewStyle().Faint(true).Render(s)
}

func italic(s string) string {
	return lipgloss.NewStyle().Italic(true).Render(s)
}

func underline(s string) string {
	return lipgloss.NewStyle().Underline(true).Render(s)
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
				{0, join(bold("Bold"), "Normal", bold("Bold"))},
				{1, join(bold("old"), "Normal", bold("Bold"))},
				{4, join("", "Normal", bold("Bold"))},
				{5, join("Normal", bold("Bold"))},
				{10, join("l", bold("Bold"))},
				{12, join(bold("Bold"))},
				{15, join(bold("d"))},
				{16, join("")},
			},
		},

		{
			"Faint Normal Bold",
			join(faint("Faint"), "Normal", bold("Bold")),
			[]testcase{
				{0, join(faint("Faint"), "Normal", bold("Bold"))},
				{1, join(faint("aint"), "Normal", bold("Bold"))},
				{5, join("", "Normal", bold("Bold"))},
				{6, join("Normal", bold("Bold"))},
				{11, join("l", bold("Bold"))},
				{13, join(bold("Bold"))},
				{16, join(bold("d"))},
				{17, join("")},
			},
		},

		{
			"Italic Normal Bold",
			join(italic("Italic"), "Normal", bold("Bold")),
			[]testcase{
				{0, join(italic("Italic"), "Normal", bold("Bold"))},
				{1, join(italic("talic"), "Normal", bold("Bold"))},
				{6, join("", "Normal", bold("Bold"))},
				{7, join("Normal", bold("Bold"))},
				{12, join("l", bold("Bold"))},
				{14, join(bold("Bold"))},
				{17, join(bold("d"))},
				{18, join("")},
			},
		},

		// TODO: lipgloss does underlineSpaces on default ...
		//		  therefore every character is rendered seperately
		//		  with it's style set and reset ....
		{
			"Underl Normal Bold",
			join(underline("Underl"), "Normal", bold("Bold")),
			[]testcase{
				{0, join(underline("Underl"), "Normal", bold("Bold"))},
				{1, join(underline("nderl"), "Normal", bold("Bold"))},
				// {6, join("", "Normal", bold("Bold"))},
				// {7, join("Normal", bold("Bold"))},
				// {12, join("l", bold("Bold"))},
				// {14, join(bold("Bold"))},
				// {17, join(bold("d"))},
				// {18, join("")},
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

	}

	for _, group := range testcaseGroups {
		for _, testcase := range group.cases {
			t.Run(fmt.Sprintf("%s with offset %d", group.desc, testcase.offset), func(t *testing.T) {
				actual := ansiStringSlice(group.input, testcase.offset)
				if testcase.expected != actual {
					//
					t.Errorf(
						"Expected '%s'%s[%d:] to equal '%s'%s(%d) but it is '%s'%s(%d)\n"+
							"Expected: >%v%s<\nActual:   >%v%s<\n"+
							"Expected: %v\nActual:   %v\n",
						group.input,
						RESET,
						testcase.offset,
						testcase.expected,
						RESET,
						len(testcase.expected),
						actual,
						RESET,
						len(actual),
						testcase.expected,
						RESET,
						actual,
						RESET,
						[]rune(testcase.expected),
						[]rune(actual),
					)
				}
			})
		}
	}
}
