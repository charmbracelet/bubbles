package textinput

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestPaddingIsUnstyled(t *testing.T) {
	bg := lipgloss.NewStyle().Background(lipgloss.Color("#AFAFAF")).Foreground(lipgloss.Color("#000000"))

	for _, tt := range []struct {
		name        string
		placeholder string
		value       string
	}{
		{name: "text", value: "abc"},
		{name: "placeholder", placeholder: "xyz"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			styles := DefaultDarkStyles()
			styles.Focused.Text = bg
			styles.Focused.Placeholder = bg

			m := New()
			m.SetStyles(styles)
			m.Focus()
			m.SetWidth(12)
			m.Placeholder = tt.placeholder
			m.SetValue(tt.value)

			got := m.View()
			if !strings.HasSuffix(got, " ") {
				t.Fatalf("expected trailing padding to be unstyled, got %q", got)
			}
		})
	}
}

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

func Test_CurrentSuggestionAfterPrevious(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		suggestions []string
		expected    string
	}{
		{
			name: "no suggestions",
		},
		{
			name:        "empty value",
			suggestions: []string{"test1", "test2"},
		},
		{
			name:        "no matches",
			value:       "other",
			suggestions: []string{"test1", "test2"},
		},
		{
			name:        "one match",
			value:       "test",
			suggestions: []string{"test1"},
			expected:    "test1",
		},
		{
			name:        "multiple matches",
			value:       "test",
			suggestions: []string{"test1", "test2", "test3"},
			expected:    "test3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			textinput := New()
			textinput.ShowSuggestions = true
			textinput.SetValue(tt.value)
			textinput.SetSuggestions(tt.suggestions)
			textinput.Focus()

			textinput, _ = textinput.Update(tea.KeyPressMsg{Code: tea.KeyUp})

			if got := textinput.CurrentSuggestion(); got != tt.expected {
				t.Fatalf("expected current suggestion %q, got %q", tt.expected, got)
			}
		})
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

