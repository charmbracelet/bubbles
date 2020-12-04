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

	m.Wrap = 1

	// Check Cursor position
	if i, err := m.GetCursorIndex(); i != 0 || err == nil {
		t.Errorf("the cursor index of a new Model should be '0' and not: '%d' and there should be a error: %#v", i, err)
	}

	// first two swaped
	orgList := MakeStringerList([]string{"2", "1", "3", "4", "5", "6", "7", "8", "9"})
	m.AddItems(orgList)

	m.MoveCursor(1)
	// Sort them
	m.Sort()
	// swap them again
	m.MoveItem(1)
	// should be the like the beginning
	sortedItemList := m.GetAllItems()

	if len(orgList) != len(sortedItemList) {
		t.Errorf("the list should not change size")
	}

	// Process/check all orgList
	for c, item := range orgList {
		if item.String() != sortedItemList[c].String() {
			t.Errorf("the old strings should match the new, but dont: %q, %q", item.String(), sortedItemList[c].String())
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
		num := fmt.Sprintf("%d", i+1)
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

	out := m.Lines()
	wrap, sep := "│", "╭"
	num := "\x1b[7m  "
	for i := 1; i < len(out); i++ {
		line := out[i]
		if i%2 == 0 {
			num = fmt.Sprintf(" %1d", (i/2)+1)
		}
		prefix := fmt.Sprintf("%s %s %d", num, wrap, i-1)
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
	m.MarkSelectCursor(0, true)
	out := m.Lines()
	prefix := "\x1b[7m 1*╭>"
	for i, line := range out {
		if !strings.HasPrefix(line, prefix) {
			t.Errorf("The prefix of the line:\n'%s'\n with linenumber %d should be:\n'%s'\n", line, i, prefix)
		}
		prefix = "\x1b[7m  *│ "
	}
}

// TestUpdateKeys test if the ctrl-c key send to the Update function work properly
func TestUpdateKeys(t *testing.T) {
	m := NewModel()
	m.Screen = ScreenInfo{Height: 50, Width: 80}

	// Quit massages
	_, cmd := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyCtrlC}))
	if cmd() != tea.Quit() {
		t.Errorf("ctrl-c should result in Quit message, not into: %#v", cmd)
	}
}

// Movements
func TestMovementKeys(t *testing.T) {
	m := NewModel()
	m.Wrap = 1
	m.Screen = ScreenInfo{Height: 50, Width: 80}
	m.AddItems(MakeStringerList([]string{"\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n"}))

	start, finish := 0, 1
	_, err := m.MoveCursor(1)
	if m.viewPos.Cursor != finish || err != nil {
		t.Errorf("'MoveCursor(1)' should have nil error but got: '%#v' and move the Cursor to index '%d', but got: %d", err, finish, m.viewPos.Cursor)
	}
	start, finish = 15, 14
	m.viewPos.Cursor = start
	_, err = m.MoveCursor(-1)
	if m.viewPos.Cursor != finish || err != nil {
		t.Errorf("'MoveCursor(-1)' should have nil error but got: '%#v' and move the Cursor to index '%d', but got: %d", err, finish, m.viewPos.Cursor)
	}

	start, finish = 55, 56
	m.viewPos.Cursor = start
	err = m.MoveItem(1)
	if m.viewPos.Cursor != finish || err != nil {
		t.Errorf("'MoveItem(1)' should have nil error but got: '%#v' and move the Cursor to index '%d', but got: %d", err, finish, m.viewPos.Cursor)
	}
	m.viewPos.LineOffset = 15
	start, finish = 15, 14
	m.viewPos.Cursor = start
	err = m.MoveItem(-1)
	if m.viewPos.Cursor != finish || err != nil {
		t.Errorf("'MoveItem(-1)' should have nil error but got: '%#v' and move the Cursor to index '%d', but got: %d", err, finish, m.viewPos.Cursor)
	}
	if m.viewPos.LineOffset != 14 {
		t.Errorf("up movement should change the Item offset to '14' but got: %d", m.viewPos.LineOffset)
	}
	finish = m.Len() - 1
	err = m.Bottom()
	if m.viewPos.Cursor != finish || err != nil {
		t.Errorf("'Bottom()' should have nil error but got: '%#v' and move the Cursor to last index: '%d', but got: %d", err, m.Len()-1, m.viewPos.Cursor)
	}
	finish = 0
	m.viewPos.Cursor = start
	err = m.Top()
	if m.viewPos.Cursor != finish || err != nil {
		t.Errorf("'Top()' should have nil error but got: '%#v' and move the Cursor to index '%d', but got: %d", err, finish, m.viewPos.Cursor)
	}
	_, err = m.SetCursor(10)
	if m.viewPos.Cursor != 10 || err != nil {
		t.Errorf("SetCursor should set the cursor to index '10' but gut '%d' and err should be nil but got '%s'", m.viewPos.Cursor, err)
	}
}

