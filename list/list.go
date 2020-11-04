package list

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/termenv"
	"sort"
	"strings"
)

// Model is a bubbletea List of strings
type Model struct {
	focus bool

	listItems     []item
	curIndex      int                    // cursor
	visibleOffset int                    // begin of the visible lines
	less          func(k, l string) bool // function used for sorting

	CursorOffset int // offset or margin between the cursor and the viewport(visible) border

	Width  int
	Height int

	Wrap bool

	SelectedPrefix   string
	UnSelectedPrefix string
	Seperator        string
	SeperatorWrap    string
	CurrentMarker    string

	WrapPrefix bool

	Number         bool
	NumberRelative bool

	LineStyle     termenv.Style
	SelectedStyle termenv.Style
	CurrentStyle  termenv.Style
}

// Item are Items used in the list Model
// to hold the Content represented as a string
type item struct {
	selected     bool
	wrapedLines  []string
	wrapedLenght int
	wrapedto     int
	value        fmt.Stringer
}

// StringItem is just a convenience to have fast a version
// to satisfy the fmt.Stringer interface with plain strings
type StringItem string

func (s StringItem) String() string {
	return string(s)
}

// MakeStringerList is a shortcut to convert a string List to a List that satisfies the fmt.Stringer Interface
func MakeStringerList(list []string) []fmt.Stringer {
	stringerList := make([]fmt.Stringer, len(list))
	for i, item := range list {
		stringerList[i] = StringItem(item)
	}
	return stringerList
}

// genVisLines renews the wrap of the content into wrapedLines
func (i item) genVisLines(wrapTo int) item {
	i.wrapedLines = strings.Split(wordwrap.String(i.value.String(), wrapTo), "\n")
	//TODO hard wrap lines/words
	i.wrapedLenght = len(i.wrapedLines)
	i.wrapedto = wrapTo
	return i
}

// View renders the Lst to a (displayable) string
func (m *Model) View() string {
	return strings.Join(m.Lines(), "\n")
}

