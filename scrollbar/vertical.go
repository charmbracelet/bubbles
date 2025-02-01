package scrollbar

import (
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NewVertical create a new vertical scrollbar.
func NewVertical() Vertical {
	return Vertical{
		Style:      lipgloss.NewStyle().Width(1),
		ThumbStyle: lipgloss.NewStyle().SetString("█"),
		TrackStyle: lipgloss.NewStyle().SetString("░"),
	}
}

// Vertical is the base struct for a vertical scrollbar.
type Vertical struct {
	Style       lipgloss.Style
	ThumbStyle  lipgloss.Style
	TrackStyle  lipgloss.Style
	height      int
	thumbHeight int
	thumbOffset int
}

// Init initializes the scrollbar model.
func (m Vertical) Init() tea.Cmd {
	return nil
}

// Update updates the scrollbar model.
func (m Vertical) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case Msg:
		m.thumbHeight, m.thumbOffset = m.computeThumb(msg.Total, msg.Visible, msg.Offset)
	case HeightMsg:
		m.height = m.computeHeight(int(msg))
	case viewport.Model:
		m.thumbHeight, m.thumbOffset = m.computeThumb(msg.TotalLineCount(), msg.VisibleLineCount(), msg.YOffset)
	}

	return m, nil
}

func (m Vertical) computeHeight(height int) int {
	return height - m.Style.GetVerticalFrameSize()
}

func (m Vertical) computeThumb(total, visible, offset int) (int, int) {
	ratio := float64(m.height) / float64(total)

	thumbHeight := max(1, int(math.Round(float64(visible)*ratio)))
	thumbOffset := max(0, min(m.height-thumbHeight, int(math.Round(float64(offset)*ratio))))

	return thumbHeight, thumbOffset
}

// View renders the scrollbar to a string.
func (m Vertical) View() string {
	bar := strings.TrimRight(
		strings.Repeat(m.TrackStyle.String()+"\n", m.thumbOffset)+
			strings.Repeat(m.ThumbStyle.String()+"\n", m.thumbHeight)+
			strings.Repeat(m.TrackStyle.String()+"\n", max(0, m.height-m.thumbOffset-m.thumbHeight)),
		"\n",
	)

	return m.Style.Render(bar)
}
