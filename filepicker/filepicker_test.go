package filepicker

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"strings"
	"testing"
	"testing/fstest"
)

func TestFS(t *testing.T) {
	fp := New()
	fp.FS = fstest.MapFS{
		"bubbles/help.txt": {Data: []byte("1")},
		"bubbles/list.txt": {Data: []byte("1")},
		"charm.sh":         {Data: []byte("   4")},
		"hello.txt":        {Data: []byte(" 2")},
		"huh.txt":          {Data: []byte("  3")},
	}

	fp.SetHeight(10)

	cmd := fp.Init()
	fp, _ = fp.Update(cmd())

	lines := strings.Split(fp.View(), "\n")
	for lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	expected := []string{"bubbles", "charm.sh", "hello.txt", "huh.txt"}
	if len(lines) != len(expected) {
		t.Fatalf("len(lines) != len(expected): got %d, want %d", len(lines), len(expected))
	}
	for i, line := range lines {
		contains := expected[i]
		if got := line; !strings.Contains(got, contains) {
			t.Errorf("View() line %d = %v; must contains %v", i, got, contains)
		}
	}

	expected = []string{"0B", "4B", "2B", "3B"}
	if len(lines) != len(expected) {
		t.Fatalf("len(lines) != len(expected): got %d, want %d", len(lines), len(expected))
	}
	for i, line := range lines {
		contains := expected[i]
		if got := line; !strings.Contains(got, contains) {
			t.Errorf("View() line %d = %v; must contains %v", i, got, contains)
		}
	}

	fp, cmd = fp.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	fp, _ = fp.Update(cmd())

	lines = strings.Split(fp.View(), "\n")
	for lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	expected = []string{"help.txt", "list.txt"}
	if len(lines) != len(expected) {
		t.Fatalf("len(lines) != len(expected): got %d, want %d", len(lines), len(expected))
	}
	for i, line := range lines {
		contains := expected[i]
		if got := line; !strings.Contains(got, contains) {
			t.Errorf("View() line %d = %v; must contains %v", i, got, contains)
		}
	}
}
