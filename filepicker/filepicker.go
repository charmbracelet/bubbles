package filepicker

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	listWidth, listHeight, _ = terminal.GetSize(0)
	docStyle                 = lipgloss.NewStyle().Margin(1, 2)
	itemStyle                = lipgloss.NewStyle().PaddingLeft(4)

	helpStyle = lipgloss.NewStyle().
			Height(5).
			Width(listWidth)

	singlePaneStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, false).
			MarginRight(2).
			Height(listHeight - 2).
			Width(listWidth - 2)

	dualPaneStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			MarginRight(2).
			Height(listHeight - 2).
			Width(listWidth/2 - 2)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
)

type (
	// Model contains the state of the filepicker view.
	Model struct {
		lFileInfo          []listItem // files in the parent directory
		rFileInfo          []listItem // files in the directory
		directory          string
		dirErr             error
		lList              list.Model
		rList              list.Model
		SelectionCompleted bool
		SelectedFileInfo   listItem
		fileExt            string
		dualPane           bool
	}

	listItem struct {
		value string
		Entry os.DirEntry
		Dir   string
	}

	itemDelegate struct {
		styles      FileNameStyles
		defDelegate list.DefaultDelegate
		dualPane    bool
	}
)

const (
	permDenied = "<permission denied>"
	dirEmpty   = "<directory empty>"
	bullet     = "•"
	ellipsis   = "…"
)

func (l listItem) FilterValue() string {
	return l.value
}

func (d itemDelegate) Height() int { return 1 }

func (d itemDelegate) Spacing() int { return 0 }

func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, lstItem list.Item) {
	d.styles = newFileNameStyles()
	s := &d.styles
	var matchedRunes []int
	i, ok := lstItem.(listItem)
	if !ok {
		return
	}
	fileName := i.value

	if m.Width() <= 0 {
		return
	}

	// Prevent text from exceeding list width
	var textwidth uint
	if d.dualPane {
		textwidth = uint(m.Width() - dualPaneStyle.GetPaddingLeft() - dualPaneStyle.GetPaddingRight())
	} else {
		textwidth = uint(m.Width() - singlePaneStyle.GetPaddingLeft() - singlePaneStyle.GetPaddingRight())
	}
	fileName = truncate.StringWithTail(fileName, textwidth, ellipsis)

	// Conditions
	var (
		isSelected  = index == m.Index()
		emptyFilter = m.FilterState() == list.Filtering &&
			m.FilterValue() == ""
		isFiltered = m.FilterState() == list.Filtering ||
			m.FilterState() == list.FilterApplied
	)

	if isFiltered && index < len(m.VisibleItems()) {
		// Get indices of matched characters
		matchedRunes = m.MatchesForItem(index)
	}
	if emptyFilter {
	} else if isSelected && m.FilterState() != list.Filtering {
		if isFiltered {
			// Highlight matches
			unmatched := s.FileSelected.Inline(true)
			matched := unmatched.Copy().Inherit(s.FilterMatch)
			fileName = lipgloss.StyleRunes(fileName, matchedRunes, matched, unmatched)
		}
		fileName = s.FileSelected.Render(fileName)
	} else {
		if isFiltered {
			// Highlight matches
			unmatched := s.FileRegular.Inline(true)
			matched := unmatched.Copy().Inherit(s.FilterMatch)
			fileName = lipgloss.StyleRunes(fileName, matchedRunes, matched, unmatched)
		}
		fileName = s.FileRegular.Render(fileName)
	}

	var fn func(str string) string
	if i.Entry == nil {
		fn = s.FileRegular.Render
		fmt.Fprint(w, fn(fileName))
		return
	}
	switch mode := i.Entry.Type(); {
	case isDir(i.Entry, i.Dir):
		fn = s.FileDirectory.Render
	case mode&os.ModeSymlink != 0:
		fn = s.FileSymLink.Render
	case mode&os.ModeDevice != 0:
		fn = s.FileBlockDevice.Render
	default:
		fn = s.FileRegular.Render
	}
	if index == m.Index() {
		fileName = fn(fileName)
		fn = s.FileSelected.Render
	}

	fmt.Fprint(w, fn(fileName))
}

func (m Model) WithFileExt(fileExt string) Model {
	m.lFileInfo = []listItem{}
	m.rFileInfo = []listItem{}
	m.fileExt = strings.TrimSpace(fileExt)
	return m.getFileinfo()
}

func (m Model) WithInitDir(initDir string) Model {
	var err error
	m.lFileInfo = []listItem{}
	m.rFileInfo = []listItem{}
	m.directory = strings.TrimSpace(initDir)
	m.directory, err = filepath.Abs(m.directory)
	if err != nil {
		m.dirErr = err
	}
	m.directory = strings.TrimSpace(m.directory)

	return m.getFileinfo()
}

func (m Model) WithDualPane(dualPane bool) Model {
	m.dualPane = dualPane
	return m
}

