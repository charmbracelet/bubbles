// Package scrollbar provides a scrollbar component for Bubble Tea applications.
package scrollbar

import (
	"cmp"
	"math"
	"strings"
)

// ScrollState is the state of the scrollbar.
type ScrollState struct {
	TotalLength int // Total length of the scrollbar.
	ThumbOffset int // Offset of the thumb.
	ThumbLength int // Length of the thumb, including the start and end characters.
}

// ContentState is the state of the content being tracked.
type ContentState struct {
	Length        int // Length of the content.
	VisibleLength int // Visible length of the content.
	Offset        int // Offset of the content.
}

// Position is the rendered position of the scrollbar.
type Position int

// Available positions for the scrollbar.
const (
	Vertical Position = iota
	Horizontal
)

// Option is used to set options in New. For example:
//
//	scrollbar := New(WithPosition(Vertical))
type Option func(*Model)

// WithPosition sets the position of the scrollbar.
func WithPosition(position Position) Option {
	return func(m *Model) {
		m.position = position

		switch position {
		case Vertical:
			m.width = 1
		case Horizontal:
			m.height = 1
		}
	}
}

// WithType sets the type of the scrollbar.
func WithType(t Type) Option {
	return func(m *Model) {
		m.barType = t
	}
}

// Model is the Bubble Tea model for this user interface.
type Model struct {
	position Position
	barType  Type
	styles   Styles
	width    int
	height   int

	// Content-specific fields that are set by the caller.
	contentLength        int
	contentVisibleLength int
	contentOffset        int
}

// New creates a new model with default settings.
func New(opts ...Option) Model {
	m := Model{
		width:    0,
		height:   0,
		styles:   DefaultDarkStyles(),
		position: Vertical,
		barType:  SlimBar(),
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&m)
	}

	return m
}

// Styles returns the current set of styles.
func (m Model) Styles() Styles {
	return m.styles
}

// SetStyles sets the styles for the scrollbar.
func (m *Model) SetStyles(s Styles) {
	m.styles = s
}

// Position returns the position of the scrollbar.
func (m Model) Position() Position {
	return m.position
}

// Width returns the width of the scrollbar.
func (m Model) Width() int {
	if m.position == Horizontal && m.width == 0 {
		return m.contentVisibleLength
	}
	return m.width
}

// SetWidth sets the width of the scrollbar. If the scrollbar is vertical, this
// is a no-op, as it is dependent on the bar type used.
func (m *Model) SetWidth(w int) {
	if m.position == Vertical {
		return
	}
	m.width = max(0, w)
	if m.contentVisibleLength == 0 {
		m.contentVisibleLength = m.width
	}
}

// Height returns the height of the scrollbar.
func (m Model) Height() int {
	if m.position == Vertical && m.height == 0 {
		return m.contentVisibleLength
	}
	return m.height
}

// SetHeight sets the height of the scrollbar. If the scrollbar is horizontal,
// this is a no-op, as it is dependent on the bar type used.
func (m *Model) SetHeight(h int) {
	if m.position == Horizontal {
		return
	}
	m.height = max(0, h)
	if m.contentVisibleLength == 0 {
		m.contentVisibleLength = m.height
	}
}

// SetContentState sets the state of the scrollbar used for tracking the content
// dimensions.
//   - length: the total length of the content (height for vertical, width for
//     horizontal)
//   - visible: the visible length of the content (height for vertical, width for
//     horizontal)
//   - offset: the offset of the view within the content (typically the y-offset
//     for a vertical scrollbar or the x-offset for a horizontal scrollbar)
func (m *Model) SetContentState(length, visible, offset int) {
	m.contentLength = max(0, length)
	m.contentVisibleLength = clamp(visible, 0, length)
	m.contentOffset = clamp(offset, 0, length-visible)
}

// ContentState returns the current content state of the scrollbar.
func (m Model) ContentState() ContentState {
	return ContentState{
		Length:        m.contentLength,
		VisibleLength: m.contentVisibleLength,
		Offset:        m.contentOffset,
	}
}

// ScrollState returns the current scroll state of the scrollbar. Returns nil
// if the scrollbar is not required based on the content information provided,
// or the size of the scrollbar is too small to render correctly.
func (m Model) ScrollState() *ScrollState {
	if (m.position == Vertical && m.height < 3) ||
		(m.position == Horizontal && m.width < 3) ||
		m.contentLength == 0 || m.contentVisibleLength == 0 ||
		m.contentLength <= m.contentVisibleLength {
		return nil
	}

	var length int

	switch m.position {
	case Vertical:
		length = m.height
	case Horizontal:
		length = m.width
	}

	ratio := float64(length) / float64(m.contentLength)

	thumbLength := max(
		m.barType.MinThumbLength(m.position),
		int(math.Round(float64(m.contentVisibleLength)*ratio)),
	)
	thumbOffset := max(
		0,
		min(length-thumbLength, int(math.Round(float64(m.contentOffset)*ratio))),
	)

	return &ScrollState{
		TotalLength: length,
		ThumbOffset: thumbOffset,
		ThumbLength: thumbLength,
	}
}

// View renders the scrollbar to a string.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	state := m.ScrollState()

	if state == nil {
		switch m.position {
		case Vertical:
			return strings.TrimRight(
				strings.Repeat(m.styles.Track.Render(" ")+"\n", m.height),
				"\n",
			)
		case Horizontal:
			return strings.TrimRight(
				strings.Repeat(m.styles.Track.Render(" "), m.width),
				" ",
			)
		}
		return ""
	}

	var thumbStart, thumbMiddle, thumbEnd, track string
	switch m.position {
	case Vertical:
		thumbStart = m.styles.ThumbStart.
			Render(string(m.barType.VerticalThumbStart))
		thumbMiddle = m.styles.ThumbMiddle.
			Render(string(m.barType.VerticalThumbMiddle))
		thumbEnd = m.styles.ThumbEnd.
			Render(string(m.barType.VerticalThumbEnd))
		track = m.styles.Track.
			Render(string(m.barType.VerticalTrack))
	case Horizontal:
		thumbStart = m.styles.ThumbStart.
			Render(string(m.barType.HorizontalThumbStart))
		thumbMiddle = m.styles.ThumbMiddle.
			Render(string(m.barType.HorizontalThumbMiddle))
		thumbEnd = m.styles.ThumbEnd.
			Render(string(m.barType.HorizontalThumbEnd))
		track = m.styles.Track.
			Render(string(m.barType.HorizontalTrack))
	}

	var suffix string
	if m.position == Vertical {
		suffix = "\n"
	}

	var s strings.Builder

	s.WriteString(strings.Repeat(track+suffix, max(0, state.ThumbOffset)))

	if m.barType.MinThumbLength(m.position) == 1 {
		s.WriteString(strings.Repeat(thumbMiddle+suffix, max(0, state.ThumbLength)))
	} else {
		s.WriteString(thumbStart + suffix)
		s.WriteString(strings.Repeat(thumbMiddle+suffix, max(0, state.ThumbLength-2)))
		s.WriteString(thumbEnd + suffix)
	}

	s.WriteString(strings.Repeat(track+suffix, max(0, state.TotalLength-state.ThumbOffset-state.ThumbLength)))

	return strings.TrimRight(s.String(), suffix)
}

// Percent returns the scroll percentage of the content.
func (m Model) Percent() float64 {
	return clamp(float64(m.contentOffset)/float64(m.contentLength-m.contentVisibleLength), 0, 1)
}

func clamp[T cmp.Ordered](v, low, high T) T {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}
