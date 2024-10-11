// Package tree provides a tree component for Bubble Tea
// applications.
package tree

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	ltree "github.com/charmbracelet/lipgloss/v2/tree"

	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/viewport"
)

const spacebar = " "

// KeyMap is the key bindings for different actions within the tree.
type KeyMap struct {
	Down         key.Binding
	Up           key.Binding
	PageDown     key.Binding
	PageUp       key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	GoToTop      key.Binding
	GoToBottom   key.Binding

	Toggle key.Binding
	Open   key.Binding
	Close  key.Binding

	// Help toggle keybindings.
	ShowFullHelp  key.Binding
	CloseFullHelp key.Binding

	Quit key.Binding
}

// DefaultKeyMap returns the default set of key bindings for navigating and acting
// upon the tree.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Down: key.NewBinding(
			key.WithKeys("down", "j", "ctrl+n"),
			key.WithHelp("↓/j", "down"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k", "ctrl+p"),
			key.WithHelp("↑/k", "up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", spacebar, "f"),
			key.WithHelp("f/pgdn", "page down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "b"),
			key.WithHelp("b/pgup", "page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("d", "ctrl+d"),
			key.WithHelp("d", "½ page down"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("u", "ctrl+u"),
			key.WithHelp("u", "½ page up"),
		),
		GoToTop: key.NewBinding(
			key.WithKeys("g", "home"),
			key.WithHelp("g", "top"),
		),
		GoToBottom: key.NewBinding(
			key.WithKeys("G", "shift+g", "end"),
			key.WithHelp("G", "bottom"),
		),

		Toggle: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("⏎", "toggle"),
		),
		Open: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("→/l", "open"),
		),
		Close: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("←/h", "close"),
		),

		// Toggle help.
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more"),
		),
		CloseFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "close help"),
		),

		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// Model is the Bubble Tea model for this tree element.
type Model struct {
	showHelp bool
	// scrollOff is the minimal number of lines to keep visible above and below the selected node.
	scrollOff int
	// openCharacter is the character used to represent an open node.
	openCharacter string
	// closedCharacter is the character used to represent a closed node.
	closedCharacter string
	// cursorCharacter is the character used to represent the cursor.
	cursorCharacter string
	// KeyMap encodes the keybindings recognized by the widget.
	KeyMap KeyMap
	// styles sets the styling for the tree
	styles Styles
	Help   help.Model

	// Additional key mappings for the short and full help views. This allows
	// you to add additional key mappings to the help menu without
	// re-implementing the help component. Of course, you can also disable the
	// list's help component and implement a new one if you need more
	// flexibility.
	AdditionalShortHelpKeys func() []key.Binding
	AdditionalFullHelpKeys  func() []key.Binding

	root *Node

	enumerator *ltree.Enumerator
	indenter   *ltree.Indenter

	viewport viewport.Model
	width    int
	height   int
	// yOffset is the vertical offset of the selected node.
	yOffset int
}

// New creates a new model with default settings.
func New(t *Node, width, height int) Model {
	m := Model{
		KeyMap:          DefaultKeyMap(),
		openCharacter:   "▼",
		closedCharacter: "▶",
		cursorCharacter: "→",
		Help:            help.New(),
		scrollOff:       5,

		showHelp: true,
		root:     t,
		viewport: viewport.Model{},
	}

	if m.root == nil {
		m.root = NewNode()
	}

	m.SetStyles(DefaultDarkStyles())
	m.SetSize(width, height)
	if m.root != nil {
		m.setAttributes()
		m.updateStyles()
		m.updateViewport(0)
	}
	return m
}

