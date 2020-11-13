package list

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
	"testing"
)

// TestViewPanic runs the View on various model list model states that should yield a panic
func TestNoAreaPanic(t *testing.T) {
	m := NewModel()
	var panicMsg interface{}
	defer func() {
		panicMsg, _ = recover().(string)
		if panicMsg != "Can't display with zero width or hight of Viewport" {
			t.Errorf("No Panic or wrong panic message: %s", panicMsg)
		}
	}()
	m.View()
}

// TestNoContentSpacePanic Fails if after the Prefixer Width is subtracted there is still spaces left for contnent when there shouldent be
func TestNoContentSpacePanic(t *testing.T) {
	m := NewModel()
	m.Screen = ScreenInfo{Width: 1, Height: 50}
	m.PrefixGen = NewPrefixer()
	m.SuffixGen = NewSuffixer()
	var panicMsg interface{}
	defer func() {
		panicMsg, _ = recover().(string)
		if panicMsg != "Can't display with zero width for content" {
			t.Errorf("No Panic or wrong panic message: %s", panicMsg)
		}
	}()
	m.View()
}

// TestLines test if the models Lines methode returns the write amount of lines
func TestEmptyLines(t *testing.T) {
	m := NewModel()
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init should do nothing") // yet
	}
	m.Screen = ScreenInfo{Height: 50, Width: 80}
	if len(m.Lines()) != 0 {
		t.Error("A list with no entrys should return no lines.")
	}
	m.Sort()
	if len(m.Lines()) != 0 {
		t.Error("A list with no entrys should return no lines.")
	}
}

// TestBasicsLines test lines without linebreaks and with content shorter than the max content-width.
func TestBasicsLines(t *testing.T) {
	m := NewModel()
	m.Screen = ScreenInfo{Height: 50, Width: 80, Profile: 0} // No color
	m.PrefixGen = NewPrefixer()
	m.SuffixGen = NewSuffixer()
	// first two swaped
	itemList := MakeStringerList([]string{"2", "1", "3", "4", "5", "6", "7", "8", "9"})
	m.AddItems(itemList)
	// Sort them
	m.Sort()
	// swap them again
	m.MoveItem(1)
	// should be the like the beginning
	sorteditemList := m.GetAllItems()

	// make sure all itemList get processed
	shorter, longer := sorteditemList, itemList
	if len(itemList) > len(longer) {
		shorter, longer = itemList, sorteditemList
	}

	// Process/check all itemList
	for c, item := range longer {
		if item.String() != shorter[c].String() {
			t.Error("something basic failed")
		}
	}

	m.Top()
	out := m.Lines()
	if len(out) > 50 {
		t.Errorf("Lines should never have more (%d) lines than Screen has lines: %d", len(out), m.Screen.Height)
	}

	light := "\x1b[7m"
	cur := ">"
	for i, line := range out {
		// Check Prefixes
		num := fmt.Sprintf("%d", i)
		prefix := light + strings.Repeat(" ", 2-len(num)) + num + " ╭" + cur
		if !strings.HasPrefix(line, prefix) {
			t.Errorf("The prefix of the line:\n%s\n with linenumber %d should be:\n%s\n", line, i, prefix)
		}
		cur = " "
		light = ""
	}
}

// TestWrappedLines test a simple case of many items with linebreaks.
func TestWrappedLines(t *testing.T) {
	m := NewModel()
	m.PrefixGen = NewPrefixer()
	m.SuffixGen = NewSuffixer()
	m.Screen = ScreenInfo{Height: 50, Width: 80}
	m.AddItems(MakeStringerList([]string{"\n0", "1\n2", "3\n4", "5\n6", "7\n8"}))
	m.viewPos = ViewPos{LineOffset: 1}

	out := m.Lines()
	wrap, sep := "│", "╭"
	num := "\x1b[7m  "
	for i, line := range out {
		if i%2 == 1 {
			num = fmt.Sprintf(" %1d", (i/2)+1)
		}
		prefix := fmt.Sprintf("%s %s %d", num, wrap, i)
		if !strings.HasPrefix(line, prefix) {
			t.Errorf("The prefix of the line:\n'%s'\n with linenumber %d should be:\n'%s'\n", line, i, prefix)
		}
		wrap, sep = sep, wrap
		num = "  "
	}
}

// TestMultiLineBreaks test one selected item
func TestMultiLineBreaks(t *testing.T) {
	m := NewModel()
	m.PrefixGen = NewPrefixer()
	m.SuffixGen = NewSuffixer()
	m.Screen = ScreenInfo{Height: 50, Width: 80}
	m.AddItems(MakeStringerList([]string{"\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n"}))
	m.MarkSelected(0, true)
	out := m.Lines()
	prefix := "\x1b[7m 0*╭>"
	for i, line := range out {
		if !strings.HasPrefix(line, prefix) {
			t.Errorf("The prefix of the line:\n'%s'\n with linenumber %d should be:\n'%s'\n", line, i, prefix)
		}
		prefix = "\x1b[7m  *│ "
	}
}

// TestUpdateKeys test if the key send to the Update function work properly
func TestUpdateKeys(t *testing.T) {
	m := NewModel()
	m.Screen = ScreenInfo{Height: 50, Width: 80}

	// Quit massages
	_, cmd := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyCtrlC}))
	if cmd() != tea.Quit() {
		t.Errorf("ctrl-c should result in Quit message, not into: %#v", cmd)
	}

	_, cmd = m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'q'}}))
	if cmd() != tea.Quit() {
		t.Errorf("'q' should result in Quit message, not into: %#v", cmd)
	}

	// Movements
	m.AddItems(MakeStringerList([]string{"\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n"}))
	newModel, cmd := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'j'}}))
	m, _ = newModel.(Model)
	if m.viewPos.Cursor != 1 && cmd == nil {
		t.Errorf("key 'j' should have nil command but got: '%#v' and move the Cursor down to index one, but got: %d", cmd, m.viewPos.Cursor)
	}

}

// TestUnfocused should make sure that the update does not change anything if model is not focused
func TestUnfocused(t *testing.T) {
	m := NewModel()
	m.focus = false
	newModel, cmd := m.Update(nil)
	oldM := fmt.Sprintf("%#v", newModel)
	newM := fmt.Sprintf("%#v", m)
	if oldM != newM || cmd != nil {
		t.Errorf("Update changes unfocused Model form:\n%#v\nto:\n%#v or returns a not nil command: %#v", oldM, newM, cmd)
	}
}
