package viewport

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func Test_SetContent(t *testing.T) {
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

func Test_SetContent_WithLargeText(t *testing.T) {
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

func Test_SetContent_WithMultipleLineLargeTest(t *testing.T) {
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

func Test_SetContent_WithStyledLargeText(t *testing.T) {
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

func Test_SetContent_WithLargeTextWithViewportPadding(t *testing.T) {
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

func Test_SetContent_WithLargeTextWithWideCharacter(t *testing.T) {
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
