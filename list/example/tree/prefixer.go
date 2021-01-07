package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/termenv"
	"strings"
)

// TreePrefixer is the default struct used for Prefixing a line
type TreePrefixer struct {
	PrefixWrap bool

	// Make clear where a item begins and where it ends
	Seperator     string
	SeperatorWrap string

	// Mark it so that even without color support all is explicit
	CurrentMarker string

	// enable Linenumber
	Number         bool
	NumberRelative bool

	prefixWidth int
	viewPos     list.ViewPos

	sepWidth  int
	markWidth int
	numWidth  int

	currentIndex int

	level       int
	LevelPadder func(int) string
}

// NewPrefixer returns a DefautPrefixer with default values
func NewPrefixer() *TreePrefixer {
	return &TreePrefixer{
		PrefixWrap: true,

		// Make clear where a item begins and where it ends
		Seperator:     "╭",
		SeperatorWrap: "│",

		// Mark it so that even without color support all is explicit
		CurrentMarker: ">",

		// enable Linenumber
		Number:         true,
		NumberRelative: false,
		LevelPadder:    padLevel,
	}
}

// InitPrefixer sets up all strings used to prefix a given line later by Prefix()
func (d *TreePrefixer) InitPrefixer(value fmt.Stringer, currentItemIndex int, position list.ViewPos, screen list.ScreenInfo) int {
	d.currentIndex = currentItemIndex
	d.viewPos = position

	offset := position.Cursor - position.LineOffset
	if offset < 0 {
		offset = 0
	}

	// Get max separators width
	d.sepWidth = ansi.PrintableRuneWidth(d.Seperator)
	if widthWrap := ansi.PrintableRuneWidth(d.SeperatorWrap); widthWrap > d.sepWidth {
		d.sepWidth = widthWrap
	}

	// get widest possible number, for padding
	// TODO handle wrap, cause only correct when wrap off:
	d.numWidth = len(fmt.Sprintf("%d", offset+screen.Height))

	// pad right of prefix, with length of current pointer

	d.markWidth = ansi.PrintableRuneWidth(d.CurrentMarker)

	n, ok := value.(node)
	if ok {
		d.level = len(n.parentIDs) - 1
	}

	// Get the hole prefix width
	d.prefixWidth = d.numWidth + d.sepWidth + d.markWidth

	return d.prefixWidth
}

// Prefix prefixes a given line
func (d *TreePrefixer) Prefix(lineIndex, allLines int) string {
	// pad all separators to the same width for easy exchange
	sepItem := strings.Repeat(" ", d.sepWidth-ansi.PrintableRuneWidth(d.Seperator)) + d.Seperator
	sepWrap := strings.Repeat(" ", d.sepWidth-ansi.PrintableRuneWidth(d.SeperatorWrap)) + d.SeperatorWrap

	// if a number is set, prepend first line with number and both with enough spaces
	firstPad := strings.Repeat(" ", d.numWidth)
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
	wrapPad := strings.Repeat(" ", d.numWidth)

	// Current: handle highlighting of current item/first-line
	curPad := strings.Repeat(" ", d.markWidth)
	if d.currentIndex == d.viewPos.Cursor {
		curPad = d.CurrentMarker
	}

	// join all prefixes
	linePrefix := strings.Join([]string{firstPad, sepItem, curPad}, "")
	if lineIndex > 0 {
		linePrefix = strings.Join([]string{wrapPad, sepWrap, strings.Repeat(" ", ansi.PrintableRuneWidth(curPad))}, "") // don't prefix wrap lines with CurrentMarker (unmark)
	}
	if d.LevelPadder != nil {
		linePrefix += d.LevelPadder(d.level)
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

func padLevel(level int) string {
	if level > 0 {
		color := termenv.ColorProfile().Color("#0000ff")
		sty := termenv.Style{}
		sty = sty.Foreground(color)
		sty = sty.Background(color)
		return fmt.Sprintf("%s %s", strings.Repeat("  ", level-1), sty.Styled(" "))
	}
	return ""
}
