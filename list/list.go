package list

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/termenv"
	"sort"
	"strings"
)

// Model is a bubbletea List of strings
type Model struct {
	focus bool

	listItems []item

	less   func(fmt.Stringer, fmt.Stringer) bool // function used for sorting
	equals func(fmt.Stringer, fmt.Stringer) bool // used after sorting, to be set from the user

	CursorOffset int // offset or margin between the cursor and the viewport(visible) border

	Screen  ScreenInfo
	viewPos ViewPos

	// Wrap changes the number of lines which get displayed. 0 means unlimited lines.
	Wrap int

	PrefixGen Prefixer
	SuffixGen Suffixer

	LineStyle     termenv.Style
	SelectedStyle termenv.Style
	CurrentStyle  termenv.Style

	// Channels to create unique ids for all added/new items
	requestID chan<- struct{}
	resultID  <-chan int
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
		viewPos:      ViewPos{LineOffset: 5},

		// show all lines
		Wrap: 0,

		SelectedStyle: selStyle,
		CurrentStyle:  curStyle,
	}
}

// Init does nothing
func (m Model) Init() tea.Cmd {
	return nil
}

// View renders the List to a (displayable) string
// since a empty string gets not displayed return something to overwrite the last removed item
func (m Model) View() string {
	lines := strings.Join(m.Lines(), "\n")
	if lines == "" {
		// TODO make empty string handling better, custom empty function?
		return "empty"
	}
	return lines
}

// Lines returns the Visible lines of the list items
// used to display the current user interface
func (m *Model) Lines() []string {
	// get public variables as locals so they can't change while using
	height := m.Screen.Height
	width := m.Screen.Width
	// check visible area
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

	linesBefor := make([]string, 0, height)
	// loop to add the item(-lines) befor the cursor to the return lines
	for c := 1; // dont add cursor item
	m.viewPos.Cursor-c >= 0; c++ {
		index := m.viewPos.Cursor - c
		item := m.listItems[index]

		contentLines := m.itemLines(item)
		// append the lines in reverse, to add them in correct order later
		for c := len(contentLines) - 1; c >= 0 && len(linesBefor) < m.viewPos.LineOffset; c-- {
			lineContent := contentLines[c]
			// Surrounding lineContent
			var linePrefix, lineSuffix string
			if m.PrefixGen != nil {
				linePrefix = m.PrefixGen.Prefix(index, c, item.selected)
			}
			if m.SuffixGen != nil {
				lineSuffix = m.SuffixGen.Suffix(index, c, item.selected)
				if lineSuffix != "" {
					free := contentWidth - ansi.PrintableRuneWidth(lineContent)
					if free < 0 {
						free = 0 // TODO is this nessecary after adding hardwrap?
					}
					lineSuffix = fmt.Sprintf("%s%s", strings.Repeat(" ", free), lineSuffix)
				}
			}

			// Join all
			line := fmt.Sprintf("%s%s%s", linePrefix, lineContent, lineSuffix)

			// Highlighting of selected lines
			style := m.LineStyle
			if item.selected {
				style = m.SelectedStyle
			}

			// Highlight and write wrapped line
			linesBefor = append(linesBefor, style.Styled(line))
		}

	}

	// append lines (befor cursor) in correct order to allLines
	allLines := make([]string, 0, height)
	for c := len(linesBefor) - 1; c >= 0; c-- {
		allLines = append(allLines, linesBefor[c])
	}

	var visLines int
	// Handle list items, start at cursor and go till end of list or visible (break)
	for index := m.viewPos.Cursor; index < m.Len(); index++ {
		item := m.listItems[index]

		lines := m.itemLines(item)

		// append all visibles lines since the cursor
		for c := 0; c < len(lines) && len(allLines) < height; c++ {
			lineContent := lines[c]
			// Surrounding content
			var linePrefix, lineSuffix string
			if m.PrefixGen != nil {
				linePrefix = m.PrefixGen.Prefix(index, c, item.selected)
			}
			if m.SuffixGen != nil {
				free := contentWidth - ansi.PrintableRuneWidth(lineContent)
				if free < 0 {
					free = 0 // TODO is this nessecary?
				}
				lineSuffix = fmt.Sprintf("%s%s", strings.Repeat(" ", free), m.SuffixGen.Suffix(index, c, item.selected))
			}

			// Join all
			line := fmt.Sprintf("%s%s%s", linePrefix, lineContent, lineSuffix)

			// Highlighting of selected and current lines
			style := m.LineStyle
			if item.selected {
				style = m.SelectedStyle
			}
			if index == m.viewPos.Cursor {
				style = m.CurrentStyle
			}

			// Highlight and write wrapped line
			allLines = append(allLines, style.Styled(line))
			visLines++
		}
	}
	return allLines
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
		// Quit
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		switch msg.String() {
		case "q":
			return m, tea.Quit

		// Move
		case "down", "j":
			m.MoveCursor(1)
			return m, nil
		case "up", "k":
			m.MoveCursor(-1)
			return m, nil
		case "t", "home":
			m.Top()
			return m, nil
		case "b", "end":
			m.Bottom()
			return m, nil
		case "K":
			m.MoveItem(-1)
			return m, nil
		case "J":
			m.MoveItem(1)
			return m, nil

		// Select
		case " ":
			m.ToggleSelect(1)
			m.MoveCursor(1)
			return m, nil
		case "v": // inVert
			m.ToggleAllSelected()
			return m, nil
		case "m": // mark
			m.MarkSelected(1, true)
			return m, nil
		case "M": // mark All
			m.MarkAllSelected(true)
			return m, nil
		case "u": // unmark
			m.MarkSelected(1, false)
			return m, nil
		case "U": // unmark All
			m.MarkAllSelected(false)
			return m, nil

		// Order changing
		case "s":
			m.Sort()
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
			m.MoveCursor(-1)
			return m, nil

		case tea.MouseWheelDown:
			m.MoveCursor(1)
			return m, nil
		}
	}
	return m, nil
}

