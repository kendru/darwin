package lookup

import (
	"testing"

	"github.com/kendru/darwin/go/recursive-query/db"
	"github.com/kendru/darwin/go/recursive-query/table"
	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	db := db.NewSimpleDatabase()
	tbl := table.NewTable("people", table.WithPrimaryKey("id"))
	db.RegisterTable(tbl)

	tbl.Insert(table.Row{"id": "1", "name": "Andrew", "age": 30, "spouse": "2"})
	tbl.Insert(table.Row{"id": "2", "name": "Diana", "age": 33, "spouse": "1"})
	tbl.Insert(table.Row{"id": "3", "name": "Audrey", "age": 8})
	tbl.Insert(table.Row{"id": "4", "name": "Arwen", "age": 8})
	tbl.Insert(table.Row{"id": "5", "name": "Jonah", "age": 6})
	tbl.Insert(table.Row{"id": "6", "name": "Abel", "age": 5})

	q := Query{
		Table: "people",
		ID: "1",
		Root: EntityNode{
			Children: []QueryNode{
				PropertyNode{Property: "name"},
				PropertyNode{Property: "age"},
				ReferenceNode{
					ForeignKey: "spouse",
					ForeignTable: "people",
					ChildNode: EntityNode{
						Children: []QueryNode{
							PropertyNode{Property: "name"},
							PropertyNode{Property: "age"},
						},
					},
				},
			},
		},
	}
	assert.Equal(t, "people.1 [.name .age {people.spouse: [.name .age]}]", q.String())

	e := NewExecutor(db)
	res, err := e.Execute(q)
	assert.NoError(t, err)

	assert.Equal(t, map[string]interface{}{
		"name": "Andrew",
		"age": 30,
		"spouse": map[string]interface{}{
			"name": "Diana",
			"age": 33,
		},
	}, res)
}
