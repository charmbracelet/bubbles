// Package filepicker provides a file picker component for Bubble Tea
// applications.
package filepicker

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dustin/go-humanize"
)

var lastID int64

func nextID() int {
	return int(atomic.AddInt64(&lastID, 1))
}

// New returns a new file picker model with default styling and key bindings.
func New() Model {
	ti := textinput.New()
	ti.Prompt = "/ "
	ti.CharLimit = 64

	return Model{
		id:               nextID(),
		CurrentDirectory: ".",
		Cursor:           ">",
		AllowedTypes:     []string{},
		selected:         0,
		ShowPermissions:  true,
		ShowSize:         true,
		ShowHidden:       false,
		DirAllowed:       false,
		FileAllowed:      true,
		AutoHeight:       true,
		height:           0,
		maxIdx:           0,
		minIdx:           0,
		selectedStack:    newStack(),
		minStack:         newStack(),
		maxStack:         newStack(),
		KeyMap:           DefaultKeyMap(),
		Styles:           DefaultStyles(),
		searchInput:      ti,
		Filter:           DefaultFilter,
	}
}

type errorMsg struct {
	err error
}

type readDirMsg struct {
	id      int
	entries []os.DirEntry
}

const (
	marginBottom  = 5
	fileSizeWidth = 7
	paddingLeft   = 2
)

// KeyMap defines key bindings for each user action.
type KeyMap struct {
	GoToTop  key.Binding
	GoToLast key.Binding
	Down     key.Binding
	Up       key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Back     key.Binding
	Open     key.Binding
	Select   key.Binding
	Search   key.Binding
}

// DefaultKeyMap defines the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		GoToTop:  key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "first")),
		GoToLast: key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "last")),
		Down:     key.NewBinding(key.WithKeys("j", "down", "ctrl+n"), key.WithHelp("j", "down")),
		Up:       key.NewBinding(key.WithKeys("k", "up", "ctrl+p"), key.WithHelp("k", "up")),
		PageUp:   key.NewBinding(key.WithKeys("K", "pgup"), key.WithHelp("pgup", "page up")),
		PageDown: key.NewBinding(key.WithKeys("J", "pgdown"), key.WithHelp("pgdown", "page down")),
		Back:     key.NewBinding(key.WithKeys("h", "backspace", "left", "esc"), key.WithHelp("h", "back")),
		Open:     key.NewBinding(key.WithKeys("l", "right", "enter"), key.WithHelp("l", "open")),
		Select:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Search:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
	}
}

// Styles defines the possible customizations for styles in the file picker.
type Styles struct {
	DisabledCursor   lipgloss.Style
	Cursor           lipgloss.Style
	Symlink          lipgloss.Style
	Directory        lipgloss.Style
	File             lipgloss.Style
	DisabledFile     lipgloss.Style
	Permission       lipgloss.Style
	Selected         lipgloss.Style
	DisabledSelected lipgloss.Style
	FileSize         lipgloss.Style
	EmptyDirectory   lipgloss.Style
}

// DefaultStyles defines the default styling for the file picker.
func DefaultStyles() Styles {
	return Styles{
		DisabledCursor:   lipgloss.NewStyle().Foreground(lipgloss.Color("247")),
		Cursor:           lipgloss.NewStyle().Foreground(lipgloss.Color("212")),
		Symlink:          lipgloss.NewStyle().Foreground(lipgloss.Color("36")),
		Directory:        lipgloss.NewStyle().Foreground(lipgloss.Color("99")),
		File:             lipgloss.NewStyle(),
		DisabledFile:     lipgloss.NewStyle().Foreground(lipgloss.Color("243")),
		DisabledSelected: lipgloss.NewStyle().Foreground(lipgloss.Color("247")),
		Permission:       lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
		Selected:         lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true),
		FileSize:         lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Width(fileSizeWidth).Align(lipgloss.Right),
		EmptyDirectory:   lipgloss.NewStyle().Foreground(lipgloss.Color("240")).PaddingLeft(paddingLeft).SetString("Bummer. No Files Found."),
	}
}

