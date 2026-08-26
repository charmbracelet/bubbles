package textinput

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func Test_CurrentSuggestion(t *testing.T) {
	textinput := New()
	textinput.ShowSuggestions = true

	suggestion := textinput.CurrentSuggestion()
	expected := ""
	if suggestion != expected {
		t.Fatalf("Error: expected no current suggestion but was %s", suggestion)
	}

	textinput.SetSuggestions([]string{"test1", "test2", "test3"})
	suggestion = textinput.CurrentSuggestion()
	expected = ""
	if suggestion != expected {
		t.Fatalf("Error: expected no current suggestion but was %s", suggestion)
	}

	textinput.SetValue("test")
	textinput.updateSuggestions()
	textinput.nextSuggestion()
	suggestion = textinput.CurrentSuggestion()
	expected = "test2"
	if suggestion != expected {
		t.Fatalf("Error: expected first suggestion but was %s", suggestion)
	}

	textinput.Blur()
	if strings.HasSuffix(textinput.View(), "test2") {
		t.Fatalf("Error: suggestions should not be rendered when input isn't focused. expected \"> test\" but got \"%s\"", textinput.View())
	}
}

func Test_SlicingOutsideCap(t *testing.T) {
	textinput := New()
	textinput.Placeholder = "作業ディレクトリを指定してください"
	textinput.SetWidth(32)
	textinput.View()
}

func TestChinesePlaceholder(t *testing.T) {
	t.Skip("Skipping flaky test, the returned view seems incorrect. TODO: Needs investigation.")
	textinput := New()
	textinput.Placeholder = "输入消息..."
	textinput.SetWidth(20)

	got := textinput.View()
	expected := "> 输入消息...       "
	if got != expected {
		t.Fatalf("expected %q but got %q", expected, got)
	}
}

func TestPlaceholderTruncate(t *testing.T) {
	t.Skip("Skipping flaky test, the returned view seems incorrect. TODO: Needs investigation.")
	textinput := New()
	textinput.Placeholder = "A very long placeholder, or maybe not so much"
	textinput.SetWidth(10)

	got := textinput.View()
	expected := "> A very …"
	if got != expected {
		t.Fatalf("expected %q but got %q", expected, got)
	}
}

func ExampleValidateFunc() {
	creditCardNumber := New()
	creditCardNumber.Placeholder = "4505 **** **** 1234"
	creditCardNumber.Focus()
	creditCardNumber.CharLimit = 20
	creditCardNumber.SetWidth(30)
	creditCardNumber.Prompt = ""
	// This anonymous function is a valid function for ValidateFunc.
	creditCardNumber.Validate = func(s string) error {
		// Credit Card Number should a string less than 20 digits
		// It should include 16 integers and 3 spaces
		if len(s) > 16+3 {
			return fmt.Errorf("CCN is too long")
		}

		if len(s) == 0 || len(s)%5 != 0 && (s[len(s)-1] < '0' || s[len(s)-1] > '9') {
			return fmt.Errorf("CCN is invalid")
		}

		// The last digit should be a number unless it is a multiple of 4 in which
		// case it should be a space
		if len(s)%5 == 0 && s[len(s)-1] != ' ' {
			return fmt.Errorf("CCN must separate groups with spaces")
		}

		// The remaining digits should be integers
		c := strings.ReplaceAll(s, " ", "")
		_, err := strconv.ParseInt(c, 10, 64)

		return err
	}
}

func keyPress(key rune) tea.Msg {
	return tea.KeyPressMsg{Code: key, Text: string(key)}
}

func sendString(m Model, str string) Model {
	for _, k := range str {
		m, _ = m.Update(keyPress(k))
	}

	return m
}

// Cursor() must report the scroll-adjusted column (pos-offset) like View() when
// the value overflows the width.
func TestCursorXAccountsForScrollOffset(t *testing.T) {
	m := New()
	m.Focus()
	m.SetVirtualCursor(false)
	m.CharLimit = 200
	m.SetWidth(20)
	m.SetValue(strings.Repeat("a", 30))

	// Scroll to the end, then move into the middle of the viewport, where the
	// width clamp doesn't mask the wrong column.
	m.CursorEnd()
	m.SetCursor(25)

	if m.offset == 0 {
		t.Fatalf("test setup: expected a non-zero scroll offset for an overflowing input, got 0")
	}

	cur := m.Cursor()
	if cur == nil {
		t.Fatal("Cursor() returned nil")
	}

	// Oracle: the cursor must land on the column where View renders the glyph
	// for value[pos] — the prompt width plus the rendered width of the visible
	// text before the cursor (value[offset:pos]). Measured from the rendered
	// text, so it's independent of Cursor()'s own arithmetic.
	want := lipgloss.Width(m.promptView()) + lipgloss.Width(string(m.value[m.offset:m.pos]))
	if cur.X != want {
		t.Errorf("Cursor().X = %d, want %d (column where View renders the cursor glyph; pos=%d, offset=%d)",
			cur.X, want, m.pos, m.offset)
	}
}
