package filepicker

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type fileNameStyles struct {
	fileNotSelected lipgloss.Style
	fileRegular     lipgloss.Style
	fileDirectory   lipgloss.Style
	fileSymLink     lipgloss.Style
	fileBlockDevice lipgloss.Style
	fileSelected    lipgloss.Style
	filterMatch     lipgloss.Style
}

func newFileNameStyles() (s fileNameStyles) {
	s.fileNotSelected = lipgloss.NewStyle()
	s.fileRegular = lipgloss.NewStyle()
	s.fileDirectory = lipgloss.NewStyle().Bold(true).
		Foreground(lipgloss.Color("32"))
	s.fileSymLink = lipgloss.NewStyle().Bold(true).
		Foreground(lipgloss.Color("36"))
	s.fileBlockDevice = lipgloss.NewStyle().Bold(true).
		Foreground(lipgloss.Color("33")).
		Background(lipgloss.Color("40"))
	s.fileSelected = lipgloss.NewStyle().
		Background(lipgloss.Color("#FFFF00"))
	s.filterMatch = lipgloss.NewStyle().Underline(true)
	return s
}

func newItemDelegate(keys *delegateKeyMap) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		var title string

		if i, ok := m.SelectedItem().(listItem); ok {
			title = i.value
		} else {
			return nil
		}
		statusMessageStyle := lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).Render

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keys.choose):
				return m.NewStatusMessage(statusMessageStyle("You chose " + title))

			case key.Matches(msg, keys.remove):
				index := m.Index()
				m.RemoveItem(index)
				if len(m.Items()) == 0 {
					keys.remove.SetEnabled(false)
				}
				return m.NewStatusMessage(statusMessageStyle("Deleted " + title))
			}
		}

		return nil
	}

	help := []key.Binding{keys.choose, keys.remove}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}

type delegateKeyMap struct {
	choose key.Binding
	remove key.Binding
}

// Additional short help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		d.choose,
		d.remove,
	}
}

// Additional full help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.choose,
			d.remove,
		},
	}
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose"),
		),
		remove: key.NewBinding(
			key.WithKeys("x", "backspace"),
			key.WithHelp("x", "delete"),
		),
	}
}
