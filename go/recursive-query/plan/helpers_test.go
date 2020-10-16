package plan

import "github.com/kendru/darwin/go/recursive-query/table"

func newStubProducerNode(rows []table.Row) *stubProducerNode {
	return &stubProducerNode{rows: rows}
}

type stubProducerNode struct {
	rows []table.Row
	idx int
	fetched bool
}

func (n *stubProducerNode) Type() NodeType {
	return 99999
}

func (n *stubProducerNode) Next() (Result, error) {
	maxIdx := len(n.rows)-1
	hasMore := n.idx < maxIdx
	var val table.Row
	if n.idx <= len(n.rows) {
		val = n.rows[n.idx]
		n.idx++
	}

	return Result{val, hasMore}, nil
}

func collectAllRows(n Node) ([]table.Row, error) {
	var rows []table.Row
	for ;; {
		res, err := n.Next()
		if err != nil {
			return nil, err
		}
		rows = append(rows, res.Val)
		if !res.hasMore {
			break
		}
	}
	return rows, nil
}
