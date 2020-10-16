package plan

import "github.com/kendru/darwin/go/recursive-query/table"


type Node interface {
	Type() NodeType
	Next() (Result, error)
}

type NodeType int

type Result struct {
	Val table.Row
	hasMore bool
}

const (
	NodeTableScan NodeType = iota
	NodeIndexScan
	NodeProject
	NodeFilter
)
