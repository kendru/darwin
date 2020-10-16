package plan

import (
	"fmt"

	"github.com/kendru/darwin/go/recursive-query/predicate"
	"github.com/kendru/darwin/go/recursive-query/table"
)

type FilterExpr interface {
	Pred(row table.Row) predicate.Predicate
}

type PropEqualsString struct {
	Prop string
	// This property is mutated for each tuple processed.
	PartialEq predicate.StringEquals
}

func (fe *PropEqualsString) Pred(row table.Row) predicate.Predicate {
	prop, ok := row[fe.Prop]
	if !ok {
		return predicate.ConstantFalse
	}

	rh, ok := prop.(string)
	if !ok {
		return predicate.ConstantFalse
	}

	fe.PartialEq.Rh = rh

	return fe.PartialEq
}

type And struct {
	FilterExprs []FilterExpr
}

func (fe *And) Pred(row table.Row) predicate.Predicate {
	preds := make([]predicate.Predicate, len(fe.FilterExprs))
	for i, filterExpr := range fe.FilterExprs {
		preds[i] = filterExpr.Pred(row)
	}
	return predicate.And{Predicates: preds}
}

type Filter struct {
	Src Node
	Expr FilterExpr
}

func (n *Filter) Type() NodeType {
	return NodeFilter
}

func (n *Filter) Next() (Result, error) {
	for ;; {
		next, err := n.Src.Next()
		if err != nil {
			return Result{}, fmt.Errorf("error getting next for filter: %w", err)
		}

		if n.Expr.Pred(next.Val).Test() {
			return next, nil
		}

		if !next.hasMore {
			return Result{hasMore: false}, nil
		}
	}
}

func (n *Filter) String() string {
	return fmt.Sprintf("(filter %s TODO)", n.Src)
}
