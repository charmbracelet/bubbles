package tree

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"

	"charm.land/bubbles/v2/key"
)

func TestTree(t *testing.T) {
	m := New(Root("~/charm").
		Child(
			"ayman",
			Root("bash").
				Child(
					Root("tools").
						Child("zsh",
							"doom-emacs",
						),
				),
			Root("carlos").
				Child(
					Root("emotes").
						Child(
							"chefkiss.png",
							"kekw.png",
						),
				),
			"maas",
		), 70, 13)

	t.Run("default tree", func(t *testing.T) {
		s := m.View()
		s = ansi.Strip(s)
		golden.RequireEqual(t, []byte(s))
	})
}

func TestTreeMultiLineNodes(t *testing.T) {
	m := New(Root("~/charm").
		Child(
			"ayman",
			"two\nlines",
			Root("bash").
				Child(
					"zsh",
					"doom\nemacs",
				),
			"maas",
		), 70, 20)

	// walk down node by node and assert the cursor row matches the
	// selected node's line offset, even when nodes span multiple lines
	for i := 0; i < m.Root().Size()-1; i++ {
		m.Down()
		node := m.NodeAtCurrentOffset()
		if node == nil {
			t.Fatalf("no node at yOffset %d", m.YOffset())
		}

		view := ansi.Strip(m.viewport.View())
		lines := strings.Split(view, "\n")
		cursorLine := -1
		for j, line := range lines {
			if strings.Contains(line, "→") {
				cursorLine = j
				break
			}
		}

		want := node.LineOffset() - m.viewport.YOffset()
		if cursorLine != want {
			t.Errorf("step %d: cursor on line %d, want %d (node %q, lineOffset %d)",
				i, cursorLine, want, node.GivenValue(), node.LineOffset())
		}
	}
}

func TestTreeAdditionalHelp(t *testing.T) {
	m := New(Root("~/charm").
		Child(
			"ayman",
			Root("bash").
				Child(
					Root("tools").
						Child("zsh",
							"doom-emacs",
						),
				),
			Root("carlos").
				Child(
					Root("emotes").
						Child(
							"chefkiss.png",
							"kekw.png",
						),
				),
			"maas",
		), 70, 13)
	m.SetAdditionalShortHelpKeys(func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("v"),
				key.WithHelp("v", "select"),
			),
		}
	})

	t.Run("additional help", func(t *testing.T) {
		s := m.View()
		s = ansi.Strip(s)
		golden.RequireEqual(t, []byte(s))
	})
}
