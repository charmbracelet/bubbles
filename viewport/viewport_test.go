package viewport

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestSetContent(t *testing.T) {
	// Normal case
	m := New(10, 10) // Create a new Model with width 10 and height 10
	m.SetContent("This is a test string")
	if len(m.lines) != 3 {
		t.Errorf("Expected 3 line, but got %d", len(m.lines))
	}

	// Edge case: empty string
	m.SetContent("")
	if len(m.lines) != 1 {
		t.Errorf("Expected 1 lines, but got %d", len(m.lines))
	}

	// Edge case: single newline
	m.SetContent("\n")
	if len(m.lines) != 2 {
		t.Errorf("Expected 2 line, but got %d", len(m.lines))
	}

	// Edge case: multiple newlines
	m.SetContent("\n\n\n")
	if len(m.lines) != 4 {
		t.Errorf("Expected 4 lines, but got %d", len(m.lines))
	}

	// Extreme case: very long string
	longString := strings.Repeat("This is a test string ", 1000)
	m.SetContent(longString)
	if len(m.lines) != (3 * 1000) { // Depending on the width of the Model, this might wrap to multiple lines
		t.Errorf("Expected 3000 lines, but got %d", len(m.lines))
	}

	// Extreme case: string with ANSI escape codes
	ansiString := "\x1b[31mThis is a test string\x1b[0m"
	m.SetContent(ansiString)
	if len(m.lines) != 3 {
		t.Errorf("Expected 3 line, but got %d", len(m.lines))
	}

	// Extreme case: 2-width characters (Japanese characters)
	japaneseString :=
		"this is a really long text that should wrap around the viewport\n" +
			"これはビューポートを囲む必要がある非常に長い日本語のテキストです"
	m.SetContent(japaneseString)
	if len(m.lines) != 15 {
		t.Errorf("Expected 15 line, but got %d", len(m.lines))
	}
}

func escapeString(s string) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	s = strings.ReplaceAll(s, " ", "\\s")
	return s
}

func TestView(t *testing.T) {
	// Normal case
	m := New(10, 10) // Create a new Model with width 10 and height 10
	m.SetContent("This is a test string")
	view := m.View()
	expected := "" +
		"This is a \n" +
		"test      \n" +
		"string    "
	if !strings.HasPrefix(view, expected) {
		t.Errorf("Expected view to start with a newline, but got %s", escapeString(view))
	}

	// Edge case: empty content
	m.SetContent("")
	view = m.View()
	expected = "" +
		"          \n"
	if !strings.HasPrefix(view, expected) {
		t.Errorf("Expected empty view, but got %s", escapeString(view))
	}

	// Edge case: single newline
	m.SetContent("\n")
	view = m.View()
	expected = "" +
		"          \n"
	if !strings.HasPrefix(view, expected) {
		t.Errorf("Expected view to be a single newline, but got %s", escapeString(view))
	}

	// Edge case: multiple newlines
	m.SetContent("\n\n\n")
	view = m.View()
	expected = "" +
		"          \n"
	if !strings.HasPrefix(view, expected) {
		t.Errorf("Expected view to contain multiple newlines, but got %s", escapeString(view))
	}

	// Extreme case: very long content
	longString := strings.Repeat("This is a test string ", 1000)
	m.SetContent(longString)
	view = m.View()
	if len(view) < (10 * 10) { // Depending on the width of the Model, this might wrap to multiple lines
		t.Errorf("Expected view to be at least 100 characters long, but got %s", escapeString(view))
	}

	// Extreme case: content with ANSI escape codes
	ansiString := "\x1b[31mThis is a test string\x1b[0m"
	m.SetContent(ansiString)
	expected = "" +
		"\x1b[31mThis is a \n" +
		"test      \n" +
		"string\x1b[0m    "
	view = m.View()
	if !strings.HasPrefix(view, expected) {
		t.Errorf("Expected view to contain ANSI escape codes, but got %s", escapeString(view))
	}

	// Test case: 2-width characters (Japanese characters)
	japaneseString :=
		"this is a really long text that should wrap around the viewport\n" +
			"これはビューポートを囲む必要がある非常に長い日本語のテキストです"
	m.SetContent(japaneseString)
	view = m.View()
	expected = "" +
		"this is a \n" +
		"really    \n" +
		"long text \n" +
		"that      \n" +
		"should    \n" +
		"wrap      \n" +
		"around the\n" +
		"viewport  \n" +
		"これはビュ\n" +
		"ーポートを"
	if !strings.HasPrefix(view, expected) {
		t.Errorf("Expected view to contain Japanese characters, but got %s", escapeString(view))
	}
}

