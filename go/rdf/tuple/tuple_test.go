package tuple

import (
	"bytes"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoundtrip(t *testing.T) {
	tup := New(uint64(123), "Hello, \x00world", uint64(456))
	buf := tup.Serialize()

	t.Log(buf)

	tupIn, err := Deserialize(buf)
	assert.NoError(t, err)
	assert.Equal(t, tup, tupIn)
}

func TestSerializeInt64(t *testing.T) {
	tup := New(int64(100))
	buf := tup.Serialize()
	assert.Len(t, buf, 9)
	tupIn, err := Deserialize(buf)
	assert.NoError(t, err)
	assert.Equal(t, tup, tupIn)
}

func TestComparableBytes(t *testing.T) {
	cases := []struct {
		lh                 *Tuple
		rh                 *Tuple
		expectedComparison int
		message            string
	}{
		{New(uint64(100)), New(uint64(200)), -1, "(100) < (200)"},
		{New("Apple"), New("Banana"), -1, `("Apple") < ("Banana")`},
		{New("a", uint64(math.MaxUint64)), New("aa", uint64(0)), -1, `("a", MAX) < ("aa", 0)`},
	}

	for i, testcase := range cases {
		assert.Equal(
			t,
			testcase.expectedComparison,
			bytes.Compare(testcase.lh.Serialize(), testcase.rh.Serialize()),
			fmt.Sprintf("Expected in case %d: %s", i, testcase.message),
		)
	}
}