// WindowMsg
func TestWindowMsg(t *testing.T) {
	m := NewModel()

	newModel, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 50})
	m, _ = newModel.(Model)

	// Because within the Update the termenv.Profile will be set, when reciving the Windowszie, depending on currently running terminal
	// we overwrite it her to have a reproduceable test-result
	m.Screen.Profile = 0

	if cmd != nil {
		t.Errorf("comand should be nil and not: '%#v'", cmd)
	}
	soll := ScreenInfo{Width: 80, Height: 50}
	if m.Screen != soll {
		t.Errorf("Screen should be %#v and not: %#v", soll, m.Screen)
	}

}

// TestSelectKeys test the keys that change the select status of an item(s).
func TestSelectKeys(t *testing.T) {
	m := NewModel()
	m.Screen = ScreenInfo{Height: 50, Width: 80}
	m.AddItems(MakeStringerList([]string{"\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n"}))

	// Mark one and move one down
	err := m.ToggleSelectCursor(1)
	if len(m.GetAllSelected()) != 1 {
		t.Errorf("ToggleSelectCursor(1) should mark exactly one items as marked not: '%d'", len(m.GetAllSelected()))
	}
	if sel, _ := m.IsSelected(0); !sel || err != nil {
		t.Errorf("ToggleSelectCursor(1) should mark the current Index, but did not or command was not nil: %#v", err)
	}

	// invert all mark stats
	m.ToggleAllSelected()
	if len(m.GetAllSelected()) != m.Len()-1 {
		t.Errorf("All items but one should be marked but '%d' from '%d' are marked", len(m.GetAllSelected()), m.Len())
	}

	// deselect all and move to top
	m.ToggleAllSelected()
	m.Top()
	// mark the first item
	err = m.MarkSelectCursor(1, true)
	if len(m.GetAllSelected()) != 1 {
		t.Errorf("MarkSelectCursor(1, true) should mark exactly one items as marked not: '%d'", len(m.GetAllSelected()))
	}
	if sel, _ := m.IsSelected(0); !sel || err != nil {
		t.Errorf("MarkSelectCursor(1, true) should mark the current Index, but did not or error was not nil: %#v", err)
	}

	// Move back to top
	m.MoveCursor(-1)
	// Unmark previous marked item
	m.MarkSelectCursor(1, false)
	if len(m.GetAllSelected()) != 0 {
		t.Errorf("no selected items should be left, but '%d' are", len(m.GetAllSelected()))
	}
}

// TestUnfocused should make sure that the update does not change anything if model is not focused
func TestUnfocused(t *testing.T) {
	m := NewModel()
	m.Focus(true)
	if !m.Focused() {
		t.Error("model should be focused but isn't")
	}
	m.Focus(false)
	// Check Cursor position
	if i, err := m.GetCursorIndex(); i != 0 || err == nil {
		t.Errorf("the cursor index of a new Model should be '0' and not: '%d' and there should be a NotFocused error: %#v", i, err)
	}

	newModel, cmd := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'j'}}))
	oldM := fmt.Sprintf("%#v", newModel)
	newM := fmt.Sprintf("%#v", m)
	if oldM != newM || cmd != nil {
		t.Errorf("Update changes unfocused Model form:\n%#v\nto:\n%#v or returns a not nil command: %#v", oldM, newM, cmd)
	}
}

