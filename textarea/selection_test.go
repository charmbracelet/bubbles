package textarea

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestTextSelection(t *testing.T) {
	ta := New()
	ta.SetValue("Hello world")
	
	// Test SelectAll
	ta.SelectAll()
	assert.True(t, ta.HasSelection())
	assert.Equal(t, "Hello world", ta.GetSelectedText())
	
	// Test ClearSelection
	ta.ClearSelection()
	assert.False(t, ta.HasSelection())
	
	// Test SetSelection
	ta.SetSelection(
		Position{Row: 0, Col: 0},
		Position{Row: 0, Col: 5},
	)
	assert.Equal(t, "Hello", ta.GetSelectedText())
	
	// Test DeleteSelection
	ta.DeleteSelection()
	assert.Equal(t, " world", ta.Value())
}

func TestMouseToPosition(t *testing.T) {
	ta := New()
	ta.SetValue("Line 1\nLine 2\nLine 3")
	ta.SetWidth(20)
	
	// Test position calculation
	pos := ta.mouseToPosition(0, 0)
	assert.Equal(t, Position{Row: 0, Col: 0}, pos)
	
	pos = ta.mouseToPosition(5, 1)
	assert.Equal(t, Position{Row: 1, Col: 5}, pos)
}

func TestWordSelection(t *testing.T) {
	ta := New()
	ta.SetValue("Hello beautiful world")
	
	// Select "beautiful"
	ta.SelectWord(Position{Row: 0, Col: 8})
	assert.Equal(t, "beautiful", ta.GetSelectedText())
}

func TestLineSelection(t *testing.T) {
	ta := New()
	ta.SetValue("Line 1\nLine 2\nLine 3")
	
	// Select second line
	ta.SelectLine(1)
	assert.Equal(t, "Line 2", ta.GetSelectedText())
}

func TestMultiLineSelection(t *testing.T) {
	ta := New()
	ta.SetValue("Line 1\nLine 2\nLine 3")
	
	// Select from middle of first line to middle of second line
	ta.SetSelection(
		Position{Row: 0, Col: 3},
		Position{Row: 1, Col: 3},
	)
	assert.Equal(t, "e 1\nLin", ta.GetSelectedText())
}

func TestSelectionBounds(t *testing.T) {
	ta := New()
	ta.SetValue("Test")
	
	// Test out of bounds selection
	ta.SetSelection(
		Position{Row: 0, Col: -1},
		Position{Row: 0, Col: 100},
	)
	assert.Equal(t, "Test", ta.GetSelectedText())
	
	// Test inverted selection (end before start)
	ta.SetSelection(
		Position{Row: 0, Col: 3},
		Position{Row: 0, Col: 1},
	)
	assert.Equal(t, "es", ta.GetSelectedText())
}