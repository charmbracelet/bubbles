package filepicker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestDefaultHeightRendersOpenedDirectory(t *testing.T) {
	root := t.TempDir()
	subdir := filepath.Join(root, "subdir")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subdir, "file.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := New()
	m.CurrentDirectory = root
	m = runFilepickerCmd(t, m, m.Init())

	if view := m.View(); !strings.Contains(view, "subdir") {
		t.Fatalf("expected initial view to show subdir, got %q", view)
	}

	var cmd tea.Cmd
	m, cmd = m.Update(tea.KeyPressMsg{Code: 'l', Text: "l"})
	m = runFilepickerCmd(t, m, cmd)

	if view := m.View(); !strings.Contains(view, "file.txt") {
		t.Fatalf("expected opened directory view to show file.txt, got %q", view)
	}
}

func runFilepickerCmd(t *testing.T, m Model, cmd tea.Cmd) Model {
	t.Helper()
	if cmd == nil {
		t.Fatal("expected command")
	}

	next, _ := m.Update(cmd())
	return next
}
