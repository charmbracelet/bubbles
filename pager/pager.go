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
	Arabic PagerType = iota
	Dots
)

// Model is the Tea model for this user interface
type Model struct {
	Type             PagerType
	Page             int
	PerPage          int
	TotalPages       int
	ActiveDot        string
	InactiveDot      string
	ArabicFormat     string
	UseLeftRightKeys bool
	UseUpDownKeys    bool
	UseHLKeys        bool
	UseJKKeys        bool
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

// GetSliceBounds is a helper function for paginating slices. Pass the length
// of the slice you're rendering and you'll receive the start and end bounds
// corresponding the to pagination. For example:
//
//     bunchOfStuff := []stuff{...}
//     start, end := model.GetSliceBounds(len(bunchOfStuff))
//     sliceToRender := bunchOfStuff[start:end]
//
func (m *Model) GetSliceBounds(length int) (start int, end int) {
	start = m.Page * m.PerPage
	end = min(m.Page*m.PerPage+m.PerPage, length)
	return start, end
}

func (m *Model) prevPage() {
	if m.Page > 0 {
		m.Page--
	}
}

func (m *Model) nextPage() {
	if m.Page < m.TotalPages-1 {
		m.Page++
	}
}

// NewModel creates a new model with defaults
func NewModel() Model {
	return Model{
		Type:             Arabic,
		Page:             0,
		PerPage:          1,
		TotalPages:       1,
		ActiveDot:        "•",
		InactiveDot:      "○",
		ArabicFormat:     "%d/%d",
		UseLeftRightKeys: true,
		UseUpDownKeys:    false,
		UseHLKeys:        true,
		UseJKKeys:        false,
	}
}

// Update is the Tea update function which binds keystrokes to pagination
func Update(msg tea.Msg, m Model) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.UseLeftRightKeys {
			switch msg.String() {
			case "left":
				m.prevPage()
			case "right":
				m.nextPage()
			}
		}
		if m.UseUpDownKeys {
			switch msg.String() {
			case "up":
				m.prevPage()
			case "down":
				m.nextPage()
			}
		}
		if m.UseHLKeys {
			switch msg.String() {
			case "h":
				m.prevPage()
			case "l":
				m.nextPage()
			}
		}
		if m.UseJKKeys {
			switch msg.String() {
			case "j":
				m.prevPage()
			case "k":
				m.nextPage()
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
	switch m.Type {
	case Dots:
		return dotsView(m)
	default:
		return arabicView(m)
	}
}

func dotsView(m Model) string {
	var s string
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
	return fmt.Sprintf(m.ArabicFormat, m.Page+1, m.TotalPages)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
