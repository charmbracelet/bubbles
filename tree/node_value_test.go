package tree

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

type nodeValueStringer string

func (s nodeValueStringer) String() string { return string(s) }

func TestNodeSetValue(t *testing.T) {
	for _, test := range []struct {
		name  string
		value any
		want  string
	}{
		{"string", "after", "after"},
		{"number", 42, "42"},
		{"stringer", nodeValueStringer("formatted"), "formatted"},
	} {
		t.Run(test.name, func(t *testing.T) {
			node := Root("before").Child("child")
			node.SetValue(test.value)
			if got := node.GivenValue(); got != test.value {
				t.Errorf("GivenValue() = %v, want %v", got, test.value)
			}
			if got := ansi.Strip(node.Value()); got != test.want {
				t.Errorf("Value() = %q, want %q", got, test.want)
			}
			wantTree := test.want + "\n└── child"
			if got := ansi.Strip(node.String()); got != wantTree {
				t.Errorf("String() = %q, want %q", got, wantTree)
			}
		})
	}
}

func TestTreeSetChildValue(t *testing.T) {
	root := Root("root").Child("before")
	model := New(root, 40, 10)
	model.SetShowHelp(false)
	root.ChildNodes()[0].SetValue("after")
	model.SetNodes(root)
	view := ansi.Strip(model.View())
	if !strings.Contains(view, "after") || strings.Contains(view, "before") {
		t.Errorf("View() should render the updated child value: %q", view)
	}
}
