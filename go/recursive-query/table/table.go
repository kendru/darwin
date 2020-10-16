package table

import (
	"errors"
	"fmt"

	"github.com/google/btree"
	"github.com/google/uuid"
)

// TODO: Abstract data source from table

type Row map[string]interface{}

type Table struct {
	Name string
	PKey *string

	// Main storage is a map of string primary key to entities
	storage map[string]Row
	// Indexes are sorted lists of (val, pk) pairs
	indexes map[string]*btree.BTree
}

func NewTable(name string, opts... TableOpt) *Table {
	t := &Table{
		Name: name,
		storage: make(map[string]Row),
		indexes: make(map[string]*btree.BTree),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

func (t *Table) Insert(row Row) error {
	id, err := t.getPkey(row)
	if err != nil {
		return err
	}
	t.storage[id] = row
	t.insertIndexes(id, row)

	return nil
}

func (t *Table) Get(id string) Row {
	row, ok := t.storage[id]
	if !ok {
		return nil
	}

	return row
}

func (t *Table) Lookup(index string, key interface{}) (Row, error) {
	idx, ok := t.indexes[index]
	if !ok {
		return nil, fmt.Errorf("Index does not exist: %v", index)
	}

	found := idx.Get(newIndexLookup(key))
	if found == nil {
		return nil, nil
	}

	item, ok := found.(*indexItem)
	if !ok {
		panic("Found non-index item in index. Check insertion code.")
	}

	if len(item.ids) == 0 {
		// Index entry exists, but element removed and entry has not been cleaned up
		// TODO: Remove index entry when all ids have been removed.
		return nil, nil
	}

	return t.Get(item.ids[0]), nil
}

func (t *Table) Scan(index string, min, max interface{}) ([]Row, error) {
	idx, ok := t.indexes[index]
	if !ok {
		return nil, fmt.Errorf("Index does not exist: %v", index)
	}

	var ids []string
	visit := func(i btree.Item) bool {
		item, ok := i.(*indexItem)
		if !ok {
			panic("Found non-index item in index. Check insertion code.")
		}

		ids = append(ids, item.ids...)

		return true
	}

	if min == nil && max == nil {
		idx.Ascend(visit)
	} else if max == nil {
		idx.AscendGreaterOrEqual(newIndexLookup(min), visit)
	} else if min == nil {
		idx.AscendLessThan(newIndexLookup(max), visit)
	} else {
		idx.AscendRange(newIndexLookup(min), newIndexLookup(max), visit)
	}

	results := make([]Row, len(ids))
	for i, id := range ids {
		results[i] = t.Get(id)
	}

	return results, nil
}

func (t *Table) FetchAll() []Row {
	rows := make([]Row, len(t.storage))
	i := 0
	for _, row := range t.storage {
		rows[i] = row
		i++
	}

	return rows
}

func (t *Table) getPkey(row Row) (string, error) {
	if t.PKey == nil {
		if _, exists := row["_id"]; exists {
			return "", errors.New("Reserved _id field exists on row without primary key")
		}

		uuid, err := uuid.NewRandom()
		if err != nil {
			return "", fmt.Errorf("Could not generate new ID: %w", err)
		}

		id := uuid.String()
		row["_id"] = id
		return id, nil
	}

	id, ok := row[*t.PKey]
	if !ok {
		return "", fmt.Errorf("No primary key found on row. Expected %v", *t.PKey)
	}

	idStr, ok := id.(string)
	if !ok {
		return "", fmt.Errorf("Expected primary key to be string, but got %T instead", id)
	}

	return idStr, nil
}

func (t *Table) insertIndexes(id string, row Row) {
	for key, val := range row {
		if idx, ok := t.indexes[key]; ok {
			var item *indexItem
			if existing := idx.Get(newIndexLookup(val)); existing != nil {
				item = existing.(*indexItem)
				item.ids = append(item.ids, id)
			} else {
				item = newIndexItem(val, []string{id})
			}
			idx.ReplaceOrInsert(item)
		}
	}
}

type TableOpt func(*Table)

func WithIndexes(properties... string) TableOpt {
	return func(t *Table) {
		for _, prop := range properties {
			t.indexes[prop] = btree.New(2)
		}
	}
}

func WithPrimaryKey(property string) TableOpt {
	return func(t *Table) {
		t.PKey = &property
	}
}
