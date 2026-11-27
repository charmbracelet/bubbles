package textarea

import (
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestSetCursorColumnSnapsToGraphemeStart(t *testing.T) {
	m := newTextArea()
	m.SetValue("A👩‍💻B")
	m.SetCursorColumn(2)
	if m.col != 1 {
		t.Fatalf("cursor inside emoji: got %d want 1", m.col)
	}
}

func TestVerticalCursorSnapsToGraphemeStart(t *testing.T) {
	m := newTextArea()
	m.SetValue("ab\n👩‍💻x")
	m.SetCursorColumn(0)
	m.CursorUp()
	m.SetCursorColumn(1)
	m.CursorDown()
	if start, _ := graphemeAt(m.value[m.row], m.col); start != m.col {
		t.Fatalf("vertical cursor split cluster at %d", m.col)
	}
}

func TestBackspaceRemovesMarkWithoutConsonant(t *testing.T) {
	for _, value := range []string{"น้", "कि"} {
		m := newTextArea()
		m.SetValue(value)
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
		want := string([]rune(value)[:len([]rune(value))-1])
		if m.Value() != want {
			t.Fatalf("backspace %q: got %q want %q", value, m.Value(), want)
		}
	}
}

func TestWordOperationsAdvanceOverGraphemes(t *testing.T) {
	for _, value := range []string{"नमस्ते rest", "น้ำ rest", "👩‍💻 rest", "e\u0301 rest"} {
		t.Run(value, func(t *testing.T) {
			m := newTextArea()
			m.SetValue(value)
			m.CursorStart()
			m.wordRight()
			if m.col == 0 {
				t.Fatal("wordRight made no progress")
			}
			if start, _ := graphemeAt(m.value[m.row], m.col); start != m.col {
				t.Fatalf("wordRight split cluster at %d", m.col)
			}
			m.CursorStart()
			m.deleteWordRight()
			if m.Value() != " rest" {
				t.Fatalf("deleteWordRight: got %q", m.Value())
			}
		})
	}
}

func TestBackspaceRemovesEmojiSequenceWhole(t *testing.T) {
	for value, want := range map[string]string{
		"x👍🏽":              "x",
		"x👨\u200d👩\u200d👧": "x",
		"x🇺🇸":              "x",
		"x❤\ufe0f":         "x",
		"x1\ufe0f\u20e3":   "x",
		"x🏴\U000e0067\U000e0062\U000e0073\U000e0063\U000e0074\U000e007f": "x",
		"x👨\u200d":  "x👨",
		"क्\u200d":  "क्",
		"x🧑\u200d🎄": "x",
		"क्\u200dष": "क्\u200d",
		"क्ष":       "क्",
	} {
		m := newTextArea()
		m.SetValue(value)
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
		if m.Value() != want {
			t.Errorf("backspace %q: got %q want %q", value, m.Value(), want)
		}
	}
}

func TestClickMapsToGraphemeStart(t *testing.T) {
	m := newTextArea()
	m.SetValue("A👩‍💻BC")
	for x, want := range []int{0, 1, 1, 4, 5, 6} {
		if got := m.PositionAt(x+m.gutterWidth(), 0).Col; got != want {
			t.Errorf("click at cell %d: got col %d want %d", x, got, want)
		}
	}
}

func TestVerticalCursorUsesGraphemeWidth(t *testing.T) {
	m := newTextArea()
	m.SetValue("abcdef\n👩‍💻xyz")
	m.CursorUp()
	m.SetCursorColumn(4)
	m.CursorDown()
	if m.col != 5 {
		t.Fatalf("down from cell 4: got col %d want 5", m.col)
	}
}

func TestTransposeSwapsGraphemes(t *testing.T) {
	m := newTextArea()
	m.SetValue("ab👩‍💻")
	m.transposeLeft()
	if got, want := m.Value(), "a👩‍💻b"; got != want {
		t.Fatalf("transpose: got %q want %q", got, want)
	}
}

func TestTypingComposesBeforeExistingText(t *testing.T) {
	for _, tc := range []struct{ value, typed, want string }{
		{"नमस्ते", "क्ष", "क्षनमस्ते"},
		{"🇸x", "🇺🇸", "🇺🇸🇸x"},
	} {
		m := newTextArea()
		m.SetValue(tc.value)
		m.CursorStart()
		for _, r := range tc.typed {
			m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		}
		if m.Value() != tc.want || m.col != len([]rune(tc.typed)) {
			t.Errorf("typing %q before %q: got %q col %d, want %q col %d",
				tc.typed, tc.value, m.Value(), m.col, tc.want, len([]rune(tc.typed)))
		}
	}
}

func TestWrapKeepsGraphemesWhole(t *testing.T) {
	for _, value := range []string{
		strings.Repeat("👩‍💻", 12),
		strings.Repeat("e\u0301", 30),
		strings.Repeat("नमस्ते", 8),
	} {
		runes := []rune(value)
		offset := 0
		for _, line := range wrap(runes, 10) {
			if start, _ := graphemeAt(runes, offset); start != offset {
				t.Errorf("wrap split a grapheme of %q at rune %d", value, offset)
			}
			offset += len(line)
		}
	}
}

func TestDeleteWordLeftKeepsSpaceGrapheme(t *testing.T) {
	m := newTextArea()
	m.SetValue("ab \u0301cd")
	m.deleteWordLeft()
	if got, want := m.Value(), "ab \u0301"; got != want {
		t.Fatalf("deleteWordLeft: got %q want %q", got, want)
	}
}

func TestIndicConjunctIsOneGrapheme(t *testing.T) {
	m := newTextArea()
	m.SetValue("क्षत्रिय")
	var stops []int
	for m.col > 0 {
		m.characterLeft(false)
		stops = append(stops, m.col)
	}
	if want := []int{7, 3, 0}; !slices.Equal(stops, want) {
		t.Fatalf("left stops: got %v want %v", stops, want)
	}
}
