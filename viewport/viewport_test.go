package viewport

import (
	"fmt"
	"testing"
)

const (
	ANSI_CURSOR_CONTROL = "\033[#C"
	ANSI_ERASE          = "\033[1J"

	ANSI_RESET        = "\033[0m"
	ANSI_COLOR        = "\033[34m" // using blue
	ANSI_BOLD         = "\033[1m"
	ANSI_FAINT        = "\033[2m"
	ANSI_ITALIC       = "\033[3m"
	ANSI_UNDERLINE    = "\033[4m"
	ANSI_BLINKING     = "\033[5m"
	ANSI_INVERSE      = "\033[6m"
	ANSI_HIDDEN       = "\033[7m"
	ANSI_STRIKETROUGH = "\033[8m"
)

func mkBold(s string) string {
	// generate test using lipgloss??
	return fmt.Sprintf("%s%s%s", "\033[1m", s, "\033[22m")
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

	testcaseGroups := []testcaseGroup{
		// english alphabet word cutting
		// also here we test that we can cut runes
		// with a width of 1
		{
			"Chineese",
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
		{
			"Cursor Control Sequence",
			ANSI_CURSOR_CONTROL + "Test",
			[]testcase{
				{0, ANSI_CURSOR_CONTROL + "Test"},
				{1, ANSI_CURSOR_CONTROL + "est"},
				{3, ANSI_CURSOR_CONTROL + "t"},
				{5, ""},
			},
		},

		// keeping styling even tho we cut into the string
		{
			"Bold Text",
			mkBold("Test"),
			[]testcase{
				{0, mkBold("Test")},
				{1, mkBold("est")},
				{2, mkBold("st")},
				{3, mkBold("t")},
				{4, ""},
				{5, ""},
			},
		},

		// due to the naive "ansi sequence preprending" logic
		// we keep the control sequence altough there is no
		// text inbetween
		{
			"Multiple Bold Text Sections",
			fmt.Sprintf("%s %s", mkBold("Bold"), "Normal"),
			[]testcase{
				{0, fmt.Sprintf("%s %s", mkBold("Bold"), "Normal")},
				{1, fmt.Sprintf("%s %s", mkBold("old"), "Normal")},
				{4, fmt.Sprintf("%s %s", mkBold(""), "Normal")},
				{5, fmt.Sprintf("%s%s", mkBold(""), "Normal")},
				{9, fmt.Sprintf("%s%s", mkBold(""), "al")},
				{11, ""},
				{12, ""},
			},
		},
	}

	for _, group := range testcaseGroups {
		for _, testcase := range group.cases {
			t.Run(fmt.Sprintf("%s with offset %d", group.desc, testcase.offset), func(t *testing.T) {
				actual := ansiStringSlice(group.input, testcase.offset)
				if testcase.expected != actual {
					t.Errorf(
						"Expected '%s'[%d:] to equal '%s'(%d) but it is '%s'(%d)",
						group.input,
						testcase.offset,
						testcase.expected,
						len(testcase.expected),
						actual,
						len(actual),
					)
				}
			})
		}
	}
}
