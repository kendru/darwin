package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnify(t *testing.T) {
	db := NewDB()

	assert.NoError(t, Unify(1, 1).Run(db))
	assert.Error(t, Unify(1, 2).Run(db))

	a := db.Fresh()
	b := db.Fresh()
	assert.NoError(t, RunAll(db,
		Unify(a, b),
		Unify(b, 12),
	))

	assert.NoError(t, Unify(a, 12).Run(db))
	assert.NoError(t, Unify(12, a).Run(db))
	assert.Error(t, Unify(14, a).Run(db))
}
