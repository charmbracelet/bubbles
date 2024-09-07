package viewport

import (
	"testing"
)

func TestWrap(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		lines []string
		width int
		want  []string
	}{
		"empty slice": {
			lines: []string{},
			width: 3,
			want:  []string{},
		},
		"all lines are within width": {
			lines: []string{"aaa", "bbb", "ccc"},
			width: 3,
			want:  []string{"aaa", "bbb", "ccc"},
		},
		"some lines exceeds width": {
			lines: []string{"aaaaaa", "bbbbbbbb", "ccc"},
			width: 3,
			want:  []string{"aaa", "aaa", "bbb", "bbb", "bb", "ccc"},
		},
	}

	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := wrap(tt.lines, tt.width)

			if len(got) != len(tt.want) {
				t.Errorf("expected len is %d but got %d", len(tt.want), len(got))
			}
			for i := range tt.want {
				if tt.want[i] != got[i] {
					t.Errorf("expected %s but got %s", tt.want[i], got[i])
				}
			}
		})
	}
}
