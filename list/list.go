package list

import (
	"bytes"
	"fmt"
	"github.com/charmbracelet/bubbles/viewport"
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

	listItems        []item
	curIndex         int                    // curser
	visibleOffset    int                    // begin of the visible lines
	lineCurserOffset int                    // offset or margin between the cursor and the viewport(visible) border
	less             func(k, l string) bool // function used for sorting

	jump int // maybe buffer for jumping multiple lines

	Viewport viewport.Model
	Wrap     bool

	SelectedPrefix string
	Seperator      string
	SeperatorWrap  string
	CurrentMarker  string

	Number         bool
	NumberRelative bool

	LineForeGroundStyle     termenv.Style
	LineBackGroundStyle     termenv.Style
	SelectedForeGroundStyle termenv.Style
	SelectedBackGroundStyle termenv.Style
}

// Item are Items used in the list Model
// to hold the Content representat as a string
type item struct {
	selected     bool
	content      string
	wrapedLines  []string
	wrapedLenght int
	wrapedto     int
	userValue    interface{}
}

// genVisLines renews the wrap of the content into wraplines
func (i item) genVisLines(wrapTo int) item {
	i.wrapedLines = strings.Split(wordwrap.String(i.content, wrapTo), "\n")
	//TODO hardwrap lines/words
	i.wrapedLenght = len(i.wrapedLines)
	i.wrapedto = wrapTo
	return i
}

// View renders the Lst to a (displayable) string
func (m *Model) View() string {
	width := m.Viewport.Width

	// check visible area
	height := m.Viewport.Height - 1 // TODO question: why does the first line get cut of, if i ommit the -1?
	width := m.Viewport.Width
	offset := m.visibleOffset
	if height*width <= 0 {
		panic("Can't display with zero width or hight of Viewport")
	}

	// Get max seperator width
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
	prefix := m.SelectedPrefix
	preWidth := ansi.PrintableRuneWidth(prefix)
	prepad := strings.Repeat(" ", preWidth)

	// pad all seperators to the same width for easy exchange
	sepItem := strings.Repeat(" ", sepWidth-widthItem) + m.Seperator
	sepWrap := strings.Repeat(" ", sepWidth-widthWrap) + m.SeperatorWrap

	// pad right of prefix, with lenght of current pointer
	suffix := m.CurrentMarker
	sufWidth := ansi.PrintableRuneWidth(suffix)
	sufpad := strings.Repeat(" ", sufWidth)

	// Get the hole prefix width
	holePrefixWidth := numWidth + preWidth + sepWidth + sufWidth

	// Get actual content width
	contentWidth := width - holePrefixWidth

	// Check if there is space for the content left
	if contentWidth <= 0 {
		panic("Can't display with zero width for content")
	}

	// renew wrap of all items TODO check if to slow
	for i := range m.listItems {
		m.listItems[i] = m.listItems[i].genVisLines(contentWidth)
	}

	var visLines int
	var holeString bytes.Buffer
out:
	// Handle list items, start at first visible and go till end of list or visible (break)
	for index := offset; index < len(m.listItems); index++ {
		if index >= len(m.listItems) || index < 0 {
			// TODO log error
			break
		}

		item := m.listItems[index]
		if item.wrapedLenght <= 0 {
			panic("cant display item with no visible content")
		}

		// if a number is set, prepend firstline with number and both with enough spaces
		firstPad := strings.Repeat(" ", numWidth)
		var wrapPad string
		if m.Number {
			lineNum := lineNumber(m.NumberRelative, m.curIndex, index)
			number := fmt.Sprintf("%d", lineNum)
			// since diggets are only singel bytes, len is sufficent:
			firstPad = strings.Repeat(" ", numWidth-len(number)) + number
			// pad wraped lines
			wrapPad = strings.Repeat(" ", numWidth)
		}

		// Selecting: handel highlighting and prefixing of selected lines
		selString := prepad       // assume not selected
		style := termenv.String() // create empty style

		if item.selected {
			style = m.SelectedBackGroundStyle // fill style
			selString = prefix                // change if selected
		}

		// Current: handel highlighting of current item/first-line
		curPad := sufpad
		if index == m.curIndex {
			style = style.Reverse()
			curPad = suffix
		}

		// join all prefixes
		linePrefix := strings.Join([]string{firstPad, selString, sepItem, curPad}, "")
		wrapPrefix := strings.Join([]string{wrapPad, selString, sepWrap, sufpad}, "") // dont prefix wrap lines with CurrentMarker (suffix)

		// join pad and first line content
		// NOTE linebreak is not added here because it would mess with the highlighting
		line := fmt.Sprintf("%s%s", linePrefix, item.wrapedLines[0])

		// Highlight and write first line
		holeString.WriteString(style.Styled(line))
		holeString.WriteString("\n")
		visLines++

		// Only write lines that are visible
		if visLines >= height {
			break out
		}

		// Dont write wraped lines if not set
		if !m.Wrap || item.wrapedLenght <= 1 {
			continue
		}

		// Write wraped lines
		for _, line := range item.wrapedLines[1:] {
			// Pad left of line
			// NOTE linebreak is not added here because it would mess with the highlighting
			padLine := fmt.Sprintf("%s%s", wrapPrefix, line)

			// Highlight and write wraped line
			holeString.WriteString(style.Styled(padLine))
			holeString.WriteString("\n")
			visLines++

			// Only write lines that are visible
			if visLines >= height {
				break out
			}
		}
	}
	return holeString.String()
}