// Lines returns the Visible lines of the list items
// used to display the current user interface
func (m *Model) Lines() []string {
	// get public variables as locals so they can't change while using
	// check visible area
	height := m.Height
	width := m.Width
	if height*width <= 0 {
		panic("Can't display with zero width or hight of Viewport")
	}
	offset := m.visibleOffset

	// Get separators width
	widthItem := ansi.PrintableRuneWidth(m.Seperator)
	widthWrap := ansi.PrintableRuneWidth(m.SeperatorWrap)

	// Find max width
	sepWidth := widthItem
	if widthWrap > sepWidth {
		sepWidth = widthWrap
	}

	// get widest *displayed* number, for padding
	numWidth := len(fmt.Sprintf("%d", len(m.listItems)-1))
	localMaxWidth := len(fmt.Sprintf("%d", offset+height-1))
	if localMaxWidth < numWidth {
		numWidth = localMaxWidth
	}

	// pad all prefixes to the same width for easy exchange
	selected := m.SelectedPrefix
	unselect := m.UnSelectedPrefix
	selWid := ansi.PrintableRuneWidth(selected)
	tmpWid := ansi.PrintableRuneWidth(unselect)

	selectWidth := selWid
	if tmpWid > selectWidth {
		selectWidth = tmpWid
	}
	selected = strings.Repeat(" ", selectWidth-selWid) + selected

	wrapPrePad := unselect
	if !m.WrapPrefix {
		wrapPrePad = strings.Repeat(" ", selectWidth)
	}

	unselect = strings.Repeat(" ", selectWidth-tmpWid) + unselect

	// pad all separators to the same width for easy exchange
	sepItem := strings.Repeat(" ", sepWidth-widthItem) + m.Seperator
	sepWrap := strings.Repeat(" ", sepWidth-widthWrap) + m.SeperatorWrap

	// pad right of prefix, with length of current pointer
	mark := m.CurrentMarker
	markWidth := ansi.PrintableRuneWidth(mark)
	unmark := strings.Repeat(" ", markWidth)

	// Get the hole prefix width
	holePrefixWidth := numWidth + selectWidth + sepWidth + markWidth

	// Get actual content width
	contentWidth := width - holePrefixWidth

	// Check if there is space for the content left
	if contentWidth <= 0 {
		panic("Can't display with zero width for content")
	}

	// If set
	wrap := m.Wrap
	if wrap {
		// renew wrap of all items
		for i := range m.listItems {
			m.listItems[i] = m.listItems[i].genVisLines(contentWidth)
		}
	}

	var visLines int
	stringLines := make([]string, 0, height)
out:
	// Handle list items, start at first visible and go till end of list or visible (break)
	for index := offset; index < len(m.listItems); index++ {
		if index >= len(m.listItems) || index < 0 {
			// TODO log error
			break
		}

		item := m.listItems[index]
		if wrap && item.wrapedLenght <= 0 {
			panic("cant display item with no visible content")
		}

		// if a number is set, prepend first line with number and both with enough spaces
		firstPad := strings.Repeat(" ", numWidth)
		var wrapPad string
		if m.Number {
			lineNum := lineNumber(m.NumberRelative, m.curIndex, index)
			number := fmt.Sprintf("%d", lineNum)
			// since digits are only single bytes, len is sufficient:
			firstPad = strings.Repeat(" ", numWidth-len(number)) + number
			// pad wrapped lines
			wrapPad = strings.Repeat(" ", numWidth)
		}

		// Selecting: handle highlighting and prefixing of selected lines
		selString := unselect
		style := m.LineStyle

		if item.selected {
			style = m.SelectedStyle
			selString = selected
		}

		// Current: handle highlighting of current item/first-line
		curPad := unmark
		if index == m.curIndex {
			style = m.CurrentStyle
			curPad = mark
		}

		// join all prefixes
		var wrapPrefix, linePrefix string

		linePrefix = strings.Join([]string{firstPad, selString, sepItem, curPad}, "")
		if wrap {
			wrapPrefix = strings.Join([]string{wrapPad, wrapPrePad, sepWrap, unmark}, "") // don't prefix wrap lines with CurrentMarker (unmark)
		}

		content := item.wrapedLines[0]
		if !wrap {
			content = item.value.String()
			// TODO hard limit the string length
		}

		// join pad and first line content
		// NOTE line break is not added here because it would mess with the highlighting
		line := fmt.Sprintf("%s%s", linePrefix, content)

		// Highlight and write first line
		stringLines = append(stringLines, style.Styled(line))
		visLines++

		// Only write lines that are visible
		if visLines >= height {
			break out
		}

		// Don't write wrapped lines if not set
		if !wrap || item.wrapedLenght <= 1 {
			continue
		}

		// Write wrapped lines
		for _, line := range item.wrapedLines[1:] {
			// Pad left of line
			// NOTE line break is not added here because it would mess with the highlighting
			padLine := fmt.Sprintf("%s%s", wrapPrefix, line)

			// Highlight and write wrapped line
			stringLines = append(stringLines, style.Styled(padLine))
			visLines++

			// Only write lines that are visible
			if visLines > height {
				break out
			}
		}
	}
	return stringLines
}

// lineNumber returns line number of the given index
// and if relative is true the absolute difference to the cursor
func lineNumber(relativ bool, curser, current int) int {
	if !relativ || curser == current {
		return current
	}

	diff := curser - current
	if diff < 0 {
		diff *= -1
	}
	return diff
}

// Update changes the Model of the List according to the messages recieved
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.focus {
		return m, nil
	}
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Ctrl+c exits
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "down", "j":
			m.Move(1)
			return m, nil
		case "up", "k":
			m.Move(-1)
			return m, nil
		case " ":
			m.ToggleSelect(1)
			m.Down()
			return m, nil
		case "g":
			m.Top()
			return m, nil
		case "G":
			m.Bottom()
			return m, nil
		case "s":
			m.Sort()
			return m, nil
		case "+":
			m.MoveItem(-1)
			return m, nil
		case "-":
			m.MoveItem(1)
			return m, nil
		case "v": // inVert
			m.ToggleAllSelected()
			return m, nil
		case "m": // mark
			m.MarkSelected(1, true)
			return m, nil
		case "M": // mark False
			m.MarkSelected(1, false)
			return m, nil
		}

	case tea.WindowSizeMsg:

		m.Width = msg.Width
		m.Height = msg.Height

		return m, cmd

	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseWheelUp:
			m.Up()

		case tea.MouseWheelDown:
			m.Down()
		}
	}
	return m, nil
}

