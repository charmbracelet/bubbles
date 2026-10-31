package textarea

import (
	"slices"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// LineHighlighter is an optional hook that applies per-token styling to
// rendered lines. It is consulted once per logical line per render and
// returns style ranges over the line's runes: range Start (inclusive) and
// End (exclusive) are rune offsets within the line, and Style is composed
// with the line's base style by the renderer.
//
// The hook runs after wrapping: the textarea asks for ranges on the raw
// logical line, then intersects those ranges with each wrapped segment at
// render time. Implementations must therefore return ranges in rune offsets
// of the logical line they are given — never byte offsets, and never ranges
// derived from a different string.
//
// The cell under the virtual cursor is never styled by ranges: the renderer
// splits cursor-line content into pre- and post-cursor segments and renders
// the cursor cell itself separately. An active selection also wins over
// highlighting, so a selected segment renders in the selection style alone.
//
// Keeping this interface range-based, rather than a func(string) string that
// rewrites the rendered line, means implementations never have to parse or
// preserve escape sequences, and the textarea's wrap and cursor math always
// operates on raw runes.
type LineHighlighter interface {
	// Highlight returns the style ranges for the given logical line.
	// lineIdx is the zero-based logical line number; line is the raw rune
	// content (no trailing newline). Implementations should return nil for
	// lines with nothing to style.
	//
	// Ranges need not be sorted and may overlap or run outside the line: the
	// renderer clips them to the line, orders them, and keeps the first of
	// any overlapping pair.
	Highlight(lineIdx int, line []rune) []lipgloss.Range
}

// SetHighlighter installs an optional line highlighter. Pass nil to disable
// highlighting. The highlighter is consulted at render time only; it does
// not affect editing, wrapping, or cursor positioning.
func (m *Model) SetHighlighter(h LineHighlighter) {
	m.highlighter = h
}

// highlightRanges returns the highlighter's ranges for a logical line, or
// nil when no highlighter is installed.
func (m *Model) highlightRanges(lineIdx int, line []rune) []lipgloss.Range {
	if m.highlighter == nil {
		return nil
	}
	return m.highlighter.Highlight(lineIdx, line)
}

// styleSegment renders a wrapped-line segment, applying any highlighter
// ranges that intersect it. segStart is the rune offset of the segment
// within its logical line. The cell under the virtual cursor never reaches
// this function: the renderer splits cursor-line content into pre- and
// post-cursor segments and renders the cursor cell itself separately, so
// token styling and cursor styling can never fight over the same cell.
func styleSegment(
	base lipgloss.Style,
	seg []rune,
	segStart int,
	lineRanges []lipgloss.Range,
) string {
	if len(lineRanges) == 0 || len(seg) == 0 {
		return base.Render(string(seg))
	}

	segEnd := segStart + len(seg)
	var ranges []lipgloss.Range
	for _, r := range lineRanges {
		// Intersect the token range with this segment.
		start := max(r.Start, segStart) - segStart
		end := min(r.End, segEnd) - segStart
		if start >= end {
			continue
		}
		ranges = append(ranges, lipgloss.NewRange(start, end, r.Style))
	}

	if len(ranges) == 0 {
		return base.Render(string(seg))
	}
	ranges = normalizeRanges(ranges)

	// StyleRanges composes each range's style over the base-rendered string,
	// addressing it in display cells rather than runes; map rune offsets
	// through display widths to stay correct for wide and clustered runes.
	return lipgloss.StyleRanges(base.Render(string(seg)), toStyleRanges(seg, ranges, base)...)
}

// normalizeRanges sorts ranges by start offset and drops any that overlap an
// earlier one.
//
// [lipgloss.StyleRanges] walks its ranges in the order given and tracks how
// far it has consumed, so an out-of-order or overlapping range makes it emit
// a stretch of the line twice and skip another. The highlighter hook is
// public, so ranges arrive in whatever order an implementation produced them
// — normalizing here keeps a careless highlighter from corrupting the render
// rather than merely mis-colouring it.
func normalizeRanges(ranges []lipgloss.Range) []lipgloss.Range {
	if len(ranges) < 2 {
		return ranges
	}

	slices.SortStableFunc(ranges, func(a, b lipgloss.Range) int {
		return a.Start - b.Start
	})

	out := ranges[:1]
	for _, r := range ranges[1:] {
		if r.Start < out[len(out)-1].End {
			continue
		}
		out = append(out, r)
	}
	return out
}

// toStyleRanges converts rune-offset ranges into the display-cell ranges
// lipgloss.StyleRanges expects: zero-based and half-open over the cells of
// the ANSI-stripped string. Each token style inherits the base style's
// attributes so tokens only override what they set.
func toStyleRanges(seg []rune, ranges []lipgloss.Range, base lipgloss.Style) []lipgloss.Range {
	out := make([]lipgloss.Range, 0, len(ranges))
	for _, r := range ranges {
		startCell := runeDisplayWidth(seg[:r.Start])
		endCell := startCell + runeDisplayWidth(seg[r.Start:r.End])
		style := r.Style.Inherit(base)
		out = append(out, lipgloss.NewRange(startCell, endCell, style))
	}
	return out
}

// runeDisplayWidth returns the total display width of runes.
//
// It must measure the same way lipgloss.StyleRanges resolves the cell
// offsets it is handed — that goes through ansi.Cut, which segments by
// grapheme cluster. Summing per-rune widths instead disagrees on every
// cluster built from more than one rune: an emoji with a skin-tone modifier
// scores 4 rather than 2, a ZWJ family 7 rather than 2, and the token
// highlight lands that many cells to the right of the token it belongs to.
func runeDisplayWidth(r []rune) int {
	return ansi.StringWidth(string(r))
}
