package tree

import (
	"charm.land/lipgloss/v2"
	ltree "charm.land/lipgloss/v2/tree"
)

// Node is a a node in the tree.
// Node implements lipgloss's tree.Node.
type Node struct {
	// tree is used as the renderer layer.
	tree *ltree.Tree

	// yOffset is the vertical offset of the selected node.
	yOffset int

	// depth is the depth of the node in the tree.
	depth int

	// isRoot is true for every Node which was added with tree.Root.
	isRoot        bool
	initialClosed bool
	open          bool

	// value is the root value of the node.
	value any

	opts itemOptions
}

// IsSelected returns whether this item is selected.
func (t *Node) IsSelected() bool {
	return t.yOffset == t.opts.treeYOffset
}

// Depth returns the depth of the node in the tree.
func (t *Node) Depth() int {
	return t.depth
}

// Size returns the number of nodes in the tree.
// Note that if a child isn't open, its size is 1.
func (t *Node) Size() int {
	return len(t.AllNodes())
}

// YOffset returns the vertical offset of the Node.
func (t *Node) YOffset() int {
	return t.yOffset
}

// Close closes the node.
func (t *Node) Close() *Node {
	t.initialClosed = true
	t.open = false
	// Reset the offset to 0,0 first.
	t.tree.Offset(0, 0)
	t.tree.Offset(t.tree.Children().Length(), 0)
	return t
}

// Open opens the node.
func (t *Node) Open() *Node {
	t.open = true
	t.tree.Offset(0, 0)
	return t
}

// IsOpen returns whether the node is open.
func (t *Node) IsOpen() bool {
	return t.open
}

type itemOptions struct {
	openCharacter   string
	closedCharacter string
	treeYOffset     int
	styles          Styles
}

// Used to print the Node's tree.
func (t *Node) String() string {
	s := t.opts.styles.OpenIndicatorStyle
	if t.open {
		return s.Render(t.opts.openCharacter+" ") + t.tree.String()
	}
	return s.Render(t.opts.closedCharacter+" ") + t.tree.String()
}

func (t *Node) getStyle() lipgloss.Style {
	s := t.opts.styles
	if t.yOffset == t.opts.treeYOffset {
		return s.selectedNodeFunc(Nodes{t}, 0)
	} else if t.yOffset == 0 {
		return s.rootNodeFunc(Nodes{t}, 0)
	} else if t.isRoot {
		return s.parentNodeFunc(Nodes{t}, 0)
	}

	return s.nodeFunc(Nodes{t}, 0)
}

// Value returns the root name of this node.
func (t *Node) Value() string {
	s := t.opts.styles
	ns := t.getStyle()
	v := ns.Render(t.tree.Value())

	if t.isRoot {
		if t.open {
			return s.OpenIndicatorStyle.Render(t.opts.openCharacter+" ") + v
		}
		return s.OpenIndicatorStyle.Render(t.opts.closedCharacter+" ") + v
	}

	// Leaf.
	return v
}

// GivenValue returns the value passed to the node.
func (t *Node) GivenValue() any {
	return t.value
}

// SetValue sets the value of the node.
func (t *Node) SetValue(value any) {
	t.value = value
}

// Children returns the children of an item.
func (t *Node) Children() ltree.Children {
	return t.tree.Children()
}

// ChildNodes returns the children of an item.
func (t *Node) ChildNodes() []*Node {
	res := []*Node{}
	children := t.tree.Children()
	for i := 0; i < children.Length(); i++ {
		child := children.At(i)
		res = append(res, child.(*Node))
	}
	return res
}

// AllNodes returns all descendant nodes as a flat list.
func (t *Node) AllNodes() []*Node {
	res := []*Node{t}
	children := t.tree.Children()
	for i := 0; i < children.Length(); i++ {
		child := children.At(i)
		res = append(res, child.(*Node).AllNodes()...)
	}
	return res
}

// Hidden returns whether this item is hidden.
func (t *Node) Hidden() bool {
	return t.tree.Hidden()
}

// SetHidden hides/shows a Tree node.
func (t *Node) SetHidden(hidden bool) {
	t.tree.SetHidden(hidden)
}

// Nodes are a list of tree nodes.
type Nodes []*Node

// At returns the node at the index.
func (t Nodes) At(index int) *Node {
	return t[index]
}

// Length returns the number of nodes.
func (t Nodes) Length() int {
	return len(t)
}

// ItemStyle sets a static style for all items.
func (t *Node) ItemStyle(s lipgloss.Style) *Node {
	t.tree.ItemStyle(s)
	return t
}

