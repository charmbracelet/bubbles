package textarea

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNew(t *testing.T) {
	textarea := newTextArea()
	view := textarea.View()

	if !strings.Contains(view, ">") {
		t.Log(view)
		t.Error("Text area did not render the prompt")
	}

	if !strings.Contains(view, "World!") {
		t.Log(view)
		t.Error("Text area did not render the placeholder")
	}
}

func TestInput(t *testing.T) {
	textarea := newTextArea()

	input := "foo"

	for _, k := range []rune(input) {
		textarea, _ = textarea.Update(keyPress(k))
	}

	view := textarea.View()

	if !strings.Contains(view, input) {
		t.Log(view)
		t.Error("Text area did not render the input")
	}

	if textarea.col != len(input) {
		t.Log(view)
		t.Error("Text area did not move the cursor to the correct position")
	}
}

func TestSoftWrap(t *testing.T) {
	textarea := newTextArea()
	textarea.Prompt = ""
	textarea.ShowLineNumbers = false
	textarea.SetWidth(5)
	textarea.SetHeight(5)
	textarea.CharLimit = 60

	textarea, _ = textarea.Update(nil)

	input := "foo bar baz"

	for _, k := range []rune(input) {
		textarea, _ = textarea.Update(keyPress(k))
	}

	view := textarea.View()

	for _, word := range strings.Split(input, " ") {
		if !strings.Contains(view, word) {
			t.Log(view)
			t.Error("Text area did not render the input")
		}
	}

	// Due to the word wrapping, each word will be on a new line and the
	// text area will look like this:
	//
	// > foo
	// > bar
	// > baz█
	//
	// However, due to soft-wrapping the column will still be at the end of the line.
	if textarea.row != 0 || textarea.col != len(input) {
		t.Log(view)
		t.Error("Text area did not move the cursor to the correct position")
	}
}

func TestCharLimit(t *testing.T) {
	textarea := newTextArea()

	// First input (foo bar) should be accepted as it will fall within the
	// CharLimit. Second input (baz) should not appear in the input.
	input := []string{"foo bar", "baz"}
	textarea.CharLimit = len(input[0])

	for _, k := range []rune(strings.Join(input, " ")) {
		textarea, _ = textarea.Update(keyPress(k))
	}

	view := textarea.View()
	if strings.Contains(view, input[1]) {
		t.Log(view)
		t.Error("Text area should not include input past the character limit")
	}
}

func TestVerticalScrolling(t *testing.T) {
	textarea := newTextArea()
	textarea.Prompt = ""
	textarea.ShowLineNumbers = false
	textarea.SetHeight(1)
	textarea.SetWidth(20)
	textarea.CharLimit = 100

	textarea, _ = textarea.Update(nil)

	input := "This is a really long line that should wrap around the text area."

	for _, k := range []rune(input) {
		textarea, _ = textarea.Update(keyPress(k))
	}

	view := textarea.View()

	// The view should contain the first "line" of the input.
	if !strings.Contains(view, "This is a really") {
		t.Log(view)
		t.Error("Text area did not render the input")
	}

	// But we should be able to scroll to see the next line.
	// Let's scroll down for each line to view the full input.
	lines := []string{
		"long line that",
		"should wrap around",
		"the text area.",
	}
	for _, line := range lines {
		textarea.viewport.LineDown(1)
		view = textarea.View()
		if !strings.Contains(view, line) {
			t.Log(view)
			t.Error("Text area did not render the correct scrolled input")
		}
	}
}

func TestWordWrapOverflowing(t *testing.T) {
	// An interesting edge case is when the user enters many words that fill up
	// the text area and then goes back up and inserts a few words which causes
	// a cascading wrap and causes an overflow of the last line.
	//
	// In this case, we should not let the user insert more words if, after the
	// entire wrap is complete, the last line is overflowing.
	textarea := newTextArea()

	textarea.SetHeight(3)
	textarea.SetWidth(20)
	textarea.CharLimit = 500

	textarea, _ = textarea.Update(nil)

	input := "Testing Testing Testing Testing Testing Testing Testing Testing"

	for _, k := range []rune(input) {
		textarea, _ = textarea.Update(keyPress(k))
		textarea.View()
	}

	// We have essentially filled the text area with input.
	// Let's see if we can cause wrapping to overflow the last line.
	textarea.row = 0
	textarea.col = 0

	input = "Testing"

	for _, k := range []rune(input) {
		textarea, _ = textarea.Update(keyPress(k))
		textarea.View()
	}

	lastLineWidth := textarea.LineInfo().Width
	if lastLineWidth > 20 {
		t.Log(lastLineWidth)
		t.Log(textarea.View())
		t.Fail()
	}
}