// FilterFunc is a function that returns matching file indices for a query.
type FilterFunc func(query string, entries []os.DirEntry) []int

// DefaultFilter performs a case-insensitive substring match on file names.
func DefaultFilter(query string, entries []os.DirEntry) []int {
	lower := strings.ToLower(query)
	var matches []int
	for i, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Name()), lower) {
			matches = append(matches, i)
		}
	}
	return matches
}

// Model represents a file picker.
type Model struct {
	id int

	// Path is the path which the user has selected with the file picker.
	Path string

	// CurrentDirectory is the directory that the user is currently in.
	CurrentDirectory string

	// AllowedTypes specifies which file types the user may select.
	// If empty the user may select any file.
	AllowedTypes []string

	KeyMap          KeyMap
	files           []os.DirEntry
	ShowPermissions bool
	ShowSize        bool
	ShowHidden      bool
	DirAllowed      bool
	FileAllowed     bool

	FileSelected  string
	selected      int
	selectedStack stack

	minIdx   int
	maxIdx   int
	maxStack stack
	minStack stack

	height     int
	AutoHeight bool

	Cursor string
	Styles Styles

	// Search/filter state
	searchInput   textinput.Model
	searchActive  bool // user is typing in the search bar
	filterApplied bool // a filter is active (search confirmed or still typing)
	filteredFiles []os.DirEntry
	filteredMap   []int // maps filtered index -> original files index

	// Filter is the function used to filter files. Defaults to case-insensitive
	// substring matching. Set this to customize filtering behavior.
	Filter FilterFunc
}

type stack struct {
	Push   func(int)
	Pop    func() int
	Length func() int
}

func newStack() stack {
	slice := make([]int, 0)
	return stack{
		Push: func(i int) {
			slice = append(slice, i)
		},
		Pop: func() int {
			res := slice[len(slice)-1]
			slice = slice[:len(slice)-1]
			return res
		},
		Length: func() int {
			return len(slice)
		},
	}
}

func (m *Model) pushView(selected, minimum, maximum int) {
	m.selectedStack.Push(selected)
	m.minStack.Push(minimum)
	m.maxStack.Push(maximum)
}

func (m *Model) popView() (int, int, int) {
	return m.selectedStack.Pop(), m.minStack.Pop(), m.maxStack.Pop()
}

func (m Model) readDir(path string, showHidden bool) tea.Cmd {
	return func() tea.Msg {
		dirEntries, err := os.ReadDir(path)
		if err != nil {
			return errorMsg{err}
		}

		sort.Slice(dirEntries, func(i, j int) bool {
			if dirEntries[i].IsDir() == dirEntries[j].IsDir() {
				return dirEntries[i].Name() < dirEntries[j].Name()
			}
			return dirEntries[i].IsDir()
		})

		if showHidden {
			return readDirMsg{id: m.id, entries: dirEntries}
		}

		var sanitizedDirEntries []os.DirEntry
		for _, dirEntry := range dirEntries {
			isHidden, _ := IsHidden(dirEntry.Name())
			if isHidden {
				continue
			}
			sanitizedDirEntries = append(sanitizedDirEntries, dirEntry)
		}
		return readDirMsg{id: m.id, entries: sanitizedDirEntries}
	}
}

// SetHeight sets the height of the file picker.
func (m *Model) SetHeight(h int) {
	m.height = h
	if m.maxIdx > m.height-1 {
		m.maxIdx = m.minIdx + m.height - 1
	}
}

// Height returns the height of the file picker.
func (m Model) Height() int {
	return m.height
}

// Init initializes the file picker model.
func (m Model) Init() tea.Cmd {
	return m.readDir(m.CurrentDirectory, m.ShowHidden)
}

// visibleFiles returns the currently visible files (filtered or all).
func (m Model) visibleFiles() []os.DirEntry {
	if m.filterApplied && len(m.filteredFiles) > 0 {
		return m.filteredFiles
	}
	return m.files
}