// NoItems is a error returned when the list is empty
type NoItems error

// NotFound gets return if the search does not yield a result
type NotFound error

// OutOfBounds is return if and index is outside the list bounderys
type OutOfBounds error

// MultipleMatches gets return if the search yield more result
type MultipleMatches error

// ConfigError is return if there is a error with the configuration of the list Modul
type ConfigError error

// NotFocused is a error return if the action can only be applied to a focused list.
type NotFocused error

// NilValue is returned if there was a request to set nil as value of a list item.
type NilValue error

// ValidIndex returns a error when the list has no items, is not focused, the index is out of bounds.
// And the nearest valid index in case of OutOfBounds error, else the index it self.
func (m *Model) ValidIndex(index int) (int, error) {
	if m.Len() <= 0 {
		return 0, NoItems(fmt.Errorf("the list has no items"))
	}
	if !m.focus {
		return 0, NotFocused(fmt.Errorf("the list is not focused"))
	}
	if index < 0 {
		return 0, OutOfBounds(fmt.Errorf("the requested index (%d) is infront the list begin (%d)", index, 0))
	}
	if index > m.Len()-1 {
		return m.Len() - 1, OutOfBounds(fmt.Errorf("the requested index (%d) is beyond the list end (%d)", index, m.Len()-1))
	}
	return index, nil
}