// New creates a new filepicker view with some useful defaults.
func New() Model {
	var (
		m   Model
		err error
	)
	m.dualPane = true
	if m.directory == "" {
		m.directory, err = os.Getwd()
		if err != nil {
			m.dirErr = err
		}
	}
	m.directory, err = filepath.Abs(m.directory)
	if err != nil {
		m.dirErr = err
	}
	m.directory = strings.TrimSpace(m.directory)
	m.lList = list.New([]list.Item{}, itemDelegate{dualPane: m.dualPane}, 0, 0)
	m.lList.SetShowHelp(false)
	m.rList = list.New([]list.Item{}, itemDelegate{dualPane: m.dualPane}, 0, 0)
	m.rList.SetShowHelp(false)
	m.rList.KeyMap = list.DefaultKeyMap()
	m.rList.KeyMap.NextPage = key.NewBinding(key.WithKeys("pgdown"),
		key.WithHelp("pgdn", "next page"))
	m.rList.KeyMap.PrevPage = key.NewBinding(key.WithKeys("pgup"),
		key.WithHelp("pgup", "prev page"))
	m.rList.SetShowTitle(false)
	m.lList.SetShowTitle(false)
	return m.getFileinfo()
}

func (m Model) Init() tea.Cmd {
	if m.dirErr != nil {
		return tea.Quit
	}
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "right":
			// Permission to be in current dir?
			if m.dirErr == nil {
				// Is current dir empty?
				if len(m.rFileInfo) != 0 {
					// assigning var in case filtering is active
					vi := m.rList.VisibleItems()[m.rList.Index()].(listItem)
					// Is selected item a dir or symlink to a dir?
					if isDir(vi.Entry, m.directory) {
						m.rList, _ = m.rList.Update(msg)
						m = m.runKeyRight(msg)
					}
				}
			}
		case "left":
			// Is filter being used or already applied?
			if m.rList.SettingFilter() {
				break
			}
			if m.rList.IsFiltered() {
				m.rList.ResetFilter()
			}

			rootDir, _ := filepath.Abs("/")
			if m.directory == rootDir {
				break
			}
			m.dirErr = nil
			m.rFileInfo = m.lFileInfo
			// place active selection at former parent directory
			m.rList.Select(m.lList.Index())
			m.directory = filepath.Dir(m.directory)
			// read file entries in current parent dir
			files, err := os.ReadDir(filepath.Dir(m.directory))
			if err != nil {
				return m, nil
			}
			m.lFileInfo = nil
			for _, f := range files {
				li := listItem{Entry: f, Dir: filepath.Dir(m.directory)}
				if isDir(f, filepath.Dir(m.directory)) {
					m.lFileInfo = append(m.lFileInfo, li)
				} else if strings.HasSuffix(f.Name(), m.fileExt) {
					m.lFileInfo = append(m.lFileInfo, li)
				}
			}
			m = m.fillPanes(msg)
			for i, f := range m.lList.Items() {
				if f.FilterValue() == filepath.Base(m.directory) {
					m.lList.Select(i)
				}
			}

		case "enter":
			filterVal := m.rList.Items()[0].FilterValue()
			// Cannot select file when in a dir without permissions
			// or when dir is empty
			if filterVal == permDenied || filterVal == dirEmpty {
				break
			}

			// if setting filter, enable filter
			if m.rList.SettingFilter() {
				m.rList.SetFilteringEnabled(true)
				m.rList, cmd = m.rList.Update(msg)
				return m, cmd
			}
			// setting selected file info
			m.SelectionCompleted = true
			if li, ok := m.rList.VisibleItems()[m.rList.Index()].(listItem); ok {
				for _, f := range m.rFileInfo {
					if f.Entry.Name() == li.Entry.Name() {
						m.SelectedFileInfo = f
					}
				}
			}
			return m, tea.Quit

		case "ctrl+c":
			return m, tea.Quit

		default:
			m.rList, cmd = m.rList.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		singlePaneStyle = singlePaneStyle.Width(msg.Width - 2).Height(msg.Height - 16)
		dualPaneStyle = dualPaneStyle.Width(msg.Width/2 - 2).Height(msg.Height - 16)
		helpStyle = helpStyle.Width(msg.Width - 1)
		m.lList.SetSize(msg.Width/2-2, msg.Height-16)
		if m.dualPane {
			m.rList.SetSize(msg.Width/2-2, msg.Height-16)
		} else {
			m.rList.SetSize(msg.Width-2, msg.Height-16)
		}
		return m, nil
	}

	m.rList, _ = m.rList.Update(msg)
	return m, cmd
}

