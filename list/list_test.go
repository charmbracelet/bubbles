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

	m.Wrap = false

	// Check Cursor position
	if i, err := m.GetCursorIndex(); i != 0 || err == nil {
		t.Errorf("the cursor index of a new Model should be '0' and not: '%d' and there should be a error: %#v", i, err)
	}

	// first two swaped
	itemList := MakeStringerList([]string{"2", "1", "3", "4", "5", "6", "7", "8", "9"})
	m.AddItems(itemList)

	m.Move(1)
	m.SetEquals(func(a, b fmt.Stringer) bool { return a.String() == b.String() })
	// Sort them
	newModel, _ := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'s'}}))
	m, _ = newModel.(Model)
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
			t.Errorf("this strings should match but dont: %q, %q", item.String(), shorter[c].String())
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
}

// Movements
func TestMovementKeys(t *testing.T) {
	m := NewModel()
	m.Wrap = false
	m.Screen = ScreenInfo{Height: 50, Width: 80}
	m.AddItems(MakeStringerList([]string{"\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n"}))

	start, finish := 0, 1
	newModel, cmd := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'j'}}))
	m, _ = newModel.(Model)
	if m.viewPos.Cursor != finish || cmd != nil {
		t.Errorf("key 'j' should have nil command but got: '%#v' and move the Cursor to index '%d', but got: %d", cmd, finish, m.viewPos.Cursor)
	}
	start, finish = 15, 14
	m.viewPos.Cursor = start
	newModel, cmd = m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'k'}}))
	m, _ = newModel.(Model)
	if m.viewPos.Cursor != finish || cmd != nil {
		t.Errorf("key 'k' should have nil command but got: '%#v' and move the Cursor to index '%d', but got: %d", cmd, finish, m.viewPos.Cursor)
	}

	start, finish = 55, 56
	m.viewPos.Cursor = start
	newModel, cmd = m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'-'}}))
	m, _ = newModel.(Model)
	if m.viewPos.Cursor != finish || cmd != nil {
		t.Errorf("key '-' should have nil command but got: '%#v' and move the Cursor to index '%d', but got: %d", cmd, finish, m.viewPos.Cursor)
	}
	m.viewPos.ItemOffset = 10
	start, finish = 15, 14
	m.viewPos.Cursor = start
	newModel, cmd = m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'+'}}))
	m, _ = newModel.(Model)
	if m.viewPos.Cursor != finish || cmd != nil {
		t.Errorf("key '+' should have nil command but got: '%#v' and move the Cursor to index '%d', but got: %d", cmd, finish, m.viewPos.Cursor)
	}
	if m.viewPos.ItemOffset != 9 {
		t.Errorf("up movement should change the Item offset to '9' but got: %d", m.viewPos.ItemOffset)
	}
	finish = m.Len() - 1
	newModel, cmd = m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'G'}}))
	m, _ = newModel.(Model)
	if m.viewPos.Cursor != finish || cmd != nil {
		t.Errorf("key 'G' should have nil command but got: '%#v' and move the Cursor to last index: '%d', but got: %d", cmd, m.Len()-1, m.viewPos.Cursor)
	}
	finish = 0
	m.viewPos.Cursor = start
	newModel, cmd = m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'g'}}))
	m, _ = newModel.(Model)
	if m.viewPos.Cursor != finish || cmd != nil {
		t.Errorf("key 'g' should have nil command but got: '%#v' and move the Cursor to index '%d', but got: %d", cmd, finish, m.viewPos.Cursor)
	}
	m.SetCursor(10)
	if m.viewPos.Cursor != 10 {
		t.Errorf("SetCursor should set the cursor to index '10' but gut '%d'", m.viewPos.Cursor)
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
	newModel, cmd := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{' '}}))
	m, _ = newModel.(Model)
	if len(m.GetSelected()) != 1 {
		t.Errorf("key ' ' should mark exactly one items as marked not: '%d'", len(m.GetSelected()))
	}
	if sel, _ := m.IsSelected(0); !sel || cmd != nil {
		t.Errorf("key ' ' should mark the current Index, but did not or command was not nil: %#v", cmd)
	}

	// invert all mark stats
	newModel, cmd = m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'v'}}))
	m, _ = newModel.(Model)
	if len(m.GetSelected()) != m.Len()-1 {
		t.Errorf("All items but one should be marked but '%d' from '%d' are marked", len(m.GetSelected()), m.Len())
	}

	// deselect all and move to top
	m.ToggleAllSelected()
	m.Top()
	// mark the first item
	newModel, cmd = m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'m'}}))
	m, _ = newModel.(Model)
	if len(m.GetSelected()) != 1 {
		t.Errorf("key 'm' should mark exactly one items as marked not: '%d'", len(m.GetSelected()))
	}
	if sel, _ := m.IsSelected(0); !sel || cmd != nil {
		t.Errorf("key 'm' should mark the current Index, but did not or command was not nil: %#v", cmd)
	}

	// Move back to top
	m.Move(-1)
	// Unmark previous marked item
	newModel, cmd = m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'M'}}))
	m, _ = newModel.(Model)
	if len(m.GetSelected()) != 0 {
		t.Errorf("no selected items should be left, but '%d' are", len(m.GetSelected()))
	}
}

