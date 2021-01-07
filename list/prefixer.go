package list

import (
	"fmt"
	"github.com/muesli/reflow/ansi"
	"strings"
)

// Prefixer is used to prefix all visible Lines.
// Init gets called ones on the beginning of the Lines methode
// and then Prefix ones, per line to draw, to generate according prefixes.
type Prefixer interface {
	InitPrefixer(currentItem fmt.Stringer, currentItemIndex int, viewPos ViewPos, screenInfo ScreenInfo) int
	Prefix(currentLine, allLines int) string
}

// DefaultPrefixer is the default struct used for Prefixing a line
type DefaultPrefixer struct {
	PrefixWrap bool

	// Make clear where a item begins and where it ends
	FirstSep      string
	Seperator     string
	SeperatorWrap string

	// Mark it so that even without color support all is explicit
	CurrentMarker string

	// enable Linenumber
	Number         bool
	NumberRelative bool

	prefixWidth int
	viewPos     ViewPos

	markWidth int
	numWidth  int

	unmark string
	mark   string

	sepItem string
	sepWrap string

	currentIndex int
}

// NewPrefixer returns a DefautPrefixer with default values
func NewPrefixer() *DefaultPrefixer {
	return &DefaultPrefixer{
		PrefixWrap: true,

		// Make clear where a item begins and where it ends
		FirstSep:      "╭",
		Seperator:     "├",
		SeperatorWrap: "│",

		// Mark it so that even without color support all is explicit
		CurrentMarker: ">",

		// enable Linenumber
		Number:         true,
		NumberRelative: false,
	}
}

// InitPrefixer sets up all strings used to prefix a given line later by Prefix()
func (d *DefaultPrefixer) InitPrefixer(value fmt.Stringer, currentItemIndex int, position ViewPos, screen ScreenInfo) int {
	// TODO adapt to per item call
	d.currentIndex = currentItemIndex
	d.viewPos = position

	offset := position.Cursor - position.LineOffset
	if offset < 0 {
		offset = 0
	}
	seperator := d.Seperator
	if currentItemIndex == 0 {
		seperator = d.FirstSep
	}

	// Get separators width
	widthItem := ansi.PrintableRuneWidth(seperator)
	widthWrap := ansi.PrintableRuneWidth(d.SeperatorWrap)

	// Find max width
	sepWidth := widthItem
	if widthWrap > sepWidth {
		sepWidth = widthWrap
	}

	// get widest possible number, for padding
	// TODO handle wrap, cause only correct when wrap off:
	d.numWidth = len(fmt.Sprintf("%d", offset+screen.Height))

	// pad all prefixes to the same width for easy exchange
	// pad all separators to the same width for easy exchange
	d.sepItem = strings.Repeat(" ", sepWidth-widthItem) + seperator
	d.sepWrap = strings.Repeat(" ", sepWidth-widthWrap) + d.SeperatorWrap

	// pad right of prefix, with length of current pointer
	d.mark = d.CurrentMarker
	d.markWidth = ansi.PrintableRuneWidth(d.mark)
	d.unmark = strings.Repeat(" ", d.markWidth)

	// Get the hole prefix width
	d.prefixWidth = d.numWidth + sepWidth + d.markWidth

	return d.prefixWidth
}

// Prefix prefixes a given line
func (d *DefaultPrefixer) Prefix(lineIndex, allLines int) string {
	// if a number is set, prepend first line with number and both with enough spaces
	firstPad := strings.Repeat(" ", d.numWidth)
	var wrapPad string
	var lineNum int
	if d.Number {
		lineNum = lineNumber(d.NumberRelative, d.viewPos.Cursor, d.currentIndex)
	}
	number := fmt.Sprintf("%d", lineNum)
	// since digits are only single bytes, len is sufficient:
	padTo := d.numWidth - len(number)
	if padTo < 0 {
		// TODO log error
		padTo = 0
	}
	firstPad = strings.Repeat(" ", padTo) + number
	// pad wrapped lines
	wrapPad = strings.Repeat(" ", d.numWidth)

	// Current: handle highlighting of current item/first-line
	curPad := d.unmark
	if d.currentIndex == d.viewPos.Cursor {
		curPad = d.mark
	}

	// join all prefixes
	linePrefix := strings.Join([]string{firstPad, d.sepItem, curPad}, "")
	if lineIndex > 0 {
		linePrefix = strings.Join([]string{wrapPad, d.sepWrap, d.unmark}, "") // don't prefix wrap lines with CurrentMarker (unmark)
	}

	return linePrefix
}

// lineNumber returns line number of the given index
// and if relative is true the absolute difference to the cursor
// or if on the cursor the absolute line number
func lineNumber(relativ bool, curser, current int) int {
	if !relativ || curser == current {
		return current + 1
	}

	diff := curser - current
	if diff < 0 {
		diff *= -1
	}
	return diff
}
