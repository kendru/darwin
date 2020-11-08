package postingslist

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScan(t *testing.T) {
	i := New()
	i.Insert([]byte("aardvark"), 10)
	i.Insert([]byte("apple"), 11)
	i.Insert([]byte("apples"), 12)
	i.Insert([]byte("apple"), 13)
	i.Insert([]byte("banana"), 14)

	assert.Equal(
		t,
		[]*Entry{
			{[]byte("apple"), []interface{}{11, 13}},
			{[]byte("apples"), []interface{}{12}},
		},
		i.Scan([]byte("apple")),
	)
}