// lineNumber returns line number of the given index
// and if relative is true the absolute difference to the curser
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
			m.Down()
			return m, nil
		case "up", "k":
			m.Up()
			return m, nil
		case " ":
			m.ToggleSelect()
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
		}

	case tea.WindowSizeMsg:

		m.Viewport.Width = msg.Width
		m.Viewport.Height = msg.Height

		// Because we're using the viewport's default update function (with pager-
		// style navigation) it's important that the viewport's update function:
		//
		// * Recieves messages from the Bubble Tea runtime
		// * Returns commands to the Bubble Tea runtime
		//

		m.Viewport, cmd = viewport.Update(msg, m.Viewport)

		return m, cmd
	}
	return m, nil
}

// Init does nothing
func (m Model) Init() tea.Cmd {
	return nil
}

// AddItems addes the given Items to the list Model
// Without performing updating the View TODO
func (m *Model) AddItems(itemList []string) {
	for _, i := range itemList {
		m.listItems = append(m.listItems, item{
			selected: false,
			content:  i},
		)
	}
}

// Down moves the "cursor" or current line down.
// If the end is allready reached err is not nil.
func (m *Model) Down() error {
	length := len(m.listItems) - 1
	if m.curIndex >= length {
		m.curIndex = length
		return fmt.Errorf("Can't go beyond last item")
	}
	m.curIndex++
	// move visible part of list if Curser is going beyond border.
	lowerBorder := m.Viewport.Height + m.visibleOffset - m.lineCurserOffset
	if m.curIndex > lowerBorder {
		m.visibleOffset = m.curIndex - (m.Viewport.Height - m.lineCurserOffset)
	}
	return nil
}

// Up moves the "cursor" or current line up.
// If the start is allready reached err is not nil.
func (m *Model) Up() error {
	if m.curIndex <= 0 {
		m.curIndex = 0
		return fmt.Errorf("Can't go infront of first item")
	}
	m.curIndex--
	// move visible part of list if Curser is going beyond border.
	upperBorder := m.visibleOffset + m.lineCurserOffset
	if m.visibleOffset > 0 && m.curIndex <= upperBorder {
		m.visibleOffset = m.curIndex - m.lineCurserOffset
	}
	return nil
}

// ToggleSelect toggles the selected status of the current Index
func (m *Model) ToggleSelect() {
	m.listItems[m.curIndex].selected = !m.listItems[m.curIndex].selected
}

// NewModel returns a Model with some save/sane defaults
func NewModel() Model {
	p := termenv.ColorProfile()
	style := termenv.Style{}.Background(p.Color("#ff0000"))
	return Model{
		lineCurserOffset: 5,

		Wrap: true,

		Seperator:      "╭",
		SeperatorWrap:  "│",
		CurrentMarker:  ">",
		SelectedPrefix: "*",
		Number:         true,

		less: func(k, l string) bool {
			return k < l
		},

		SelectedBackGroundStyle: style,
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
	maxVisItems := m.Viewport.Height - m.lineCurserOffset
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
func (m *Model) GetSelected() []string {
	var selected []string
	for _, item := range m.listItems {
		if item.selected {
			selected = append(selected, item.content)
		}
	}
	return selected
}

// Less is a Proxy to the less function, set from the user.
// Swap is used to fullfill the Sort-interface
func (m *Model) Less(i, j int) bool {
	return m.less(m.listItems[i].content, m.listItems[j].content)
}

// Swap is used to fullfill the Sort-interface
func (m *Model) Swap(i, j int) {
	m.listItems[i], m.listItems[j] = m.listItems[j], m.listItems[i]
}

// Len is used to fullfill the Sort-interface
func (m *Model) Len() int {
	return len(m.listItems)
}

// SetLess sets the internal less function used for sorting the list items
func (m *Model) SetLess(less func(string, string) bool) {
	m.less = less
}

// Sort sorts the listitems acording to the set less function
func (m *Model) Sort() {
	sort.Sort(m)
}
