package filepicker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// newTestDir creates a directory containing a subdirectory with a file in it.
func newTestDir(tb testing.TB) string {
	tb.Helper()
	root := tb.TempDir()
	sub := filepath.Join(root, "subdir")
	if err := os.Mkdir(sub, 0o755); err != nil {
		tb.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "file.go"), []byte("package main"), 0o644); err != nil {
		tb.Fatal(err)
	}
	return root
}

// update sends msg to the model and runs the returned command, if any, feeding
// its message back into the model.
func update(tb testing.TB, m Model, msg tea.Msg) Model {
	tb.Helper()
	m, cmd := m.Update(msg)
	if cmd != nil {
		m, _ = m.Update(cmd())
	}
	return m
}

// TestNavigateWithoutWindowSize asserts that a file picker which never receives
// a tea.WindowSizeMsg still renders its entries after entering a directory.
func TestNavigateWithoutWindowSize(t *testing.T) {
	m := New()
	m.CurrentDirectory = newTestDir(t)

	m = update(t, m, m.Init()())
	if got := m.View(); !strings.Contains(got, "subdir") {
		t.Fatalf("expected view to contain %q, got %q", "subdir", got)
	}

	m = update(t, m, tea.KeyPressMsg{Code: 'l', Text: "l"})
	if !strings.HasSuffix(m.CurrentDirectory, "subdir") {
		t.Fatalf("expected to descend into subdir, got %q", m.CurrentDirectory)
	}
	if m.maxIdx < m.minIdx {
		t.Fatalf("maxIdx (%d) must not be less than minIdx (%d)", m.maxIdx, m.minIdx)
	}
	if got := m.View(); !strings.Contains(got, "file.go") {
		t.Fatalf("expected view to contain %q, got %q", "file.go", got)
	}
}

// TestDidSelectFileRequiresPath asserts that a selection is only reported once
// a path has actually been selected.
func TestDidSelectFileRequiresPath(t *testing.T) {
	m := New()
	m.CurrentDirectory = filepath.Join(newTestDir(t), "subdir")
	m.AllowedTypes = []string{".txt"}

	m = update(t, m, m.Init()())

	enter := tea.KeyPressMsg{Code: tea.KeyEnter}
	if didSelect, path := m.DidSelectFile(enter); didSelect {
		t.Fatalf("expected no selection for a disallowed type, got %q", path)
	}
	if didSelect, path := m.DidSelectDisabledFile(enter); didSelect {
		t.Fatalf("expected no disabled selection before a path is set, got %q", path)
	}

	m.AllowedTypes = []string{".go"}
	m = update(t, m, enter)
	didSelect, path := m.DidSelectFile(enter)
	if !didSelect {
		t.Fatal("expected file.go to be selected")
	}
	if !strings.HasSuffix(path, "file.go") {
		t.Fatalf("expected path to end in file.go, got %q", path)
	}
}
