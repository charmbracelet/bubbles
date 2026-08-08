package textarea

import (
	"math/rand"
	"strings"
	"testing"
	"unicode"

	rw "github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
)

// oldWrap is the pre-optimization wrap() implementation, kept verbatim as a
// byte-level oracle for the builder-based rewrite. The new wrap() must
// produce identical output for every input.
func oldWrap(runes []rune, width int) [][]rune {
	var (
		lines  = [][]rune{{}}
		word   = []rune{}
		row    int
		spaces int
	)

	// Word wrap the runes
	for _, r := range runes {
		if unicode.IsSpace(r) {
			spaces++
		} else {
			word = append(word, r)
		}

		if spaces > 0 {
			if uniseg.StringWidth(string(lines[row]))+uniseg.StringWidth(string(word))+spaces > width {
				row++
				lines = append(lines, []rune{})
				lines[row] = append(lines[row], word...)
				lines[row] = append(lines[row], oldRepeatSpaces(spaces)...)
				spaces = 0
				word = nil
			} else {
				lines[row] = append(lines[row], word...)
				lines[row] = append(lines[row], oldRepeatSpaces(spaces)...)
				spaces = 0
				word = nil
			}
		} else {
			// If the last character is a double-width rune, then we may not be able to add it to this line
			// as it might cause us to go past the width.
			lastCharLen := rw.RuneWidth(word[len(word)-1])
			if uniseg.StringWidth(string(word))+lastCharLen > width {
				// If the current line has any content, let's move to the next
				// line because the current word fills up the entire line.
				if len(lines[row]) > 0 {
					row++
					lines = append(lines, []rune{})
				}
				lines[row] = append(lines[row], word...)
				word = nil
			}
		}
	}

	if uniseg.StringWidth(string(lines[row]))+uniseg.StringWidth(string(word))+spaces >= width {
		lines = append(lines, []rune{})
		lines[row+1] = append(lines[row+1], word...)
		// We add an extra space at the end of the line to account for the
		// trailing space at the end of the previous soft-wrapped lines so that
		// behaviour when navigating is consistent and so that we don't need to
		// continually add edges to handle the last line of the wrapped input.
		spaces++
		lines[row+1] = append(lines[row+1], oldRepeatSpaces(spaces)...)
	} else {
		lines[row] = append(lines[row], word...)
		spaces++
		lines[row] = append(lines[row], oldRepeatSpaces(spaces)...)
	}

	return lines
}

func oldRepeatSpaces(n int) []rune {
	return []rune(strings.Repeat(string(' '), n))
}

// sameRows reports whether two wrap results are byte-for-byte identical.
func sameRows(a, b [][]rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if string(a[i]) != string(b[i]) {
			return false
		}
	}
	return true
}

// TestWrapParityCorpus checks the builder rewrite against oldWrap on a
// hand-picked corpus: ASCII word boundaries, leading/trailing spaces, long
// unbroken words, CJK and ambiguous-width runes, empty input, widths 1..80.
func TestWrapParityCorpus(t *testing.T) {
	corpus := []string{
		"",
		" ",
		"   ",
		"a",
		"word",
		"a b c d e f g",
		"  leading spaces",
		"trailing spaces  ",
		"  both sides  ",
		strings.Repeat("a", 200), // long unbroken word
		strings.Repeat("supercalifragilisticexpialidocious", 10),
		"word " + strings.Repeat("x", 120) + " word", // long word mid-sentence
		"世",
		"世界你好世界你好世界你好",
		"hello 世界 world 你好",
		"αβγ αβγ αβγ", // ambiguous-width Greek
		"· · · ·",     // ambiguous-width middle dot
		"─ ─ ─",       // box drawing, ambiguous-width
		"🙂🙂🙂🙂🙂",
		"a🙂b🙂c",
		"  世 界 你 好  ",
		"tab\tseparated\twords",
		"newline\nseparated",
		"word " + strings.Repeat("世", 50) + " tail",
		"widths of\nmultiple\nlines",
		strings.Repeat("ab ", 100) + ".",
	}

	for _, s := range corpus {
		for width := 1; width <= 80; width++ {
			runes := []rune(s)
			got := wrap(runes, width)
			want := oldWrap(runes, width)
			if !sameRows(got, want) {
				t.Fatalf("parity mismatch for %q width=%d\nnew: %q\nold: %q",
					s, width, rowsToStrings(got), rowsToStrings(want))
			}
		}
	}
}

