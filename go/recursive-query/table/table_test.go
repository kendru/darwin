package table

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTable(t *testing.T) {
	_ = NewTable("test")
}

func TestInsertAndGet(t *testing.T) {
	tbl := NewTable("test", WithPrimaryKey("id"))
	e1 := Row{"id": "e1"}
	e2 := Row{"id": "e2"}

	var err error
	err = tbl.Insert(e1)
	assert.NoError(t, err)
	err = tbl.Insert(e2)
	assert.NoError(t, err)

	assert.Equal(t, e1, tbl.Get("e1"))
	assert.Equal(t, e2, tbl.Get("e2"))
	assert.Nil(t, tbl.Get("e3"))
}

func TestInsertAndLookupByIndex(t *testing.T) {
	tbl := NewTable("test", WithPrimaryKey("id"), WithIndexes("name"))
	e1 := Row{"id": "1", "name": "Andrew"}
	e2 := Row{"id": "2", "name": "Diana"}

	var err error
	err = tbl.Insert(e1)
	assert.NoError(t, err)
	err = tbl.Insert(e2)
	assert.NoError(t, err)

	found, err := tbl.Lookup("name", "Andrew")
	assert.NoError(t, err)
	assert.Equal(t, e1, found)
}

func TestScan(t *testing.T) {
	tbl := NewTable("test", WithIndexes("name"))

	tbl.Insert(Row{"name": "Andrew"})
	tbl.Insert(Row{"name": "Diana"})
	tbl.Insert(Row{"name": "Audrey"})
	tbl.Insert(Row{"name": "Arwen"})
	tbl.Insert(Row{"name": "Jonah"})
	tbl.Insert(Row{"name": "Abel"})

	results, err := tbl.Scan("name", "A", "B")
	assert.NoError(t, err)

	foundNames := make([]string, len(results))
	for i, result := range results {
		foundNames[i] = result["name"].(string)
	}

	assert.Equal(t, []string{
		"Abel",
		"Andrew",
		"Arwen",
		"Audrey",
	}, foundNames)
}

func TestCannotInsertWithNonStringPkey(t *testing.T) {
	tbl := NewTable("test", WithPrimaryKey("id"))
	err := tbl.Insert(Row{"id": 1})
	assert.Error(t, err)
}

