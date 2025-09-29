package scrollbar

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
)

func TestScrollbar(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
		length  int
		visible int
		offset  int
	}{
		{
			name:    "vertical-10perc-start",
			options: []Option{WithPosition(Vertical)},
			length:  100, visible: 10, offset: 0,
		},
		{
			name:    "vertical-10perc-middle",
			options: []Option{WithPosition(Vertical)},
			length:  100, visible: 10, offset: 49,
		},
		{
			name:    "vertical-10perc-end",
			options: []Option{WithPosition(Vertical)},
			length:  100, visible: 10, offset: 91,
		},
		{
			name:    "horizontal-10perc-start",
			options: []Option{WithPosition(Horizontal)},
			length:  100, visible: 10, offset: 0,
		},
		{
			name:    "horizontal-10perc-middle",
			options: []Option{WithPosition(Horizontal)},
			length:  100, visible: 10, offset: 49,
		},
		{
			name:    "horizontal-10perc-end",
			options: []Option{WithPosition(Horizontal)},
			length:  100, visible: 10, offset: 91,
		},
		{
			name:    "vertical-33perc-start",
			options: []Option{WithPosition(Vertical)},
			length:  30, visible: 9, offset: 0,
		},
		{
			name:    "vertical-33perc-middle",
			options: []Option{WithPosition(Vertical)},
			length:  30, visible: 9, offset: 9,
		},
		{
			name:    "vertical-33perc-end",
			options: []Option{WithPosition(Vertical)},
			length:  30, visible: 9, offset: 21,
		},
		{
			name:    "horizontal-33perc-start",
			options: []Option{WithPosition(Horizontal)},
			length:  30, visible: 9, offset: 0,
		},
		{
			name:    "horizontal-33perc-middle",
			options: []Option{WithPosition(Horizontal)},
			length:  30, visible: 9, offset: 9,
		},
		{
			name:    "horizontal-33perc-end",
			options: []Option{WithPosition(Horizontal)},
			length:  30, visible: 9, offset: 21,
		},
	}

	for _, tc := range tests {
		// basic block bar.
		t.Run("block-"+tc.name, func(t *testing.T) {
			model := New(append(tc.options, WithType(BlockBar()))...)
			switch model.Position() {
			case Vertical:
				model.SetHeight(tc.visible)
			case Horizontal:
				model.SetWidth(tc.visible)
			}

			model.SetContentState(tc.length, tc.visible, tc.offset)
			golden.RequireEqual(t, ansi.Strip(model.View()))
		})

		// slim circles bar.
		t.Run("slim-circles-"+tc.name, func(t *testing.T) {
			model := New(append(tc.options, WithType(SlimCirclesBar()))...)
			switch model.Position() {
			case Vertical:
				model.SetHeight(tc.visible)
			case Horizontal:
				model.SetWidth(tc.visible)
			}

			model.SetContentState(tc.length, tc.visible, tc.offset)
			golden.RequireEqual(t, ansi.Strip(model.View()))
		})
	}
}
