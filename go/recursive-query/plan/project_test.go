package plan

import (
	"testing"

	"github.com/kendru/darwin/go/recursive-query/table"
	"github.com/stretchr/testify/assert"
)

func TestProject(t *testing.T) {
	src := newStubProducerNode([]table.Row{
		{"id": "1", "first_name": "Bill", "last_name": "Williams", "age": "68"},
		{"id": "2", "first_name": "Jack", "last_name": "Jackson",  "age": "43"},
	})
	n := &Project{Src: src, Fields: []string{"first_name", "age"}}

	rows, err := collectAllRows(n)
	assert.NoError(t, err)

	assert.Equal(t, []table.Row{
		{"first_name": "Bill", "age": "68"},
		{"first_name": "Jack", "age": "43"},
	}, rows)
}

