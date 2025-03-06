package scrollbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"testing"
)

func TestVerticalView(t *testing.T) {
	tests := []struct {
		name    string
		total   int
		visible int
		offset  int
		view    string
	}{
		{
			name:    "ThirdTop",
			total:   9,
			visible: 3,
			offset:  0,
			view:    "█\n░\n░",
		},
		{
			name:    "ThirdMiddle",
			total:   9,
			visible: 3,
			offset:  3,
			view:    "░\n█\n░",
		},
		{
			name:    "ThirdBottom",
			total:   9,
			visible: 3,
			offset:  6,
			view:    "░\n░\n█",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var scrollbar tea.Model
			scrollbar = NewVertical()
			scrollbar, _ = scrollbar.Update(HeightMsg(test.visible))
			scrollbar, _ = scrollbar.Update(Msg{
				Total:   test.total,
				Visible: test.visible,
				Offset:  test.offset,
			})
			view := scrollbar.View()

			if view != test.view {
				t.Errorf("expected:\n%s\ngot:\n%s", test.view, view)
			}
		})
	}
}
