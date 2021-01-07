package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/muesli/reflow/ansi"
	"strings"
)

// SelectPrefixer is the default struct used for Prefixing a line
type SelectPrefixer struct {
	PrefixWrap bool

	// Make clear where a item begins and where it ends
	Seperator     string
	SeperatorWrap string

	// Mark it so that even without color support all is explicit
	CurrentMarker string

	// Mark if item is selected or not
	Selected string
	UnSelect string

	// enable Linenumber
	Number         bool
	NumberRelative bool

	prefixWidth int
	viewPos     list.ViewPos

	markWidth int
	numWidth  int

	unmark string
	mark   string

	sepItem string
	sepWrap string

	selecStr string
	unselStr string

	selWidth int
	unselWid int

	currentIndex int
	model        list.Model
	value        item
}

// NewPrefixer returns a DefautPrefixer with default values
func NewPrefixer() *SelectPrefixer {
	return &SelectPrefixer{
		PrefixWrap: false,

		// Make clear where a item begins and where it ends
		Seperator:     "•",
		SeperatorWrap: " ",

		// Mark it so that even without color support all is explicit
		CurrentMarker: ">",

		Selected: "[✓]",
		UnSelect: "[ ]",

		// enable Linenumber
		Number:         true,
		NumberRelative: false,
	}
}

// InitPrefixer sets up all strings used to prefix a given line later by Prefix()
func (s *SelectPrefixer) InitPrefixer(value fmt.Stringer, currentItemIndex int, position list.ViewPos, screen list.ScreenInfo) int {
	// TODO adapt to per item call
	n, ok := value.(item)
	if ok {
		s.value = n
	}
	s.currentIndex = currentItemIndex
	s.viewPos = position

	offset := position.Cursor - position.LineOffset
	if offset < 0 {
		offset = 0
	}

	// get widest possible number, for padding
	// TODO handle wrap, cause only correct when wrap off:
	s.numWidth = len(fmt.Sprintf("%d", offset+screen.Height))

	seWidth := ansi.PrintableRuneWidth(s.Selected)
	unWidth := ansi.PrintableRuneWidth(s.UnSelect)
	s.selWidth = seWidth
	if unWidth > s.selWidth {
		s.selWidth = unWidth
	}
	// pad the selectStrings incase they have different lenght
	s.selecStr = s.Selected + strings.Repeat(" ", s.selWidth-seWidth)
	s.unselStr = s.UnSelect + strings.Repeat(" ", s.selWidth-unWidth)

	// Get separators width
	widthItem := ansi.PrintableRuneWidth(s.Seperator)
	widthWrap := ansi.PrintableRuneWidth(s.SeperatorWrap)
	// Find max width
	sepWidth := widthItem
	if widthWrap > sepWidth {
		sepWidth = widthWrap
	}
	// pad all separators to the same width for easy exchange
	s.sepItem = strings.Repeat(" ", sepWidth-widthItem) + s.Seperator
	s.sepWrap = strings.Repeat(" ", sepWidth-widthWrap) + s.SeperatorWrap

	// pad right of prefix, with length of current pointer
	s.mark = s.CurrentMarker
	s.markWidth = ansi.PrintableRuneWidth(s.mark)
	s.unmark = strings.Repeat(" ", s.markWidth)

	// Get the hole prefix width
	s.prefixWidth = s.numWidth + s.selWidth + sepWidth + s.markWidth

	return s.prefixWidth
}

// Prefix prefixes a given line
func (s *SelectPrefixer) Prefix(lineIndex int) string {
	// if a number is set, prepend first line with number and both with enough spaces
	firstPad := strings.Repeat(" ", s.numWidth)
	var wrapPad string
	var lineNum int
	if s.Number {
		lineNum = lineNumber(s.NumberRelative, s.viewPos.Cursor, s.currentIndex)
	}
	number := fmt.Sprintf("%d", lineNum)
	// since digits are only single bytes, len is sufficient:
	padTo := s.numWidth - len(number)
	if padTo < 0 {
		// TODO log error
		padTo = 0
	}
	firstPad = strings.Repeat(" ", padTo) + number
	// pad wrapped lines
	wrapPad = strings.Repeat(" ", s.numWidth)

	// add un/selected string
	selPad := s.unselStr

	if s.value.selected {
		selPad = s.selecStr
	}

	// Current: handle marking of current item/first-line
	curPad := s.unmark
	if s.currentIndex == s.viewPos.Cursor {
		curPad = s.mark
	}

	// join all prefixes
	linePrefix := strings.Join([]string{firstPad, selPad, s.sepItem, curPad}, "")
	if lineIndex > 0 && !s.PrefixWrap {
		linePrefix = strings.Join([]string{wrapPad, strings.Repeat(" ", s.selWidth), s.sepWrap, s.unmark}, "") // don't prefix wrap lines with CurrentMarker (unmark)
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