func (m *Model) validOffset(newCursor int) (int, error) {
	if m.CursorOffset*2 > m.Screen.Height {
		return 0, ConfigError(fmt.Errorf("CursorOffset must be less than have the screen height"))
	}
	newCursor, err := m.ValidIndex(newCursor)
	if m.Len() <= 0 {
		return m.CursorOffset, err
	}
	amount := newCursor - m.viewPos.Cursor
	if amount == 0 {
		if m.viewPos.LineOffset < m.CursorOffset {
			return m.CursorOffset, nil
		}
		return m.viewPos.LineOffset, nil
	}
	newOffset := m.viewPos.LineOffset + amount

	if m.Wrap == 0 || m.Wrap > 1 {
		// assume down (positiv) movement
		start := 0
		stop := amount - 1 // exclude target item (-lines)

		d := 1
		if amount < 0 {
			d = -1
			stop = amount * d
			start = 1 // exclude old cursor position
		}

		var lineSum int
		for i := start; i <= stop; i++ {
			lineSum += strings.Count(m.listItems[m.viewPos.Cursor+i*d].value.String(), "\n") + 1
		}
		newOffset = m.viewPos.LineOffset + lineSum*d
	}

	if newOffset < m.CursorOffset {
		newOffset = m.CursorOffset
	} else if newOffset > m.Screen.Height-m.CursorOffset-1 {
		newOffset = m.Screen.Height - m.CursorOffset - 1
	}
	return newOffset, err
}

// ViewPos is used for holding the information about the View parameters
type ViewPos struct {
	LineOffset int
	Cursor     int
}

// ScreenInfo holds all information about the screen Area
type ScreenInfo struct {
	Width   int
	Height  int
	Profile termenv.Profile
}

// MoveCursor moves the cursor by amount and returns OutOfBounds error if amount go's beyond list borders
// or if the CursorOffset is greater than half of the display height returns ConfigError
func (m *Model) MoveCursor(amount int) (int, error) {
	target := m.viewPos.Cursor + amount

	newOffset, err := m.validOffset(target)
	target, err = m.ValidIndex(target)

	m.viewPos.Cursor = target
	m.viewPos.LineOffset = newOffset
	return target, err
}

// SetCursor set the cursor to the specified index if possible,
// if not the nearest end of the list, will be used and OutOfBounds error is returned
func (m *Model) SetCursor(target int) (int, error) {
	newOffset, err := m.validOffset(target)
	target, err = m.ValidIndex(target)
	m.viewPos.Cursor = target
	m.viewPos.LineOffset = newOffset
	return target, err
}

// Top moves the cursor to the first line
func (m *Model) Top() {
	m.viewPos.Cursor = 0
	m.viewPos.LineOffset = m.CursorOffset
}

// Bottom moves the cursor to the last line
func (m *Model) Bottom() {
	end := len(m.listItems) - 1
	m.viewPos.LineOffset = m.Screen.Height - m.CursorOffset
	m.MoveCursor(end)
}

// AddItems adds the given Items to the list Model
// and if a costum less function is provided, they get sorted.
// if a entry of itemList is nil it will get skiped
func (m *Model) AddItems(itemList []fmt.Stringer) error {
	oldLenght := m.Len()
	for _, i := range itemList {
		if i != nil {
			m.listItems = append(m.listItems, item{
				selected: false,
				value:    i,
				id:       m.getID(),
			},
			)
		}
	}
	// only sort if user set less function
	if m.less != nil {
		// Sort will take care of the correct position of Cursor and Offset
		m.Sort()
	}
	var err error
	if m.Len() < oldLenght+len(itemList) {
		err = NilValue(fmt.Errorf("there where '%d' nil values which where not added", m.Len()-oldLenght+len(itemList)))
	}
	return err
}

// RemoveIndex returns a error if the index is not valid,
// and if valid, returns the item while removing it from the list.
func (m *Model) RemoveIndex(index int) (fmt.Stringer, error) {
	index, err := m.ValidIndex(index)
	if m.Len() == 0 {
		m.viewPos.Cursor = 0
		return nil, err
	}
	var rest []item
	itemValue, _ := m.GetItem(index)
	if index+1 < m.Len() {
		rest = m.listItems[index+1:]
	}
	m.listItems = append(m.listItems[:index], rest...)
	newCursor, _ := m.ValidIndex(index)
	newOffset, _ := m.validOffset(newCursor)
	m.viewPos.Cursor = newCursor
	m.viewPos.LineOffset = newOffset
	return itemValue, err
}