// TestGetIndex sets a equals function and searches After the index of a specific item with GetIndex
func TestGetIndex(t *testing.T) {
	m := NewModel()
	_, err := m.GetIndex(StringItem("z"))
	if err == nil {
		t.Errorf("Get Index should return a error but got nil")
	}
	m.AddItems(MakeStringerList([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}))
	m.SetEquals(func(a, b fmt.Stringer) bool { return a.String() == b.String() })
	index, err := m.GetIndex(StringItem("z"))
	if err != nil {
		t.Errorf("GetIndex should not return error: %s", err)
	}
	if index != m.Len()-1 {
		t.Errorf("GetIndex returns wrong index: '%d' instead of '%d'", index, m.Len()-1)
	}
}

// TestWithinBorder test if indexes are within the listborders
func TestWithinBorder(t *testing.T) {
	m := NewModel()
	_, err := m.ValidIndex(0)
	if _, ok := err.(NoItems); !ok {
		t.Errorf("a empty list has no item '0', should return a NoItems error, but got: %#v", err)
	}
}

// TestCopy test if if Copy returns a deep copy
func TestCopy(t *testing.T) {
	org := NewModel()
	sec := org.Copy()

	org.SetLess(func(a, b fmt.Stringer) bool { return a.String() < b.String() })

	if &org == sec {
		t.Errorf("Copy should return a deep copy but has the same pointer:\norginal: '%p', copy: '%p'", &org, sec)
	}

	if org.focus != sec.focus ||
		fmt.Sprintf("%#v", org.listItems) != fmt.Sprintf("%#v", sec.listItems) ||

		// All should be the same except the changed less function
		fmt.Sprintf("%p", org.less) == fmt.Sprintf("%p", sec.less) ||
		fmt.Sprintf("%p", org.equals) != fmt.Sprintf("%p", sec.equals) ||

		fmt.Sprintf("%#v", org.CursorOffset) != fmt.Sprintf("%#v", sec.CursorOffset) ||

		fmt.Sprintf("%#v", org.Screen) != fmt.Sprintf("%#v", sec.Screen) ||
		fmt.Sprintf("%#v", org.viewPos) != fmt.Sprintf("%#v", sec.viewPos) ||

		fmt.Sprintf("%#v", org.Wrap) != fmt.Sprintf("%#v", sec.Wrap) ||

		fmt.Sprintf("%#v", org.PrefixGen) != fmt.Sprintf("%#v", sec.PrefixGen) ||
		fmt.Sprintf("%#v", org.SuffixGen) != fmt.Sprintf("%#v", sec.SuffixGen) ||

		fmt.Sprintf("%#v", org.LineStyle) != fmt.Sprintf("%#v", sec.LineStyle) ||
		fmt.Sprintf("%#v", org.SelectedStyle) != fmt.Sprintf("%#v", sec.SelectedStyle) ||
		fmt.Sprintf("%#v", org.CurrentStyle) != fmt.Sprintf("%#v", sec.CurrentStyle) {

		t.Errorf("Copy should have same string repesentation except different less function pointer:\n orginal: '%#v'\n    copy: '%#v'", org, sec)
	}
}

