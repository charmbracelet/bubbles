package table

import (
	"strings"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

// Much of the following code is credited to https://github.com/MichaelMure/go-term-text.

const (
	escapeSequenceStart = '\x1b'
	escapeSequenceEnd   = 'm'
)

// escapeSequence represents a terminal escape sequence from a string.
type escapeSequence struct {
	content string // the literal escape sequence
	pos     int    // the position of the escape sequence in the un-escaped string
}

// truncate truncates a string to a given width, replacing the last character with an ellipsis
// and taking into account escape sequences. The width is the number of cells, not runes.
func truncate(s string, w int) string {
	if len(s) == 0 {
		return s
	}

	if w <= 0 {
		// Assumption: we do not want to wrap the ellipsis in the original escape sequences.
		return "…"
	}

	l := lengthWithoutEscapeSequences(s)
	if l <= w || l == 0 {
		return s
	}

	cleaned, escapes := extractEscapeSequences(s)
	truncated := runewidth.Truncate(cleaned, w-1, "")

	// Assumption: we do not want to wrap the ellipsis in the original escape sequences.
	return applyEscapeSequences(truncated, escapes) + "…"
}

// lengthWithoutEscapeSequences return the length of a string in a terminal, while ignoring terminal escape sequences.
func lengthWithoutEscapeSequences(s string) int {
	l := 0
	inSequence := false

	for _, char := range s {
		if char == escapeSequenceStart {
			inSequence = true
		}
		if !inSequence {
			l += runewidth.RuneWidth(char)
		}
		if char == escapeSequenceEnd {
			inSequence = false
		}
	}

	return l
}

// extractEscapeSequences extracts terminal escape sequences from a string and returns the string without
// escape sequences and a slice of escape sequences. The provided string should not contain '\n' characters.
func extractEscapeSequences(s string) (string, []escapeSequence) {
	var sequences []escapeSequence
	var sb strings.Builder

	pos := 0
	item := ""
	occupiedRuneCount := 0
	inSequence := false
	for i, r := range []rune(s) {
		if r == escapeSequenceStart {
			pos = i
			item = string(r)
			inSequence = true
			continue
		}
		if inSequence {
			item += string(r)
			if r == escapeSequenceEnd {
				sequences = append(sequences, escapeSequence{item, pos - occupiedRuneCount})
				occupiedRuneCount += utf8.RuneCountInString(item)
				inSequence = false
			}
			continue
		}
		sb.WriteRune(r)
	}

	return sb.String(), sequences
}

// applyEscapeSequences apply the extracted terminal escape sequences to the edited string.
// Escape sequences need to be ordered by their position.
// If the position is < 0, the sequence is applied at the beginning of the string.
// If the position is > len(line), the sequence is applied at the end of the string.
func applyEscapeSequences(s string, escapes []escapeSequence) string {
	if len(escapes) == 0 {
		return s
	}

	var sb strings.Builder

	currPos := 0
	currItem := 0
	for _, r := range s {
		for currItem < len(escapes) && currPos >= escapes[currItem].pos {
			sb.WriteString(escapes[currItem].content)
			currItem++
		}
		sb.WriteRune(r)
		currPos++
	}

	// Don't forget the trailing escapes, if any.
	for currItem < len(escapes) {
		sb.WriteString(escapes[currItem].content)
		currItem++
	}

	return sb.String()
}
