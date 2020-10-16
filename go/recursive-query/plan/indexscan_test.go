package plan

import (
	"testing"

	"github.com/kendru/darwin/go/recursive-query/table"
	"github.com/stretchr/testify/assert"
)

func TestIndexScan(t *testing.T) {
	tbl := table.NewTable("test", table.WithPrimaryKey("id"), table.WithIndexes("prop"))
	tbl.Insert(table.Row{"id": "1", "prop": "foo"})
	tbl.Insert(table.Row{"id": "2", "prop": "bar"})
	tbl.Insert(table.Row{"id": "3", "prop": "baz"})
	n := &IndexScan{table: tbl, index: "prop"}

	rows, err := collectAllRows(n)
	assert.NoError(t, err)

	assert.Equal(t, []table.Row{
		{"id": "2", "prop": "bar"},
		{"id": "3", "prop": "baz"},
		{"id": "1", "prop": "foo"},
	}, rows)
}
