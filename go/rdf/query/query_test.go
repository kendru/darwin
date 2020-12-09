package query

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kendru/darwin/go/rdf/database"
)

func TestFreshIsFresh(t *testing.T) {
	assert.False(t, Fresh().isBound)
}

func TestSimplePattern(t *testing.T) {
	v := Fresh()
	r := NewRule(v, "name", "Fred")
	p := NewPattern(v)
	q := NewQuery(p, r)

	db := database.New()
	db.Observe(database.MustNewFact(db, 1, "name", "Fred"))
	db.Observe(database.MustNewFact(db, 2, "name", "Wilma"))
	db.Observe(database.MustNewFact(db, 3, "name", "Fred"))
	db.Observe(database.MustNewFact(db, 4, "name", "Ethel"))

	res, err := q.Execute(db)
	assert.NoError(t, err)

	assert.Equal(t, []interface{}{1, 3}, res.Rows)
}
