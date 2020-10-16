package plan

import (
	"fmt"
	"strings"

	"github.com/kendru/darwin/go/recursive-query/table"
)

type Project struct {
	Src Node
	Fields []string
}

func (n *Project) Type() NodeType {
	return NodeProject
}

func (n *Project) Next() (Result, error) {
	next, err := n.Src.Next()
	if err != nil {
		return Result{}, fmt.Errorf("error getting next for projection: %w", err)
	}

	return Result{
		Val: n.project(next.Val),
		hasMore: next.hasMore,
	}, nil
}

func (n *Project) project(row table.Row) table.Row {
	projected := make(table.Row)
	for _, field := range n.Fields {
		if val, ok := row[field]; ok {
			projected[field] = val
		}
	}

	return projected
}

func (n *Project) String() string {
	return fmt.Sprintf("(project %s [%s])", n.Src, strings.Join(n.Fields, ", "))
}
