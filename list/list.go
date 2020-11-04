package list

import (
	"bytes"
	"fmt"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	runewidth "github.com/mattn/go-runewidth"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/termenv"
	"strings"
)

// Model is a bubbletea List of strings
type Model struct {
	focus bool

	listItems        []item
	curIndex         int
	visibleOffset    int
	lineCurserOffset int

	Viewport viewport.Model
	Wrap     bool

	Seperator        string
	SeperatorWrap    string
	CurrentSeperator string
	RelativeNumber   bool
	AbsoluteNumber   bool

	jump int // maybe buffer for jumping multiple lines

	LineForeGroundColor     string
	LineBackGroundColor     string
	SelectedForeGroundColor string
	SelectedBackGroundColor string
}

// Item are Items used in the list Model
// to hold the Content representat as a string
type item struct {
	selected     bool
	content      string
	wrapedLines  []string
	wrapedLenght int
	wrapedto     int
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
	sepWidth := maxRuneWidth(m.Seperator, m.SeperatorWrap, m.CurrentSeperator)

	// Get actual content width
	contentWidth := width - (sepWidth + padWidth + 1)

	// Check if there is space for the content left
	if contentWidth <= 0 {
		panic("Can't display with zero width for content")
	}

	// renew wrap of all items TODO check if to slow
	for i := range m.listItems {
		m.listItems[i] = m.listItems[i].genVisLines(contentWidth)

	}

	p := termenv.ColorProfile()

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

		// handel highlighting of selected lines
		colored := termenv.String()
		if item.selected {
			colored = colored.Background(p.Color(m.SelectedBackGroundColor))
		}

		// handel highlighting of current line
		if index+m.visibleOffset == m.curIndex {
			colored = colored.Reverse()
			sepString = m.CurrentSeperator
		}

		// if set prepend firstline with linenumber
		var firstPad string
		if m.AbsoluteNumber || m.RelativeNumber {
			lineOffset := m.visibleOffset + index
			firstPad = fmt.Sprintf("%"+fmt.Sprint(padWidth)+"d%"+fmt.Sprint(sepWidth)+"s", lineOffset, sepString)
		}

		// join pad and line content
		if item.wrapedLenght == 0 {
			panic("cant display item with no visible content")
		}

		lineContent := item.wrapedLines[0]
		// NOTE linebreak is not added here because it would mess with the highlighting
		line := fmt.Sprintf("%s%s", firstPad, lineContent)

		// Highlight and write first line
		coloredLine := colored.Styled(line)
		holeString.WriteString(coloredLine)
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
			pad := strings.Repeat(" ", padWidth) + m.SeperatorWrap
			// NOTE linebreak is not added here because it would mess with the highlighting
			padLine := fmt.Sprintf("%s%s", pad, line)

			// Highlight and write wraped line
			holeString.WriteString(colored.Styled(padLine))
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			if m.jump > 0 {
				m.curIndex -= m.jump
				m.jump = 0
			} else {
				m.curIndex--
			}
		case " ":
			m.listItems[m.curIndex].selected = !m.listItems[m.curIndex].selected
		}
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
	return Model{
		lineCurserOffset: 5,

		Wrap: true,

		Seperator:        " ╭ ",
		SeperatorWrap:    " │ ",
		CurrentSeperator: " ╭>",
		AbsoluteNumber:   true,

		SelectedBackGroundColor: "#ff0000",
	}
}

// Top moves the cursor to the first line
func (m *Model) Top() {
	m.visibleOffset = 0
	m.curIndex = 0
}

// Bottom moves the cursor to the first line
func (m *Model) Bottom() {
	visLines := m.Viewport.Height - m.lineCurserOffset
	start := len(m.listItems) - visLines // FIXME acount for wraped lines
	m.visibleOffset = start
	m.curIndex = len(m.listItems) - 1
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
