package viewport

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/ansi"
)

// check if s conforms to: ESC + [ + ... + m
func isGraphicControlSequence(s string) bool {
	return strings.HasPrefix(s, "\x1B[") && strings.HasSuffix(s, "m")
}

func wrapIntoCSISeq(s string) string {
	return fmt.Sprintf("%c[%sm", ansi.Marker, s)
}

type activeStyle struct {
	s lipgloss.Style
	// lipgloss/termenv don't cover all cases
	extra  []string
	hidden bool
}

// predicate applied to a string
type strPredicate func(string) bool

// inRange returns a new strPredicate which checks
// if s ∈ [left, right]
func inRange(left, right string) func(string) bool {
	return func(s string) bool {
		return left <= s && s <= right
	}
}

// eq returns a new strPredicate which checks
// if s == against
func eq(against string) func(string) bool {
	return func(s string) bool {
		return s == against
	}
}

// any returns a new strPredicate which checks
// if s satisfies any of the strPredicates passed to it
func any(fns ...strPredicate) strPredicate {
	return func(s string) bool {
		for _, fn := range fns {
			if fn(s) {
				return true
			}
		}
		return false
	}
}

// createMatcherForParts returns a function that applies
// one predicate per arr element returning true when
// all predicates matched
func createMatcherForParts(arr []string) func(...strPredicate) bool {
	return func(fns ...strPredicate) bool {
		// if we have more to match than left in the slice -> no match is possible
		if len(arr) < len(fns) {
			return false
		}
		for i, fn := range fns {
			if !fn(arr[i]) {
				return false
			}
		}
		return true
	}

}

// maps a an ansi code to the lipgloss style function
var codeToColorChoice = map[string]func(lipgloss.Style, lipgloss.TerminalColor) lipgloss.Style{
	"38": lipgloss.Style.Foreground,
	"48": lipgloss.Style.Background,
}

var matchByte = inRange("0", "255")
var matchFgColorCode = any(inRange("30", "37"), inRange("90", "97"))
var matchBgColorCode = any(inRange("40", "47"), inRange("100", "107"))

var seqRgbColor = []strPredicate{any(eq("38"), eq("48")), eq("2"), matchByte, matchByte, matchByte}
var seq256Color = []strPredicate{any(eq("38"), eq("48")), eq("5"), matchByte}
var seq8To16ColorBright = []strPredicate{any(eq("01"), eq("1")), any(matchFgColorCode, matchBgColorCode)}

// parse the the graphical ansi sequence and update the state
func (as *activeStyle) updateStyle(s string) {
	// remove the start 'ESC[' and the termiator 'm'
	s = s[2 : len(s)-1]

	// s can be empty if: ESC + m == ESC + 0 + m
	if len(s) == 0 {
		as.s = lipgloss.NewStyle()
		as.extra = as.extra[:0]
		as.hidden = false
		return
	}
	parts := strings.Split(s, ";")

	for i := 0; i < len(parts); i++ {
		// helper for matching the parts based on patterns
		matchPartSeq := createMatcherForParts(parts[i:])

		// parse going from longest to shortest sequence,
		// as we might have to consume multiple parts
		switch {
		// rgb colors
		case matchPartSeq(seqRgbColor...):
			r, _ := strconv.Atoi(parts[i+2])
			g, _ := strconv.Atoi(parts[i+3])
			b, _ := strconv.Atoi(parts[i+4])
			hexrepr := fmt.Sprintf("#%02x%02x%02x", r, g, b)
			as.s = codeToColorChoice[parts[i]](as.s, lipgloss.Color(hexrepr))
			i += 4

		// 256 colors
		case matchPartSeq(seq256Color...):
			as.s = codeToColorChoice[parts[i]](as.s, lipgloss.Color(parts[i+2]))
			i += 2
		// 8-16 Colors with bright modifier
		case matchPartSeq(seq8To16ColorBright...):
			// bold/bright colors are not supported by lipgloss/termenv
			as.extra = append(as.extra, wrapIntoCSISeq("1;"+parts[i+1]))
			i++

		// 8-16 Colors
		case matchPartSeq(matchFgColorCode):
			as.s = as.s.Foreground(lipgloss.Color(parts[i]))
		case matchPartSeq(matchBgColorCode):
			as.s = as.s.Background(lipgloss.Color(parts[i]))

		// reset fg color only
		case matchPartSeq(eq("39")):
			as.s = as.s.UnsetForeground()
		// reset bg color only
		case matchPartSeq(eq("49")):
			as.s = as.s.UnsetBackground()

		// reset
		case matchPartSeq(eq("0")):
			as.s = lipgloss.NewStyle()
			as.extra = as.extra[:0]
			as.hidden = false

		// setting non color modifiers
		case matchPartSeq(eq("1")):
			as.s = as.s.Bold(true)
		case matchPartSeq(eq("2")):
			as.s = as.s.Faint(true)
		case matchPartSeq(eq("3")):
			as.s = as.s.Italic(true)
		case matchPartSeq(eq("4")):
			as.s = as.s.Underline(true)
		case matchPartSeq(eq("5")):
			as.s = as.s.Blink(true)
		case matchPartSeq(eq("7")):
			as.s = as.s.Reverse(true)
		case matchPartSeq(eq("8")): // not supported by lipgloss/termenv
			as.hidden = true
		case matchPartSeq(eq("9")):
			as.s = as.s.Strikethrough(true)

		// resetting non color modifiers
		case matchPartSeq(eq("22")):
			as.s = as.s.UnsetBold()
			as.s = as.s.UnsetFaint()
		case matchPartSeq(eq("23")):
			as.s = as.s.UnsetItalic()
		case matchPartSeq(eq("24")):
			as.s = as.s.UnsetUnderline()
		case matchPartSeq(eq("25")):
			as.s = as.s.UnsetBlink()
		case matchPartSeq(eq("27")):
			as.s = as.s.UnsetReverse()
		case matchPartSeq(eq("28")): // not supported by lipgloss/termenv
			as.hidden = false
		case matchPartSeq(eq("29")):
			as.s = as.s.UnsetStrikethrough()
		}
	}
}

