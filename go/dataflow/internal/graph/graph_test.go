package graph

import (
	"fmt"
	"testing"
)

func TestGraph(t *testing.T) {
	g := NewGraph()

	identity := func(x interface{}) interface{} {
		return x
	}
	inc := func(x interface{}) interface{} {
		return x.(int) + 1
	}

	a := g.NewNode(NewMapNode(identity))
	b := g.NewNode(NewMapNode(inc))
	c := g.NewNode(NewMapNode(identity))

	g.Link(a, "out", b, "in")
	g.Link(b, "out", c, "in")

	g.runNode(b, "in", []interface{}{1, 2, 3})

	fmt.Println(g.Debug())
	// Fake assertion so that the graph will print
	// assert.True(t, false)
}
