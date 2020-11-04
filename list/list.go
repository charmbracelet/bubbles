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
	listItems     []item
	curIndex      int
	visibleItems  []item
	visibleOffset int

	Viewport viewport.Model
	wrap     bool

	seperator      string
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
		panic("Can't display width zero width for content")
	}

	p := termenv.ColorProfile()

	var visLines int
	var holeString bytes.Buffer
out:
	for index, item := range m.listItems {
		//handel highlighting of current or selected lines
		colored := termenv.String()
		if item.selected {
			colored.Background(p.Color("ff1111"))
		}
		if index+m.visibleOffset == m.curIndex {
			colored.Reverse()
		}
		contentLines := strings.Split(wordwrap.String(colored.Styled(item.content), contentWidth), "\n") // QUEST why does colored.Styled needs the argument?

		// is set prepend firstline with linenumber
		if m.absoluteNumber || m.relativeNumber {
			holeString.WriteString(fmt.Sprintf("%"+fmt.Sprint(padTo)+"d"+m.seperator, m.visibleOffset+index))
		}
		// Only handel lines that are visible
		if visLines+len(contentLines) >= m.Viewport.Height {
			break out
		}
		// Write first line
		holeString.WriteString(contentLines[0])
		holeString.WriteString("\n")

		visLines++
		if len(contentLines) == 1 || !m.wrap {
			continue
		}
		// Write wraped lines
		for _, line := range contentLines[1:] {
			holeString.WriteString("\n" + strings.Repeat(" ", padTo) + m.seperator) // Pad line
			holeString.WriteString(line)                                            // write line
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
