package lookup

import (
	"testing"

	"github.com/kendru/darwin/go/recursive-query/db"
	"github.com/kendru/darwin/go/recursive-query/plan"
	"github.com/kendru/darwin/go/recursive-query/predicate"
	"github.com/kendru/darwin/go/recursive-query/table"
	"github.com/stretchr/testify/assert"
)

func TestEvaluate(t *testing.T) {
	db := db.NewSimpleDatabase()
	tbl := table.NewTable("test_tbl")
	db.RegisterTable(tbl)
	exec := NewExecutor(db)

	exec.Push(FetchEntity{"test_tbl", "some_id"})
	exec.Push(SelectField{"id"})
	exec.Push(SelectField{"title"})

	exec.Evaluate()

	assert.Equal(t, &plan.Project{
		Src: &plan.Filter{
			Src: &plan.TableScan{
				Table: tbl,
			},
			Expr: &plan.PropEqualsString{
				Prop: "id",
				PartialEq: predicate.StringEquals{
					Lh: "some_id",
				},
			},
		},
		Fields: []string{"id", "title"},
	}, exec.root)
}
