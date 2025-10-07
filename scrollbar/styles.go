package scrollbar

import "github.com/charmbracelet/lipgloss/v2"

// Styles are the styles for the scrollbar. For vertical scrollbars, the start
// thumb is the top. For horizontal scrollbars, the start thumb is the left.
type Styles struct {
	ThumbStart  lipgloss.Style
	ThumbMiddle lipgloss.Style
	ThumbEnd    lipgloss.Style
	Track       lipgloss.Style
}

// DefaultStyles returns the default styles for the scrollbar.
func DefaultStyles(isDark bool) Styles {
	lightDark := lipgloss.LightDark(isDark)

	var s Styles

	s.ThumbStart = lipgloss.NewStyle().
		Foreground(lightDark(lipgloss.Color("16"), lipgloss.Color("252")))
	s.ThumbMiddle = s.ThumbStart
	s.ThumbEnd = s.ThumbStart
	s.Track = lipgloss.NewStyle().
		Foreground(lightDark(lipgloss.Color("248"), lipgloss.Color("240")))

	return s
}

// DefaultLightStyles returns the default styles for a light background.
func DefaultLightStyles() Styles {
	return DefaultStyles(false)
}

// DefaultDarkStyles returns the default styles for a dark background.
func DefaultDarkStyles() Styles {
	return DefaultStyles(true)
}

// Type represents a scrollbars appearance as characters. All characters must
// be single-width.
type Type struct {
	VerticalThumbStart  rune
	VerticalThumbMiddle rune
	VerticalThumbEnd    rune
	VerticalTrack       rune

	HorizontalThumbStart  rune
	HorizontalThumbMiddle rune
	HorizontalThumbEnd    rune
	HorizontalTrack       rune
}

// MinThumbLength returns the minimum length of the thumb for the given position.
// This will dynamically change based on if the start/end characters are the same.
func (t Type) MinThumbLength(pos Position) int {
	switch pos {
	case Vertical:
		if t.VerticalThumbStart == t.VerticalThumbEnd &&
			t.VerticalThumbStart == t.VerticalThumbMiddle {
			return 1
		}
	case Horizontal:
		if t.HorizontalThumbStart == t.HorizontalThumbEnd &&
			t.HorizontalThumbStart == t.HorizontalThumbMiddle {
			return 1
		}
	}

	return 3 // Thumb start + thumb middle + thumb end.
}

// SlimBar returns a scrollbar that uses slim/thin bars for both vertical and
// horizontal scrollbars.
func SlimBar() Type {
	return Type{
		VerticalThumbStart:  '┃',
		VerticalThumbMiddle: '┃',
		VerticalThumbEnd:    '┃',
		VerticalTrack:       '┃',

		HorizontalThumbStart:  '▁',
		HorizontalThumbMiddle: '▁',
		HorizontalThumbEnd:    '▁',
		HorizontalTrack:       '▁',
	}
}

// SlimDottedBar returns a scrollbar that uses slim bars for thumbs, and dotted
// bars for the track.
func SlimDottedBar() Type {
	return Type{
		VerticalThumbStart:  '┃',
		VerticalThumbMiddle: '┃',
		VerticalThumbEnd:    '┃',
		VerticalTrack:       '┇',

		HorizontalThumbStart:  '▬',
		HorizontalThumbMiddle: '▬',
		HorizontalThumbEnd:    '▬',
		HorizontalTrack:       '▬',
	}
}

// SlimCirclesBar returns a scrollbar that uses slim bars for thumbs and tracks,
// but uses circles as the start and end thumb characters.
func SlimCirclesBar() Type {
	return Type{
		VerticalThumbStart:  '◉',
		VerticalThumbMiddle: '┃',
		VerticalThumbEnd:    '◉',
		VerticalTrack:       '┃',

		HorizontalThumbStart:  '◉',
		HorizontalThumbMiddle: '━',
		HorizontalThumbEnd:    '◉',
		HorizontalTrack:       '━',
	}
}

// BlockBar returns a scrollbar that uses full block bars for both vertical and
// horizontal scrollbars.
func BlockBar() Type {
	return Type{
		VerticalThumbStart:  '█',
		VerticalThumbMiddle: '█',
		VerticalThumbEnd:    '█',
		VerticalTrack:       '░',

		HorizontalThumbStart:  '█',
		HorizontalThumbMiddle: '█',
		HorizontalThumbEnd:    '█',
		HorizontalTrack:       '░',
	}
}

// DottedBar returns a scrollbar that uses a mix of dots and bars for both
// vertical and horizontal scrollbars.
func DottedBar() Type {
	return Type{
		VerticalThumbStart:  '⣿',
		VerticalThumbMiddle: '⣿',
		VerticalThumbEnd:    '⣿',
		VerticalTrack:       '⣿',

		HorizontalThumbStart:  '⣤',
		HorizontalThumbMiddle: '⣤',
		HorizontalThumbEnd:    '⣤',
		HorizontalTrack:       '⣤',
	}
}

// ASCIIBar returns a scrollbar that uses basic ASCII characters for both
// vertical and horizontal scrollbars.
func ASCIIBar() Type {
	return Type{
		VerticalThumbStart:  '|',
		VerticalThumbMiddle: '|',
		VerticalThumbEnd:    '|',
		VerticalTrack:       '|',

		HorizontalThumbStart:  '-',
		HorizontalThumbMiddle: '-',
		HorizontalThumbEnd:    '-',
		HorizontalTrack:       '-',
	}
}