// TestSetCursor tests if the LineOffset and Cursor positions are correct
func TestSetCursor(t *testing.T) {
	m := NewModel()
	m.Screen = ScreenInfo{Height: 50, Width: 80}
	m.AddItems(MakeStringerList([]string{"\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n", ""}))
	type test struct {
		oldView ViewPos
		target  int
		newView ViewPos
	}
	toTest := []test{
		// forwards
		{ViewPos{0, 0}, -2, ViewPos{5, 0}},
		{ViewPos{0, 0}, 2, ViewPos{5, 2}},
		{ViewPos{0, 4}, 8, ViewPos{8, 8}},
		{ViewPos{0, 5}, 0, ViewPos{5, 0}},
		{ViewPos{0, 0}, 19, ViewPos{38, 19}},
		{ViewPos{0, 0}, 25, ViewPos{44, 25}},
		{ViewPos{0, 0}, 100, ViewPos{44, 72}},
		// backwards
		{ViewPos{45, m.Len() - 1}, -2, ViewPos{5, 0}},
		{ViewPos{45, m.Len() - 1}, 2, ViewPos{5, 2}},
		{ViewPos{45, m.Len() - 1}, 8, ViewPos{5, 8}},
		{ViewPos{45, m.Len() - 1}, 0, ViewPos{5, 0}},
		{ViewPos{45, m.Len() - 1}, 19, ViewPos{5, 19}},
		{ViewPos{45, m.Len() - 1}, 25, ViewPos{5, 25}},
		{ViewPos{45, m.Len() - 1}, 100, ViewPos{45, 72}},
	}
	for i, tCase := range toTest {
		m.viewPos = tCase.oldView
		m.SetCursor(tCase.target)
		if m.viewPos != tCase.newView {
			t.Errorf("In Test number: %d, the returned ViewPos is wrong:\n'%#v' and should be:\n'%#v' after requesting target: %d", i, m.viewPos, tCase.newView, tCase.target)
		}
	}
}

// TestSelectFunctions test if the function that handel the selected state of items work proper
func TestSelectFunctions(t *testing.T) {
	m := NewModel()
	err1 := m.ToggleSelectCursor(-1)
	err2 := m.MarkSelectCursor(-1, true)
	if err1 == nil || err2 == nil {
		t.Error("cant toggle no items")
	}
	m.AddItems(MakeStringerList([]string{""}))
	err3 := m.ToggleSelectCursor(0)
	if ok, err4 := m.IsSelected(0); !ok || err3 != nil || err4 != nil {
		t.Errorf("Item should be selected after toggle or no error should be returned: '%#v' or '%#v'", err3, err4)
	}
	err5 := m.MarkSelectCursor(-1, false)
	sel, err6 := m.IsSelected(0)
	if err5 == nil || err6 != nil || sel {
		t.Errorf("Item should not be selected after marking it false, error should be not nil: '%#v' and other error should be be nil '%#v'", err5, err6)
	}
	err7 := m.MarkSelectCursor(m.Len()+1, false)
	if err7 == nil {
		t.Error("MarkSelectCursor should fail if position is beyond list end")
	}
	_, err8 := m.IsSelected(m.Len())
	if err8 == nil {
		t.Error("error Should not be nil after trying to check selected state beyond list end")
	}
	m.viewPos.Cursor = m.Len() - 1
	err9 := m.ToggleSelectCursor(1)
	sel, _ = m.IsSelected(m.Len() - 1)
	if _, ok := err9.(OutOfBounds); !ok || !sel {
		t.Errorf("marking the last item should give a OutOfBounds error, but got: '%s'\nand after it, it should be marked: '%t'", err9, sel)
	}
}

// TestMoveItem test wrong arguments
func TestMoveItem(t *testing.T) {
	m := NewModel()
	err := m.MoveItem(0)
	_, ok := err.(OutOfBounds)
	if !ok {
		t.Errorf("MoveItem called on a empty list should return a OutOfBounds error, but got: %s", err)
	}
	m.AddItems(MakeStringerList([]string{""}))
	err = m.MoveItem(0)
	if err != nil {
		t.Errorf("MoveItem(0) should not not return a error on a not empty list")
	}
	err = m.MoveItem(1)
	_, ok = err.(OutOfBounds)
	if !ok {
		t.Errorf("MoveItem should return a OutOfBounds error if traget is beyond list border, but got: '%s'", err)
	}
}
