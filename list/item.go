package list

import (
	"fmt"
	"github.com/muesli/reflow/wordwrap"
	"strings"
)

// Item are Items used in the list Model
// to hold the Content represented as a string
type item struct {
	selected bool
	value    fmt.Stringer
	id       int
}

// itemLines returns the lines of the item string value wrapped to the according content-width
func (m *Model) itemLines(i item) []string {
	var preWidth, sufWidth int
	if m.PrefixGen != nil {
		preWidth = m.PrefixGen.InitPrefixer(m.viewPos, m.Screen)
	}
	if m.SuffixGen != nil {
		sufWidth = m.SuffixGen.InitSuffixer(m.viewPos, m.Screen)
	}
	contentWith := m.Screen.Width - preWidth - sufWidth
	// TODO hard limit the string length
	lines := strings.Split(wordwrap.String(i.value.String(), contentWith), "\n")
	if m.Wrap != 0 && len(lines) > m.Wrap {
		return lines[:m.Wrap]
	}
	return lines
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
