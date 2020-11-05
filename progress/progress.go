package progress

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
)

// Value is explicit type of, you know, progression. Can be 0.0 < x < 1.0
type Value float64

type Model struct {
	// Left side color of progress bar. By default, it's #00dbde
	StartColor colorful.Color

	// Left side color of progress bar. By default, it's #fc00ff
	EndColor colorful.Color

	// Width of progress bar in symbols.
	Width int

	// Which part of bar need to visualise. Can be 0.0 < Progress < 1.0, if value is bigger or smaller, it'll
	// be mapped to this values
	Progress float64

	// filament rune of done part of progress bar. it's █ by default
	FilamentSymbol rune

	// empty rune of pending part of progress bar. it's ░ by default
	EmptySymbol rune

	// if true, gradient will be setted from start to end of filled part. Instead, it'll work
	// on all proggress bar length
	FullGradientMode bool
}

// NewModel returns a model with default values.
func NewModel(size int) *Model {
	startColor, _ := colorful.Hex("#00dbde")
	endColor, _ := colorful.Hex("#fc00ff")
	return &Model{
		StartColor:     startColor,
		EndColor:       endColor,
		Width:         size,
		FilamentSymbol: '█',
		EmptySymbol:    '░',
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch cmd := msg.(type) {
	case Value:
		if cmd > 1 {
			cmd = 1
		}
		if cmd < 0 {
			cmd = 0
		}

		m.Progress = float64(cmd)
	}

	return m, nil
}

func (m *Model) View() string {
	ramp := make([]string, int(float64(m.Width)*m.Progress))
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
