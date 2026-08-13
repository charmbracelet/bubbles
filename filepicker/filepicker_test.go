package filepicker

import (
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestCWDDefault tests that when CWDDefault is true, the filepicker opens
// in the parent directory with the original CWD selected.
func TestCWDDefault(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create subdirectories
	projectDir := filepath.Join(tmpDir, "myproject")
	err := os.Mkdir(projectDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	otherDir := filepath.Join(tmpDir, "otherproject")
	err = os.Mkdir(otherDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	anotherDir := filepath.Join(tmpDir, "anotherproject")
	err = os.Mkdir(anotherDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Test with CWDDefault enabled
	t.Run("CWDDefault enabled", func(t *testing.T) {
		m := New()
		m.CurrentDirectory = projectDir
		m.DirAllowed = true
		m.FileAllowed = false
		m.SetHeight(10)
		m.SetCWDDefault(true) // Call SetCWDDefault before Init

		// Initialize the model
		cmd := m.Init()

		// Execute the command to read the directory
		msg := cmd()
		m, _ = m.Update(msg)

		// Verify the current directory is now the parent
		if m.CurrentDirectory != tmpDir {
			t.Errorf("Expected CurrentDirectory to be %s, got %s", tmpDir, m.CurrentDirectory)
		}

		// Verify that the original directory is selected
		if m.FileSelected != "myproject" {
			t.Errorf("Expected FileSelected to be 'myproject', got '%s'", m.FileSelected)
		}

		// Verify the selected index points to the right file
		selectedFile := m.files[m.selected]
		if selectedFile.Name() != "myproject" {
			t.Errorf("Expected selected file to be 'myproject', got '%s'", selectedFile.Name())
		}

		// Test that we can select the directory with Enter key
		keyMsg := tea.KeyPressMsg{
			Code: tea.KeyEnter,
		}
		m, _ = m.Update(keyMsg)

		// Verify the path was set correctly
		expectedPath := filepath.Join(tmpDir, "myproject")
		if m.Path != expectedPath {
			t.Errorf("Expected Path to be %s, got %s", expectedPath, m.Path)
		}

		// Verify DidSelectFile returns true
		didSelect, path := m.DidSelectFile(keyMsg)
		if !didSelect {
			t.Error("Expected DidSelectFile to return true")
		}
		if path != expectedPath {
			t.Errorf("Expected selected path to be %s, got %s", expectedPath, path)
		}
	})

	// Test with CWDDefault disabled (normal behavior)
	t.Run("CWDDefault disabled", func(t *testing.T) {
		m := New()
		m.CurrentDirectory = projectDir
		m.DirAllowed = true
		m.FileAllowed = false
		m.SetHeight(10)

		// Initialize the model
		cmd := m.Init()
		if cmd == nil {
			t.Fatal("Init() should return a command")
		}

		// Execute the command to read the directory
		msg := cmd()
		m, _ = m.Update(msg)

		// Verify the current directory is still the project directory
		if m.CurrentDirectory != projectDir {
			t.Errorf("Expected CurrentDirectory to be %s, got %s", projectDir, m.CurrentDirectory)
		}

		// Verify nothing is pre-selected (or first item is selected by default)
		if m.selected != 0 {
			t.Errorf("Expected selected to be 0, got %d", m.selected)
		}
	})

	// Test navigation after CWDDefault selection
	t.Run("Navigation after CWDDefault", func(t *testing.T) {
		m := New()
		m.CurrentDirectory = projectDir
		m.DirAllowed = true
		m.FileAllowed = false
		m.SetHeight(10)
		m.SetCWDDefault(true) // Call SetCWDDefault before Init

		// Initialize
		cmd := m.Init()
		msg := cmd()
		m, _ = m.Update(msg)

		// Navigate down
		downKey := tea.KeyPressMsg{
			Code: tea.KeyDown,
		}
		m, _ = m.Update(downKey)

		// Should now be on "otherproject"
		if m.files[m.selected].Name() != "otherproject" {
			t.Errorf("Expected to navigate to 'otherproject', got '%s'", m.files[m.selected].Name())
		}

		// Navigate up
		upKey := tea.KeyPressMsg{
			Code: tea.KeyUp,
		}
		m, _ = m.Update(upKey)

		// Should be back on myproject
		if m.files[m.selected].Name() != "myproject" {
			t.Errorf("Expected to navigate back to 'myproject', got '%s'", m.files[m.selected].Name())
		}
	})
}

// TestFileSelected tests that FileSelected is updated correctly during navigation.
func TestFileSelected(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create multiple subdirectories and files
	dirs := []string{"aaa_dir", "bbb_dir", "ccc_dir", "ddd_dir", "eee_dir"}
	for _, dir := range dirs {
		err := os.Mkdir(filepath.Join(tmpDir, dir), 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory %s: %v", dir, err)
		}
	}

	// Create some files
	files := []string{"file1.txt", "file2.go", "file3.md"}
	for _, file := range files {
		err := os.WriteFile(filepath.Join(tmpDir, file), []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	t.Run("FileSelected updates on navigation", func(t *testing.T) {
		m := New()
		m.CurrentDirectory = tmpDir
		m.DirAllowed = true
		m.FileAllowed = true
		m.SetHeight(10)

		// Initialize
		cmd := m.Init()
		msg := cmd()
		m, _ = m.Update(msg)

		// Initially should have first directory selected (directories sort first)
		if m.FileSelected != "aaa_dir" {
			t.Errorf("Expected initial FileSelected to be 'aaa_dir', got '%s'", m.FileSelected)
		}

		// Navigate down
		downKey := tea.KeyPressMsg{Code: tea.KeyDown}
		m, _ = m.Update(downKey)
		if m.FileSelected != "bbb_dir" {
			t.Errorf("After down, expected FileSelected to be 'bbb_dir', got '%s'", m.FileSelected)
		}

		// Navigate down again
		m, _ = m.Update(downKey)
		if m.FileSelected != "ccc_dir" {
			t.Errorf("After second down, expected FileSelected to be 'ccc_dir', got '%s'", m.FileSelected)
		}

		// Navigate up
		upKey := tea.KeyPressMsg{Code: tea.KeyUp}
		m, _ = m.Update(upKey)
		if m.FileSelected != "bbb_dir" {
			t.Errorf("After up, expected FileSelected to be 'bbb_dir', got '%s'", m.FileSelected)
		}

		// Go to top
		topKey := tea.KeyPressMsg{Code: 'g', Text: "g"}
		m, _ = m.Update(topKey)
		if m.FileSelected != "aaa_dir" {
			t.Errorf("After goto top, expected FileSelected to be 'aaa_dir', got '%s'", m.FileSelected)
		}

		// Go to last
		lastKey := tea.KeyPressMsg{Code: 'G', Text: "G"}
		m, _ = m.Update(lastKey)
		if m.FileSelected != "file3.md" {
			t.Errorf("After goto last, expected FileSelected to be 'file3.md', got '%s'", m.FileSelected)
		}
	})

	t.Run("FileSelected updates on page navigation", func(t *testing.T) {
		m := New()
		m.CurrentDirectory = tmpDir
		m.DirAllowed = true
		m.FileAllowed = true
		m.SetHeight(3) // Small height to test paging

		// Initialize
		cmd := m.Init()
		msg := cmd()
		m, _ = m.Update(msg)

		// Initially should have first item
		if m.FileSelected != "aaa_dir" {
			t.Errorf("Expected initial FileSelected to be 'aaa_dir', got '%s'", m.FileSelected)
		}

		// Page down
		pageDownKey := tea.KeyPressMsg{Code: tea.KeyPgDown}
		m, _ = m.Update(pageDownKey)

		// Should have jumped by height (3 items)
		expectedAfterPageDown := "ddd_dir"
		if m.FileSelected != expectedAfterPageDown {
			t.Errorf("After page down, expected FileSelected to be '%s', got '%s'", expectedAfterPageDown, m.FileSelected)
		}

		// Page up
		pageUpKey := tea.KeyPressMsg{Code: tea.KeyPgUp}
		m, _ = m.Update(pageUpKey)

		// Should have jumped back
		expectedAfterPageUp := "aaa_dir"
		if m.FileSelected != expectedAfterPageUp {
			t.Errorf("After page up, expected FileSelected to be '%s', got '%s'", expectedAfterPageUp, m.FileSelected)
		}
	})

	t.Run("FileSelected is empty when no files", func(t *testing.T) {
		emptyDir := filepath.Join(tmpDir, "empty")
		err := os.Mkdir(emptyDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create empty directory: %v", err)
		}

		m := New()
		m.CurrentDirectory = emptyDir
		m.DirAllowed = true
		m.FileAllowed = true
		m.SetHeight(10)

		// Initialize
		cmd := m.Init()
		msg := cmd()
		m, _ = m.Update(msg)

		// FileSelected should be empty
		if m.FileSelected != "" {
			t.Errorf("Expected FileSelected to be empty in empty directory, got '%s'", m.FileSelected)
		}

		// Try to navigate (should not crash)
		downKey := tea.KeyPressMsg{Code: tea.KeyDown}
		m, _ = m.Update(downKey)
		if m.FileSelected != "" {
			t.Errorf("Expected FileSelected to remain empty after navigation in empty directory, got '%s'", m.FileSelected)
		}
	})

	t.Run("FileSelected persists across directory navigation", func(t *testing.T) {
		m := New()
		m.CurrentDirectory = tmpDir
		m.DirAllowed = true
		m.FileAllowed = true
		m.SetHeight(10)

		// Initialize
		cmd := m.Init()
		msg := cmd()
		m, _ = m.Update(msg)

		// Navigate to second directory
		downKey := tea.KeyPressMsg{Code: tea.KeyDown}
		m, _ = m.Update(downKey)
		if m.FileSelected != "bbb_dir" {
			t.Errorf("Expected FileSelected to be 'bbb_dir', got '%s'", m.FileSelected)
		}

		// Enter the directory
		enterKey := tea.KeyPressMsg{Code: tea.KeyEnter}
		m, cmd = m.Update(enterKey)

		// Execute the readDir command
		if cmd != nil {
			msg := cmd()
			m, _ = m.Update(msg)
		}

		// Now we're inside bbb_dir (empty), FileSelected should be empty
		if m.FileSelected != "" {
			t.Errorf("Expected FileSelected to be empty inside bbb_dir, got '%s'", m.FileSelected)
		}

		// Navigate back
		backKey := tea.KeyPressMsg{Code: tea.KeyEsc}
		m, cmd = m.Update(backKey)

		// Execute the readDir command for parent
		if cmd != nil {
			msg := cmd()
			m, _ = m.Update(msg)
		}

		// Should be back in tmpDir with bbb_dir selected (from stack)
		if m.FileSelected != "bbb_dir" {
			t.Errorf("After going back, expected FileSelected to be 'bbb_dir', got '%s'", m.FileSelected)
		}
	})
}