func TestPlaceholderFullRenderWithoutWidth(t *testing.T) {
	// Regression test for https://github.com/charmbracelet/bubbles/issues/779:
	// when Width is 0 (default), only the first character of the placeholder was shown.
	ti := New()
	ti.Placeholder = "Nickname"

	// Strip ANSI escape codes before checking, since the cursor and the rest
	// of the placeholder are styled separately.
	plain := ansi.Strip(ti.View())
	if !strings.Contains(plain, "Nickname") {
		t.Fatalf("expected placeholder %q to appear in full, got: %q", "Nickname", plain)
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

func TestShiftBackspaceDeletesCharacterBackward(t *testing.T) {
	ti := New()
	ti.Focus()
	ti.SetValue("abc")

	ti, _ = ti.Update(tea.KeyPressMsg{Code: tea.KeyBackspace, Mod: tea.ModShift})

	if got, want := ti.Value(), "ab"; got != want {
		t.Errorf("after shift+backspace got %q, want %q", got, want)
	}
}

func keyPress(key rune) tea.Msg {
	return tea.KeyPressMsg{Code: key, Text: string(key)}
}

func TestDeleteWordForwardAtWordEnd(t *testing.T) {
	tests := []struct {
		name, value, want string
		cursor            int
	}{
		{"inside word", "abc d", "a d", 1},
		{"last rune of word", "ab c", "a c", 1},
		{"last rune of value", "ab", "a", 1},
		{"on whitespace", "a bc", "a", 1},
		{"unicode word", "a界 c", "a c", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			m.Focus()
			m.SetValue(tt.value)
			m.SetCursor(tt.cursor)
			m, _ = m.Update(tea.KeyPressMsg{Code: 'd', Mod: tea.ModAlt})
			if got := m.Value(); got != tt.want {
				t.Errorf("alt+d at cursor %d in %q = %q, want %q", tt.cursor, tt.value, got, tt.want)
			}
		})
	}
}

func sendString(m Model, str string) Model {
	for _, k := range str {
		m, _ = m.Update(keyPress(k))
	}

	return m
}

type customClipboardErr struct {
	msg string
}

func (e *customClipboardErr) Error() string {
	return e.msg
}

func TestPasteErrMsgUnwrap(t *testing.T) {
	underlyingErr := &customClipboardErr{msg: "clipboard read error"}
	m := New()
	m.Focus()

	m, _ = m.Update(pasteErrMsg{underlyingErr})

	if m.Err == nil {
		t.Fatal("expected m.Err to be set, got nil")
	}

	if !errors.Is(m.Err, underlyingErr) {
		t.Errorf("expected errors.Is(m.Err, underlyingErr) to be true")
	}

	if unwrapped := errors.Unwrap(m.Err); unwrapped != underlyingErr {
		t.Errorf("expected errors.Unwrap(m.Err) to be %v, got %v", underlyingErr, unwrapped)
	}

	var target *customClipboardErr
	if !errors.As(m.Err, &target) || target != underlyingErr {
		t.Errorf("expected errors.As(m.Err, &target) to match underlying error")
	}
}

func Test_ViewRespectsWidthWithSuggestions(t *testing.T) {
	m := New()
	m.SetWidth(20)
	m.SetValue("hel")
	m.ShowSuggestions = true
	m.SetSuggestions([]string{"hello world this is a very long suggestion overflow", "second"})
	m.Focus()
	m, _ = m.Update(keyPress('l'))

	got := ansi.StringWidth(m.View())
	// prompt (2) + width (20) + cursor (1)
	if want := 23; got != want {
		t.Fatalf("expected rendered width %d but got %d: %q", want, got, m.View())
	}
}

func Test_CursorPositionWithWideCharacters(t *testing.T) {
	m := New()
	m.SetVirtualCursor(false)
	m.Focus()
	m.SetWidth(40)
	m.SetValue("あいう")

	c := m.Cursor()
	if c == nil {
		t.Fatal("expected a cursor")
	}

	// The default prompt is "> ", and the three hiragana characters occupy
	// six terminal columns, so the cursor belongs at column 8.
	const want = 2 + 6
	if c.X != want {
		t.Fatalf("expected cursor X to be %d, got %d", want, c.X)
	}
}

func Test_CursorPositionWithScrolledValue(t *testing.T) {
	m := New()
	m.SetVirtualCursor(false)
	m.Focus()
	m.SetWidth(5)
	m.SetValue("abcdefghij")
	m.SetCursor(8)

	c := m.Cursor()
	if c == nil {
		t.Fatal("expected a cursor")
	}

	// The value is wider than the input, so the view is scrolled: the
	// visible text is "fghij" and the cursor sits after "fgh", which is
	// three columns wide.
	const want = 2 + 3
	if c.X != want {
		t.Fatalf("expected cursor X to be %d, got %d", want, c.X)
	}
}

func TestSetValueRefreshesSuggestions(t *testing.T) {
	textinput := New()
	textinput.ShowSuggestions = true
	textinput.SetSuggestions([]string{"hello"})
	textinput.Focus()
	textinput, _ = textinput.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})

	textinput.SetValue("ab")
	if got := textinput.MatchedSuggestions(); len(got) != 0 {
		t.Fatalf("expected no matched suggestions but got %v", got)
	}

	textinput, _ = textinput.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if got := textinput.Value(); got != "ab" {
		t.Fatalf("expected value %q to be kept but was %q", "ab", got)
	}

	textinput.SetValue("he")
	textinput, _ = textinput.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if got := textinput.Value(); got != "hello" {
		t.Fatalf("completion = %q, want hello", got)
	}
}

func TestAcceptSuggestionAfterSetValue(t *testing.T) {
	textinput := New()
	textinput.ShowSuggestions = true
	textinput.SetSuggestions([]string{"hello"})
	textinput.Focus()
	textinput, _ = textinput.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})

	// Before this fix, SetValue left the old match shorter than the new
	// value.
	value := "a much longer value than any suggestion"
	textinput.SetValue(value)
	textinput, _ = textinput.Update(tea.KeyPressMsg{Code: tea.KeyTab})

	if got := textinput.Value(); got != value {
		t.Fatalf("expected value %q to be kept but was %q", value, got)
	}
}

func TestTypedPrefixStillAcceptsSuggestion(t *testing.T) {
	t.Parallel()
	m := New()
	m.ShowSuggestions = true
	m.SetSuggestions([]string{"hello"})
	m.Focus()
	m = sendString(m, "he")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if got := m.Value(); got != "hello" {
		t.Fatalf("completion = %q, want hello", got)
	}
}

func TestSetValueDoesNotCompleteUnmatchedInput(t *testing.T) {
	t.Parallel()
	m := New()
	m.ShowSuggestions = true
	m.SetSuggestions([]string{"hello"})
	m.Focus()
	m.SetValue("other")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if got := m.Value(); got != "other" {
		t.Fatalf("unmatched value = %q, want other", got)
	}
}