func TestValueSoftWrap(t *testing.T) {
	textarea := newTextArea()
	textarea.SetWidth(16)
	textarea.SetHeight(10)
	textarea.CharLimit = 500

	textarea, _ = textarea.Update(nil)

	input := "Testing Testing Testing Testing Testing Testing Testing Testing"

	for _, k := range []rune(input) {
		textarea, _ = textarea.Update(keyPress(k))
		textarea.View()
	}

	value := textarea.Value()
	if value != input {
		t.Log(value)
		t.Log(input)
		t.Fatal("The text area does not have the correct value")
	}
}

func TestSetValue(t *testing.T) {
	textarea := newTextArea()
	textarea.SetValue(strings.Join([]string{"Foo", "Bar", "Baz"}, "\n"))

	if textarea.row != 2 && textarea.col != 3 {
		t.Log(textarea.row, textarea.col)
		t.Fatal("Cursor Should be on row 2 column 3 after inserting 2 new lines")
	}

	value := textarea.Value()
	if value != "Foo\nBar\nBaz" {
		t.Fatal("Value should be Foo\nBar\nBaz")
	}

	// SetValue should reset text area
	textarea.SetValue("Test")
	value = textarea.Value()
	if value != "Test" {
		t.Log(value)
		t.Fatal("Text area was not reset when SetValue() was called")
	}
}

func TestInsertString(t *testing.T) {
	textarea := newTextArea()

	// Insert some text
	input := "foo baz"

	for _, k := range []rune(input) {
		textarea, _ = textarea.Update(keyPress(k))
	}

	// Put cursor in the middle of the text
	textarea.col = 4

	textarea.InsertString("bar ")

	value := textarea.Value()
	if value != "foo bar baz" {
		t.Log(value)
		t.Fatal("Expected insert string to insert bar between foo and baz")
	}
}

func TestCanHandleEmoji(t *testing.T) {
	textarea := newTextArea()
	input := "🧋"

	for _, k := range []rune(input) {
		textarea, _ = textarea.Update(keyPress(k))
	}

	value := textarea.Value()
	if value != input {
		t.Log(value)
		t.Fatal("Expected emoji to be inserted")
	}

	input = "🧋🧋🧋"

	textarea.SetValue(input)

	value = textarea.Value()
	if value != input {
		t.Log(value)
		t.Fatal("Expected emoji to be inserted")
	}

	if textarea.col != 3 {
		t.Log(textarea.col)
		t.Fatal("Expected cursor to be on the third character")
	}

	if charOffset := textarea.LineInfo().CharOffset; charOffset != 6 {
		t.Log(charOffset)
		t.Fatal("Expected cursor to be on the sixth character")
	}
}

func TestVerticalNavigationKeepsCursorHorizontalPosition(t *testing.T) {
	textarea := newTextArea()
	textarea.SetWidth(20)

	textarea.SetValue(strings.Join([]string{"你好你好", "Hello"}, "\n"))

	textarea.row = 0
	textarea.col = 2

	// 你好|你好
	// Hell|o
	// 1234|

	// Let's imagine our cursor is on the first line where the pipe is.
	// We press the down arrow to get to the next line.
	// The issue is that if we keep the cursor on the same column, the cursor will jump to after the `e`.
	//
	// 你好|你好
	// He|llo
	//
	// But this is wrong because visually we were at the 4th character due to
	// the first line containing double-width runes.
	// We want to keep the cursor on the same visual column.
	//
	// 你好|你好
	// Hell|o
	//
	// This test ensures that the cursor is kept on the same visual column by
	// ensuring that the column offset goes from 2 -> 4.

	lineInfo := textarea.LineInfo()
	if lineInfo.CharOffset != 4 || lineInfo.ColumnOffset != 2 {
		t.Log(lineInfo.CharOffset)
		t.Log(lineInfo.ColumnOffset)
		t.Fatal("Expected cursor to be on the fourth character because there are two double width runes on the first line.")
	}

	downMsg := tea.KeyMsg{Type: tea.KeyDown, Alt: false, Runes: []rune{}}
	textarea, _ = textarea.Update(downMsg)

	lineInfo = textarea.LineInfo()
	if lineInfo.CharOffset != 4 || lineInfo.ColumnOffset != 4 {
		t.Log(lineInfo.CharOffset)
		t.Log(lineInfo.ColumnOffset)
		t.Fatal("Expected cursor to be on the fourth character because we came down from the first line.")
	}
}