// ItemStyleFunc sets the item style function. Use this for conditional styling.
// For example:
//
//	t := tree.Root("root").
//		ItemStyleFunc(func(_ tree.Nodes, i int) lipgloss.Style {
//			if selected == i {
//				return lipgloss.NewStyle().Foreground(hightlightColor)
//			}
//			return lipgloss.NewStyle().Foreground(dimColor)
//		})
func (t *Node) ItemStyleFunc(f StyleFunc) *Node {
	t.tree.ItemStyleFunc(func(children ltree.Children, i int) lipgloss.Style {
		c := make(Nodes, children.Length())
		for i := 0; i < children.Length(); i++ {
			c[i] = children.At(i).(*Node)
		}
		return f(c, i)
	})
	return t
}

// Enumerator sets the enumerator implementation. This can be used to change the
// way the branches indicators look.  Lipgloss includes predefined enumerators
// for a classic or rounded tree. For example, you can have a rounded tree:
//
//	tree.New().
//		Enumerator(ltree.RoundedEnumerator)
func (t *Node) Enumerator(enumerator ltree.Enumerator) *Node {
	t.tree.Enumerator(enumerator)
	return t
}

// Indenter sets the indenter implementation. This is used to change the way
// the tree is indented. The default indentor places a border connecting sibling
// elements and no border for the last child.
//
//	└── Foo
//	    └── Bar
//	        └── Baz
//	            └── Qux
//	                └── Quux
//
// You can define your own indenter.
//
//	func ArrowIndenter(children tree.Children, index int) string {
//		return "→ "
//	}
//
//	→ Foo
//	→ → Bar
//	→ → → Baz
//	→ → → → Qux
//	→ → → → → Quux
func (t *Node) Indenter(indenter ltree.Indenter) *Node {
	t.tree.Indenter(indenter)
	return t
}

// EnumeratorStyle sets a static style for all enumerators.
//
// Use EnumeratorStyleFunc to conditionally set styles based on the tree node.
func (t *Node) EnumeratorStyle(style lipgloss.Style) *Node {
	t.tree.EnumeratorStyle(style)
	return t
}

// EnumeratorStyleFunc sets the enumeration style function. Use this function
// for conditional styling.
//
//	t := tree.Root("root").
//		EnumeratorStyleFunc(func(_ tree.Children, i int) lipgloss.Style {
//		    if selected == i {
//		        return lipgloss.NewStyle().Foreground(hightlightColor)
//		    }
//		    return lipgloss.NewStyle().Foreground(dimColor)
//		})
func (t *Node) EnumeratorStyleFunc(f func(children ltree.Children, i int) lipgloss.Style) *Node {
	t.tree.EnumeratorStyleFunc(f)
	return t
}

// IndenterStyle sets a static style for all indenters.
//
// Use IndenterStyleFunc to conditionally set styles based on the tree node.
func (t *Node) IndenterStyle(style lipgloss.Style) *Node {
	t.tree.IndenterStyle(style)
	return t
}

// IndenterStyleFunc sets the indenter style function. Use this function
// for conditional styling.
//
//	t := tree.Root("root").
//		IndenterStyleFunc(func(_ tree.Children, i int) lipgloss.Style {
//		    if selected == i {
//		        return lipgloss.NewStyle().Foreground(hightlightColor)
//		    }
//		    return lipgloss.NewStyle().Foreground(dimColor)
//		})
func (t *Node) IndenterStyleFunc(f func(children ltree.Children, i int) lipgloss.Style) *Node {
	t.tree.IndenterStyleFunc(f)
	return t
}

// RootStyle sets a style for the root element.
func (t *Node) RootStyle(style lipgloss.Style) *Node {
	t.tree.RootStyle(style)
	return t
}

// Child adds a child to this tree.
//
// If a Child Node is passed without a root, it will be parented to it's sibling
// child (auto-nesting).
//
//	tree.Root("Foo").Child(tree.Root("Bar").Child("Baz"), "Qux")
//
//	├── Foo
//	├── Bar
//	│   └── Baz
//	└── Qux
func (t *Node) Child(children ...any) *Node {
	for _, child := range children {
		switch child := child.(type) {
		case *Node:
			t.tree.Child(child)

			// Close the node again as the number of children has changed
			if t.initialClosed {
				t.Close()
			}
		default:
			item := new(Node)
			item.opts.styles = DefaultDarkStyles()
			item.tree = ltree.Root(child)
			item.open = false
			item.value = child
			t.tree.Child(item)

			// Close the node again as the number of children has changed
			if t.initialClosed {
				t.Close()
			}
		}
	}

	return t
}

// NewNode returns a new node.
func NewNode() *Node {
	t := new(Node)
	t.opts.styles = DefaultDarkStyles()
	t.open = true
	t.isRoot = true
	t.tree = ltree.New()
	return t
}

// Root returns a new tree with the root set.
func Root(root any) *Node {
	t := NewNode()
	t.value = root
	t.tree = ltree.Root(root)
	return t
}
