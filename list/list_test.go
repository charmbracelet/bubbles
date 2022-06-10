package list

import (
	"fmt"
	"io"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                          { return 1 }
func (d itemDelegate) Spacing() int                         { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m Model, index int, listItem Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)
	fmt.Fprint(w, m.Styles.TitleBar.Render(str))
}

func TestStatusBarTitle(t *testing.T) {
	assert := assert.New(t)

	list := New([]Item{
		item("foo"),
		item("bar"),
	}, itemDelegate{}, 10, 10)

	statusBar := list.statusView()

	assert.Contains(statusBar, "2 items")
}

func TestStatusBarWithoutItems(t *testing.T) {
	assert := assert.New(t)

	list := New([]Item{}, itemDelegate{}, 10, 10)

	statusBar := list.statusView()

	assert.Contains(statusBar, "No items")
}

func TestCustomStatusBarTitle(t *testing.T) {
	assert := assert.New(t)

	list := New([]Item{
		item("foo"),
		item("bar"),
	}, itemDelegate{}, 10, 10)

	list.SetStatusBarTitle("connections")
	statusBar := list.statusView()

	assert.Contains(statusBar, "2 connections")
}
