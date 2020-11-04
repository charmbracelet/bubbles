package list

import (
	"bytes"
	"fmt"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	runewidth "github.com/mattn/go-runewidth"
	"github.com/muesli/reflow/wordwrap"
	"strings"
)

// Model is a bubbletea List of strings with get wraped
type Model struct {
	seperator      string
	listItems      []listItem
	relativeNumber bool
	absoluteNumber bool
	curIndex       int
	visibleItems   []listItem
	viewport       viewport.Model
	visibleOffset  int
}

type listItem struct {
	selected bool
	content  string
}

// View renders the Lst to a (displayable) string
func (m *Model) View() string {
	width := m.viewport.Width
	max := m.viewport.Height
	if m.curIndex > max {
		max = m.curIndex
	}

	padTo := runewidth.StringWidth(fmt.Sprintf("%d", max))
	sep := runewidth.StringWidth(fmt.Sprintf(m.seperator))

	contentWidth := width - (padTo + sep)
	if contentWidth <= 0 {
		panic("Can't display with zero width for content")
	}
	var holeString bytes.Buffer
	for index, item := range m.visibleItems {
		content := wordwrap.String(item.content, contentWidth)
		contentLines := strings.SplitN(content, "\n", 1)                                                  //split into first line and the rest
		holeString.WriteString(fmt.Sprintf("%"+fmt.Sprint(padTo)+"d"+m.seperator, m.visibleOffset+index)) // prepend firstline with linenumber
		holeString.WriteString(contentLines[0])                                                           //write firstline
		if len(contentLines) == 1 {
			continue
		}
		holeString.WriteString(strings.ReplaceAll(contentLines[1], "\n", "\n"+strings.Repeat(" ", padTo)+m.seperator)) // Pad all remaning lines

	}
	return holeString.String()
}

// Update changes the Model of the List according to the messages recieved
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		}
	}
	return m, nil
}

// Init does nothing
func (m Model) Init() tea.Cmd {
	return nil
}