// Update is the Bubble Tea update loop.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Down):
			m.Down()
		case key.Matches(msg, m.KeyMap.Up):
			m.Up()
		case key.Matches(msg, m.KeyMap.PageDown):
			m.PageDown()
		case key.Matches(msg, m.KeyMap.PageUp):
			m.PageUp()
		case key.Matches(msg, m.KeyMap.HalfPageDown):
			m.HalfPageDown()
		case key.Matches(msg, m.KeyMap.HalfPageUp):
			m.HalfPageUp()
		case key.Matches(msg, m.KeyMap.GoToTop):
			m.GoToTop()
		case key.Matches(msg, m.KeyMap.GoToBottom):
			m.GoToBottom()

		case key.Matches(msg, m.KeyMap.Toggle):
			m.ToggleCurrentNode()
		case key.Matches(msg, m.KeyMap.Open):
			m.OpenCurrentNode()
		case key.Matches(msg, m.KeyMap.Close):
			m.CloseCurrentNode()

		case key.Matches(msg, m.KeyMap.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.KeyMap.ShowFullHelp):
			fallthrough
		case key.Matches(msg, m.KeyMap.CloseFullHelp):
			m.Help.ShowAll = !m.Help.ShowAll
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the component.
func (m Model) View() string {
	treeView := m.viewport.View()

	var help string
	if m.showHelp {
		help = m.helpView()
	}

	return lipgloss.JoinVertical(lipgloss.Left, treeView, help)
}

// SetScrollOff sets the minimal number of lines to keep visible above and below the selected node.
func (m *Model) SetScrollOff(val int) {
	m.scrollOff = val
}

// SetOpenCharacter sets the character used to represent an open node.
func (m *Model) SetOpenCharacter(character string) {
	m.openCharacter = character
}

// SetClosedCharacter sets the character used to represent a closed node.
func (m *Model) SetClosedCharacter(character string) {
	m.closedCharacter = character
}

// SetCursorCharacter sets the character used to represent the cursor.
func (m *Model) SetCursorCharacter(character string) {
	m.cursorCharacter = character
}

// SetNodes sets the tree to the given root node.
func (m *Model) SetNodes(t *Node) {
	m.root = t
	if m.enumerator != nil {
		m.root.Enumerator(*m.enumerator)
	}
	if m.indenter != nil {
		m.root.Indenter(*m.indenter)
	}
	m.setAttributes()
	m.updateStyles()
	m.updateViewport(0)
}

// Down moves the selection down by one item.
func (m *Model) Down() {
	m.updateViewport(1)
}

// Up moves the selection up by one item.
func (m *Model) Up() {
	m.updateViewport(-1)
}

// PageDown moves the selection down by one page.
func (m *Model) PageDown() {
	m.updateViewport(m.viewport.Height())
}

// PageUp moves the selection up by one page.
func (m *Model) PageUp() {
	m.updateViewport(-m.viewport.Height())
}

// HalfPageDown moves the selection down by half a page.
func (m *Model) HalfPageDown() {
	m.updateViewport(m.viewport.Height() / 2)
}

// HalfPageUp moves the selection up by half a page.
func (m *Model) HalfPageUp() {
	m.updateViewport(-m.viewport.Height() / 2)
}

// GoToTop moves the selection to the top of the tree.
func (m *Model) GoToTop() {
	m.updateViewport(-m.yOffset)
}

// GoToBottom moves the selection to the bottom of the tree.
func (m *Model) GoToBottom() {
	m.updateViewport(m.root.Size())
}

// ToggleCurrentNode toggles the current node open/close state.
func (m *Model) ToggleCurrentNode() {
	node := findNode(m.root, m.yOffset)
	if node == nil {
		return
	}
	m.toggleNode(node, !node.IsOpen())
}

// OpenCurrentNode opens the currently selected node.
func (m *Model) OpenCurrentNode() {
	node := findNode(m.root, m.yOffset)
	if node == nil {
		return
	}
	m.toggleNode(node, true)
}

// CloseCurrentNode closes the currently selected node.
func (m *Model) CloseCurrentNode() {
	node := findNode(m.root, m.yOffset)
	if node == nil {
		return
	}
	m.toggleNode(node, false)
}

func (m *Model) toggleNode(node *Node, open bool) {
	node.open = open

	// reset the offset to 0,0 first
	node.tree.Offset(0, 0)
	if !open {
		node.tree.Offset(node.tree.Children().Length(), 0)
	}
	m.setAttributes()
	m.updateViewport(m.yOffset - node.yOffset)
}

