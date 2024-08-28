package table

import "testing"

func TestTruncate(t *testing.T) {
	tt := map[string]struct {
		in    string
		width int
		want  string
	}{
		"empty string": {
			in:    "",
			width: 0,
			want:  "",
		},
		"string without escape sequence; not truncated": {
			in:    "hello",
			width: 5,
			want:  "hello",
		},
		"string without escape sequence; truncated": {
			in:    "hello",
			width: 3,
			want:  "he…",
		},
		"string with escape sequence; not truncated": {
			in:    "\x1b[31mhello\x1b[0m",
			width: 5,
			want:  "\u001B[31mhello\u001B[0m",
		},
		"string with escape sequence; truncated": {
			// TODO(?): should the ellipsis be included inside the escape sequence?

			in:    "\x1b[31mhello\x1b[0m",
			width: 3,
			want:  "\u001B[31mhe\u001B[0m…",
		},
		"string with escape sequence and emoji; not truncated": {
			in:    "\x1b[31m👋\x1b[0m",
			width: 2,
			want:  "\u001B[31m👋\u001B[0m",
		},
		// TODO: fix the following emoji test case
		//"string with escape sequence, emoji, and combining character; truncated": {
		//	// NOTE: The combining character may not be rendered correctly in the test output.
		//	in:    "\x1b[31m👋\u0308\x1b[0m",
		//	width: 2,
		//	want:  "\x1b[31m👋\x1b[0m…",
		//},
		//"string with escape sequence, emoji, and non-print character": {
		//	in:    "\x1b[31m👋\u0000\x1b[0m",
		//	width: 1,
		//	want:  "\x1b[31m👋\x1b[0m…",
		//},
		"string with double escape sequences; truncated between": {
			in:    "\x1b[31mhello\x1b[0m\x1b[31mhello\x1b[0m",
			width: 6,
			want:  "\u001B[31mhello\u001B[0m\u001B[31m\u001B[0m…",
		},
		"string with double escape sequences; truncated inside first": {
			in:    "\x1b[31mhello\x1b[0m\x1b[31mhello\x1b[0m",
			width: 3,
			want:  "\u001B[31mhe\u001B[0m\u001B[31m\u001B[0m…",
		},
		"string with double escape sequences; truncated inside second": {
			in:    "\x1b[31mhello\x1b[0m\x1b[31mhello\x1b[0m",
			width: 7,
			want:  "\u001B[31mhello\u001B[0m\u001B[31mh\u001B[0m…",
		},
		"string with newline; truncated before": {
			in:    "hello\nworld",
			width: 4,
			want:  "hel…",
		},
		"string with newline; truncated after": {
			in:    "hello\nworld",
			width: 8,
			want:  "hello\nwo…",
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			got := truncate(tc.in, tc.width)

			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLengthWithoutEscapeSequences(t *testing.T) {
	tt := map[string]struct {
		in   string
		want int
	}{
		"empty string": {
			in:   "",
			want: 0,
		},
		"string without escape sequence": {
			in:   "hello",
			want: 5,
		},
		"string with escape sequence": {
			in:   "\x1b[31mhello\x1b[0m",
			want: 5,
		},
		"string with escape sequence and emoji": {
			in:   "\x1b[31m👋\x1b[0m",
			want: 2,
		},
		"string with escape sequence, emoji, and combining character": {
			in:   "\x1b[31m👋\u0308\x1b[0m",
			want: 2,
		},
		"string with escape sequence, emoji, and non-print character": {
			in:   "\x1b[31m👋\u0000\x1b[0m",
			want: 2,
		},
		"string with double escape sequences": {
			in:   "\x1b[31mhello\x1b[0m\x1b[31mhello\x1b[0m",
			want: 10,
		},
		"string with newline": {
			in:   "hello\nworld",
			want: 10,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			got := lengthWithoutEscapeSequences(tc.in)

			if got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestExtractEscapeSequences(t *testing.T) {
	tt := map[string]struct {
		in            string
		wantStr       string
		wantSequences []escapeSequence
	}{
		"empty string": {
			in:            "",
			wantStr:       "",
			wantSequences: []escapeSequence{},
		},
		"string without escape sequence": {
			in:            "hello",
			wantStr:       "hello",
			wantSequences: []escapeSequence{},
		},
		"string with escape sequence": {
			in:      "\x1b[31mhello\x1b[0m",
			wantStr: "hello",
			wantSequences: []escapeSequence{
				{
					content: "\x1b[31m",
					pos:     0,
				},
				{
					content: "\x1b[0m",
					pos:     5,
				},
			},
		},
		"string with escape sequence and emoji": {
			in:      "\x1b[31m👋\x1b[0m",
			wantStr: "👋",
			wantSequences: []escapeSequence{
				{
					content: "\x1b[31m",
					pos:     0,
				},
				{
					content: "\x1b[0m",
					pos:     1,
				},
			},
		},
		"string with escape sequence, emoji, and combining character": {
			// NOTE: The combining character may not be rendered correctly in the test output.
			in:      "\x1b[31m👋\u0308\x1b[0m",
			wantStr: "👋\u0308",
			wantSequences: []escapeSequence{
				{
					content: "\x1b[31m",
					pos:     0,
				},
				{
					content: "\x1b[0m",
					pos:     2,
				},
			},
		},
		"string with escape sequence, emoji, and non-print character": {
			in:      "\x1b[31m👋\u0000\x1b[0m",
			wantStr: "👋\u0000",
			wantSequences: []escapeSequence{
				{
					content: "\x1b[31m",
					pos:     0,
				},
				{
					content: "\x1b[0m",
					pos:     2,
				},
			},
		},
		"string with double escape sequences": {
			in:      "\x1b[31mhello\x1b[0m\x1b[31mhello\x1b[0m",
			wantStr: "hellohello",
			wantSequences: []escapeSequence{
				{
					content: "\x1b[31m",
					pos:     0,
				},
				{
					content: "\x1b[0m",
					pos:     5,
				},
				{
					content: "\x1b[31m",
					pos:     5,
				},
				{
					content: "\x1b[0m",
					pos:     10,
				},
			},
		},
		"string with newline": {
			in:            "hello\nworld",
			wantStr:       "hello\nworld",
			wantSequences: []escapeSequence{},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			gotStr, gotSequences := extractEscapeSequences(tc.in)

			if gotStr != tc.wantStr {
				t.Errorf("string: got %q, want %q", gotStr, tc.wantStr)
			}

			if len(gotSequences) != len(tc.wantSequences) {
				t.Errorf("escape sequences: got %d, want %d", len(gotSequences), len(tc.wantSequences))
			}

			for i := range gotSequences {
				if gotSequences[i].content != tc.wantSequences[i].content {
					t.Errorf("sequence %d content: got %q, want %q", i, gotSequences[i].content, tc.wantSequences[i].content)
				}

				if gotSequences[i].pos != tc.wantSequences[i].pos {
					t.Errorf("sequence %d pos: got %d, want %d", i, gotSequences[i].pos, tc.wantSequences[i].pos)
				}
			}
		})
	}
}

func TestApplyEscapeSequences(t *testing.T) {
	tt := map[string]struct {
		in        string
		sequences []escapeSequence
		want      string
	}{
		"empty string": {
			in:        "",
			sequences: []escapeSequence{},
			want:      "",
		},
		"string without escape sequence": {
			in:        "hello",
			sequences: []escapeSequence{},
			want:      "hello",
		},
		"string with escape sequence": {
			in: "hello",
			sequences: []escapeSequence{
				{
					content: "\x1b[31m",
					pos:     0,
				},
				{
					content: "\x1b[0m",
					pos:     5,
				},
			},
			want: "\u001B[31mhello\u001B[0m",
		},
		"string with escape sequence and emoji": {
			in: "👋",
			sequences: []escapeSequence{
				{
					content: "\x1b[31m",
					pos:     0,
				},
				{
					content: "\x1b[0m",
					pos:     1,
				},
			},
			want: "\u001B[31m👋\u001B[0m",
		},
		"string with escape sequence, emoji, and combining character": {
			// NOTE: The combining character may not be rendered correctly in the test output.
			in: "👋\u0308",
			sequences: []escapeSequence{
				{
					content: "\x1b[31m",
					pos:     0,
				},
				{
					content: "\x1b[0m",
					pos:     2,
				},
			},
			want: "\u001B[31m👋̈\u001B[0m",
		},
		"string with escape sequence, emoji, and non-print character": {
			in: "👋\u0000",
			sequences: []escapeSequence{
				{
					content: "\x1b[31m",
					pos:     0,
				},
				{
					content: "\x1b[0m",
					pos:     2,
				},
			},
			want: "\u001B[31m👋\u0000\u001B[0m",
		},
		"string with double escape sequences": {
			in: "hellohello",
			sequences: []escapeSequence{
				{
					content: "\x1b[31m",
					pos:     0,
				},
				{
					content: "\x1b[0m",
					pos:     5,
				},
				{
					content: "\x1b[31m",
					pos:     5,
				},
				{
					content: "\x1b[0m",
					pos:     10,
				},
			},
			want: "\u001B[31mhello\u001B[0m\u001B[31mhello\u001B[0m",
		},
		"string with newline": {
			in:        "hello\nworld",
			sequences: []escapeSequence{},
			want:      "hello\nworld",
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			got := applyEscapeSequences(tc.in, tc.sequences)

			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
