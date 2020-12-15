package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringHashKey(t *testing.T) {
	a1 := &String{Value: "hello world"}
	a2 := &String{Value: "hello world"}
	b1 := &String{Value: "jalapeno"}
	b2 := &String{Value: "jalapeno"}

	assert.Equal(t, a1.HashKey(), a2.HashKey())
	assert.Equal(t, b1.HashKey(), b2.HashKey())
	assert.NotEqual(t, a1.HashKey(), b1.HashKey())
}