func (m *Model) updateViewport(movement int) {
	if m.root == nil {
		return
	}

	m.yOffset = max(min(m.root.Size()-1, m.yOffset+movement), 0)
	m.updateStyles()

	cursor := m.cursorView()
	m.viewport.SetContent(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			cursor,
			m.styles.TreeStyle.Render(m.root.String()),
		),
	)

	// if this is the initial render, make sure we show the root node
	if m.yOffset == 0 && movement == 0 {
		return
	}

	// make sure there are enough lines above and below the selected node
	height := m.viewport.VisibleLineCount()
	scrolloff := min(m.scrollOff, height/2)
	minTop := max(m.yOffset-scrolloff, 0)
	minBottom := min(m.viewport.TotalLineCount()-1, m.yOffset+scrolloff)

	if m.viewport.YOffset() > minTop { // reveal more lines above
		m.viewport.SetYOffset(minTop)
	} else if m.viewport.YOffset()+height < minBottom+1 { // reveal more lines below
		m.viewport.SetYOffset(minBottom - height + 1)
	}
}

// SetStyles sets the styles for this component.
func (m *Model) SetStyles(styles Styles) {
	if styles.NodeStyleFunc != nil {
		styles.nodeFunc = styles.NodeStyleFunc
	} else {
		styles.nodeFunc = func(_ Nodes, _ int) lipgloss.Style {
			return styles.NodeStyle
		}
	}
	if styles.SelectedNodeStyleFunc != nil {
		styles.selectedNodeFunc = styles.SelectedNodeStyleFunc
	} else {
		styles.selectedNodeFunc = func(_ Nodes, _ int) lipgloss.Style {
			return styles.SelectedNodeStyle
		}
	}
	if styles.ParentNodeStyleFunc != nil {
		styles.parentNodeFunc = styles.ParentNodeStyleFunc
	} else {
		styles.parentNodeFunc = func(_ Nodes, _ int) lipgloss.Style {
			return styles.ParentNodeStyle
		}
	}
	if styles.RootNodeStyleFunc != nil {
		styles.rootNodeFunc = styles.RootNodeStyleFunc
	} else {
		styles.rootNodeFunc = func(_ Nodes, _ int) lipgloss.Style {
			return styles.RootNodeStyle
		}
	}

	if m.root != nil {
		m.root.EnumeratorStyle(styles.EnumeratorStyle)
		m.root.IndenterStyle(styles.IndenterStyle)
		m.root.ItemStyleFunc(func(children Nodes, i int) lipgloss.Style {
			child := children.At(i)
			return child.getStyle()
		})
	}

	m.styles = styles
	// call SetSize as it takes into account width/height of the styles frame sizes
	m.SetSize(m.width, m.height)
	m.updateViewport(0)
}

// SetShowHelp shows or hides the help view.
func (m *Model) SetShowHelp(v bool) {
	m.showHelp = v
	m.SetSize(m.width, m.height)
}

// Width returns the current width setting.
func (m Model) Width() int {
	return m.width
}

// Height returns the current height setting.
func (m Model) Height() int {
	return m.height
}

// SetWidth sets the width of this component.
func (m *Model) SetWidth(width int) {
	m.SetSize(width, m.height)
}

// SetHeight sets the height of this component.
func (m *Model) SetHeight(height int) {
	m.SetSize(m.width, height)
}

// SetSize sets the width and height of this component.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.root.tree.Width(width - lipgloss.Width(m.cursorView()) - m.styles.TreeStyle.GetHorizontalFrameSize())

	m.viewport.SetWidth(width)
	hv := 0
	if m.showHelp {
		hv = lipgloss.Height(m.helpView())
	}
	m.viewport.SetHeight(height - hv)
	m.Help.Width = width
}

// ShortHelp returns bindings to show in the abbreviated help view.
// It's part of the help.KeyMap interface.
func (m Model) ShortHelp() []key.Binding {
	kb := []key.Binding{
		m.KeyMap.Down,
		m.KeyMap.Up,
		m.KeyMap.Toggle,
	}

	if m.AdditionalShortHelpKeys != nil {
		kb = append(kb, m.AdditionalShortHelpKeys()...)
	}

	kb = append(kb, m.KeyMap.Quit, m.KeyMap.ShowFullHelp)

	return kb
}

