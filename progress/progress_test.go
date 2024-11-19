package progress

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

const (
	AnsiReset = "\x1b[0m"
)

func TestSolid(t *testing.T) {
	r := lipgloss.DefaultRenderer()
	r.SetColorProfile(termenv.TrueColor)

	tests := []struct {
		name     string
		width    int
		expected string
	}{
		{
			name:     "width 3",
			width:    3,
			expected: `[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;96;96;96m░[0m`,
		},
		{
			name:     "width 5",
			width:    5,
			expected: `[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m`,
		},
		{
			name:     "width 50",
			width:    50,
			expected: `[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;117;113;249m█[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m[38;2;96;96;96m░[0m`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := New(
				WithFillStyles(
					r.NewStyle().Foreground(lipgloss.Color("#7571F9")),
					r.NewStyle().Foreground(lipgloss.Color("#606060")),
				),
				WithoutPercentage(),
			)
			p.Width = test.width
			res := p.ViewAs(0.5)

			if res != test.expected {
				t.Errorf("expected view %q, instead got %q", test.expected, res)
			}
		})
	}
}

func TestGradient(t *testing.T) {

	r := lipgloss.DefaultRenderer()
	r.SetColorProfile(termenv.TrueColor)

	colA := "#FF0000"
	colB := "#00FF00"

	var p Model
	var descr string

	for _, scale := range []bool{false, true} {
		opts := []Option{
			WithFillStyles(
				r.NewStyle().Foreground(lipgloss.Color("#7571F9")),
				r.NewStyle().Foreground(lipgloss.Color("#606060")),
			),
			WithoutPercentage(),
		}
		if scale {
			descr = "progress bar with scaled gradient"
			opts = append(opts, WithScaledGradient(colA, colB))
		} else {
			descr = "progress bar with gradient"
			opts = append(opts, WithGradient(colA, colB))
		}

		t.Run(descr, func(t *testing.T) {
			p = New(opts...)

			// build the expected colors by colorizing an empty string and then cutting off the following reset sequence
			sb := strings.Builder{}
			sb.WriteString(r.NewStyle().Foreground(lipgloss.Color(colA)).Render(""))
			expFirst := strings.Split(sb.String(), AnsiReset)[0]
			sb.Reset()
			sb.WriteString(r.NewStyle().Foreground(lipgloss.Color(colB)).Render(""))
			expLast := strings.Split(sb.String(), AnsiReset)[0]

			for _, width := range []int{3, 5, 50} {
				p.Width = width
				res := p.ViewAs(1.0)

				// extract colors from the progress bar by splitting at p.Full+AnsiReset, leaving us with just the color sequences
				colors := strings.Split(res, string(p.Full)+AnsiReset)

				// discard the last color, because it is empty (no new color comes after the last char of the bar)
				colors = colors[0 : len(colors)-1]

				if expFirst != colors[0] {
					t.Errorf("expected first color of bar to be first gradient color %q, instead got %q", expFirst, colors[0])
				}

				if expLast != colors[len(colors)-1] {
					t.Errorf("expected last color of bar to be second gradient color %q, instead got %q", expLast, colors[len(colors)-1])
				}
			}
		})
	}

}