// View renders the help view's current state.
func (m Model) View() string {
	if m.dirErr != nil {
		if !os.IsPermission(m.dirErr) {
			dirErrPane := lipgloss.JoinHorizontal(lipgloss.Top,
				helpStyle.Render(fmt.Sprintf("%v", m.dirErr)))
			return dirErrPane
		}
	}
	if !m.dualPane {
		filePane := singlePaneStyle.Render(
			m.rList.View(),
		)
		m.rList.Help.Width = listWidth
		windowPane := lipgloss.JoinVertical(lipgloss.Top, filePane,
			helpStyle.Render(m.rList.Help.View(m.rList)),
		)
		return windowPane
	}
	filePane := lipgloss.JoinHorizontal(lipgloss.Top,
		dualPaneStyle.Render(
			m.lList.View(),
		),
		dualPaneStyle.Render(
			m.rList.View(),
		),
	)
	m.rList.Help.Width = listWidth
	windowPane := lipgloss.JoinVertical(lipgloss.Top,
		filePane,
		helpStyle.Render(m.rList.Help.View(m.rList)),
	)

	return windowPane
}

func (m Model) getFileinfo() Model {
	var msg tea.Msg
	lFiles, lErr := os.ReadDir(filepath.Dir(m.directory))
	if lErr != nil {
		m.dirErr = lErr
		return m
	}
	for _, f := range lFiles {
		li := listItem{Entry: f, Dir: m.directory}
		// include directories in filter regardless of name
		if isDir(f, filepath.Dir(m.directory)) {
			m.lFileInfo = append(m.lFileInfo, li)
			// include only files with matching suffix (extension)
		} else if strings.HasSuffix(f.Name(), m.fileExt) {
			m.lFileInfo = append(m.lFileInfo, li)
		}
	}
	rFiles, rErr := os.ReadDir(m.directory)
	if rErr != nil {
		m.dirErr = rErr
		return m
	}
	for _, f := range rFiles {
		li := listItem{Entry: f, Dir: m.directory}
		if isDir(f, m.directory) {
			m.rFileInfo = append(m.rFileInfo, li)
		} else if strings.HasSuffix(f.Name(), m.fileExt) {
			m.rFileInfo = append(m.rFileInfo, li)
		}
	}

	m = m.fillPanes(msg)
	m.rList.ResetSelected()
	for i, f := range m.lList.Items() {
		if f.FilterValue() == filepath.Base(m.directory) {
			m.lList.Select(i)
		}
	}
	return m
}

func (m Model) fillPanes(msg tea.Msg) Model {
	leftPaneItems := []list.Item{}
	for _, choice := range m.lFileInfo {
		li := listItem{value: choice.Entry.Name(), Entry: choice.Entry, Dir: filepath.Dir(m.directory)}
		leftPaneItems = append(leftPaneItems, li)
	}
	rightPaneItems := []list.Item{}
	for _, choice := range m.rFileInfo {
		li := listItem{value: choice.Entry.Name(), Entry: choice.Entry, Dir: m.directory}
		rightPaneItems = append(rightPaneItems, li)
	}
	if os.IsPermission(m.dirErr) {
		rightPaneItems = []list.Item{listItem{value: permDenied}}
	}
	if len(rightPaneItems) == 0 {
		rightPaneItems = []list.Item{listItem{value: dirEmpty}}
	}
	rootDir, _ := filepath.Abs("/")
	if m.directory == rootDir {
		leftPaneItems = []list.Item{listItem{value: rootDir}}
	}
	m.rList.SetItems(rightPaneItems)
	m.lList.SetItems(leftPaneItems)
	m.rList.Title = m.directory
	return m
}

func isSymLinkToDir(f os.DirEntry, parent string) bool {
	base := f.Name()
	linkDest, err := filepath.EvalSymlinks(filepath.Join(parent, base))
	if err != nil {
		return false
	}
	linkDir, dirErr := os.Lstat(linkDest)
	if dirErr != nil {
		return false
	}
	return linkDir.IsDir()
}

// isDir returns true if fileEntry is a dir or a symlink to a dir
func isDir(fileEntry fs.DirEntry, parent string) bool {
	if fileEntry.Type()&os.ModeSymlink != 0 {
		return isSymLinkToDir(fileEntry, parent)
	}
	return fileEntry.IsDir()
}

func (m Model) runKeyRight(msg tea.Msg) Model {
	m.lFileInfo = m.rFileInfo
	vi := m.rList.VisibleItems()[m.rList.Index()].(listItem).Entry.Name()
	m.directory = filepath.Join(m.directory, vi)
	files, err := os.ReadDir(m.directory)
	m.dirErr = err
	m.rFileInfo = nil
	for _, f := range files {
		li := listItem{Entry: f, Dir: m.directory}
		if isDir(f, m.directory) {
			m.rFileInfo = append(m.rFileInfo, li)
		} else if strings.HasSuffix(f.Name(), m.fileExt) {
			m.rFileInfo = append(m.rFileInfo, li)
		}
	}

	// If filter is being used, reset filter
	m.rList.ResetFilter()
	m = m.fillPanes(msg)
	m.rList.ResetSelected()
	for i, f := range m.lList.Items() {
		if f.FilterValue() == filepath.Base(m.directory) {
			m.lList.Select(i)
		}
	}
	return m
}
