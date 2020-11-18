package progress

import (
	"fmt"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/termenv"
)

var color func(string) termenv.Color = termenv.ColorProfile().Color

type Model struct {
	// Left side color of progress bar. By default, it's #00dbde
	StartColor colorful.Color

	// Left side color of progress bar. By default, it's #fc00ff
	EndColor colorful.Color

	// Width of progress bar in symbols.
	Width int

	// filament rune of done part of progress bar. it's █ by default
	Full rune

	// empty rune of pending part of progress bar. it's ░ by default
	Empty rune

	// if true, gradient will be setted from start to end of filled part. Instead, it'll work
	// on all proggress bar length
	FullGradientMode bool

	ShowPercent   bool
	PercentFormat string
}

// NewModel returns a model with default values.
func NewModel() Model {
	startColor, _ := colorful.Hex("#00dbde")
	endColor, _ := colorful.Hex("#fc00ff")
	return Model{
		StartColor:    startColor,
		EndColor:      endColor,
		Width:         40,
		Full:          '█',
		Empty:         '░',
		ShowPercent:   true,
		PercentFormat: " %3.0f%%",
	}
}

func (m Model) View(percent float64) string {
	if m.ShowPercent {
		s := fmt.Sprintf(m.PercentFormat, percent*100)
		w := ansi.PrintableRuneWidth(s)
		return m.bar(percent, w) + s
	}
	return m.bar(percent, 0)
}

func (m Model) bar(percent float64, textWidth int) string {
	w := m.Width - textWidth

	ramp := make([]string, int(float64(w)*percent))
	for i := 0; i < len(ramp); i++ {
		gradientPart := float64(w)
		if m.FullGradientMode {
			gradientPart = float64(len(ramp))
		}
		percent := float64(i) / gradientPart
		c := m.StartColor.BlendLuv(m.EndColor, percent)
		ramp[i] = c.Hex()
	}

	var fullCells string
	for i := 0; i < len(ramp); i++ {
		fullCells += termenv.String(string(m.Full)).Foreground(color(ramp[i])).String()
	}

	fullCells += strings.Repeat(string(m.Empty), w-len(ramp))
	return fullCells
}