// Init does nothing
func (m Model) Init() tea.Cmd {
	return nil
}

// AddItems adds the given Items to the list Model
// Without performing updating the View TODO
func (m *Model) AddItems(itemList []fmt.Stringer) {
	for _, i := range itemList {
		m.listItems = append(m.listItems, item{
			selected: false,
			value:    i},
		)
	}
}

// Down moves the "cursor" or current line down.
// If the end is already reached err is not nil.
func (m *Model) Down() error {
	return m.Move(1)
}

// Up moves the "cursor" or current line up.
// If the start is already reached, err is not nil.
func (m *Model) Up() error {
	return m.Move(-1)
}

// Move moves the cursor by amount, does nothing if amount is 0
// and returns error != nil if amount go's beyond list borders
// or if the CursorOffset is greater than half of the display height
func (m *Model) Move(amount int) error {
	// do nothing
	if amount == 0 {
		return nil
	}
	var err error
	curOff := m.CursorOffset
	visOff := m.visibleOffset
	height := m.Height
	if curOff >= height/2 {
		curOff = 0
		err = fmt.Errorf("cursor offset must be less than halfe of the display height: setting it to zero")
	}

	target := m.curIndex + amount
	if !m.CheckWithinBorder(target) {
		return fmt.Errorf("Cant move outside the list: %d", target)
	}
	// move visible part of list if Cursor is going beyond border.
	lowerBorder := height + visOff - curOff
	upperBorder := visOff + curOff

	direction := 1
	if amount < 0 {
		direction = -1
	}

	// visible Down movement
	if direction > 0 && target > lowerBorder {
		visOff = target - (height - curOff)
	}
	// visible Up movement
	if direction < 0 && target < upperBorder {
		visOff = target - curOff
	}
	// don't go in front of list begin
	if visOff < 0 {
		visOff = 0
	}
	m.curIndex = target
	m.visibleOffset = visOff
	return err
}

// NewModel returns a Model with some save/sane defaults
func NewModel() Model {
	p := termenv.ColorProfile()
	selStyle := termenv.Style{}.Background(p.Color("#ff0000"))
	// just reverse colors to keep there information
	curStyle := termenv.Style{}.Reverse()
	return Model{
		// Accept key presses
		focus: true,

		// Try to keep $CursorOffset lines between Cursor and screen Border
		CursorOffset: 5,

		// Wrap lines to have no loss of information
		Wrap: true,

		// Make clear where a item begins and where it ends
		Seperator:     "╭",
		SeperatorWrap: "│",

		// Mark it so that even without color support all is explicit
		CurrentMarker:  ">",
		SelectedPrefix: "*",

		// enable Linenumber
		Number: true,

		less: func(k, l string) bool {
			return k < l
		},

		SelectedStyle: selStyle,
		CurrentStyle:  curStyle,
	}
}

// ToggleSelect toggles the selected status
// of the current Index if amount is 0
// returns err != nil when amount lands outside list and safely does nothing
// else if amount is not 0 toggles selected amount items
// excluding the item on which the cursor lands
func (m *Model) ToggleSelect(amount int) error {
	if amount == 0 {
		m.listItems[m.curIndex].selected = !m.listItems[m.curIndex].selected
	}

	direction := 1
	if amount < 0 {
		direction = -1
	}

	cur := m.curIndex
	target := cur + amount - direction
	if !m.CheckWithinBorder(target) {
		return fmt.Errorf("Cant go beyond list borders: %d", target)
	}
	for c := 0; c < amount*direction; c++ {
		m.listItems[cur+c].selected = !m.listItems[cur+c].selected
	}
	m.curIndex = target - direction
	m.Move(direction)
	return nil
}