// updateFilter recomputes the filtered files based on the current search query.
func (m *Model) updateFilter() {
	query := m.searchInput.Value()
	if query == "" {
		m.filterApplied = false
		m.filteredFiles = nil
		m.filteredMap = nil
		return
	}

	indices := m.Filter(query, m.files)
	m.filteredFiles = make([]os.DirEntry, 0, len(indices))
	m.filteredMap = make([]int, 0, len(indices))
	for _, idx := range indices {
		m.filteredFiles = append(m.filteredFiles, m.files[idx])
		m.filteredMap = append(m.filteredMap, idx)
	}
	m.filterApplied = true

	// Reset selection if it's out of bounds
	if m.selected >= len(m.filteredFiles) {
		m.selected = max(0, len(m.filteredFiles)-1)
	}
	if m.minIdx > m.selected {
		m.minIdx = m.selected
	}
	m.maxIdx = m.minIdx + m.Height() - 1
}

// resetSearch clears the search state and restores the full file list.
func (m *Model) resetSearch() {
	m.searchActive = false
	m.filterApplied = false
	m.filteredFiles = nil
	m.filteredMap = nil
	m.searchInput.SetValue("")
	m.searchInput.Blur()
	m.selected = 0
	m.minIdx = 0
	m.maxIdx = m.Height() - 1
}

// Update handles user interactions within the file picker model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case readDirMsg:
		if msg.id != m.id {
			break
		}
		m.files = msg.entries
		m.maxIdx = max(m.maxIdx, m.Height()-1)
		// Reapply filter if active
		if m.filterApplied {
			m.updateFilter()
		}
	case tea.WindowSizeMsg:
		if m.AutoHeight {
			m.SetHeight(msg.Height - marginBottom)
		}
		m.maxIdx = m.Height() - 1
	case tea.KeyPressMsg:
		// When search is active, handle search-specific keys first
		if m.searchActive {
			var cmd tea.Cmd
			switch {
			case msg.String() == "esc":
				if m.filterApplied && m.searchInput.Value() != "" {
					// Filter was applied: clear everything
					m.resetSearch()
				} else {
					// No filter: just exit search mode
					m.searchActive = false
					m.searchInput.Blur()
				}
				return m, nil
			case msg.String() == "enter":
				// Confirm search: exit search mode but keep filter
				m.searchActive = false
				m.searchInput.Blur()
				return m, nil
			default:
				// Forward to text input
				m.searchInput, cmd = m.searchInput.Update(msg)
				m.updateFilter()
				return m, cmd
			}
		}

		switch {
		case key.Matches(msg, m.KeyMap.Search):
			// Enter search mode
			m.searchActive = true
			m.searchInput.Focus()
			return m, textinput.Blink

		case key.Matches(msg, m.KeyMap.GoToTop):
			m.selected = 0
			m.minIdx = 0
			m.maxIdx = m.Height() - 1
		case key.Matches(msg, m.KeyMap.GoToLast):
			vis := m.visibleFiles()
			m.selected = len(vis) - 1
			m.minIdx = len(vis) - m.Height()
			m.maxIdx = len(vis) - 1
		case key.Matches(msg, m.KeyMap.Down):
			vis := m.visibleFiles()
			m.selected++
			if m.selected >= len(vis) {
				m.selected = len(vis) - 1
			}
			if m.selected > m.maxIdx {
				m.minIdx++
				m.maxIdx++
			}
		case key.Matches(msg, m.KeyMap.Up):
			m.selected--
			if m.selected < 0 {
				m.selected = 0
			}
			if m.selected < m.minIdx {
				m.minIdx--
				m.maxIdx--
			}
		case key.Matches(msg, m.KeyMap.PageDown):
			vis := m.visibleFiles()
			m.selected += m.Height()
			if m.selected >= len(vis) {
				m.selected = len(vis) - 1
			}
			m.minIdx += m.Height()
			m.maxIdx += m.Height()

			if m.maxIdx >= len(vis) {
				m.maxIdx = len(vis) - 1
				m.minIdx = m.maxIdx - m.Height()
			}
		case key.Matches(msg, m.KeyMap.PageUp):
			m.selected -= m.Height()
			if m.selected < 0 {
				m.selected = 0
			}
			m.minIdx -= m.Height()
			m.maxIdx -= m.Height()

			if m.minIdx < 0 {
				m.minIdx = 0
				m.maxIdx = m.minIdx + m.Height()
			}
		case key.Matches(msg, m.KeyMap.Back):
			// Esc has layered behavior:
			// 1. If filter is applied, clear filter first
			// 2. Otherwise, go to parent directory
			if m.filterApplied {
				m.resetSearch()
				return m, nil
			}

			m.CurrentDirectory = filepath.Dir(m.CurrentDirectory)
			if m.selectedStack.Length() > 0 {
				m.selected, m.minIdx, m.maxIdx = m.popView()
			} else {
				m.selected = 0
				m.minIdx = 0
				m.maxIdx = m.Height() - 1
			}
			return m, m.readDir(m.CurrentDirectory, m.ShowHidden)
		case key.Matches(msg, m.KeyMap.Open):
			vis := m.visibleFiles()
			if len(vis) == 0 {
				break
			}

			f := vis[m.selected]
			info, err := f.Info()
			if err != nil {
				break
			}
			isSymlink := info.Mode()&os.ModeSymlink != 0
			isDir := f.IsDir()

			if isSymlink {
				symlinkPath, _ := filepath.EvalSymlinks(filepath.Join(m.CurrentDirectory, f.Name()))
				info, err := os.Stat(symlinkPath)
				if err != nil {
					break
				}
				if info.IsDir() {
					isDir = true
				}
			}

			if (!isDir && m.FileAllowed) || (isDir && m.DirAllowed) {
				if key.Matches(msg, m.KeyMap.Select) {
					// Select the current path as the selection
					m.Path = filepath.Join(m.CurrentDirectory, f.Name())
				}
			}

			if !isDir {
				break
			}

			// Entering a directory: clear search
			if m.filterApplied {
				m.searchInput.SetValue("")
				m.filterApplied = false
				m.filteredFiles = nil
				m.filteredMap = nil
			}

			m.CurrentDirectory = filepath.Join(m.CurrentDirectory, f.Name())
			m.pushView(m.selected, m.minIdx, m.maxIdx)
			m.selected = 0
			m.minIdx = 0
			m.maxIdx = m.Height() - 1
			return m, m.readDir(m.CurrentDirectory, m.ShowHidden)
		}
	}
	return m, nil
}

