package colorpicker

import (
	"fmt"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

type mode int

const (
	modeRGB mode = iota
	modeHSV
)

type Styles struct {
	Base         lipgloss.Style
	Frame        lipgloss.Style
	FocusedValue Value
	BlurredValue Value
	ActiveValue  Value
}

type Value struct {
	Label lipgloss.Style
	Value lipgloss.Style
	Unit  lipgloss.Style
	Decr  lipgloss.Style
	Incr  lipgloss.Style
}

func DefaultStyles() (s Styles) {
	s.Base = lipgloss.NewStyle().
		Background(lipgloss.Color("#191919")).
		Foreground(lipgloss.Color("#F2F2F2"))
	s.Frame = s.Base.
		Margin(2, 4).
		Padding(2, 4)

		// Blurred.
	var v Value
	v.Label = s.Base.
		Foreground(lipgloss.Color("241"))
	v.Value = s.Base.
		Foreground(lipgloss.Color("241"))
	v.Unit = s.Base.
		Foreground(lipgloss.Color("241"))
	v.Decr = s.Base.
		Foreground(lipgloss.Color("#3e3e3e")).
		SetString("◀ ")
	v.Incr = v.Decr.
		SetString(" ▶")
	s.BlurredValue = v

	// Focused
	v.Value = v.Value.
		Foreground(lipgloss.Color("#a1a1a1"))
	v.Unit = v.Unit.
		Foreground(lipgloss.Color("#878787"))
	v.Label = v.Unit
	s.FocusedValue = v

	// Active
	v.Incr = v.Incr.Foreground(lipgloss.Color("#ff6e9e"))
	v.Decr = v.Incr
	s.ActiveValue = v

	return s
}

type keypress int

const (
	pressNone keypress = iota
	pressDecr
	pressIncr
)

type Model struct {
	Mode                 mode
	Styles               Styles
	color                colorful.Color
	labels               []string
	inputs               []float64
	index                int
	keydown              keypress
	keyboardEnhancements bool
}

func (m Model) Init() (tea.Model, tea.Cmd) {
	return m.InitAs()
}

func (m Model) InitAs() (Model, tea.Cmd) {
	m.Styles = DefaultStyles()
	m.inputs = make([]float64, 3)
	m.SetColor(colorful.Color{R: 1, G: 0, B: 0})
	return m, nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.UpdateAs(msg)
}

func (m Model) UpdateAs(msg tea.Msg) (Model, tea.Cmd) {
	m.keydown = pressNone

	switch msg := msg.(type) {
	case tea.KeyboardEnhancementsMsg:
		m.keyboardEnhancements = msg.SupportsKeyReleases()
		return m, tea.Println("Keyboard enhancements supported. Release? %v", m.keyboardEnhancements)
	case tea.KeyPressMsg:
		switch {
		case msg.Code == tea.KeyTab:
			m.index++
			if m.index >= len(m.inputs) {
				m.index = 0
			}
			return m, nil

		case msg.Code == tea.KeySpace:
			m.updateColorFromInputs()
			switch m.Mode {
			case modeRGB:
				// The user is switching from HSV to RGB.
				m.Mode = modeHSV
				m.Mode = modeHSV
				h, s, v := m.color.Hsv()
				m.inputs[0], m.inputs[1], m.inputs[2] = h, s*100, v*100
			case modeHSV:
				// The user is switching from RGB to HSV.
				m.Mode = modeRGB
				r, g, b := m.color.RGB255()
				m.inputs[0] = float64(r)
				m.inputs[1] = float64(g)
				m.inputs[2] = float64(b)
			}
			m.updateColorFromInputs()
			return m, nil

		case msg.Code == tea.KeyTab && msg.Mod == tea.ModShift:
			m.inputs[m.index]++
			return m, nil

		case msg.Code == tea.KeyLeft || msg.Code == 'h':
			if m.keyboardEnhancements {
				m.keydown = pressDecr
			}

			var decr float64 = 1
			if msg.Mod == tea.ModShift {
				decr = 10
			}
			m.inputs[m.index] -= decr

			if m.inputs[m.index] < 0 {
				m.inputs[m.index] = 0
			}

			m.updateColorFromInputs()
			return m, nil

		case msg.Code == tea.KeyRight || msg.Code == 'l':
			if m.keyboardEnhancements {
				m.keydown = pressIncr
			}

			var incr float64 = 1
			if msg.Mod == tea.ModShift {
				incr = 10
			}
			m.inputs[m.index] += incr

			switch m.Mode {
			case modeRGB:
				if m.inputs[m.index] > 255 {
					m.inputs[m.index] = 255
				}
			case modeHSV:
				switch m.index {
				case 0:
					if m.inputs[m.index] > 359 {
						m.inputs[m.index] = 359
					}
				case 1, 2:
					if m.inputs[m.index] > 100 {
						m.inputs[m.index] = 100
					}
				}
			}

			m.updateColorFromInputs()
			return m, nil

		case msg.Code == tea.KeyEnter:
			return m, tea.SetClipboard(m.Hex())
		}
	}

	return m, nil
}

func (m *Model) updateColorFromInputs() {
	switch m.Mode {
	case modeRGB:
		// Colorful.Color uses RGB values between 0 and 1, so we need to
		// convert to a 0-255 scale.
		r, g, b := m.inputs[0], m.inputs[1], m.inputs[2]
		m.color = colorful.Color{R: r / 255, G: g / 255, B: b / 255}
	case modeHSV:
		h, s, v := m.inputs[0], m.inputs[1], m.inputs[2]
		m.color = colorful.Hsv(h, s/100, v/100)
	}
}

func (i *Model) SetColor(c colorful.Color) {
	i.color = c
	switch i.Mode {
	case modeRGB:
		r, g, b := c.RGB255()
		i.inputs[0] = float64(r)
		i.inputs[1] = float64(g)
		i.inputs[2] = float64(b)
	case modeHSV:
		i.inputs[0], i.inputs[1], i.inputs[2] = i.color.Hsv()
	}
	return
}

func (m Model) View() string {
	labels := make([]string, len(m.inputs))
	switch m.Mode {
	case modeRGB:
		labels = []string{"R", "G", "B"}
	case modeHSV:
		labels = []string{"H", "S", "V"}
	}

	var (
		b strings.Builder
		s = m.Styles
		p = func(args ...any) {
			fmt.Fprint(&b, args...)
		}
		ansi256Val = fmt.Sprintf("%d", colorprofile.ANSI256.Convert(m.color))
		ansiVal    = fmt.Sprintf("%d", colorprofile.ANSI.Convert(m.color))
	)

	const block = "        "

	// TrueColor
	var truecolor strings.Builder
	{
		c := lipgloss.Color(m.Hex())
		fg := s.Base.Foreground(c)
		bg := s.Base.Background(c)
		truecolor.WriteString("TrueColor      \n\n")
		truecolor.WriteString(bg.Render(block))
		truecolor.WriteString("\n")
		truecolor.WriteString(fg.Render(m.Hex()))
	}

	// ANSI256
	var ansi256 strings.Builder
	{
		c := lipgloss.Color(ansi256Val)
		fg := s.Base.Foreground(c).Render
		bg := s.Base.Background(c).Render
		ansi256.WriteString("ANSI 256    \n\n")
		ansi256.WriteString(bg(block))
		ansi256.WriteString("\n")
		ansi256.WriteString(fg(ansi256Val))
	}

	// ANSI
	var ansi strings.Builder
	{
		c := lipgloss.Color(ansiVal)
		fg := s.Base.Foreground(c).Render
		bg := s.Base.Background(c).Render
		ansi.WriteString("ANSI 4-Bit\n\n")
		ansi.WriteString(bg(block))
		ansi.WriteString("\n")
		ansi.WriteString(fg(ansiVal))
	}

	p(lipgloss.JoinHorizontal(lipgloss.Top, truecolor.String(), ansi256.String(), ansi.String()))

	b.WriteString("\n\n")

	// Adjustment UI.
	for j, in := range m.inputs {
		var v Value

		if m.index == j {
			v = s.FocusedValue
			p(m.decrView())
		} else {
			v = s.BlurredValue
			p(s.Base.Render("  "))
		}
		p(v.Label.Render(labels[j] + " "))
		p(v.Value.Render(fmt.Sprintf("%3.f", in)))
		if m.Mode == modeHSV {
			switch j {
			case 0:
				p(v.Unit.Render("°"))
			case 1, 2:
				p(v.Unit.Render("%"))
			}
		} else {
			p(v.Unit.Render(" "))
		}

		// Incrementer.
		var renderIncrementer bool
		if m.index == j {
			switch m.Mode {
			case modeRGB:
				renderIncrementer = m.inputs[m.index] < 255
			case modeHSV:
				switch j {
				case 0:
					renderIncrementer = m.inputs[m.index] < 359
				case 1, 2:
					renderIncrementer = m.inputs[m.index] < 100
				}
			}
		}
		if renderIncrementer {
			incr := v.Incr
			if m.keydown == pressIncr {
				incr = s.ActiveValue.Incr
			}
			p(incr.String())
		} else {
			p(s.Base.Render("  "))
		}
	}

	// Spectrum.
	p("\n\n" + m.spectrumView() + "\n\n")

	return s.Frame.Render(b.String())
}

// decrView renders the decrementer for the active value.
func (m Model) decrView() string {
	s := m.Styles
	if m.inputs[m.index] <= 0 {
		return s.Base.Render("  ")
	}
	decr := s.FocusedValue.Decr
	if m.keydown == pressDecr {
		decr = s.ActiveValue.Decr
	}
	return decr.String()
}

func (m Model) incrView() string {
	var (
		s  = m.Styles
		ok bool
	)

	for i, v := range m.inputs {
		if m.index == i {
			switch m.Mode {
			case modeRGB:
				ok = v < 255
			case modeHSV:
				switch i {
				case 0:
					ok = v < 359
				case 1, 2:
					ok = v < 100
				}
			}
		}
	}

	if ok {
		return s.FocusedValue.Incr.String()
	}
	return s.Base.Render("  ")
}

func (m Model) spectrumView() string {
	h, s, v := m.color.Hsv()

	colorStops := []colorful.Color{
		colorful.Hsv(0, s, v),
		colorful.Hsv(90, s, v),
		colorful.Hsv(180, s, v),
		colorful.Hsv(270, s, v),
		colorful.Hsv(359.9, s, v),
	}

	const width = 40
	builders := make([]strings.Builder, 4)
	sectionWidth := width / len(builders)

	for i := range builders {
		for j := 0; j < sectionWidth; j++ {
			color := colorStops[i].BlendHsv(colorStops[i+1], float64(j)/float64(sectionWidth)).Hex()
			builders[i].WriteString(lipgloss.NewStyle().Background(lipgloss.Color(color)).Render(" "))
		}
	}

	pos := int(math.Floor(h / 360 * width))
	mark := strings.Repeat(" ", pos) + "▲"

	var view strings.Builder
	for _, b := range builders {
		view.WriteString(b.String())
	}
	view.WriteString("\n" + mark)
	return view.String()
}

func (i Model) Hex() string {
	return i.color.Hex()
}
