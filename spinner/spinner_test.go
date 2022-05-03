package spinner_test

import (
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
)

func TestSpinnerNew(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		s := spinner.New()

		exp, got := spinner.Line, s.Spinner

		if exp.FPS != got.FPS {
			t.Errorf("expecting %d FPS, got %d", exp.FPS, got.FPS)
		}

		if e, g := len(exp.Frames), len(got.Frames); e != g {
			t.Fatalf("expecting %d frames, got %d", e, g)
		}

		for i, e := range exp.Frames {
			if g := got.Frames[i]; e != g {
				t.Errorf("expecting frame index %d with value %q, got %q", i, e, g)
			}
		}
	})
}