// View returns the view of the file picker.
func (m Model) View() string {
	var s strings.Builder

	// Show search bar when active
	if m.searchActive {
		s.WriteString(m.searchInput.View())
		s.WriteRune('\n')
	}

	vis := m.visibleFiles()

	// Show filter status when filter is applied but not actively typing
	if m.filterApplied && !m.searchActive && m.searchInput.Value() != "" {
		status := fmt.Sprintf("  filter: %s (%d/%d)", m.searchInput.Value(), len(vis), len(m.files))
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(status))
		s.WriteRune('\n')
	}

	if len(vis) == 0 {
		if m.filterApplied {
			noMatch := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).PaddingLeft(paddingLeft).SetString("No matches.")
			s.WriteString(noMatch.Height(m.Height()).MaxHeight(m.Height()).String())
		} else {
			s.WriteString(m.Styles.EmptyDirectory.Height(m.Height()).MaxHeight(m.Height()).String())
		}
		return s.String()
	}

	for i, f := range vis {
		if i < m.minIdx || i > m.maxIdx {
			continue
		}

		var symlinkPath string
		info, err := f.Info()
		if err != nil {
			continue
		}
		isSymlink := info.Mode()&os.ModeSymlink != 0
		size := strings.Replace(humanize.Bytes(uint64(info.Size())), " ", "", 1) //nolint:gosec
		name := f.Name()

		if isSymlink {
			symlinkPath, _ = filepath.EvalSymlinks(filepath.Join(m.CurrentDirectory, name))
		}

		disabled := !m.canSelect(name) && !f.IsDir()

		if m.selected == i { //nolint:nestif
			selected := ""
			if m.ShowPermissions {
				selected += " " + info.Mode().String()
			}
			if m.ShowSize {
				selected += fmt.Sprintf("%"+strconv.Itoa(m.Styles.FileSize.GetWidth())+"s", size)
			}
			selected += " " + name
			if isSymlink {
				selected += " → " + symlinkPath
			}
			if disabled {
				s.WriteString(m.Styles.DisabledCursor.Render(m.Cursor) + m.Styles.DisabledSelected.Render(selected))
			} else {
				s.WriteString(m.Styles.Cursor.Render(m.Cursor) + m.Styles.Selected.Render(selected))
			}
			s.WriteRune('\n')
			continue
		}

		style := m.Styles.File
		if f.IsDir() {
			style = m.Styles.Directory
		} else if isSymlink {
			style = m.Styles.Symlink
		} else if disabled {
			style = m.Styles.DisabledFile
		}

		fileName := style.Render(name)
		s.WriteString(m.Styles.Cursor.Render(" "))
		if isSymlink {
			fileName += " → " + symlinkPath
		}
		if m.ShowPermissions {
			s.WriteString(" " + m.Styles.Permission.Render(info.Mode().String()))
		}
		if m.ShowSize {
			s.WriteString(m.Styles.FileSize.Render(size))
		}
		s.WriteString(" " + fileName)
		s.WriteRune('\n')
	}

	for i := lipgloss.Height(s.String()); i <= m.Height(); i++ {
		s.WriteRune('\n')
	}

	return s.String()
}

