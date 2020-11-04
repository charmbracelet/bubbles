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

	listItems     []item
	curIndex      int
	visibleItems  []item
	visibleOffset int

	Viewport viewport.Model
	wrap     bool

	seperator      string
	seperatorWrap  string
	relativeNumber bool
	absoluteNumber bool

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
	max := m.Viewport.Height
	if m.curIndex > max {
		max = m.curIndex
	}

	padTo := runewidth.StringWidth(fmt.Sprintf("%d", max))
	sep := runewidth.StringWidth(fmt.Sprintf(m.seperator))

	contentWidth := width - (padTo + sep)
	if contentWidth <= 0 {
		panic("Can't display with zero width for content")
	}

	m.seperator = " ╭ "
	m.seperatorWrap = " │ "

	p := termenv.ColorProfile()

	var visLines int
	var holeString bytes.Buffer
out:
	// Handle list items
	for index, item := range m.listItems {
		sep := m.seperator
		// handel highlighting of current or selected lines
		colored := termenv.String()
		if item.selected {
			colored = colored.Background(p.Color("#ff0000"))
		}
		if index+m.visibleOffset == m.curIndex {
			colored = colored.Reverse()
			sep = " ╭>"
		}
		contentLines := strings.Split(wordwrap.String(colored.Styled(item.content), contentWidth), "\n")

		var firstPad string
		// if set prepend firstline with linenumber
		if m.absoluteNumber || m.relativeNumber {
			firstPad = colored.Styled(fmt.Sprintf("%"+fmt.Sprint(padTo)+"d"+sep, m.visibleOffset+index))
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
	if m.Viewport.Height >= len(m.listItems) {
		m.visibleItems = m.listItems
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
	if m.curIndex > len(m.listItems) {
		return fmt.Errorf("Can't go beyond last item")
	}
	m.curIndex++
	return nil
}

// Up moves the "cursor" or current line up.
// If the start is allready reached err is not nil.
func (m *Model) Up() error {
	if m.curIndex <= 0 {
		return fmt.Errorf("Can't go infront of first item")
	}
	m.curIndex--
	return nil
}

// ToggleSelect toggles the selected status of the current Index
func (m *Model) ToggleSelect() {
	m.listItems[m.curIndex].selected = !m.listItems[m.curIndex].selected
}