// TestWrapParityFuzz deterministically fuzzes wrap() against oldWrap over an
// alphabet that mixes ASCII, whitespace, CJK, emoji and ambiguous-width
// runes.
func TestWrapParityFuzz(t *testing.T) {
	const alphabet = "abc XYZ  \t\n!?世 界 你 好 🙂 α β · ─ ,."
	rng := rand.New(rand.NewSource(20260808))
	for i := 0; i < 5000; i++ {
		n := rng.Intn(160)
		var sb strings.Builder
		for j := 0; j < n; j++ {
			sb.WriteRune(rune(alphabet[rng.Intn(len(alphabet))]))
		}
		width := 1 + rng.Intn(80)
		runes := []rune(sb.String())
		got := wrap(runes, width)
		want := oldWrap(runes, width)
		if !sameRows(got, want) {
			t.Fatalf("parity mismatch (iter %d) for %q width=%d\nnew: %q\nold: %q",
				i, sb.String(), width, rowsToStrings(got), rowsToStrings(want))
		}
	}
}

// TestWrapCursorParity adopts the go-Whale methodology: for every cursor
// position in the line, LineInfo must report the same height and row offset
// that the wrap() output implies, i.e. the soft-wrap grid and the cursor
// navigation logic agree.
func TestWrapCursorParity(t *testing.T) {
	inputs := []string{
		"a",
		"hello world",
		"  padded   ",
		strings.Repeat("word ", 20),
		"world 世界 你好",
		"supercalifragilisticexpialidocious supercalifragilisticexpialidocious",
		"🙂 🙂 🙂 🙂 🙂",
		"tab\there\tand\there",
	}
	for _, s := range inputs {
		for width := 1; width <= 40; width++ {
			m := New()
			m.SetWidth(width)
			m.SetValue(s)
			// m.width accounts for prompt, line numbers and borders reserved
			// by SetWidth, and SetValue sanitizes the input; the wrap grid
			// must be computed from exactly what LineInfo will see.
			line := m.value[0]
			grid := wrap(line, m.width)
			for k := 0; k <= len(line); k++ {
				m.SetCursorColumn(k)
				info := m.LineInfo()
				wantRow, wantHeight := gridRowForCursor(grid, k)
				if info.RowOffset != wantRow || info.Height != wantHeight {
					t.Fatalf("cursor parity mismatch for %q width=%d k=%d: got row=%d height=%d, want row=%d height=%d (grid %q)",
						s, width, k, info.RowOffset, info.Height, wantRow, wantHeight, rowsToStrings(grid))
				}
			}
		}
	}
}

// gridRowForCursor computes the (row offset, height) that LineInfo derives
// from a wrap() grid for a cursor at column k, mirroring LineInfo's counting
// logic exactly.
func gridRowForCursor(grid [][]rune, k int) (row, height int) {
	var counter int
	for i, line := range grid {
		if counter+len(line) == k && i+1 < len(grid) {
			return i + 1, len(grid)
		}
		if counter+len(line) >= k {
			return i, len(grid)
		}
		counter += len(line)
	}
	return 0, len(grid)
}

func rowsToStrings(rows [][]rune) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = string(r)
	}
	return out
}

// BenchmarkWrap measures the soft-wrap hot path. A 20k-rune line used to cost
// roughly 13k allocations per call (one per row flush); the builder rewrite
// materializes only the final [][]rune.
func BenchmarkWrap(b *testing.B) {
	cases := []struct {
		name  string
		runes []rune
		width int
	}{
		{"20k", []rune(strings.Repeat("word ", 4000)), 40},
		{"20kNoSpaces", []rune(strings.Repeat("a", 20000)), 40},
		{"20kCJK", []rune(strings.Repeat("世", 20000)), 40},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = wrap(tc.runes, tc.width)
			}
		})
	}
}

// BenchmarkLineInfoCached measures LineInfo on a 20k line whose wrap result
// is already cached: every lookup must hit memoizedWrap without allocating,
// which is what the FNV-1a hash key enables.
func BenchmarkLineInfoCached(b *testing.B) {
	m := New()
	m.SetWidth(40)
	m.SetValue(strings.Repeat("word ", 4000))
	m.LineInfo() // warm the cache

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.LineInfo()
	}
}
