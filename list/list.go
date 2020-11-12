package list

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jinzhu/copier"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/termenv"
	"sort"
	"strings"
)

// Model is a bubbletea List of strings
type Model struct {
	focus bool

	listItems []item

	less   func(string, string) bool             // function used for sorting
	equals func(fmt.Stringer, fmt.Stringer) bool // used after sorting, to be set from the user

	CursorOffset int // offset or margin between the cursor and the viewport(visible) border

	Screen  ScreenInfo
	viewPos ViewPos

	Wrap bool

	PrefixGen Prefixer
	SuffixGen Suffixer

	LineStyle     termenv.Style
	SelectedStyle termenv.Style
	CurrentStyle  termenv.Style
}

// NewModel returns a Model with some save/sane defaults
// design to transfer as much internal information to the user
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

		less: func(k, l string) bool {
			return k < l
		},

		SelectedStyle: selStyle,
		CurrentStyle:  curStyle,
	}
}

// Init does nothing
func (m Model) Init() tea.Cmd {
	return nil
}

// View renders the List to a (displayable) string
func (m Model) View() string {
	return strings.Join(m.Lines(), "\n")
}

// Lines returns the Visible lines of the list items
// used to display the current user interface
func (m *Model) Lines() []string {
	// get public variables as locals so they can't change while using

	// check visible area
	height := m.Screen.Height
	width := m.Screen.Width
	if height*width <= 0 {
		panic("Can't display with zero width or hight of Viewport")
	}

	// Get the Width of each suf/prefix
	var prefixWidth, suffixWidth int
	if m.PrefixGen != nil {
		prefixWidth = m.PrefixGen.InitPrefixer(m.viewPos, m.Screen)
	}
	if m.SuffixGen != nil {
		suffixWidth = m.SuffixGen.InitSuffixer(m.viewPos, m.Screen)
	}

	// Get actual content width
	contentWidth := width - prefixWidth - suffixWidth

	// Check if there is space for the content left
	if contentWidth <= 0 {
		panic("Can't display with zero width for content")
	}

	lineOffset := m.viewPos.LineOffset
	offset := m.viewPos.ItemOffset

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

		lines := m.itemLines(item)

		var ignoreLines bool
		if len(lines) > 1 && lineOffset > 0 && index == offset {
			ignoreLines = true
		}

		// Write lines
		for i, line := range lines {
			// skip unvisible leading lines
			if ignoreLines && lineOffset > 0 {
				lineOffset--
				continue
			}

			// Surrounding content
			var linePrefix, lineSuffix string
			if m.PrefixGen != nil {
				linePrefix = m.PrefixGen.Prefix(index, i, item.selected)
			}
			if m.SuffixGen != nil {
				lineSuffix = fmt.Sprintf("%s%s", strings.Repeat(" ", contentWidth-ansi.PrintableRuneWidth(line)), m.SuffixGen.Suffix(index, i, item.selected))
			}

			// Join all
			line := fmt.Sprintf("%s%s%s", linePrefix, line, lineSuffix)

			// Highlighting of selected and current lines
			style := m.LineStyle
			if item.selected {
				style = m.SelectedStyle
			}
			if index == m.viewPos.Cursor {
				style = m.CurrentStyle
			}

			// Highlight and write wrapped line
			stringLines = append(stringLines, style.Styled(line))
			visLines++

			// Only write lines that are visible
			if visLines >= height {
				break out
			}
		}
	}
	lenght := len(stringLines)
	if lenght > m.Screen.Height {
		panic(fmt.Sprintf("can't display %d lines when screen has %d lines.", lenght, m.Screen.Height))
	}
	return stringLines
}

// Update changes the Model of the List according to the messages received
// if the list is focused, else does nothing.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.focus {
		return m, nil
	}

	if m.PrefixGen == nil {
		// use default
		m.PrefixGen = NewPrefixer()
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
			m.Move(1)
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

		m.Screen.Width = msg.Width
		m.Screen.Height = msg.Height
		m.Screen.Profile = termenv.ColorProfile()

		return m, cmd

	case tea.MouseMsg:
		switch msg.Type {
		case tea.MouseWheelUp:
			m.Move(-1)

		case tea.MouseWheelDown:
			m.Move(1)
		}
	}
	return m, nil
}

// NotFound gets return if the search does not yield a result
type NotFound error

// OutOfBounds is return if and index is outside the list bounderys
type OutOfBounds error

