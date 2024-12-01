package tree

import (
	"github.com/charmbracelet/lipgloss"
	ltree "github.com/charmbracelet/lipgloss/tree"
)

// Node is a a node in the tree
// Node implements lipgloss's tree.Node
type Node struct {
	// tree is used as the renderer layer
	tree *ltree.Tree

	// yOffset is the vertical offset of the selected node.
	yOffset int

	// depth is the depth of the node in the tree
	depth int

	// isRoot is true for every Node which was added with tree.Root
	isRoot        bool
	initialClosed bool
	open          bool

	// value is the root value of the node
	value any

	opts itemOptions

	// TODO: move to lipgloss.Tree?
	size int
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
// Note that if a child isn't open, its size is 1
func (t *Node) Size() int {
	return t.size
}

// YOffset returns the vertical offset of the Node
func (t *Node) YOffset() int {
	return t.yOffset
}

// Close closes the node.
func (t *Node) Close() *Node {
	t.initialClosed = true
	t.open = false
	// reset the offset to 0,0 first
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

// Used to print the Node's tree
// TODO: Value is not called on the root node, so we need to repeat the open/closed character
// Should this be fixed in lipgloss?
func (t *Node) String() string {
	// s := t.opts.styles.OpenIndicatorStyle
	// if t.open {
	// 	return s.Render(t.opts.openCharacter+" ") + t.tree.String()
	// }
	// return s.Render(t.opts.closedCharacter+" ") + t.tree.String()
	return t.tree.String()
}

// Value returns the root name of this node.
func (t *Node) Value() string {
	s := t.opts.styles
	var ns lipgloss.Style

	if t.yOffset == t.opts.treeYOffset {
		ns = s.selectedNodeFunc(Nodes{t}, 0)
	} else if t.yOffset == 0 {
		ns = s.rootNodeFunc(Nodes{t}, 0)
	} else if t.isRoot {
		ns = s.parentNodeFunc(Nodes{t}, 0)
	} else {
		ns = s.nodeFunc(Nodes{t}, 0)
	}

	v := ns.Render(t.tree.Value())

	if t.isRoot {
		if t.open {
			return s.OpenIndicatorStyle.Render(t.opts.openCharacter+" ") + v
		}
		return s.OpenIndicatorStyle.Render(t.opts.closedCharacter+" ") + v
	}

	// leaf
	return v
}

// GivenValue returns the value passed to the node.
func (t *Node) GivenValue() any {
	return t.value
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

// FlatNodes returns all descendant items in as a flat list.
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

// Nodes are a list of tree nodes.
type Nodes []*Node

// Children returns the children of an item.
func (t Nodes) At(index int) *Node {
	return t[index]
}

// Children returns the children of an item.
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
//
// TODO: currently unused as this is set in the Styles struct.
func (t *Node) ItemStyleFunc(f StyleFunc) *Node {
	t.tree.ItemStyleFunc(func(children ltree.Children, i int) lipgloss.Style {
		c := make(Nodes, children.Length())
		// TODO: if we expose Depth and Size in lipgloss, we can avoid this
		for i := 0; i < children.Length(); i++ {
			c[i] = children.At(i).(*Node)
		}
		return f(c, i)
	})
	return t
}

// TODO: support IndentStyleFunc in lipgloss so we can have a full background for the item

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
// TODO: currently unused as this is set in the Styles struct.
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
//
// TODO: currently unused as this is set in the Styles struct.
func (t *Node) EnumeratorStyleFunc(f func(children ltree.Children, i int) lipgloss.Style) *Node {
	t.tree.EnumeratorStyleFunc(f)
	return t
}

// IndenterStyle sets a static style for all indenters.
//
// Use IndenterStyleFunc to conditionally set styles based on the tree node.
// TODO: currently unused as this is set in the Styles struct.
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
//
// TODO: currently unused as this is set in the Styles struct.
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
			t.size = t.size + child.size
			t.tree.Child(child)

			// Close the node again as the number of children has changed
			if t.initialClosed {
				t.Close()
			}
		default:
			item := new(Node)
			item.opts.styles = DefaultStyles()
			item.tree = ltree.Root(child)
			item.size = 1
			item.open = false
			item.value = child
			t.size = t.size + item.size
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
	t.opts.styles = DefaultStyles()
	t.size = 1
	t.open = true
	t.isRoot = true
	t.tree = ltree.New()
	return t
}

// Root returns a new tree with the root set.
func Root(root any) *Node {
	t := new(Node)
	t.opts.styles = DefaultStyles()
	t.size = 1
	t.value = root
	t.open = true
	t.isRoot = true
	switch root := root.(type) {
	case *Node:
		t.tree = ltree.Root(root.Value())
	default:
		item := new(Node)
		item.value = root
		item.opts.styles = DefaultStyles()
		item.size = 1
		item.open = true
		item.isRoot = true
		t.tree = ltree.Root(item)
	}

	return t
}
