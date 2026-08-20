package textarea

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// tokenHighlighter is a test LineHighlighter that marks every "@"-prefixed
// word with a fixed style.
type tokenHighlighter struct {
	style lipgloss.Style
}

func (h tokenHighlighter) Highlight(_ int, line []rune) []lipgloss.Range {
	var ranges []lipgloss.Range
	i := 0
	for i < len(line) {
		if line[i] == '@' {
			end := i + 1
			for end < len(line) && line[end] != ' ' {
				end++
			}
			ranges = append(ranges, lipgloss.NewRange(i, end, h.style))
			i = end
			continue
		}
		i++
	}
	return ranges
}

func TestStyleSegmentAppliesRanges(t *testing.T) {
	t.Parallel()

	base := lipgloss.NewStyle()
	tokenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))

	seg := []rune("see @foo.go here")
	ranges := []lipgloss.Range{lipgloss.NewRange(4, 11, tokenStyle)}

	got := styleSegment(base, seg, 0, ranges)

	if plain := base.Render(string(seg)); got == plain {
		t.Error("styled segment must differ from the plain render")
	}
	if want := "see @foo.go here"; !strings.Contains(ansi.Strip(got), want) {
		t.Errorf("stripped output = %q, want it to contain %q", ansi.Strip(got), want)
	}
}

// TestStyleSegmentStylesExactlyTheToken pins which cells the token style
// lands on. Regression: the rune-to-cell conversion used to add one, so
// every token rendered shifted a cell to the right — the trigger character
// stayed unstyled and the following space picked the style up instead.
func TestStyleSegmentStylesExactlyTheToken(t *testing.T) {
	t.Parallel()

	base := lipgloss.NewStyle()
	tokenStyle := lipgloss.NewStyle().Bold(true)

	seg := []rune("ab @foo cd")
	got := styleSegment(base, seg, 0, []lipgloss.Range{lipgloss.NewRange(3, 7, tokenStyle)})

	if want := "ab \x1b[1m@foo\x1b[m cd"; got != want {
		t.Errorf("styleSegment() = %q, want %q", got, want)
	}
}

// TestStyleSegmentStylesExactlyTheTokenWideRunes is the same pin with
// double-width runes ahead of the token, where a rune offset and a cell
// offset diverge.
func TestStyleSegmentStylesExactlyTheTokenWideRunes(t *testing.T) {
	t.Parallel()

	base := lipgloss.NewStyle()
	tokenStyle := lipgloss.NewStyle().Bold(true)

	// "世界" is two runes but four cells.
	seg := []rune("世界 @foo")
	got := styleSegment(base, seg, 0, []lipgloss.Range{lipgloss.NewRange(3, 7, tokenStyle)})

	if want := "世界 \x1b[1m@foo\x1b[m"; got != want {
		t.Errorf("styleSegment() = %q, want %q", got, want)
	}
}

func TestStyleSegmentOffsetsBySegStart(t *testing.T) {
	t.Parallel()

	base := lipgloss.NewStyle()
	tokenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))

	// Logical line: "see @foo.go here"; the wrapped segment is "foo.go here"
	// starting at rune 7, so the token range (4,11) intersects as (0,4).
	seg := []rune("foo.go here")
	ranges := []lipgloss.Range{lipgloss.NewRange(4, 11, tokenStyle)}

	got := styleSegment(base, seg, 7, ranges)

	if want := "foo.go here"; !strings.Contains(ansi.Strip(got), want) {
		t.Errorf("stripped output = %q, want it to contain %q", ansi.Strip(got), want)
	}
	if plain := base.Render(string(seg)); got == plain {
		t.Error("a range intersecting the segment must change the render")
	}
}

func TestStyleSegmentNoRangesIsPlainRender(t *testing.T) {
	t.Parallel()

	base := lipgloss.NewStyle()
	seg := []rune("plain text")
	want := base.Render(string(seg))

	if got := styleSegment(base, seg, 0, nil); got != want {
		t.Errorf("nil ranges: got %q, want plain render %q", got, want)
	}

	outside := []lipgloss.Range{lipgloss.NewRange(50, 60, lipgloss.NewStyle().Bold(true))}
	if got := styleSegment(base, seg, 0, outside); got != want {
		t.Errorf("range outside segment: got %q, want plain render %q", got, want)
	}
}