// DidSelectFile returns whether a user has selected a file (on this msg).
func (m Model) DidSelectFile(msg tea.Msg) (bool, string) {
	didSelect, path := m.didSelectFile(msg)
	if didSelect && m.canSelect(path) {
		return true, path
	}
	return false, ""
}

// DidSelectDisabledFile returns whether a user tried to select a disabled file
// (on this msg). This is necessary only if you would like to warn the user that
// they tried to select a disabled file.
func (m Model) DidSelectDisabledFile(msg tea.Msg) (bool, string) {
	didSelect, path := m.didSelectFile(msg)
	if didSelect && !m.canSelect(path) {
		return true, path
	}
	return false, ""
}

func (m Model) didSelectFile(msg tea.Msg) (bool, string) {
	vis := m.visibleFiles()
	if len(vis) == 0 {
		return false, ""
	}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// If the msg does not match the Select keymap then this could not have been a selection.
		if !key.Matches(msg, m.KeyMap.Select) {
			return false, ""
		}

		// The key press was a selection, let's confirm whether the current file could
		// be selected or used for navigating deeper into the stack.
		f := vis[m.selected]
		info, err := f.Info()
		if err != nil {
			return false, ""
		}
		isSymlink := info.Mode()&os.ModeSymlink != 0
		isDir := f.IsDir()

		if isSymlink {
			symlinkPath, _ := filepath.EvalSymlinks(filepath.Join(m.CurrentDirectory, f.Name()))
			info, err := os.Stat(symlinkPath)
			if err != nil {
				break
			}
			if info.IsDir() {
				isDir = true
			}
		}

		if (!isDir && m.FileAllowed) || (isDir && m.DirAllowed) && m.Path != "" {
			return true, m.Path
		}

		// If the msg was not a KeyPressMsg, then the file could not have been selected this iteration.
		// Only a KeyPressMsg can select a file.
	default:
		return false, ""
	}
	return false, ""
}

func (m Model) canSelect(file string) bool {
	if len(m.AllowedTypes) <= 0 {
		return true
	}

	for _, ext := range m.AllowedTypes {
		if strings.HasSuffix(file, ext) {
			return true
		}
	}
	return false
}

// HighlightedPath returns the path of the currently highlighted file or directory.
func (m Model) HighlightedPath() string {
	vis := m.visibleFiles()
	if len(vis) == 0 || m.selected < 0 || m.selected >= len(vis) {
		return ""
	}
	return filepath.Join(m.CurrentDirectory, vis[m.selected].Name())
}