// MultipleMatches gets return if the search yield more result
type MultipleMatches error

// ConfigError is return if there is a error with the configuration of the list Modul
type ConfigError error

// NotFocused is a error return if the action can only be applied to a focused list
type NotFocused error

// ViewPos is used for holding the information about the View parameters
type ViewPos struct {
	Cursor     int
	ItemOffset int
	LineOffset int
}

// ScreenInfo holds all information about the screen Area
type ScreenInfo struct {
	Width   int
	Height  int
	Profile termenv.Profile
}

// Move moves the cursor by amount and returns OutOfBounds error if amount go's beyond list borders
// or if the CursorOffset is greater than half of the display height returns ConfigError
// if amount is 0 the Curser will get set within the view bounds
func (m *Model) Move(amount int) (int, error) {
	target := m.viewPos.Cursor + amount
	newPos, err := m.KeepVisible(target)
	m.viewPos = newPos
	return newPos.Cursor, err
}

// Top moves the cursor to the first line
func (m *Model) Top() {
	m.viewPos.Cursor = 0
	m.viewPos.ItemOffset = 0
	m.viewPos.LineOffset = 0
}

// Bottom moves the cursor to the last line
func (m *Model) Bottom() {
	end := len(m.listItems) - 1
	m.Move(end)
}

// KeepVisible will set the Cursor within the visible area of the list
// and if CursorOffset is != 0 will set it within this bounderys
// if CursorOffset is bigger than half the screen hight error will be of type ConfigError
// If the cursor would be outside of the list, it will be set to the according nearest value
// and error will be of type OutOfBounds. The return int is the absolut item number on which the cursor gets set
func (m *Model) KeepVisible(target int) (ViewPos, error) {
	var err error
	// Check if Cursor would be beyond list
	if length := len(m.listItems); target >= length {
		target = length - 1
		errMsg := "requested cursor position was behind of the list"
		err = OutOfBounds(fmt.Errorf(errMsg))
	}

	// Check if Cursor would be infront of list
	if target < 0 {
		target = 0
		errMsg := "requested cursor position was infront of the list"
		err = OutOfBounds(fmt.Errorf(errMsg))
	}

	if target == 0 {
		return ViewPos{}, nil
	}

	if m.Wrap {
		return m.keepVisibleWrap(target)
	}

	m.viewPos.LineOffset = 0

	visItemsBeforCursor := target - m.viewPos.ItemOffset

	// Visible Area and Cursor are at beginning of List -> cant move further up.
	if m.viewPos.ItemOffset <= 0 && visItemsBeforCursor <= m.CursorOffset {
		return ViewPos{Cursor: target}, err
	}

	// Cursor is infront of Boundry -> move visible Area up
	if visItemsBeforCursor < m.CursorOffset {
		return ViewPos{Cursor: target, ItemOffset: target - m.CursorOffset}, err
	}

	// Cursor Position is within bounds -> all good
	if visItemsBeforCursor >= m.CursorOffset && visItemsBeforCursor < m.Screen.Height-m.CursorOffset {
		return ViewPos{Cursor: target, ItemOffset: m.viewPos.ItemOffset}, err
	}

	// Cursor is beyond boundry -> move visibel Area down
	lowerOffset := m.viewPos.ItemOffset - (m.Screen.Height - m.CursorOffset - visItemsBeforCursor - 1)
	return ViewPos{Cursor: target, ItemOffset: lowerOffset}, err
}

