package colorpicker

import (
	"fmt"
	"image/color"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	colorconv "github.com/charmbracelet/x/exp/color"
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

	var v Value
	v.Label = s.Base.
		Foreground(lipgloss.Color("241"))
	v.Value = s.Base.
		Foreground(lipgloss.Color("241"))
	v.Unit = s.Base.
		Foreground(lipgloss.Color("241"))
	v.Decr = s.Base.
		Foreground(lipgloss.Color("205")).
		SetString("◀ ")
	v.Incr = s.Base.
		Foreground(lipgloss.Color("82")).
		SetString(" ▶")

	s.FocusedValue = v
	s.BlurredValue = v

	return s
}

type Model struct {
	Mode   mode
	Styles Styles
	color  colorconv.Color
	labels []string
	inputs []int
	index  int
}

func (m Model) Init() (tea.Model, tea.Cmd) {
	return m.InitAs()
}

func (m Model) InitAs() (Model, tea.Cmd) {
	m.Styles = DefaultStyles()
	m.inputs = make([]int, 3)
	m.SetColor(color.RGBA{0, 0, 0, 255})
	return m, nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.UpdateAs(msg)
}

func (m Model) UpdateAs(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
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
				r, g, b, _ := m.color.RGBA()
				m.inputs[0] = int(r)
				m.inputs[1] = int(g)
				m.inputs[2] = int(b)
				m.Mode = modeHSV
			case modeHSV:
				h, s, v := m.color.HSV()
				m.inputs[0] = int(h)
				m.inputs[1] = int(s * 100)
				m.inputs[2] = int(v * 100)
				m.Mode = modeRGB
			}
			return m, nil

		case msg.Code == tea.KeyTab && msg.Mod == tea.ModShift:
			m.inputs[m.index]++
			return m, nil

		case msg.Code == tea.KeyLeft || msg.Code == 'h':
			decr := 1
			if msg.Mod == tea.ModShift {
				decr = 10
			}
			m.inputs[m.index] -= decr

			switch m.Mode {
			case modeRGB:
				if m.inputs[m.index] < 0 {
					m.inputs[m.index] = 255
				}
			case modeHSV:
				switch m.index {
				case 0:
					if m.inputs[m.index] < 0 {
						m.inputs[m.index] = 360
					}
				case 1, 2:
					if m.inputs[m.index] < 0 {
						m.inputs[m.index] = 100
					}
				}
			}

			m.updateColorFromInputs()
			return m, nil

		case msg.Code == tea.KeyRight || msg.Code == 'l':
			incr := 1
			if msg.Mod == tea.ModShift {
				incr = 10
			}
			m.inputs[m.index] += incr

			switch m.Mode {
			case modeRGB:
				if m.inputs[m.index] > 255 {
					m.inputs[m.index] = 0
				}
			case modeHSV:
				switch m.index {
				case 0:
					if m.inputs[m.index] > 360 {
						m.inputs[m.index] = 0
					}
				case 1, 2:
					if m.inputs[m.index] > 100 {
						m.inputs[m.index] = 0
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
		m.color.FromRGB(
			uint8(m.inputs[0]),
			uint8(m.inputs[1]),
			uint8(m.inputs[2]),
		)
	case modeHSV:
		m.color.FromHSV(
			float64(m.inputs[0]),
			float64(m.inputs[1])/100,
			float64(m.inputs[2])/100,
		)
	}
}

func (i *Model) SetColor(c color.Color) {
	r, g, b, _ := c.RGBA()
	i.color.FromRGB(uint8(r), uint8(g), uint8(b))
	switch i.Mode {
	case modeRGB:
		r, g, b, _ := c.RGBA()
		i.inputs[0] = int(r)
		i.inputs[1] = int(g)
		i.inputs[2] = int(b)
	case modeHSV:
		h, s, v := i.color.HSV()
		i.inputs[0] = int(h)
		i.inputs[1] = int(s)
		i.inputs[2] = int(v)
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

	// ansi256 := colorprofile.ANSI256.Convert(i.color.Color)
	// ansi := colorprofile.ANSI.Convert(i.color.Color)

	var (
		b strings.Builder
		s = m.Styles
		p = func(args ...any) {
			fmt.Fprint(&b, args...)
		}
	)

	const block = "        "

	c := lipgloss.Color(m.Hex())
	fg := s.Base.Foreground(c)
	bg := s.Base.Background(c)
	p(bg.Render(block))
	p(fg.Render(" " + m.Hex()))

	// b.WriteString(lipgloss.NewStyle().Background(lipgloss.Color(ansi256)).Render(block))
	// b.WriteString(" ")
	// b.WriteString(lipgloss.NewStyle().Background(lipgloss.Color(ansi)).Render(block))

	b.WriteString("\n\n")

	for j, in := range m.inputs {
		var v Value
		if m.index == j {
			v = s.FocusedValue
			p(s.FocusedValue.Decr.String())
		} else {
			v = s.BlurredValue
			p(s.Base.Render("  "))
		}
		p(v.Label.Render(labels[j] + " "))
		p(v.Value.Render(fmt.Sprintf("%3d", in)))
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
		if m.index == j {
			p(v.Incr.String())
		} else {
			p(s.Base.Render("  "))
		}
	}
	return s.Frame.Render(b.String())
}

func (i Model) Hex() string {
	return i.color.Hex()
}
