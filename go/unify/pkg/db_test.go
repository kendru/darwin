package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtendWalk(t *testing.T) {
	db := NewDB()

	a := db.Fresh()
	b := db.Fresh()

	assert.Equal(t, a, db.Walk(a))
	assert.Equal(t, "a value", db.Walk("a value"))

	db.Extend(a, "foo")
	assert.Equal(t, "foo", db.Walk(a))
	assert.Equal(t, b, db.Walk(b))

	db.Extend(b, a)
	assert.Equal(t, "foo", db.Walk(b))
	assert.Equal(t, "foo", db.Walk(a))

	c := db.Fresh()
	d := db.Fresh()
	db.Extend(c, d)
	assert.Equal(t, d, db.Walk(c))
}