func (m *Model) keepVisibleWrap(target int) (ViewPos, error) {

	if !m.CheckWithinBorder(target) {
		return ViewPos{}, OutOfBounds(fmt.Errorf("can't move beyond list bonderys, with requested cursor position: %d", target))
	}

	if target == 0 {
		return ViewPos{}, nil
	}

	direction := 1
	diff := target - m.viewPos.Cursor
	if diff < 0 {
		direction = -1
	}

	type beforCursor struct {
		listIndex  int
		linesBefor int
	}

	lineCount := make([]beforCursor, 0, m.Screen.Height)

	var lineSum int
	if direction >= 0 {
		lineSum = 1 // Cursorline is not counted in the following loop, so do it here
	}

	var lower, upper bool // Visible lower/upper
	upperBorder := m.CursorOffset
	lowerBorder := m.Screen.Height - m.CursorOffset
	// calculate how much space/lines the items befor the requested cursor position occupy
	for c := target - 1; c >= 0 && c > target-m.Screen.Height; c-- {
		lineSum += len(m.itemLines(m.listItems[c]))
		lineCount = append(lineCount, beforCursor{c, lineSum})

		// if new target infront of old visible offset dont mark borders
		// TODO here is a bug: when there is a list item with more than Screen.Height-m.CursorOffset lines
		// the up movement below this item will move to the wrong position, no solution yet
		if target-1 < m.viewPos.ItemOffset+m.CursorOffset {
			continue
		}

		// mark the pass of a border
		if !upper && lineSum > upperBorder {
			upper = true
		}
		if !lower && lineSum >= lowerBorder && c >= m.viewPos.ItemOffset {
			lower = true
		}
	}

	// Can't Move visible infront of list begin
	if direction < 0 && len(lineCount) > 0 && // possible upwards movement
		lineCount[len(lineCount)-1].linesBefor < m.CursorOffset && // beyond upper border
		m.viewPos.ItemOffset <= 0 && m.viewPos.LineOffset <= 0 { // but allready at beginning of list

		return ViewPos{Cursor: target}, nil
	}

	var lastOffset, lineOffset int
	for _, count := range lineCount {
		lastOffset = count.listIndex // Visible Offset
		// can't Move beyond list end, setting offsets accordingly
		if target >= len(m.listItems)-1 && count.linesBefor > lowerBorder {
			lineOffset = count.linesBefor - lowerBorder
			return ViewPos{ItemOffset: lastOffset, LineOffset: lineOffset, Cursor: len(m.listItems) - 1}, nil
		}
		// infront upper border -> Move up
		if direction < 0 && !upper && count.linesBefor > upperBorder {
			lineOffset = count.linesBefor - upperBorder
			return ViewPos{ItemOffset: lastOffset, LineOffset: lineOffset, Cursor: target}, nil
		}
		// beyond lower border -> Moving Down
		if direction >= 0 && lower && count.linesBefor >= lowerBorder {
			lastOffset = count.listIndex // Visible Offset
			lineOffset = count.linesBefor - lowerBorder
			return ViewPos{ItemOffset: lastOffset, LineOffset: lineOffset, Cursor: target}, nil
		}
	}

	// Within bounds: only change cursor
	return ViewPos{ItemOffset: m.viewPos.ItemOffset, LineOffset: m.viewPos.LineOffset, Cursor: target}, nil
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

// ToggleSelect toggles the selected status
// of the current Index if amount is 0
// returns err != nil when amount lands outside list and safely does nothing
// else if amount is not 0 toggles selected amount items
// excluding the item on which the cursor would land
func (m *Model) ToggleSelect(amount int) error {
	if amount == 0 {
		m.listItems[m.viewPos.Cursor].selected = !m.listItems[m.viewPos.Cursor].selected
	}

	direction := 1
	if amount < 0 {
		direction = -1
	}

	cur := m.viewPos.Cursor

	target, err := m.Move(amount)
	start, end := cur, target
	if direction < 0 {
		start, end = target+1, cur+1
	}
	// mark/start at first item
	if cur+amount < 0 {
		start = 0
	}
	// mark last item when trying to go beyond list
	if cur+amount >= len(m.listItems) {
		end++
	}
	for c := start; c < end; c++ {
		m.listItems[c].selected = !m.listItems[c].selected
	}
	return err
}

// MarkSelected selects or unselects depending on 'mark'
// amount = 0 changes the current item but does not move the cursor
// if amount would be outside the list error is from type OutOfBounds
// else all items till but excluding the end cursor position gets (un-)marked
func (m *Model) MarkSelected(amount int, mark bool) error {
	cur := m.viewPos.Cursor
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
		return OutOfBounds(fmt.Errorf("Cant go beyond list borders: %d", target))
	}
	for c := 0; c < amount*direction; c++ {
		m.listItems[cur+c].selected = mark
	}
	m.viewPos.Cursor = target
	m.Move(direction)
	return nil
}