func TestVerticalNavigationShouldRememberPositionWhileTraversing(t *testing.T) {
	textarea := newTextArea()
	textarea.SetWidth(40)

	// Let's imagine we have a text area with the following content:
	//
	// Hello
	// World
	// This is a long line.
	//
	// If we are at the end of the last line and go up, we should be at the end
	// of the second line.
	// And, if we go up again we should be at the end of the first line.
	// But, if we go back down twice, we should be at the end of the last line
	// again and not the fifth (length of second line) character of the last line.
	//
	// In other words, we should remember the last horizontal position while
	// traversing vertically.

	textarea.SetValue(strings.Join([]string{"Hello", "World", "This is a long line."}, "\n"))

	// We are at the end of the last line.
	if textarea.col != 20 || textarea.row != 2 {
		t.Log(textarea.col)
		t.Fatal("Expected cursor to be on the 20th character of the last line")
	}

	// Let's go up.
	upMsg := tea.KeyMsg{Type: tea.KeyUp, Alt: false, Runes: []rune{}}
	textarea, _ = textarea.Update(upMsg)

	// We should be at the end of the second line.
	if textarea.col != 5 || textarea.row != 1 {
		t.Log(textarea.col)
		t.Fatal("Expected cursor to be on the 5th character of the second line")
	}

	// And, again.
	textarea, _ = textarea.Update(upMsg)

	// We should be at the end of the first line.
	if textarea.col != 5 || textarea.row != 0 {
		t.Log(textarea.col)
		t.Fatal("Expected cursor to be on the 5th character of the first line")
	}

	// Let's go down, twice.
	downMsg := tea.KeyMsg{Type: tea.KeyDown, Alt: false, Runes: []rune{}}
	textarea, _ = textarea.Update(downMsg)
	textarea, _ = textarea.Update(downMsg)

	// We should be at the end of the last line.
	if textarea.col != 20 || textarea.row != 2 {
		t.Log(textarea.col)
		t.Fatal("Expected cursor to be on the 20th character of the last line")
	}

	// Now, for correct behavior, if we move right or left, we should forget
	// (reset) the saved horizontal position. Since we assume the user wants to
	// keep the cursor where it is horizontally. This is how most text areas
	// work.

	textarea, _ = textarea.Update(upMsg)
	leftMsg := tea.KeyMsg{Type: tea.KeyLeft, Alt: false, Runes: []rune{}}
	textarea, _ = textarea.Update(leftMsg)

	if textarea.col != 4 || textarea.row != 1 {
		t.Log(textarea.col)
		t.Fatal("Expected cursor to be on the 5th character of the second line")
	}

	// Going down now should keep us at the 4th column since we moved left and
	// reset the horizontal position saved state.
	textarea, _ = textarea.Update(downMsg)
	if textarea.col != 4 || textarea.row != 2 {
		t.Log(textarea.col)
		t.Fatal("Expected cursor to be on the 4th character of the last line")
	}
}

func TestRendersEndOfLineBuffer(t *testing.T) {
	textarea := newTextArea()
	textarea.ShowLineNumbers = true
	textarea.SetWidth(20)

	view := textarea.View()
	if !strings.Contains(view, "~") {
		t.Log(view)
		t.Fatal("Expected to see a tilde at the end of the line")
	}
}

func newTextArea() Model {
	textarea := New()

	textarea.Prompt = "> "
	textarea.Placeholder = "Hello, World!"

	textarea.Focus()

	textarea, _ = textarea.Update(nil)

	return textarea
}

func keyPress(key rune) tea.Msg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}, Alt: false}
}
