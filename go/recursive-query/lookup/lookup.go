package lookup

import (
	"errors"
	"fmt"

	"github.com/kendru/darwin/go/recursive-query/db"
	"github.com/kendru/darwin/go/recursive-query/plan"
	"github.com/kendru/darwin/go/recursive-query/predicate"
	"github.com/kendru/darwin/go/recursive-query/table"
)

type Executor struct {
	db db.Database
	root plan.Node
	ops []Operation
}

func NewExecutor(db db.Database) *Executor {
	return &Executor{db: db}
}

func (e *Executor) Push(op Operation) {
	e.ops = append(e.ops, op)
}

// Operations may pop more operations from the stack
func (e *Executor) Evaluate() error {
	for i := 0; i < len(e.ops); i++ {
		op := e.ops[i]
		switch typedOp := op.(type) {

		case SelectField:
			if e.root == nil {
				return errors.New("Cannot select a field when no entity has been fetched")
			}

			// Continue selecting properties as long as the next item on the stack is a SelectField.
			fields := []string{typedOp.Field}
			for j := 1; j < len(e.ops) - i; j++ {
				nextOp := e.ops[i+j]
				if nextSelect, ok := nextOp.(SelectField); ok {
					fields = append(fields, nextSelect.Field)
					// Consume this item from the stack.
					i++
				} else {
					break
				}
			}

			e.root = &plan.Project{Src: e.root, Fields: fields}

		case FetchEntity:
			tbl := e.db.GetTable(typedOp.Table)

			e.root = &plan.Filter{
				Src: &plan.TableScan{
					Table: tbl,
				},
				Expr: &plan.PropEqualsString{
					Prop: "id",
					PartialEq: predicate.StringEquals{
						Lh: typedOp.Id,
					},
				},
			}
		}
	}

	return nil
}

func (e *Executor) Run() (table.Row, error) {
	if e.root == nil {
		return nil, errors.New("Cannot run without a plan - try evaluating first")
	}

	res, err := e.root.Next()
	if err != nil {
		return nil, fmt.Errorf("error evaluating query: %w", err)
	}

	return res.Val, nil
}

type Operation interface {
	// Extend(db.Database, plan.Node) plan.Node
}

type FetchEntity struct {
	Table string
	Id string
}

func (f FetchEntity) String() string {
	return fmt.Sprintf("(fetch %s.%s)", f.Table, f.Id)
}

type SelectField struct {
	Field string
}

func (f SelectField) String() string {
	return fmt.Sprintf("(select $%s)", f.Field)
}
