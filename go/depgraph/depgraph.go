package depgraph

import (
	"errors"
)

// A node in this graph is just a string, so a nodeset is a map whose
// keys are the nodes that are present.
type nodeset map[string]struct{}

// depmap tracks the nodes that have some dependency relationship to
// some other node, represented by the key of the map.
type depmap map[string]nodeset

type Graph struct {
	// Maintain dependency relationships in both directions.
	// `dependencies` tracks child -> parents, and `dependents` tracks parent -> children.
	dependencies, dependents depmap
	nodes                    nodeset
}

func New() *Graph {
	return &Graph{
		dependencies: make(depmap),
		dependents:   make(depmap),
		nodes:        make(nodeset),
	}
}

func (g *Graph) DependOn(child, parent string) error {
	if child == parent {
		return errors.New("self-referential dependencies not allowed")
	}

	if err := g.addEdge(child, parent); err != nil {
		return err
	}

	g.nodes[parent] = struct{}{}
	g.nodes[child] = struct{}{}

	return nil
}

func (g *Graph) addEdge(child, parent string) error {
	if g.DependsOn(parent, child) {
		return errors.New("circular dependencies not allowed")
	}

	addNodeToNodeset(g.dependents, parent, child)
	addNodeToNodeset(g.dependencies, child, parent)

	return nil
}

func (g *Graph) DependsOn(child, parent string) bool {
	deps := g.Dependencies(child)
	_, ok := deps[parent]
	return ok
}

func (g *Graph) HasDependent(parent, child string) bool {
	deps := g.Dependents(parent)
	_, ok := deps[child]
	return ok
}

func (g *Graph) Leaves() []string {
	out := make([]string, 0)
	for node := range g.nodes {
		dependencies, ok := g.dependencies[node]
		if !ok {
			out = append(out, node)
		} else {
			// Additionally, if no dependencies exist in the graph, consider this a leaf.
			var foundReference bool
			for referencedID := range dependencies {
				_, foundReference = g.nodes[referencedID]
				if foundReference {
					break
				}
			}
			if !foundReference {
				out = append(out, node)
			}
		}
	}
	return out
}

// TopoSortedLayers returns a slice of all of the graph nodes in topological sort order. That is,
// if `B` depends on `A`, then `A` is guaranteed to come before `B` in the sorted output.
// The graph is guaranteed to be cycle-free because cycles are detected while building the
// graph. Additionally, the output is grouped into "layers", which are guaranteed to not have
// any dependencies within each layer. This is useful, e.g. when building an execution plan for
// some DAG, in which case each element within each layer could be executed in parallel. If you
// do not need this layered property, use `Graph.TopoSorted()`, which flattens all elements.
func (g *Graph) TopoSortedLayers() [][]string {
	out := [][]string{}

	shrinkingGraph := g.clone()
	for {
		leaves := shrinkingGraph.Leaves()
		if len(leaves) == 0 {
			break
		}

		out = append(out, leaves)
		for _, leafNode := range leaves {

			dependents := shrinkingGraph.dependents[leafNode]

			for dependent := range dependents {
				// Should be safe because every relationship is bidirectional.
				dependencies := shrinkingGraph.dependencies[dependent]
				if len(dependencies) == 1 {
					// The only dependent _must_ be `leafNode`, so we can delete the `dep` entry entirely.
					delete(shrinkingGraph.dependencies, dependent)
				} else {
					delete(dependencies, leafNode)
				}
			}
			delete(shrinkingGraph.dependents, leafNode)
		}

		nextLeaves := shrinkingGraph.Leaves()
		leaves = nextLeaves
	}

	return out
}

// TopoSorted returns all the nodes in the graph is topological sort order.
// See also `Graph.TopoSortedLayers()`.
func (g *Graph) TopoSorted() []string {
	nodeCount := 0
	layers := g.TopoSortedLayers()
	for _, layer := range layers {
		nodeCount += len(layer)
	}

	allNodes := make([]string, 0, nodeCount)
	for _, layer := range layers {
		for _, node := range layer {
			allNodes = append(allNodes, node)
		}
	}

	return allNodes
}

func (g *Graph) Dependencies(child string) nodeset {
	return g.buildTransitive(child, g.immediateDependencies)
}

func (g *Graph) immediateDependencies(node string) nodeset {
	return g.dependencies[node]
}

func (g *Graph) Dependents(parent string) nodeset {
	return g.buildTransitive(parent, g.immediateDependents)
}

func (g *Graph) immediateDependents(node string) nodeset {
	return g.dependents[node]
}

func (g *Graph) clone() *Graph {
	return &Graph{
		dependencies: copyDepmap(g.dependencies),
		dependents:   copyDepmap(g.dependents),
		nodes:        copyNodeset(g.nodes),
	}
}

// buildTransitive starts at `root` and continues calling `nextFn` to keep discovering more nodes until
// the graph cannot produce any more. It returns the set of all discovered nodes.
func (g *Graph) buildTransitive(root string, nextFn func(string) nodeset) nodeset {
	if _, ok := g.nodes[root]; !ok {
		return nil
	}

	out := make(nodeset)
	searchNext := []string{root}
	for len(searchNext) > 0 {
		// List of new nodes from this layer of the dependency graph. This is
		// assigned to `searchNext` at the end of the outer "discovery" loop.
		discovered := []string{}
		for _, node := range searchNext {
			// For each node to discover, find the next nodes.
			for nextNode := range nextFn(node) {
				// If we have not seen the node before, add it to the output as well
				// as the list of nodes to traverse in the next iteration.
				if _, ok := out[nextNode]; !ok {
					out[nextNode] = struct{}{}
					discovered = append(discovered, nextNode)
				}
			}
		}
		searchNext = discovered
	}

	return out
}

func copyNodeset(s nodeset) nodeset {
	out := make(nodeset, len(s))
	for k, v := range s {
		out[k] = v
	}
	return out
}

func copyDepmap(m depmap) depmap {
	out := make(depmap, len(m))
	for k, v := range m {
		out[k] = copyNodeset(v)
	}
	return out
}

func addNodeToNodeset(dm depmap, key, node string) {
	nodes, ok := dm[key]
	if !ok {
		nodes = make(nodeset)
		dm[key] = nodes
	}
	nodes[node] = struct{}{}
}
