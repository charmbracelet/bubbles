package main

import (
	"fmt"
	"testing"
)

// TestLess test the symmetric of the used less function.
// Since if the less function yields a < b == b < a
// for any possible input the sort while not be reproducible!!!
func TestLess(t *testing.T) {
	allNodes := []fmt.Stringer{
		node{
			parentIDs: []int{7},
			value:     "no children here"},
		node{
			parentIDs: []int{1},
			value:     "use '+' to unfold a node"},
		node{
			parentIDs: []int{1, 4},
			value:     "use '-' to hide all children of this node"},
		node{
			parentIDs: []int{1, 8},
			value:     "use 'up' and 'down' to move around"},
		node{
			parentIDs: []int{1, 4, 5},
			value:     "grand child\nwith a line break"},
		node{
			parentIDs: []int{3},
			value:     "parent with no grand children"},
		node{
			parentIDs: []int{3, 2},
			value:     "hänsel"},
		node{
			parentIDs: []int{3, 6},
			value:     "gretel"},
	}
	allLen := len(allNodes)
	for c := 0; c < allLen; c++ {
		for i := c + 1; i < allLen; i++ {
			if less(allNodes[c], allNodes[i]) == less(allNodes[i], allNodes[c]) {
				cIDs, _ := allNodes[c].(node)
				iIDs, _ := allNodes[i].(node)
				t.Errorf("%v, %v", cIDs.parentIDs, iIDs.parentIDs)
			}
		}
	}
}
