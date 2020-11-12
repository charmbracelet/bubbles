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
	InitPrefixer(ViewPos, ScreenInfo) int
	Prefix(currentItem, currentLine int, selected bool) string
}

// DefaultPrefixer is the default struct used for Prefixing a line
type DefaultPrefixer struct {
	PrefixWrap bool

	// Make clear where a item begins and where it ends
	Seperator     string
	SeperatorWrap string

	// Mark it so that even without color support all is explicit
	CurrentMarker  string
	SelectedPrefix string

	// enable Linenumber
	Number         bool
	NumberRelative bool

	UnSelectedPrefix string

	prefixWidth int
	viewPos     ViewPos

	markWidth int
	numWidth  int

	unmark string
	mark   string

	selectedString string
	unselect       string

	wrapSelectPad string
	wrapUnSelePad string

	sepItem string
	sepWrap string
}

// NewPrefixer returns a DefautPrefixer with default values
func NewPrefixer() *DefaultPrefixer {
	return &DefaultPrefixer{
		PrefixWrap: false,

		// Make clear where a item begins and where it ends
		Seperator:     "╭",
		SeperatorWrap: "│",

		// Mark it so that even without color support all is explicit
		CurrentMarker:    ">",
		SelectedPrefix:   "*",
		UnSelectedPrefix: "",

		// enable Linenumber
		Number:         true,
		NumberRelative: false,
	}
}

// InitPrefixer sets up all strings used to prefix a given line later by Prefix()
func (d *DefaultPrefixer) InitPrefixer(position ViewPos, screen ScreenInfo) int {
	d.viewPos = position

	offset := position.ItemOffset

	// Get separators width
	widthItem := ansi.PrintableRuneWidth(d.Seperator)
	widthWrap := ansi.PrintableRuneWidth(d.SeperatorWrap)

	// Find max width
	sepWidth := widthItem
	if widthWrap > sepWidth {
		sepWidth = widthWrap
	}

	// get widest possible number, for padding
	d.numWidth = len(fmt.Sprintf("%d", offset+screen.Height))

	// pad all prefixes to the same width for easy exchange
	d.selectedString = d.SelectedPrefix
	d.unselect = d.UnSelectedPrefix
	selWid := ansi.PrintableRuneWidth(d.selectedString)
	tmpWid := ansi.PrintableRuneWidth(d.unselect)

	selectWidth := selWid
	if tmpWid > selectWidth {
		selectWidth = tmpWid
	}
	d.selectedString = strings.Repeat(" ", selectWidth-selWid) + d.selectedString

	d.wrapSelectPad = strings.Repeat(" ", selectWidth)
	d.wrapUnSelePad = strings.Repeat(" ", selectWidth)
	if d.PrefixWrap {
		d.wrapSelectPad = strings.Repeat(" ", selectWidth-selWid) + d.selectedString
		d.wrapUnSelePad = strings.Repeat(" ", selectWidth-tmpWid) + d.unselect
	}

	d.unselect = strings.Repeat(" ", selectWidth-tmpWid) + d.unselect

	// pad all separators to the same width for easy exchange
	d.sepItem = strings.Repeat(" ", sepWidth-widthItem) + d.Seperator
	d.sepWrap = strings.Repeat(" ", sepWidth-widthWrap) + d.SeperatorWrap

	// pad right of prefix, with length of current pointer
	d.mark = d.CurrentMarker
	d.markWidth = ansi.PrintableRuneWidth(d.mark)
	d.unmark = strings.Repeat(" ", d.markWidth)

	// Get the hole prefix width
	d.prefixWidth = d.numWidth + selectWidth + sepWidth + d.markWidth

	return d.prefixWidth
}

// Prefix prefixes a given line
func (d *DefaultPrefixer) Prefix(currentIndex int, wrapIndex int, selected bool) string {
	// if a number is set, prepend first line with number and both with enough spaces
	firstPad := strings.Repeat(" ", d.numWidth)
	var wrapPad string
	var lineNum int
	if d.Number {
		lineNum = lineNumber(d.NumberRelative, d.viewPos.Cursor, currentIndex)
	}
	number := fmt.Sprintf("%d", lineNum)
	// since digits are only single bytes, len is sufficient:
	firstPad = strings.Repeat(" ", d.numWidth-len(number)) + number
	// pad wrapped lines
	wrapPad = strings.Repeat(" ", d.numWidth)
	// Selecting: handle highlighting and prefixing of selected lines
	selString := d.unselect

	wrapPrePad := d.wrapUnSelePad
	if selected {
		selString = d.selectedString
		wrapPrePad = d.wrapSelectPad
	}

	// Current: handle highlighting of current item/first-line
	curPad := d.unmark
	if currentIndex == d.viewPos.Cursor {
		curPad = d.mark
	}

	// join all prefixes
	var wrapPrefix, linePrefix string

	linePrefix = strings.Join([]string{firstPad, selString, d.sepItem, curPad}, "")
	if wrapIndex > 0 {
		wrapPrefix = strings.Join([]string{wrapPad, wrapPrePad, d.sepWrap, d.unmark}, "") // don't prefix wrap lines with CurrentMarker (unmark)
		return wrapPrefix
	}

	return linePrefix
}
