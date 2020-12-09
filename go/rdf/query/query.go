package query

import "github.com/kendru/darwin/go/rdf/database"

// Fresh creates a fresh logic variable
func Fresh() *LVar {
	return &LVar{}
}

// LVar is an anonymous logic variable
type LVar struct {
	isBound bool
}

// NewRule creates a rule that acts as a query clause. Any of subject,
// predicate, and object may be values or LVars.
func NewRule(s interface{}, p interface{}, o interface{}) *Rule {
	return &Rule{
		Subject:   s,
		Predicate: p,
		Object:    o,
	}
}

type Rule struct {
	Subject, Predicate, Object interface{}
}

func NewPattern(lVars ...*LVar) *Pattern {
	return &Pattern{lVars}
}

type Pattern struct {
	lVars []*LVar
}

func NewQuery(p *Pattern, r ...*Rule) *Query {
	return Query{
		pattern: p,
		rules:   r,
	}
}

type Query struct {
	pattern *Pattern
	rules   []*Rule
}

func (q *Query) Execute(db *database.Database) (res QueryResult, err error) {
  db.
}

type QueryResult struct {
	Rows [][]interface{}
}
