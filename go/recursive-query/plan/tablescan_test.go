package plan

import (
	"testing"

	"github.com/kendru/darwin/go/recursive-query/table"
	"github.com/stretchr/testify/assert"
)

func TestTableScan(t *testing.T) {
	tbl := table.NewTable("test", table.WithPrimaryKey("id"))
	tbl.Insert(table.Row{"id": "1", "prop": "foo"})
	tbl.Insert(table.Row{"id": "2", "prop": "bar"})
	tbl.Insert(table.Row{"id": "3", "prop": "baz"})
	n := &TableScan{Table: tbl}

	rows, err := collectAllRows(n)
	assert.NoError(t, err)

	assert.Contains(t, rows, table.Row{"id": "1", "prop": "foo"})
	assert.Contains(t, rows, table.Row{"id": "2", "prop": "bar"})
	assert.Contains(t, rows, table.Row{"id": "3", "prop": "baz"})
	assert.Len(t, rows, 3)
}