// ToggleAllSelected inverts the select state of ALL items
func (m *Model) ToggleAllSelected() {
	for i := range m.listItems {
		m.listItems[i].selected = !m.listItems[i].selected
	}
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

// Sort sorts the list items according to the set less-function
// If there is no Equals-function set (with SetEquals), the current Item will maybe change!
// Since the index of the current pointer does not change
func (m *Model) Sort() {
	equ := m.equals
	var tmp item
	if equ != nil {
		tmp = m.listItems[m.viewPos.Cursor]
	}
	sort.Sort(m)
	if equ == nil {
		return
	}
	for i, item := range m.listItems {
		if is := equ(item.value, tmp.value); is {
			m.viewPos.Cursor = i
			break // Stop when first (and hopefully only one) is found
		}
	}
	m.Move(0)

}

// Less is a Proxy to the less function, set from the user.
func (m *Model) Less(i, j int) bool {
	return m.less(m.listItems[i].value.String(), m.listItems[j].value.String())
}

// Swap swaps the items position within the list
// and is used to fulfill the Sort-interface
func (m *Model) Swap(i, j int) {
	m.listItems[i], m.listItems[j] = m.listItems[j], m.listItems[i]
}

// Len returns the amount of list-items
// and is used to fulfill the Sort-interface
func (m *Model) Len() int {
	return len(m.listItems)
}

// SetLess sets the internal less function used for sorting the list items
func (m *Model) SetLess(less func(string, string) bool) {
	m.less = less
}

// SetEquals sets the internal equals methode used if provided to set the cursor again on the same item after sorting
func (m *Model) SetEquals(equ func(first, second fmt.Stringer) bool) {
	m.equals = equ
}

// GetEquals returns the internal equals methode
// used to set the curser after sorting on the same item again
func (m *Model) GetEquals() func(first, second fmt.Stringer) bool {
	return m.equals
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
	cur := m.viewPos.Cursor
	target, err := m.Move(amount)
	if err != nil {
		return err
	}
	m.Swap(cur, target)
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

// GetIndex returns NotFound error if the Equals Methode is not set (SetEquals)
// else it returns the index of the found item
func (m *Model) GetIndex(toSearch fmt.Stringer) (int, error) {
	if m.equals == nil {
		return -1, NotFound(fmt.Errorf("no equals function provided. Use SetEquals to set it"))
	}
	tmpList := m.listItems
	matchList := make([]chan bool, len(tmpList))
	equ := m.equals

	for i, item := range tmpList {
		resChan := make(chan bool)
		matchList[i] = resChan
		go func(f, s fmt.Stringer, equ func(fmt.Stringer, fmt.Stringer) bool, res chan<- bool) {
			res <- equ(f, s)
		}(item.value, toSearch, equ, resChan)
	}

	var c, lastIndex int
	for i, resChan := range matchList {
		if <-resChan {
			c++
			lastIndex = i
		}
	}
	if c > 1 {
		return -c, MultipleMatches(fmt.Errorf("The provided equals function yields multiple matches betwen one and other fmt.Stringer's"))
	}
	return lastIndex, nil
}

// UpdateAllItems takes a function and updates with it, all items in the list
func (m *Model) UpdateAllItems(updater func(fmt.Stringer) fmt.Stringer) {
	for i, item := range m.listItems {
		m.listItems[i].value = updater(item.value)
	}
}

// UpdateSelectedItems updates all selected items within the list with given function
func (m *Model) UpdateSelectedItems(updater func(fmt.Stringer) fmt.Stringer) {
	for i, item := range m.listItems {
		if item.selected {
			m.listItems[i].value = updater(item.value)
		}
	}
}

// GetCursorIndex returns current cursor position within the List
func (m *Model) GetCursorIndex() (int, error) {
	if !m.focus {
		return m.viewPos.Cursor, NotFocused(fmt.Errorf("Model is not focused"))
	}
	if m.CheckWithinBorder(m.viewPos.Cursor) {
		return m.viewPos.Cursor, OutOfBounds(fmt.Errorf("Cursor is out auf bounds"))
	}
	// TODO handel not focused case
	return m.viewPos.Cursor, nil
}

// GetAllItems returns all items in the list in current order
func (m *Model) GetAllItems() []fmt.Stringer {
	list := m.listItems
	stringerList := make([]fmt.Stringer, len(list))
	for i, item := range list {
		stringerList[i] = item.value
	}
	return stringerList
}

// MoveByLine moves the Viewposition by one line
// not by a item
//func (m *Model) MoveByLine(amount) (ViewPos, error) {
//}

// lineNumber returns line number of the given index
// and if relative is true the absolute difference to the cursor
// or if on the cursor the absolute line number
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

// Copy returns a deep copy of the list-model
func (m *Model) Copy() *Model {
	copiedModel := Model{}
	copier.Copy(&copiedModel, &m)
	return &copiedModel
}
