package viewport

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/ansi"
)

// impl s[n:] ansi AND unicode aware
// we have to keep not resetted ansi codes
// ignores all non ansi color/graphics sequences
// regards ansi color and graphics mode (... elaborate)
func ansiStringSlice(s string, n int) string {
	if n <= 0 {
		return s
	}

	if lipgloss.Width(s) < n {
		return ""
	}

	// count n down until we have to cut but making sure that
	// if the character at n is also control we have to continue
	// reading it as this might be a reset or similar

	var i int
	var c rune

	isansi := false

	// instead of implementing the rather complex logic just
	// keep track of all ansi codes that we encountered
	ansicodes := strings.Builder{}
	for i, c = range s {
		switch {
		case c == ansi.Marker:
			isansi = true
			ansicodes.WriteRune(c)

		// at the end of the sequence or when we have multiple arguments ...
		// process what we got
		case isansi && (c == ';' || ansi.IsTerminator(c)):
			if ansi.IsTerminator(c) {
				isansi = false
			}
			ansicodes.WriteRune(c)

		// we don't count whatever is inside the ansi sequence
		case isansi:
			ansicodes.WriteRune(c)

		// outside of ansi mode we count
		// TODO: add comment
		default:
			if n == 0 {
				return ansicodes.String() + s[i:]
			}
			n--
		}
	}

	return ""
}
