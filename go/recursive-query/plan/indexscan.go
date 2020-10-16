package plan

import "github.com/kendru/darwin/go/recursive-query/table"

type IndexScan struct {
	table *table.Table
	index string
	min interface{}
	max interface{}
	rows []table.Row
	idx int
	fetched bool
}

func (n *IndexScan) Type() NodeType {
	return NodeIndexScan
}

func (n *IndexScan) Next() (Result, error) {
	var err error
	if (!n.fetched) {
		n.fetched = true
		n.rows, err = n.table.Scan(n.index, n.min, n.max)
		if err != nil {
			return Result{}, err
		}
	}

	maxIdx := len(n.rows)-1
	hasMore := n.idx < maxIdx
	var val table.Row
	if n.idx <= len(n.rows) {
		val = n.rows[n.idx]
		n.idx++
	}

	return Result{val, hasMore}, nil
}
