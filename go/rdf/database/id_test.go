package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFreshIDsNotEqual(t *testing.T) {
	id1 := Fresh()
	id2 := Fresh()
	assert.NotEqual(t, id1, id2)
}
