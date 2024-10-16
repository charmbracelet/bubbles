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
		"full sentence exceeding width": {
			lines: []string{"hello bob, I like yogurt in the mornings."},
			width: 12,
			want:  []string{"hello bob, I", "like yogurt", "in the", "mornings."},
		},
		"whitespace of head of line is preserved": {
			lines: []string{" aaa", "bbb", "ccc"},
			width: 5,
			want:  []string{" aaa", "bbb", "ccc"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := wrap(tt.lines, tt.width)

			if len(got) != len(tt.want) {
				t.Fatalf("expected len is %d but got %d", len(tt.want), len(got))
			}
			for i := range tt.want {
				if tt.want[i] != got[i] {
					t.Fatalf("expected %q but got %q", tt.want[i], got[i])
				}
			}
		})
	}
}