func TestViewWithHighlighterStylesTokens(t *testing.T) {
	t.Parallel()

	const value = "see @foo.go and @bar.md"

	m := New()
	m.SetValue(value)
	m.SetWidth(80)
	m.SetHighlighter(tokenHighlighter{style: lipgloss.NewStyle().Foreground(lipgloss.Color("2"))})

	plain := New()
	plain.SetValue(value)
	plain.SetWidth(80)

	highlighted := m.View()
	if highlighted == plain.View() {
		t.Error("highlighter should change the rendered view")
	}
	// Content must survive styling: stripping ANSI yields the same text.
	if got, want := ansi.Strip(highlighted), ansi.Strip(plain.View()); got != want {
		t.Errorf("stripped view = %q, want %q", got, want)
	}
}

func TestViewWithoutHighlighterUnchanged(t *testing.T) {
	t.Parallel()

	m := New()
	m.SetValue("see @foo.go")
	m.SetWidth(80)

	before := m.View()
	m.SetHighlighter(nil)
	if got := m.View(); got != before {
		t.Error("a nil highlighter must leave rendering unchanged")
	}
}

func TestViewHighlighterOnCursorLine(t *testing.T) {
	t.Parallel()

	m := New()
	m.Focus()
	m.SetValue("go @foo.go now")
	m.SetWidth(80)
	m.SetHighlighter(tokenHighlighter{style: lipgloss.NewStyle().Foreground(lipgloss.Color("2"))})

	// Move the cursor into the middle of the token to exercise the
	// pre/post-cursor split paths.
	m.SetCursorColumn(5)

	// Cursor-line rendering must not corrupt content.
	if want := "go @foo.go now"; !strings.Contains(ansi.Strip(m.View()), want) {
		t.Errorf("stripped view missing %q", want)
	}
}

func TestViewHighlighterTokenSurvivesWrap(t *testing.T) {
	t.Parallel()

	m := New()
	// Force a wrap inside the token: width 10 splits "ab @longtoken cd"
	// across several segments.
	m.SetValue("ab @longtoken cd")
	m.SetWidth(10)
	m.SetHeight(6)
	m.SetHighlighter(tokenHighlighter{style: lipgloss.NewStyle().Foreground(lipgloss.Color("2"))})

	highlighted := m.View()
	stripped := ansi.Strip(highlighted)

	// Every token character still renders after wrapping + styling. The wrap
	// splits the token across segments, so assert on the wrap chunks.
	for _, part := range []string{"ab", "@lon", "gtok", "en", "cd"} {
		if !strings.Contains(stripped, part) {
			t.Errorf("wrapped output missing %q", part)
		}
	}
	if !strings.Contains(highlighted, "\x1b[") {
		t.Error("expected ANSI styling in output")
	}
}

// TestRuneDisplayWidthMatchesGraphemeSegmentation pins the measurement
// contract with lipgloss.StyleRanges.
//
// StyleRanges resolves the cell offsets it is handed with ansi.Cut, which
// segments by grapheme cluster. Summing per-rune widths disagrees on every
// cluster built from more than one rune, and the highlight then lands that
// many cells right of its token.
func TestRuneDisplayWidthMatchesGraphemeSegmentation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
	}{
		{"ascii", "hello"},
		{"cjk double width", "日本語"},
		{"emoji with skin tone modifier", "👍🏽"},
		{"zwj family", "👨‍👩‍👧"},
		{"emoji then token", "👍🏽 @file.go"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got, want := runeDisplayWidth([]rune(tc.in)), ansi.StringWidth(tc.in); got != want {
				t.Errorf("runeDisplayWidth(%q) = %d, want %d (grapheme-cluster segmentation)",
					tc.in, got, want)
			}
		})
	}
}

// TestHighlightRangesLandOnTokenAfterEmoji is the user-visible form of the
// bug: with per-rune widths the styled span started two cells past the '@'.
func TestHighlightRangesLandOnTokenAfterEmoji(t *testing.T) {
	t.Parallel()

	// "👍🏽 @file.go" — the token starts after a 2-cell grapheme and a space,
	// so at cell 3. Per-rune summing put it at 5.
	const line = "👍🏽 @file.go"

	at := strings.IndexRune(line, '@')
	if at <= 0 {
		t.Fatalf("test string must contain '@', got byte index %d", at)
	}
	atRune := len([]rune(line[:at]))

	if got, want := runeDisplayWidth([]rune(line)[:atRune]), 3; got != want {
		t.Errorf("token starts at cell %d, want %d (the cell ansi.Cut would land on)", got, want)
	}
}
