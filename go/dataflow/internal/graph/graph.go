package graph

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
)

type NodeID string

func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[NodeID]*Node),
		links: make(map[graphLocation][]graphLocation),
	}
}

type Graph struct {
	nodes  map[NodeID]*Node
	links  map[graphLocation][]graphLocation
	nextID uint64
}

func (g *Graph) Link(a NodeID, from nodePort, b NodeID, to nodePort) {
	source := graphLocation{a, from}
	dest := graphLocation{b, to}

	dests, ok := g.links[source]
	if !ok {
		dests = make([]graphLocation, 0)
		g.links[source] = dests
	}

	var exists bool
	for _, loc := range dests {
		if loc == dest {
			exists = true
			break
		}
	}

	if !exists {
		g.links[source] = append(dests, dest)
	}
}

func (g *Graph) Debug() string {
	var sb strings.Builder
	sb.WriteString("Nodes:\n")
	for _, node := range g.nodes {
		sb.WriteString(fmt.Sprintf("\t%s\n", node))
	}
	sb.WriteString("Links:\n")
	for src, dests := range g.links {
		prefix := fmt.Sprintf("\t[%s]:(%s) -> ", src.node, src.port)
		pad := strings.Repeat(" ", len(prefix))
		sb.WriteString(prefix)
		for i, dest := range dests {
			sb.WriteString(fmt.Sprintf("(%s):[%s]\n", dest.port, dest.node))
			if i < len(dests)-1 {
				sb.WriteString(pad)
			}
		}
	}
	return sb.String()
}

type graphLocation struct {
	node NodeID
	port nodePort
}

type nodePort = string

func (g *Graph) NewNode(inner StatefulNode) NodeID {
	id := fmt.Sprintf("node%d", atomic.AddUint64(&g.nextID, 1))
	node := &Node{
		ID:           NodeID(id),
		StatefulNode: inner,
	}
	g.nodes[node.ID] = node

	return node.ID
}

func (g *Graph) runNode(id NodeID, port nodePort, messages []interface{}) error {
	node, ok := g.nodes[id]
	if !ok {
		// TODO: Use custom errors.
		return fmt.Errorf("Node not found in graph: %q", id)
	}

	// outputs contains a map of the output port to the messages that were emitted.
	outputs := make(map[nodePort][]interface{})
	for _, msg := range messages {
		Ingest(node, port, msg, outputs)
	}

	if jsonBytes, err := json.MarshalIndent(outputs, "", "  "); err == nil {
		fmt.Println(string(jsonBytes))
	} else {
		fmt.Println("Error printing outputs:", err)
	}

	// For each output, find all connected nodes.
	for outPort, outMessages := range outputs {
		outputLocation := graphLocation{
			node: id,
			port: outPort,
		}
		nextLocations := g.links[outputLocation]
		fmt.Printf("From %s to:", outputLocation)
		_ = outMessages
		for _, nextLocation := range nextLocations {
			fmt.Printf("%s ", nextLocation)
		}
		fmt.Println()
	}

	return nil
}

type Node struct {
	ID NodeID
	StatefulNode
}

func (n *Node) String() string {
	return fmt.Sprintf("<%s:%s>", n.Name(), n.ID)
}

type StatefulNode interface {
	Name() string
	DoIngest(port nodePort, msg interface{}, emit emitFn) error
}

type emitFn func(port nodePort, msg interface{})

func Ingest(n *Node, port nodePort, msg interface{}, outputs map[nodePort][]interface{}) error {
	return n.StatefulNode.DoIngest(port, msg, func(port nodePort, msg interface{}) {
		outputs[port] = append(outputs[port], msg)
	})
}

func NewMapNode(mapper func(interface{}) interface{}) *MapNode {
	return &MapNode{mapper}
}

type MapNode struct {
	mapper func(interface{}) interface{}
}

func (n *MapNode) Name() string {
	return "map"
}

func (n *MapNode) DoIngest(port nodePort, msg interface{}, emit emitFn) error {
	switch string(port) {
	case "in":
		emit(nodePort("out"), n.mapper(msg))
	default:
		return fmt.Errorf("Invalid port for MapNode: %s", port)
	}
	return nil
}
