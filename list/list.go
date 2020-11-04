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
	visibleItems     []item
	visibleOffset    int
	lineCurserOffset int

	Viewport viewport.Model
	wrap     bool

	seperator        string
	seperatorWrap    string
	currentSeperator string
	relativeNumber   bool
	absoluteNumber   bool

	jump int

	LineForeGroundColor     string
	LineBackGroundColor     string
	SelectedForeGroundColor string
	SelectedBackGroundColor string
}

// Item are Items used in the list Model
// to hold the Content representat as a string
type item struct {
	selected bool
	content  string
}

// View renders the Lst to a (displayable) string
func (m *Model) View() string {
	width := m.Viewport.Width

	// padding for the right amount of numbers
	max := m.Viewport.Height                       // relativ
	abs := m.visibleOffset + m.Viewport.Height - 1 // absolute
	if abs > max {
		max = abs
	}
	padTo := runewidth.StringWidth(fmt.Sprintf("%d", max))
	sep := maxRuneWidth(m.seperator, m.seperatorWrap, m.currentSeperator)

	// Check if there is space for the content left
	contentWidth := m.Viewport.Width - (padTo + sep)
	if contentWidth <= 0 {
		panic("Can't display with zero width for content")
	}

	// Set Visible lines
	begin := m.visibleOffset
	if begin < 0 {
		begin = 0
	}
	end := m.visibleOffset + m.Viewport.Height
	lenght := len(m.listItems)
	if end > lenght {
		end = len(m.listItems)
	}
	m.visibleItems = m.listItems[begin:end]

	p := termenv.ColorProfile()

	var visLines int
	var holeString bytes.Buffer
out:
	// Handle list items
	for index, item := range m.visibleItems {
		sepString := m.seperator
		// handel highlighting of current or selected lines
		colored := termenv.String()
		if item.selected {
			colored = colored.Background(p.Color(m.SelectedBackGroundColor))
		}
		if index+m.visibleOffset == m.curIndex {
			colored = colored.Reverse()
			sepString = m.currentSeperator
		}
		contentLines := strings.Split(wordwrap.String(colored.Styled(item.content), contentWidth), "\n")

		var firstPad string
		// if set prepend firstline with linenumber
		if m.absoluteNumber || m.relativeNumber {
			firstPad = colored.Styled(fmt.Sprintf("%"+fmt.Sprint(padTo)+"d%"+fmt.Sprint(sep)+"s", m.visibleOffset+index, sepString))
		}
		// Only handel lines that are visible
		if visLines+len(contentLines) >= m.Viewport.Height {
			break out
		}
		// Write first line
		holeString.WriteString(firstPad)
		holeString.WriteString(contentLines[0])
		holeString.WriteString("\n")

		visLines++
		if len(contentLines) == 1 || m.wrap {
			continue
		}

		// Write wraped lines
		for _, line := range contentLines[1:] {
			holeString.WriteString(strings.Repeat(" ", padTo) + m.seperatorWrap) // Pad line // TODO test seperator width
			holeString.WriteString(line)                                         // write line
			holeString.WriteString("\n")                                         // Write end of line
			visLines++
			// Only write lines that are visible
			if visLines >= m.Viewport.Height {
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
				m.jump = 0 // TODO check if this realy resets jump (if m is a pointer) likely pointer
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

// SetAbsNumber sets if absolute Linenumbers should be displayed
func (m *Model) SetAbsNumber(setTo bool) {
	m.absoluteNumber = setTo
}

// SetSeperator sets the seperator string
// between left border and the content of the line
func (m *Model) SetSeperator(sep string) {
	m.seperator = sep
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

		wrap: true,

		seperator:        " ╭ ",
		seperatorWrap:    " │ ",
		currentSeperator: " ╭>",
		absoluteNumber:   true,

		SelectedBackGroundColor: "#ff0000",
	}
}

// Top moves the cursor to the first line
func (m *Model) Top() {
	m.visibleOffset = 0
	m.visibleItems = m.listItems[0:m.Viewport.Height]
	m.curIndex = 0
}

// Bottom moves the cursor to the first line
func (m *Model) Bottom() {
	visLines := m.Viewport.Height - m.lineCurserOffset
	start := len(m.listItems) - visLines // FIXME acount for wraped lines
	m.visibleOffset = start
	m.visibleItems = m.listItems[start : start+visLines]
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