// TestUnfocused should make sure that the update does not change anything if model is not focused
func TestUnfocused(t *testing.T) {
	m := NewModel()
	m.Focus()
	if !m.Focused() {
		t.Error("model should be focused but isn't")
	}
	m.UnFocus()
	// Check Cursor position
	if i, err := m.GetCursorIndex(); i != 0 || err == nil {
		t.Errorf("the cursor index of a new Model should be '0' and not: '%d' and there should be a NotFocused error: %#v", i, err)
	}

	newModel, cmd := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{' '}}))
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

// TestItemUpdater test if items get updated
func TestItemUpdater(t *testing.T) {
	m := NewModel()
	old := MakeStringerList([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"})
	m.AddItems(old)
	m.UpdateAllItems(func(in fmt.Stringer) fmt.Stringer { return StringItem("-") })
	for i, content := range m.GetAllItems() {
		if content.String() != "-" {
			t.Errorf("after Updating all items should result in string '-' but got '%s' form old item: '%s'", content.String(), old[i])
		}
	}
	m.Bottom()
	m.ToggleSelect(-26)
	m.UpdateSelectedItems(func(in fmt.Stringer) fmt.Stringer { return StringItem("_") })

	for i, content := range m.GetAllItems() {
		if content.String() != "_" {
			t.Errorf("after Updating selected (all) items should result in string '_' but got '%s' form old item: '%s'", content.String(), old[i])
		}
	}
}

// TestWithinBorder test if indexes are within the listborders
func TestWithinBorder(t *testing.T) {
	m := NewModel()
	if m.CheckWithinBorder(0) {
		t.Error("a empty list has no item '0', should return 'false'")
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

// TestKeepVisibleWrap test the private helper function of KeepVisible
func TestKeepVisibleWrap(t *testing.T) {
	m := NewModel()
	m.Screen = ScreenInfo{Height: 50, Width: 80}
	m.AddItems(MakeStringerList([]string{"\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n"}))
	type test struct {
		oldView ViewPos
		target  int
		newView ViewPos
	}
	toTest := []test{
		{ViewPos{0, 0, 3}, -2, ViewPos{0, 0, 0}},     // infront of list
		{ViewPos{0, 0, 3}, 2, ViewPos{0, 0, 2}},      // begin of list and upper border
		{ViewPos{4, 0, 12}, 8, ViewPos{5, 1, 8}},     // Middel of list and upper border
		{ViewPos{5, 0, 15}, 0, ViewPos{0, 0, 0}},     // beginning
		{ViewPos{15, 1, 14}, 19, ViewPos{15, 1, 19}}, // Middel
		{ViewPos{0, 0, 0}, 25, ViewPos{3, 0, 25}},    // pass of lower border
		{ViewPos{0, 0, 0}, 100, ViewPos{49, 0, 71}},  // pass of lower border
	}
	for i, tCase := range toTest {
		m.viewPos = tCase.oldView
		if g := m.keepVisibleWrap(tCase.target); g != tCase.newView {
			t.Errorf("In Test number: %d, the returned ViewPos is wrong:\n'%#v' and should be:\n'%#v' after requesting target: %d", i, g, tCase.newView, tCase.target)
		}
	}
}

// TestSelectFunctions test if the function that handel the selected state of items work proper
func TestSelectFunctions(t *testing.T) {
	m := NewModel()
	err1 := m.ToggleSelect(-1)
	err2 := m.MarkSelected(-1, true)
	if err1 == nil || err2 == nil {
		t.Error("cant toggle no items")
	}
	m.AddItems(MakeStringerList([]string{""}))
	err3 := m.ToggleSelect(0)
	if ok, err4 := m.IsSelected(0); !ok || err3 != nil || err4 != nil {
		t.Errorf("Item should be selected after toggle or no error should be returned: '%#v' or '%#v'", err3, err4)
	}
	err5 := m.MarkSelected(-1, false)
	sel, err6 := m.IsSelected(0)
	if err5 == nil || err6 != nil || sel {
		t.Errorf("Item should not be selected after marking it false, error should be not nil: '%#v' and other error should be be nil '%#v'", err5, err6)
	}
	err7 := m.MarkSelected(m.Len()+1, false)
	if err7 == nil {
		t.Error("MarkSelected should fail if position is beyond list end")
	}
	_, err8 := m.IsSelected(m.Len())
	if err8 == nil {
		t.Error("error Should not be nil after trying to check selected state beyond list end")
	}
	m.viewPos.Cursor = m.Len() - 1
	err9 := m.ToggleSelect(1)
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
