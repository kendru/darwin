package plan

import (
	"testing"

	"github.com/kendru/darwin/go/recursive-query/predicate"
	"github.com/kendru/darwin/go/recursive-query/table"
	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	src := newStubProducerNode([]table.Row{
		{"id": "1", "first_name": "Bill", "last_name": "Williams"},
		{"id": "2", "first_name": "Jack", "last_name": "Jackson"},
		{"id": "3", "first_name": "Bill", "last_name": "Jones"},
	})
	n := &Filter{Src: src, Expr: &PropEqualsString{
		Prop: "first_name",
		PartialEq: predicate.StringEquals{Lh: "Bill"},
	}}

	rows, err := collectAllRows(n)
	assert.NoError(t, err)

	assert.Equal(t, []table.Row{
		{"id": "1", "first_name": "Bill", "last_name": "Williams"},
		{"id": "3", "first_name": "Bill", "last_name": "Jones"},
	}, rows)
}

