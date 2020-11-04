package plan

import "github.com/kendru/darwin/go/recursive-query/table"

type Node interface {
	Type() NodeType
	Next() (Result, error)

	// A property describes some physical feature of the resulting
	// data set - e.g. that it is sorted by some key.
	Properties() []Property
}

type NodeType int

const (
	NodeTableScan NodeType = iota
	NodeIndexScan
	NodeProject
	NodeFilter
)

type Result struct {
	Val     table.Row
	hasMore bool
}
