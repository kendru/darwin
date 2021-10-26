package main

import (
	"fmt"
	"strings"

	"github.com/kendru/darwin/go/depgraph"
)

func main() {
	g := depgraph.New()
	g.DependOn("cake", "eggs")
	g.DependOn("cake", "flour")
	g.DependOn("eggs", "chickens")
	g.DependOn("flour", "grain")
	g.DependOn("chickens", "feed")
	g.DependOn("chickens", "grain")
	g.DependOn("grain", "soil")

	for i, layer := range g.TopoSortedLayers() {
		fmt.Printf("%d: %s\n", i, strings.Join(layer, ", "))
	}
	// Output:
	// 0: feed, soil
	// 1: grain
	// 2: flour, chickens
	// 3: eggs
	// 4: cake
}
