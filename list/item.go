package list

import (
	"fmt"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/reflow/wordwrap"
	"strings"
)

// Item are Items used in the list Model
// to hold the Content represented as a string
type item struct {
	value fmt.Stringer
	id    int
}

// itemLines returns the lines of the item string value wrapped to the according content-width
// and the write amount of lines accoring to m.Wrap
func (m *Model) itemLines(i item, index int) []string {
	var preWidth, sufWidth int
	if m.PrefixGen != nil {
		preWidth = m.PrefixGen.InitPrefixer(i.value, index, m.viewPos, m.Screen)
	}
	if m.SuffixGen != nil {
		sufWidth = m.SuffixGen.InitSuffixer(i.value, index, m.viewPos, m.Screen)
	}
	contentWith := m.Screen.Width - preWidth - sufWidth
	// TODO hard limit the string length
	lines := strings.Split(wordwrap.String(i.value.String(), contentWith), "\n")
	if m.Wrap != 0 && len(lines) > m.Wrap {
		return lines[:m.Wrap]
	}
	return lines
}

// getItemLines surrounds the line content with the according prefix and suffix
func (m *Model) getItemLines(index, contentWidth int) ([]string, error) {
	_, err := m.ValidIndex(index)
	if err != nil {
		return nil, err
	}
	item := m.listItems[index]
	lines := m.itemLines(item, index)
	completLines := make([]string, len(lines))

	for c := 0; c < len(lines); c++ {
		lineContent := lines[c]
		// Surrounding content
		var linePrefix, lineSuffix string
		if m.PrefixGen != nil {
			linePrefix = m.PrefixGen.Prefix(c)
		}
		if m.SuffixGen != nil {
			free := contentWidth - ansi.PrintableRuneWidth(lineContent)
			if free < 0 {
				free = 0 // TODO is this nessecary?
			}
			suffix := m.SuffixGen.Suffix(c)
			if suffix != "" {
				lineSuffix = fmt.Sprintf("%s%s", strings.Repeat(" ", free), suffix)
			}
		}

		// Join all
		line := fmt.Sprintf("%s%s%s", linePrefix, lineContent, lineSuffix)

		// Highlighting of current item lines
		style := m.LineStyle
		if index == m.viewPos.Cursor {
			style = m.CurrentStyle
		}

		// Highlight and write line
		completLines[c] = style.Styled(line)
	}
	return completLines, nil
}

// StringItem is just a convenience to satisfy the fmt.Stringer interface with plain strings
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

// itemList is a struct only to make the list-Model Sortable without exposing the demanded methodes,
// cause they would undermine the event loop. (Swap would change the list without the possibility of
// returning a tea.Cmd to signal the change.
type itemList struct {
	list *[]item
	less *func(fmt.Stringer, fmt.Stringer) bool
}

func (m *itemList) Less(i, j int) bool {
	// If User does not provide less function use string comparison, but dont change m.less, to be able to see when user set one.
	if *m.less == nil {
		return (*m.list)[i].value.String() < (*m.list)[j].value.String()
	}
	return (*m.less)((*m.list)[i].value, (*m.list)[j].value)
}

func (m *itemList) Swap(i, j int) {
	(*m.list)[i], (*m.list)[j] = (*m.list)[j], (*m.list)[i]
}

func (m *itemList) Len() int {
	return len(*m.list)
}
