package textarea

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// assertMatchesFresh checks that a model whose wrap cache has been kept up to
// date incrementally renders and measures exactly like a model built by
// newModel that wrapped the same content from scratch.
func assertMatchesFresh(t *testing.T, m Model, newModel func() Model) {
	t.Helper()

	fresh := newModel()
	fresh.SetValue(m.Value())
	fresh.row = m.row
	fresh.col = m.col
	// The viewport clamps the offset against its content, so it needs content
	// before it can be scrolled to where m is scrolled.
	fresh.viewport.SetContentLines(fresh.viewLines())
	fresh.viewport.SetYOffset(m.viewport.YOffset())

	if got, want := m.Value(), fresh.Value(); got != want {
		t.Fatalf("value mismatch:\ngot:  %q\nwant: %q", got, want)
	}
	if got, want := m.totalVisualLines(), fresh.totalVisualLines(); got != want {
		t.Errorf("totalVisualLines = %d, want %d", got, want)
	}
	if got, want := m.cursorLineNumber(), fresh.cursorLineNumber(); got != want {
		t.Errorf("cursorLineNumber = %d, want %d", got, want)
	}
	if got, want := m.LineInfo(), fresh.LineInfo(); got != want {
		t.Errorf("LineInfo = %+v, want %+v", got, want)
	}
	if got, want := m.View(), fresh.View(); got != want {
		t.Errorf("view mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestWrapCacheStaysInSync(t *testing.T) {
	content := strings.Join([]string{
		"the quick brown fox jumps over the lazy dog",
		"short",
		"",
		"another line that is long enough to be soft wrapped a few times over",
		"tail",
	}, "\n")

	keys := map[string]tea.Msg{
		"enter":     tea.KeyPressMsg{Code: tea.KeyEnter},
		"backspace": tea.KeyPressMsg{Code: tea.KeyBackspace},
		"delete":    tea.KeyPressMsg{Code: tea.KeyDelete},
		"up":        tea.KeyPressMsg{Code: tea.KeyUp},
		"down":      tea.KeyPressMsg{Code: tea.KeyDown},
		"left":      tea.KeyPressMsg{Code: tea.KeyLeft},
		"right":     tea.KeyPressMsg{Code: tea.KeyRight},
		"home":      tea.KeyPressMsg{Code: tea.KeyHome},
		"end":       tea.KeyPressMsg{Code: tea.KeyEnd},
		"ctrl+k":    tea.KeyPressMsg{Code: 'k', Mod: tea.ModCtrl},
		"ctrl+u":    tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl},
		"alt+d":     tea.KeyPressMsg{Code: 'd', Mod: tea.ModAlt},
		"alt+u":     tea.KeyPressMsg{Code: 'u', Mod: tea.ModAlt},
		"ctrl+t":    tea.KeyPressMsg{Code: 't', Mod: tea.ModCtrl},
	}

	// newModel builds an empty model in the configuration the cases below
	// expect. Cases that change the configuration mid-edit set width.
	newModel := func(width int) func() Model {
		return func() Model {
			m := newTextArea()
			m.ShowLineNumbers = true
			m.SetWidth(width)
			m.SetHeight(5)
			return m
		}
	}

	for _, tc := range []struct {
		name string
		// width the model holds by the time the edit is done. Defaults to 30.
		width int
		edit  func(Model) Model
	}{
		{"split line", 0, func(m Model) Model {
			m.row, m.col = 0, 10
			m, _ = m.Update(keys["enter"])
			return m
		}},
		{"split first line at start", 0, func(m Model) Model {
			m.row, m.col = 0, 0
			m, _ = m.Update(keys["enter"])
			return m
		}},
		{"merge line above", 0, func(m Model) Model {
			m.row, m.col = 3, 0
			m, _ = m.Update(keys["backspace"])
			return m
		}},
		{"merge line below", 0, func(m Model) Model {
			m.row, m.col = 1, len(m.value[1])
			m, _ = m.Update(keys["delete"])
			return m
		}},
		{"type on a wrapped line", 0, func(m Model) Model {
			m.row, m.col = 3, 4
			return sendString(m, "hello there")
		}},
		{"type at the end of the buffer", 0, func(m Model) Model {
			m.MoveToEnd()
			return sendString(m, " and then some more words to wrap")
		}},
		{"paste multiple lines", 0, func(m Model) Model {
			m.row, m.col = 1, 2
			m, _ = m.Update(tea.PasteMsg{Content: "one\ntwo\nthree and a longer line that wraps"})
			return m
		}},
		{"backspace repeatedly", 0, func(m Model) Model {
			m.row, m.col = 0, 20
			for range 25 {
				m, _ = m.Update(keys["backspace"])
			}
			return m
		}},
		{"delete to end of line", 0, func(m Model) Model {
			m.row, m.col = 3, 12
			m, _ = m.Update(keys["ctrl+k"])
			return m
		}},
		{"delete to start of line", 0, func(m Model) Model {
			m.row, m.col = 3, 12
			m, _ = m.Update(keys["ctrl+u"])
			return m
		}},
		{"delete word forward", 0, func(m Model) Model {
			m.row, m.col = 0, 4
			m, _ = m.Update(keys["alt+d"])
			return m
		}},
		{"uppercase word", 0, func(m Model) Model {
			m.row, m.col = 0, 4
			m, _ = m.Update(keys["alt+u"])
			return m
		}},
		{"transpose characters", 0, func(m Model) Model {
			m.row, m.col = 0, 5
			m, _ = m.Update(keys["ctrl+t"])
			return m
		}},
		{"delete every line", 0, func(m Model) Model {
			m.MoveToEnd()
			for range len(m.Value()) {
				m, _ = m.Update(keys["backspace"])
			}
			return m
		}},
		{"resize after editing", 15, func(m Model) Model {
			m.row, m.col = 0, 5
			m = sendString(m, "some inserted text")
			m.SetWidth(15)
			m, _ = m.Update(nil)
			return m
		}},
		{"scroll to the bottom", 0, func(m Model) Model {
			m.MoveToEnd()
			m, _ = m.Update(nil)
			return m
		}},
		{"reset then set value", 0, func(m Model) Model {
			m.Reset()
			m.SetValue("a totally different value\nwith two lines")
			m.MoveToEnd()
			return m
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			width := tc.width
			if width == 0 {
				width = 30
			}

			m := newModel(30)()
			m.SetValue(content)
			m.MoveToBegin()

			// Prime the cache so that the edit exercises the incremental
			// path rather than a cold wrap.
			_ = m.View()

			m = tc.edit(m)
			assertMatchesFresh(t, m, newModel(width))
		})
	}
}

func TestWrapCacheAfterScroll(t *testing.T) {
	newModel := func() Model {
		m := newTextArea()
		m.ShowLineNumbers = true
		m.SetWidth(30)
		m.SetHeight(5)
		return m
	}

	m := newModel()

	lines := make([]string, 200)
	for i := range lines {
		lines[i] = fmt.Sprintf("line %d: some text that is long enough to soft wrap", i)
	}
	m.SetValue(strings.Join(lines, "\n"))
	m.MoveToBegin()
	_ = m.View()

	for _, offset := range []int{0, 7, 120, 400, 42} {
		m.viewport.SetYOffset(offset)
		m.row = m.rowForVisualLine(offset)
		m.col = 0
		assertMatchesFresh(t, m, newModel)
	}
}

// TestWrapCacheRenderBufferReuse makes sure that the render buffer handed to
// the viewport is not clobbered by a later render.
func TestWrapCacheRenderBufferReuse(t *testing.T) {
	m := newTextArea()
	m.SetWidth(30)
	m.SetHeight(4)
	m.SetValue("one\ntwo\nthree")

	m.MoveToBegin()
	m, _ = m.Update(nil)

	// The viewport holds on to the slice that viewLines returns, which is
	// reused across renders. Rendering twice in a row must not change what the
	// viewport shows.
	first := m.View()
	if got := m.View(); got != first {
		t.Errorf("view changed after re-render:\ngot:\n%s\nwant:\n%s", got, first)
	}

	// Same goes for the content the viewport was given during Update.
	m = sendString(m, "!")
	updated := m.viewport.View()
	if got := m.View(); got != updated {
		t.Errorf("view does not match viewport content set during Update:\ngot:\n%s\nwant:\n%s", got, updated)
	}
}