// ToggleSelect toggles the selected status
// of the current Index if amount is 0
// returns err != nil when amount lands outside list and safely does nothing
// else if amount is not 0 toggles selected amount items
// excluding the item on which the cursor would land
func (m *Model) ToggleSelect(amount int) error {
	if m.Len() == 0 {
		return OutOfBounds(fmt.Errorf("No Items"))
	}
	if amount == 0 {
		m.listItems[m.viewPos.Cursor].selected = !m.listItems[m.viewPos.Cursor].selected
	}

	direction := 1
	if amount < 0 {
		direction = -1
	}

	cur := m.viewPos.Cursor

	target, err := m.MoveCursor(amount)
	start, end := cur, target
	if direction < 0 {
		start, end = target+1, cur+1
	}
	// mark/start at first item
	if cur+amount < 0 {
		start = 0
	}
	// mark last item when trying to go beyond list
	if cur+amount >= m.Len() {
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
	direction := 1
	if amount < 0 {
		direction = -1
	}
	target := cur + amount - direction

	target, err := m.ValidIndex(target)
	if m.Len() == 0 {
		return err
	}
	// correct amount in case target has changed
	amount = target - cur + direction

	if amount == 0 {
		m.listItems[cur].selected = mark
		return nil
	}
	for c := 0; c < amount*direction; c++ {
		m.listItems[cur+c].selected = mark
	}
	m.viewPos.Cursor = target
	_, errSec := m.MoveCursor(direction)
	if err == nil {
		err = errSec
	}
	return err
}

// MarkAllSelected marks all items of the list according to mark
// or returns OutOfBounds if list has no Items
func (m *Model) MarkAllSelected(mark bool) error {
	_, err := m.ValidIndex(0)
	if m.Len() == 0 {
		return err
	}
	for c := range m.listItems {
		m.listItems[c].selected = mark
	}
	return err
}

// ToggleAllSelected inverts the select state of ALL items
func (m *Model) ToggleAllSelected() error {
	_, err := m.ValidIndex(0)
	if m.Len() == 0 {
		return err
	}
	for i := range m.listItems {
		m.listItems[i].selected = !m.listItems[i].selected
	}
	return err
}

// IsSelected returns true if the given Item is selected
// false otherwise. If the requested index is outside the list
// error is not nil.
func (m *Model) IsSelected(index int) (bool, error) {
	index, err := m.ValidIndex(index)
	if m.Len() == 0 {
		return false, err
	}
	return m.listItems[index].selected, err
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
// If its not set than order after string.
func (m *Model) Sort() {
	if m.Len() < 1 {
		return
	}
	old := m.listItems[m.viewPos.Cursor].id
	sort.Sort(m)
	for i, item := range m.listItems {
		if item.id == old {
			m.viewPos.Cursor = i
			break
		}
	}
}

// Less is a Proxy to the less function, set from the user.
// since the Sort-interface demands a Less Methode without a error return value
// so we sadly have to returns silently if a index is out side the list, to not panic.
func (m *Model) Less(i, j int) bool {
	_, errI := m.ValidIndex(i)
	_, errJ := m.ValidIndex(j)
	if errI != nil || errJ != nil {
		return false
	}
	// If User does not provide less function use string comparison, but dont change m.less, to be able to see when user set one.
	if m.less == nil {
		return m.listItems[i].value.String() < m.listItems[j].value.String()
	}
	return m.less(m.listItems[i].value, m.listItems[j].value)
}

// Swap swaps the items position within the list
// and is used to fulfill the Sort-interface
// since the Sort-interface demands a Swap Methode without a error return value
// so we sadly have to returns silently if a index is out side the list, to not panic.
func (m *Model) Swap(i, j int) {
	_, errI := m.ValidIndex(i)
	_, errJ := m.ValidIndex(j)
	if errI != nil || errJ != nil {
		return
	}
	m.listItems[i], m.listItems[j] = m.listItems[j], m.listItems[i]
}

// Len returns the amount of list-items
// and is used to fulfill the Sort-interface
func (m *Model) Len() int {
	return len(m.listItems)
}

// SetLess sets the internal less function used for sorting the list items
func (m *Model) SetLess(less func(a, b fmt.Stringer) bool) {
	m.less = less
}

// SetEquals sets the internal equals methode used to get the index (GetIndex) of a provided fmt.Stringer value
func (m *Model) SetEquals(equ func(first, second fmt.Stringer) bool) {
	m.equals = equ
}

// GetEquals returns the internal equals methode
// used to get the index (GetIndex) of a provided fmt.Stringer value
func (m *Model) GetEquals() func(first, second fmt.Stringer) bool {
	return m.equals
}

// MoveItem moves the current item by amount to the end
// So: MoveItem(1) Moves the Item towards the end by one
// and MoveItem(-1) Moves the Item towards the beginning
// MoveItem(0) safely does nothing
// and a amount that would result outside the list returns a error != nil
func (m *Model) MoveItem(amount int) error {
	cur := m.viewPos.Cursor
	target, err := m.ValidIndex(cur + amount)
	if m.Len() == 0 {
		return err
	}
	if amount == 0 {
		return nil
	}
	if err != nil {
		return err
	}
	d := 1
	if amount < 0 {
		d = -1
	}
	// TODO change to not O(n)
	for c := 0; c*d < amount*d; c += d {
		m.Swap(cur+c, cur+c+d)
	}
	linOff, _ := m.validOffset(target)
	m.viewPos.LineOffset = linOff
	m.viewPos.Cursor = target
	return nil
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
// else it returns the index of the first found item
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
		// TODO performance: trust User and remove check for multiple matches?
		return -c, MultipleMatches(fmt.Errorf("The provided equals function yields multiple matches betwen one and other fmt.Stringer's"))
	}
	return lastIndex, nil
}

// UpdateItem takes a indes and updates the item at the index with the given function
// or if index outside the list returns OutOfBounds error.
// If the returned fmt.Stringer value is nil, then the item gets removed from the list and a NilValue error is returned.
func (m *Model) UpdateItem(index int, updater func(fmt.Stringer) (fmt.Stringer, tea.Cmd)) (tea.Cmd, error) {
	index, err := m.ValidIndex(index)
	if m.Len() == 0 {
		return nil, err
	}
	v, cmd := updater(m.listItems[index].value)
	// remove item when value equals nil
	if v == nil {
		m.RemoveIndex(index)
		return cmd, NilValue(fmt.Errorf("cant add nil value to list"))
	}
	m.listItems[index].value = v
	return cmd, err
}

// GetCursorIndex returns current cursor position within the List
// and also NotFocused error if the Model is not focused
func (m *Model) GetCursorIndex() (int, error) {
	if m.Len() == 0 {
		return 0, OutOfBounds(fmt.Errorf("No Items"))
	}
	if !m.focus {
		return m.viewPos.Cursor, NotFocused(fmt.Errorf("Model is not focused"))
	}
	return m.viewPos.Cursor, nil
}

// GetItem returns the item if the index exists
// OutOfBounds otherwise
func (m *Model) GetItem(index int) (fmt.Stringer, error) {
	index, err := m.ValidIndex(index)
	if m.Len() == 0 {
		return nil, err
	}
	return m.listItems[index].value, nil
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

// Copy returns a deep copy of the list-model
func (m *Model) Copy() *Model {
	copiedModel := &Model{}
	*copiedModel = *m
	return copiedModel
}

// GetID returns a new for this list unique id
func (m *Model) getID() int {
	if m.requestID == nil || m.resultID == nil {
		req := make(chan struct{})
		res := make(chan int)

		m.requestID = req
		m.resultID = res

		// the id '0' is skiped to be able to distinguish zero-value and proper id TODO is this a valid/good way to go?
		go func(requ <-chan struct{}, send chan<- int) {
			for c := 2; true; c++ {
				_ = <-requ
				send <- c
			}
		}(req, res)

		return 1
	}
	var e struct{}
	m.requestID <- e
	return <-m.resultID
}
