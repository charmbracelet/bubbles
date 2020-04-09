// package pager provides a Tea package for calulating pagination and rendering
// pagination info. Note that this package does not render actual pages: it's
// purely for handling keystrokes related to pagination, and rendering
// pagination status.
package pager

import (
	"fmt"

	"github.com/charmbracelet/tea"
)

// PagerType specifies the way we render pagination
type PagerType int

// Pagination rendering options
const (
	Dots PagerType = iota
	Arabic
)

// Model is the Tea model for this user interface
type Model struct {
	Page         int
	PerPage      int
	TotalPages   int
	ActiveDot    string
	InactiveDot  string
	ArabicFormat string
	RTL          bool
}

// SetTotalPages is a helper method for calculatng the total number of pages
// from a given number of items. It's use is optional. Note that it both
// returns the number of total pages and alters the model.
func (m *Model) SetTotalPages(items int) int {
	if items == 0 {
		return 0
	}
	n := items / m.PerPage
	if items%m.PerPage > 0 {
		n += 1
	}
	m.TotalPages = n
	return n
}

// NewModel creates a new model with defaults
func NewModel() Model {
	return Model{
		Page:         0,
		PerPage:      1,
		TotalPages:   1,
		ActiveDot:    "•",
		InactiveDot:  "○",
		ArabicFormat: "%d/%d",
		RTL:          false,
	}
}

// Update is the Tea update function which binds keystrokes to pagination
func Update(msg tea.Msg, model tea.Model) (tea.Model, tea.Cmd) {
	m, ok := model.(Model)
	if !ok {
		return tea.ModelAssertionErr, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			if m.Page > 0 {
				m.Page--
			}
		case "right":
			if m.Page < m.TotalPages-1 {
				m.Page++
			}
		}
	}
	return m, nil
}

// View renders the pagination to a string
func View(model tea.Model) string {
	m, ok := model.(Model)
	if !ok {
		return ""
	}
	return dotsView(m)
}

func dotsView(m Model) string {
	var s string
	if m.RTL {
		for i := m.TotalPages; i > 0; i-- {
			if i == m.Page {
				s += m.ActiveDot
				continue
			}
			s += m.InactiveDot
		}
		return s
	}
	for i := 0; i < m.TotalPages; i++ {
		if i == m.Page {
			s += m.ActiveDot
			continue
		}
		s += m.InactiveDot
	}
	return s
}

func arabicView(m Model) string {
	if m.RTL {
		return fmt.Sprintf(m.ArabicFormat, m.TotalPages, m.Page+1)
	}
	return fmt.Sprintf(m.ArabicFormat, m.Page+1, m.TotalPages)
}
