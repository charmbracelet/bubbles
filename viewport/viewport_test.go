package viewport

import (
	"testing"
)

func Test_countHeightBasedOnWidth(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		lines []string
		width int
		want  int
	}{
		"Empty lines": {
			lines: []string{},
			width: 0,
			want:  0,
		},
		"Lines within width": {
			lines: []string{"123", "123"},
			width: 5,
			want:  2,
		},
		"Lines over width": {
			lines: []string{"1234567890", "123"},
			width: 5,
			want:  3,
		},
	}

	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := countHeightBasedOnWidth(tt.lines, tt.width)
			if tt.want != got {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
