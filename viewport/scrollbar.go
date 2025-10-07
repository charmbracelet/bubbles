package viewport

import (
	"github.com/charmbracelet/bubbles/v2/scrollbar"
	"github.com/charmbracelet/lipgloss/v2"
)

// WithScrollbars sets the scrollbar type, and whether to enable the x and y
// scrollbars. Scrollbars will still only be rendered for each axis only if
// required (i.e. if the content is longer than the viewport in that given
// axis).
func WithScrollbars(scrollbarType scrollbar.Type, x, y bool) Option {
	return func(m *Model) {
		if x {
			m.xscrollbarEnabled = x
			m.xscrollbar = scrollbar.New(
				scrollbar.WithType(scrollbarType),
				scrollbar.WithPosition(scrollbar.Horizontal),
			)
		}
		if y {
			m.yscrollbarEnabled = y
			m.yscrollbar = scrollbar.New(
				scrollbar.WithType(scrollbarType),
				scrollbar.WithPosition(scrollbar.Vertical),
			)
		}
		m.calculateScrollbar()
	}
}

// SetScrollbarStyles sets the styles for the scrollbars.
func (m *Model) SetScrollbarStyles(style scrollbar.Styles) {
	m.xscrollbar.SetStyles(style)
	m.yscrollbar.SetStyles(style)
}

// calculateScrollbar calculates if any scrollbars should be enabled, and if so,
// adjusts the rendered dimensions of the viewport, in addition to updating
// the scrollbar's content state. This ensures that the scrollbars do not take up
// any rendered space unless required.
func (m *Model) calculateScrollbar() {
	if !m.xscrollbarEnabled && !m.yscrollbarEnabled {
		m.renderedWidth = m.actualWidth
		m.renderedHeight = m.actualHeight
		return
	}

	totalLines := m.TotalLineCount()

	m.yscrollbarRender = m.yscrollbarEnabled && totalLines > m.actualHeight && m.actualWidth > 1
	if m.yscrollbarRender {
		m.renderedWidth = max(0, m.actualWidth-m.yscrollbar.Width())
	} else {
		m.renderedWidth = m.actualWidth
	}

	m.xscrollbarRender = m.xscrollbarEnabled && m.longestLineWidth > m.actualWidth && !m.SoftWrap && m.actualHeight > 1
	if m.xscrollbarRender {
		m.renderedHeight = max(0, m.actualHeight-m.xscrollbar.Height())
		// Recalculate vertical scrollbar enablement logic again, as they are
		// co-dependent on each other given each can change the rendered dimensions.
		if !m.yscrollbarRender {
			m.yscrollbarRender = m.yscrollbarEnabled && totalLines > m.renderedHeight && m.renderedWidth > 1
			if m.yscrollbarRender {
				m.renderedWidth = max(0, m.actualWidth-m.yscrollbar.Width())
			}
		}
	} else {
		m.renderedHeight = m.actualHeight
	}

	if m.yscrollbarRender {
		m.yscrollbar.SetContentState(
			m.TotalLineCount(),
			m.renderedHeight,
			m.yOffset,
		)
		m.yscrollbar.SetHeight(m.renderedHeight)
	}

	if m.xscrollbarRender {
		m.xscrollbar.SetContentState(
			m.longestLineWidth,
			m.renderedWidth,
			m.xOffset,
		)
		m.xscrollbar.SetWidth(m.renderedWidth)
	}
}

// viewWithScrollbars renders the viewport with the scrollbars included.
func (m Model) viewWithScrollbars(content string) string {
	switch {
	case !m.yscrollbarRender && !m.xscrollbarRender:
		return content
	case m.yscrollbarRender && m.xscrollbarRender:
		return lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				content,
				m.yscrollbar.View(),
			),
			m.xscrollbar.View(),
		)
	case m.yscrollbarRender:
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			content,
			m.yscrollbar.View(),
		)
	case m.xscrollbarRender:
		return lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			m.xscrollbar.View(),
		)
	}
	return content
}
