package plan

import (
	"fmt"

	"github.com/kendru/darwin/go/recursive-query/table"
)

type TableScan struct {
	Table *table.Table
	rows []table.Row
	idx int
	fetched bool
}

func (n *TableScan) Type() NodeType {
	return NodeTableScan
}

func (n *TableScan) Next() (Result, error) {
	if (!n.fetched) {
		n.fetched = true
		n.rows = n.Table.FetchAll()
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

func (n *TableScan) String() string {
	return fmt.Sprintf("(table-scan %s)", n.Table.Name)
}