// FullHelp returns bindings to show the full help view. It's part of the
// help.KeyMap interface.
func (m Model) FullHelp() [][]key.Binding {
	kb := [][]key.Binding{
		{
			m.KeyMap.Down,
			m.KeyMap.Up,
			m.KeyMap.Open,
			m.KeyMap.Close,
			m.KeyMap.Toggle,
		},
		{
			m.KeyMap.PageDown,
			m.KeyMap.PageUp,
			m.KeyMap.HalfPageDown,
			m.KeyMap.HalfPageUp,
		},
		{
			m.KeyMap.GoToTop,
			m.KeyMap.GoToBottom,
		},
	}

	if m.AdditionalFullHelpKeys != nil {
		kb = append(kb, m.AdditionalFullHelpKeys())
	}

	kb = append(kb, []key.Binding{
		m.KeyMap.Quit,
		m.KeyMap.CloseFullHelp,
	})

	return kb
}

func (m Model) cursorView() string {
	if m.cursorCharacter == "" {
		return ""
	}
	cursor := strings.Split(strings.Repeat(" ", m.root.Size()), "")
	cursor[m.yOffset] = m.cursorCharacter
	return m.styles.CursorStyle.Render(lipgloss.JoinVertical(lipgloss.Left, cursor...))
}

func (m Model) helpView() string {
	return m.styles.HelpStyle.Render(m.Help.View(m))
}

// Root returns the root node of the tree.
// Equivalent to calling `Model.Node(0)`.
func (m *Model) Root() *Node {
	return m.root
}

// AllNodes returns all nodes in the tree as a flat list.
func (m *Model) AllNodes() []*Node {
	return m.root.AllNodes()
}

func (m *Model) setAttributes() {
	setDepths(m.root, 0)
	setYOffsets(m.root)
}

// setSizes updates each Node's size.
// Note that if a child isn't open, its size is 1.
func setDepths(t *Node, depth int) {
	t.depth = depth
	children := t.tree.Children()
	for i := 0; i < children.Length(); i++ {
		child := children.At(i)
		setDepths(child.(*Node), depth+1)
	}
}

// setYOffsets updates each Node's yOffset based on how many items are "above" it.
func setYOffsets(t *Node) {
	children := t.tree.Children()
	above := 0
	for i := 0; i < children.Length(); i++ {
		child := children.At(i)
		if child, ok := child.(*Node); ok {
			child.yOffset = t.yOffset + above + i + 1
			setYOffsets(child)
			above += child.Size() - 1
		}
	}
}

// YOffset returns the vertical offset of the selected node.
// Useful for scrolling to the selected node using a viewport.
func (m *Model) YOffset() int {
	return m.yOffset
}

// Node returns the item at the given yoffset.
func (m *Model) Node(yoffset int) *Node {
	return findNode(m.root, yoffset)
}

// NodeAtCurrentOffset returns the item at the current yoffset.
func (m *Model) NodeAtCurrentOffset() *Node {
	return findNode(m.root, m.yOffset)
}

// Enumerator sets the enumerator for the tree.
func (m *Model) Enumerator(enumerator ltree.Enumerator) *Model {
	m.enumerator = &enumerator
	m.root.Enumerator(enumerator)
	return m
}

// Indenter sets the indenter for the tree.
func (m *Model) Indenter(indenter ltree.Indenter) *Model {
	m.indenter = &indenter
	m.root.Indenter(indenter)
	return m
}

// Since the selected node changes, we need to capture m.yOffset in the
// style function's closure again.
func (m *Model) updateStyles() {
	if m.root != nil {
		m.root.RootStyle(m.rootStyle())
	}

	items := m.AllNodes()
	opts := m.getItemOpts()
	for _, item := range items {
		item.opts = *opts
	}
}

func (m *Model) getItemOpts() *itemOptions {
	return &itemOptions{
		openCharacter:   m.openCharacter,
		closedCharacter: m.closedCharacter,
		treeYOffset:     m.yOffset,
		styles:          m.styles,
	}
}

func (m *Model) rootStyle() lipgloss.Style {
	if m.root.yOffset == m.yOffset {
		return m.styles.selectedNodeFunc(Nodes{m.root}, 0)
	}

	return m.styles.rootNodeFunc(Nodes{m.root}, 0)
}

// findNode starts a DFS search for the node with the given yOffset
// starting from the given item.
func findNode(t *Node, yOffset int) *Node {
	if t.yOffset == yOffset {
		return t
	}

	children := t.tree.Children()
	for i := 0; i < children.Length(); i++ {
		child := children.At(i)
		if child, ok := child.(*Node); ok {
			found := findNode(child, yOffset)
			if found != nil {
				return found
			}
		}
	}

	return nil
}