// MarkSelected selects or unselects depending on 'mark'
// amount = 0 changes the current item but does not move the cursor
// if amount would be outside the list error is not nil
// else all items till but excluding the end cursor position
func (m *Model) MarkSelected(amount int, mark bool) error {
	cur := m.curIndex
	if amount == 0 {
		m.listItems[cur].selected = mark
		return nil
	}
	direction := 1
	if amount < 0 {
		direction = -1
	}

	target := cur + amount - direction
	if !m.CheckWithinBorder(target) {
		return fmt.Errorf("Cant go beyond list borders: %d", target)
	}
	for c := 0; c < amount*direction; c++ {
		m.listItems[cur+c].selected = mark
	}
	m.curIndex = target
	m.Move(direction)
	return nil
}

// ToggleAllSelected inverts the select state of ALL items
func (m *Model) ToggleAllSelected() {
	for i := range m.listItems {
		m.listItems[i].selected = !m.listItems[i].selected
	}
}

// Top moves the cursor to the first line
func (m *Model) Top() {
	m.visibleOffset = 0
	m.curIndex = 0
}

// Bottom moves the cursor to the last line
func (m *Model) Bottom() {
	end := len(m.listItems) - 1
	m.curIndex = end
	maxVisItems := m.Height - m.CursorOffset
	var visLines, smallestVisIndex int
	for c := end; visLines < maxVisItems; c-- {
		if c < 0 {
			break
		}
		visLines += m.listItems[c].wrapedLenght
		smallestVisIndex = c
	}
	m.visibleOffset = smallestVisIndex
}

// GetSelected returns you a list of all items
// that are selected in current (displayed) order
func (m *Model) GetSelected() []fmt.Stringer {
	var selected []fmt.Stringer
	for _, item := range m.listItems {
		if item.selected {
			selected = append(selected, item.value)
		}
	}
	return selected
}

// Less is a Proxy to the less function, set from the user.
// Swap is used to fulfill the Sort-interface
func (m *Model) Less(i, j int) bool {
	return m.less(m.listItems[i].value.String(), m.listItems[j].value.String())
}

// Swap is used to fulfill the Sort-interface
func (m *Model) Swap(i, j int) {
	m.listItems[i], m.listItems[j] = m.listItems[j], m.listItems[i]
}

// Len is used to fulfill the Sort-interface
func (m *Model) Len() int {
	return len(m.listItems)
}

// SetLess sets the internal less function used for sorting the list items
func (m *Model) SetLess(less func(string, string) bool) {
	m.less = less
}

// Sort sorts the listitems according to the set less function
// The current Item will maybe change!
// Since the index of the current pointer does not change
func (m *Model) Sort() {
	sort.Sort(m)
}

// MoveItem moves the current item by amount to the end
// So: MoveItem(1) Moves the Item towards the end by one
// and MoveItem(-1) Moves the Item towards the beginning
// MoveItem(0) safely does nothing
// and a amount that would result outside the list returns a error != nil
func (m *Model) MoveItem(amount int) error {
	if amount == 0 {
		return nil
	}
	cur := m.curIndex
	target := cur + amount
	if !m.CheckWithinBorder(target) {
		return fmt.Errorf("Cant move outside the list: %d", target)
	}
	m.Swap(cur, target)
	m.curIndex = target
	return nil
}

// CheckWithinBorder returns true if the give index is within the list borders
func (m *Model) CheckWithinBorder(index int) bool {
	length := len(m.listItems)
	if index >= length || index < 0 {
		return false
	}
	return true
}

// AddDataItem adds a Item with the given interface{} value added to the List item
// So that when sorting, the connection between the string and the interfave{} value stays.
//func (m *Model) AddDataItem(content string, data interface{}) {
//	m.listItems = append(m.listItems, item{content: content, userValue: data})
//}

// Focus sets the list Model focus so it accepts key input and responds to them
func (m *Model) Focus() {
	m.focus = true
}

// UnFocus removes the focus so that the list Model does NOT respond to key presses
func (m *Model) UnFocus() {
	m.focus = false
}

// Focused returns if the list Model is focused and accepts key presses
func (m *Model) Focused() bool {
	return m.focus
}
