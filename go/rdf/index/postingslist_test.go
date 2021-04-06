package index

import (
	"testing"

	"github.com/kendru/darwin/go/rdf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestScan(t *testing.T) {
	i := NewPostingsList()
	api := NewIndexAPI()
	api.InsertInt64(i, []byte("aardvark"), 10)
	api.InsertInt64(i, []byte("apple"), 11)
	api.InsertInt64(i, []byte("apples"), 12)
	api.InsertInt64(i, []byte("apple"), 13)
	api.InsertInt64(i, []byte("banana"), 14)

	res, err := api.ScanInt64(i, ScanArgs{Prefix: []byte("apple")})
	assert.NoError(t, err, "scan should be success")

	assert.Equal(
		t,
		[]DecodedPostingsEntry{
			{[]byte("apple"), testutil.GenericInt64Slice(11, 13)},
			{[]byte("apples"), testutil.GenericInt64Slice(12)},
		},
		res,
	)
}
