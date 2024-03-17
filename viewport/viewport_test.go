package viewport

import (
	"strings"
	"testing"
	"unicode"

	"github.com/MakeNowJust/heredoc"
	"github.com/acarl005/stripansi"
	"github.com/charmbracelet/lipgloss"
)

const (
	viewportH = 8
	viewportW = 15
	content   = "line 1\nline 2\nline 3\nline 4\nline 5\nline 6\nline 7\nline 8\nline 9\nline 10"
)

func TestMaxYOffset(t *testing.T) {
	type want struct {
		maxYOffset    int
		scrollTopView string
		scrollBotView string
	}

	tests := []struct {
		name  string
		style lipgloss.Style
		want  want
	}{
		{
			name:  "no style",
			style: lipgloss.NewStyle(),
			want: want{
				maxYOffset: 2,
				scrollTopView: heredoc.Doc(`
					line 1
					line 2
					line 3
					line 4
					line 5
					line 6
					line 7
					line 8
				`),
				scrollBotView: heredoc.Doc(`
					line 3
					line 4
					line 5
					line 6
					line 7
					line 8
					line 9
					line 10
				`),
			},
		},
		{
			name:  "with single border",
			style: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()),
			want: want{
				maxYOffset: 4,
				scrollTopView: heredoc.Doc(`
					╭─────────────╮
					│line 1       │
					│line 2       │
					│line 3       │
					│line 4       │
					│line 5       │
					│line 6       │
					╰─────────────╯
				`),
				scrollBotView: heredoc.Doc(`
					╭─────────────╮
					│line 5       │
					│line 6       │
					│line 7       │
					│line 8       │
					│line 9       │
					│line 10      │
					╰─────────────╯
				`),
			},
		},
		{
			name:  "with border + padding",
			style: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 1),
			want: want{
				maxYOffset: 6,
				scrollTopView: heredoc.Doc(`
					╭─────────────╮
					│             │
					│ line 1      │
					│ line 2      │
					│ line 3      │
					│ line 4      │
					│             │
					╰─────────────╯
				`),
				scrollBotView: heredoc.Doc(`
					╭─────────────╮
					│             │
					│ line 7      │
					│ line 8      │
					│ line 9      │
					│ line 10     │
					│             │
					╰─────────────╯
				`),
			},
		},
		{
			name:  "with border + margin",
			style: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Margin(1, 0),
			want: want{
				maxYOffset: 6,
				scrollTopView: heredoc.Doc(`

					╭─────────────╮
					│line 1       │
					│line 2       │
					│line 3       │
					│line 4       │
					╰─────────────╯

				`),
				scrollBotView: heredoc.Doc(`

					╭─────────────╮ 
					│line 7       │ 
					│line 8       │ 
					│line 9       │ 
					│line 10      │ 
					╰─────────────╯

				`),
			},
		},
		{
			name:  "with border + margin + padding",
			style: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Margin(1, 0).Padding(1, 1),
			want: want{
				maxYOffset: 8,
				scrollTopView: heredoc.Doc(`

					╭─────────────╮
					│             │
					│ line 1      │
					│ line 2      │
					│             │
					╰─────────────╯

				`),
				scrollBotView: heredoc.Doc(`

					╭─────────────╮
					│             │
					│ line 9      │
					│ line 10     │
					│             │
					╰─────────────╯

				`),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viewport := New(viewportW, viewportH)
			viewport.Style = tt.style.Copy()
			viewport.SetContent(content)

			viewport, _ = viewport.Update(nil)

			maxYOffset := viewport.maxYOffset()
			if maxYOffset != tt.want.maxYOffset {
				t.Fatalf("\nWant maxYOffset:\n%v\nGot:\n%v\n", tt.want.maxYOffset, maxYOffset)
			}

			viewport.SetYOffset(0)
			viewTop := stripString(viewport.View())
			wantViewTop := stripString(tt.want.scrollTopView)

			if viewTop != wantViewTop {
				t.Fatalf("Want view (when scrolled to top):\n%v\nGot:\n%v\n", wantViewTop, viewTop)
			}

			viewport.SetYOffset(100)
			viewBot := stripString(viewport.View())
			wantViewBot := stripString(tt.want.scrollBotView)

			if viewBot != wantViewBot {
				t.Fatalf("Want view (when scrolled to bottom):\n%v\nGot:\n%v\n", wantViewBot, viewBot)
			}
		})
	}
}

func stripString(str string) string {
	s := stripansi.Strip(str)
	ss := strings.Split(s, "\n")

	var lines []string
	for _, l := range ss {
		trim := strings.TrimRightFunc(l, unicode.IsSpace)
		if trim != "" {
			lines = append(lines, trim)
		}
	}

	return strings.Join(lines, "\n")
}
