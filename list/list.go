package list

import (
	"bytes"
	"fmt"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	runewidth "github.com/mattn/go-runewidth"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/termenv"
	"sort"
	"strings"
)

// Model is a bubbletea List of strings
type Model struct {
	focus bool

	listItems        []item
	curIndex         int
	visibleOffset    int
	lineCurserOffset int
	less             func(k, l string) bool

	jump int // maybe buffer for jumping multiple lines

	Viewport viewport.Model
	Wrap     bool

	Seperator        string
	SeperatorWrap    string
	SeperatorCurrent string
	SelectedPrefix   string
	RelativeNumber   bool
	AbsoluteNumber   bool

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
	height := m.Viewport.Height
	width := m.Viewport.Width
	if height*width <= 0 {
		panic("Can't display with zero width or hight of Viewport")
	}

	// if there is something to pad
	var relWidth, absWidth, padWidth int

	if m.RelativeNumber {
		relWidth = len(fmt.Sprintf("%d", height))
	}

	if m.AbsoluteNumber {
		absWidth = len(fmt.Sprintf("%d", len(m.listItems)))
	}

	// get widest number to pad
	padWidth = relWidth
	if padWidth < absWidth {
		padWidth = absWidth
	}

	// Get max seperator width
	sepWidth := maxRuneWidth(m.Seperator, m.SeperatorWrap, m.SeperatorCurrent) + runewidth.StringWidth(m.SelectedPrefix)

	//Get hole Width
	holeWidth := sepWidth + padWidth

	// Get actual content width
	contentWidth := width - (holeWidth + 1)

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
	for index := m.visibleOffset; index < len(m.listItems); index++ {
		if index >= len(m.listItems) || index < 0 {
			break
		}

		item := m.listItems[index]

		sepString := m.Seperator
		wrapString := m.SeperatorWrap

		// handel highlighting and prefixing of selected lines
		style := termenv.String()
		if item.selected {
			style = m.SelectedBackGroundStyle
			sepString = m.SelectedPrefix + sepString
			wrapString = m.SelectedPrefix + wrapString
		}

		// handel highlighting of current line
		if index+m.visibleOffset == m.curIndex {
			style = style.Reverse()
			sepString = m.SeperatorCurrent
		}

		// if set, prepend firstline with enough space for linenumber and seperator
		// This while first create a string like: "%3d%4s"
		// Which will be than filled with linenumber and seperator string
		var firstPad string
		if m.AbsoluteNumber || m.RelativeNumber {
			lineOffset := m.visibleOffset + index
			firstPad = fmt.Sprintf("%"+fmt.Sprint(padWidth)+"d%"+fmt.Sprint(sepWidth)+"s", lineOffset, sepString)
		}

		if item.wrapedLenght == 0 {
			panic("cant display item with no visible content")
		}

		lineContent := item.wrapedLines[0]
		// join pad and line content
		// NOTE linebreak is not added here because it would mess with the highlighting
		line := fmt.Sprintf("%s%s", firstPad, lineContent)

		// Highlight and write first line
		holeString.WriteString(style.Styled(line))
		holeString.WriteString("\n")
		visLines++

		// Dont write wraped lines if not set
		if !m.Wrap || item.wrapedLenght < 1 {
			continue
		}

		// Only write lines that are visible
		if visLines >= height {
			break out
		}

		// Write wraped lines
		for _, line := range item.wrapedLines[1:] {
			// Pad left of line
			// TODO performance: do stringlength and prepending befor loop
			pad := strings.Repeat(" ", holeWidth-runewidth.StringWidth(wrapString)) + wrapString
			// NOTE linebreak is not added here because it would mess with the highlighting
			padLine := fmt.Sprintf("%s%s", pad, line)

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
	if m.curIndex >= lowerBorder {
		m.visibleOffset++
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
		m.visibleOffset--
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

		Seperator:        " ╭ ",
		SeperatorWrap:    " │ ",
		SeperatorCurrent: " ╭>",
		SelectedPrefix:   "*",
		AbsoluteNumber:   true,
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

// maxRuneWidth returns the maximal lenght of occupied space
// frome the given strings
func maxRuneWidth(words ...string) int {
	var max int
	for _, w := range words {
		width := runewidth.StringWidth(w)
		if width > max {
			max = width
		}
	}
	return max
}

// GetSelected returns you a orderd list of all items
// that are selected
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
