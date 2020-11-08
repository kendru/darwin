package database

import (
	"bytes"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyOrdering(t *testing.T) {
	f1 := Fact{
		Subject:   100,
		Predicate: "a",
	}
	f2 := Fact{
		Subject:   100,
		Predicate: "b",
	}
	assert.True(t, bytes.Compare(f1.subjectPredicateKey(), f2.subjectPredicateKey()) == -1, "Expected: (1, a) < (1, b)")
	assert.True(t, bytes.Compare(f1.predicateSubjectKey(), f2.predicateSubjectKey()) == -1, "Expected: (a, 1) < (b, 1)")

	f3 := Fact{
		Subject:   0,
		Predicate: "aa",
	}
	f4 := Fact{
		Subject:   math.MaxUint64,
		Predicate: "a",
	}
	assert.True(t, bytes.Compare(f3.predicateSubjectKey(), f4.predicateSubjectKey()) == 1, "Expected: (aa, 0) > (a, MAX)")
}
