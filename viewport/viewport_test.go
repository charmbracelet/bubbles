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

// generate test using lipgloss??
func mkBold(s string) string {
	return fmt.Sprintf("%s%s%s", "\033[1m", s, "\033[22m")
}

func TestCut(t *testing.T) {
	testcases := []struct {
		intput   string
		offset   int
		expected string
	}{
		// english alphabet word cutting
		// also here we test that we can cut runes
		// with a width of 1
		// {"Word", 0, "Word"},
		// {"Word", 1, "ord"},
		// {"Word", 2, "rd"},
		// {"Word", 3, "d"},
		// {"Word", 4, ""},
		// {"Word", 5, ""},

		// Chineese alphabet word cutting
		// also here we test that we can cut runes
		// with a width of > 1
		// {"传传传", 0, "传传传"},
		// {"传传传", 1, "传传"},
		// {"传传传", 2, "传"},
		// {"传传传", 3, ""},
		// {"传传传", 4, ""},

		// {"🍫🍬🍭", 0, "🍫🍬🍭"},
		// {"🍫🍬🍭", 1, "🍬🍭"},
		// {"🍫🍬🍭", 2, "🍭"},
		// {"🍫🍬🍭", 3, ""},
		// {"🍫🍬🍭", 4, ""},

		// test that we completely ignore control sequences

		// Test that we preserve graphical ansi control sequences
		// when we cut within an ansi sequence (TODO: reword)
		// {mkBold("Test"), 0, mkBold("Test")},
		{mkBold("Test"), 1, mkBold("est")},
		// {mkBold("Test"), 2, mkBold("st")},
		// {mkBold("Test"), 3, mkBold("t")},
		// {mkBold("Test"), 4, ""},
		// {mkBold("Test"), 5, ""},
	}

	for _, testcase := range testcases {
		t.Run(fmt.Sprintf("Graphic Rune Count Of: %s", testcase.intput), func(t *testing.T) {
			actual := cutAt(testcase.intput, testcase.offset)
			if testcase.expected != actual {
				t.Errorf("Expected '%s'[%d:] to equal '%s'(%d) but it is '%s'(%d)", testcase.intput, testcase.offset, testcase.expected, len(testcase.expected), actual, len(actual))
			}
		})
	}
}
