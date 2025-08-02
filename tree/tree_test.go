package tree

import (
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
