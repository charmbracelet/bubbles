package progress

import (
	"strings"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
)

type Model struct {
	// Left side color of progress bar. By default, it's #00dbde
	StartColor colorful.Color

	// Left side color of progress bar. By default, it's #fc00ff
	EndColor colorful.Color

	// Width of progress bar in symbols.
	Width int

	// filament rune of done part of progress bar. it's █ by default
	FilamentSymbol rune

	// empty rune of pending part of progress bar. it's ░ by default
	EmptySymbol rune

	// if true, gradient will be setted from start to end of filled part. Instead, it'll work
	// on all proggress bar length
	FullGradientMode bool
}

// NewModel returns a model with default values.
func NewModel() Model {
	startColor, _ := colorful.Hex("#00dbde")
	endColor, _ := colorful.Hex("#fc00ff")
	return Model{
		StartColor:     startColor,
		EndColor:       endColor,
		Width:          40,
		FilamentSymbol: '█',
		EmptySymbol:    '░',
	}
}

func (m *Model) View(percent float64) string {
	ramp := make([]string, int(float64(m.Width)*percent))
	for i := 0; i < len(ramp); i++ {
		gradientPart := float64(m.Width)
		if m.FullGradientMode {
			gradientPart = float64(len(ramp))
		}
		percent := float64(i) / gradientPart
		c := m.StartColor.BlendLuv(m.EndColor, percent)
		ramp[i] = c.Hex()
	}

	var fullCells string
	for i := 0; i < len(ramp); i++ {
		fullCells += termenv.String(string(m.FilamentSymbol)).Foreground(termenv.ColorProfile().Color(ramp[i])).String()
	}

	fullCells += strings.Repeat(string(m.EmptySymbol), m.Width-len(ramp))
	return fullCells
}