func TestSetContent_NormalCase(t *testing.T) {
	vp := New(0, 0)
	vp.SetContent("test")

	actual := strings.Join(vp.lines, "\n")
	expected := "test"

	if actual != expected {
		t.Errorf("Error: vp.lines = %q, want %q", actual, expected)
	}
}

func assertArray(t *testing.T, actual, expected []string) {
	if len(actual) != len(expected) {
		t.Errorf("Error: actual = %d, want %d", len(actual), len(expected))
	}

	for i, actualLine := range actual {
		if len(actualLine) != len(expected[i]) {
			t.Errorf("Error: actual = %d, want %d", len(actualLine), len(expected[i]))
		}
	}

	actualStr := strings.Join(actual, "\n")
	expectedStr := strings.Join(expected, "\n")

	if actualStr != expectedStr {
		t.Errorf("Error: actual = %q, want %q", actualStr, expectedStr)
	}
}

func TestSetContent_WithLargeText(t *testing.T) {
	content := []string{
		"this is a really long text that should wrap around the viewport",
	}
	vp := New(30, 5)
	vp.SetContent(strings.Join(content, "\n"))

	assertArray(t, vp.lines, []string{
		"this is a really long text    ",
		"that should wrap around the   ",
		"viewport                      ",
	})
}

func TestSetContent_WithMultipleLineLargeTest(t *testing.T) {
	content := []string{
		"this is a really long text that should wrap around the viewport",
		"this is a really long text that should wrap around the viewport",
	}
	vp := New(30, 5)
	vp.SetContent(strings.Join(content, "\n"))

	assertArray(t, vp.lines, []string{
		"this is a really long text    ",
		"that should wrap around the   ",
		"viewport                      ",
		"this is a really long text    ",
		"that should wrap around the   ",
		"viewport                      ",
	})
}

func TestSetContent_WithStyledLargeText(t *testing.T) {
	content := []string{
		lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("this ") + "is a really long text that should wrap around the viewport",
	}
	vp := New(30, 5)
	vp.SetContent(strings.Join(content, "\n"))

	assertArray(t, vp.lines, []string{
		"this is a really long text    ",
		"that should wrap around the   ",
		"viewport                      ",
	})
}

func TestSetContent_WithLargeTextWithViewportPadding(t *testing.T) {
	content := []string{
		"this is a really long text that should wrap around the viewport",
	}
	vp := New(30, 5)
	vp.Style = lipgloss.NewStyle().Padding(1)
	vp.SetContent(strings.Join(content, "\n"))

	assertArray(t, vp.lines, []string{
		"this is a really long text  ",
		"that should wrap around the ",
		"viewport                    ",
	})
}

func TestSetContent_WithLargeTextWithWideCharacter(t *testing.T) {
	content := []string{
		"this is a really long text that should wrap around the viewport",
		"これはビューポートを囲む必要がある非常に長い日本語のテキストです",
	}
	vp := New(30, 5)
	vp.SetContent(strings.Join(content, "\n"))

	assertArray(t, vp.lines, []string{
		"this is a really long text    ",
		"that should wrap around the   ",
		"viewport                      ",
		"これはビューポートを囲む必要が",
		"ある非常に長い日本語のテキスト",
		"です                          ",
	})
}

func TestWrapContent_WidthPaddingAndMaxWidth(t *testing.T) {
	m := Model{
		Width: 10,
		Style: lipgloss.NewStyle(),
	}

	// Test case 1: Content width is less than model width
	content := "Hello"
	expected := "Hello     "
	actual := m.wrapContent(content, false)
	if actual != expected {
		t.Errorf("Expected '%s' but got '%s'", expected, actual)
	}

	// Test case 2: Content width is equal to model width
	content = "abcdefghij"
	expected = "abcdefghij"
	actual = m.wrapContent(content, false)
	if actual != expected {
		t.Errorf("Expected '%s' but got '%s'", expected, actual)
	}

	// Test case 3: Content width is greater than model width
	content = "abcdefghijk"
	expected = "" +
		"abcdefghij\n" +
		"k         "
	actual = m.wrapContent(content, false)
	if actual != expected {
		t.Errorf("Expected '%s' but got '%s'", expected, actual)
	}
}