// render applies the current style state that was
// collected during cutting to the string ...
func (as *activeStyle) render(s string) string {
	var res strings.Builder
	if as.hidden {
		res.WriteString(wrapIntoCSISeq("8"))
	}
	for _, extra := range as.extra {
		res.WriteString(extra)
	}

	// The style application is done through lipgloss/termenv which
	// put's a reset at the end which we don't want since the string
	// might already have a reset somewhere in it ... we don't know
	// which is fine as we just made sure that we keep the correct
	// style after cutting the string. Remove the reset in case
	// render actually did something which might not be the case
	// when the style has not been altered
	rendered := as.s.Render(s)
	if len(s) != len(rendered) {
		rendered = strings.TrimSuffix(rendered, "\033[0m")
	}
	res.WriteString(rendered)
	return res.String()
}

// impl s[n:] ansi AND unicode aware
// we have to keep not resetted ansi codes
// ignores all non ansi color/graphics sequences
// regards ansi color and graphics mode (... elaborate)
func ansiStringSlice(s string, n int) string {
	if n <= 0 {
		return s
	}
	// we cannot exit out early because there could be
	// ansi sequences which deal with cursor movement etc.
	// if lipgloss.Width(s) < n { return "" }

	// we want to cut the string but keep the style up to
	// this point ...
	style := &activeStyle{}

	// all non graphic ansi sequences that we may read are
	// stored here so we can pass them on. If we don't read any
	// this is no overhead
	nonGraphicAnsiSeqences := strings.Builder{}

	currAnsiSeq := strings.Builder{}

	cells := []rune(s)
	for i := 0; i < len(cells); i++ {
		cell := cells[i]
		switch {
		// the ansi sequence is complete .. handle it based on its type
		case currAnsiSeq.Len() > 0 && ansi.IsTerminator(cell):
			currAnsiSeq.WriteRune(cell)

			ansiSeq := currAnsiSeq.String()
			// if it is a graphic control sequence we update
			// the style otherwise we MUST save it
			if isGraphicControlSequence(ansiSeq) {
				style.updateStyle(ansiSeq)
			} else {
				nonGraphicAnsiSeqences.WriteString(ansiSeq)
			}
			currAnsiSeq.Reset()

		// ansi sequence starts or we are within one
		case cell == ansi.Marker || currAnsiSeq.Len() > 0:
			currAnsiSeq.WriteRune(cell)

		// this is not a ansi sequence so we can count down
		// until we have found the point to cut at
		default:
			if n == 0 {
				return nonGraphicAnsiSeqences.String() + style.render(string(cells[i:]))
			}
			n--
		}
	}

	// s has less runes than we want to cut ...
	// let's at least return the ansi sequences which
	// have nothing to do with graphic modifications and put the
	// style back in place
	return nonGraphicAnsiSeqences.String() + style.render("")
}
