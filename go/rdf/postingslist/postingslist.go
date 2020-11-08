package postingslist

import (
	"bytes"

	"github.com/google/btree"
)

type PostingsList struct {
	idx *btree.BTree
}

func New() *PostingsList {
	return &PostingsList{
		idx: btree.New(2),
	}
}

func (p *PostingsList) Insert(key []byte, val interface{}) {
	p.insertItem(&Entry{
		Key:      key,
		Postings: []interface{}{val},
	})
}

func (p *PostingsList) Scan(prefix []byte) []*Entry {
	var entries []*Entry
	visit := func(i btree.Item) bool {
		entry, ok := i.(*Entry)
		if !ok {
			panic("Found non-index item in index. Check insertion code.")
		}

		if bytes.Compare(prefix, entry.Key[0:len(prefix)]) != 0 {
			return false
		}

		entries = append(entries, entry)

		return true
	}

	p.idx.AscendGreaterOrEqual(&Entry{Key: prefix}, visit)

	return entries
}

func (p *PostingsList) insertItem(item *Entry) {
	if existing := p.idx.Get(item); existing != nil {
		existingItem := existing.(*Entry)
		existingItem.Postings = append(existingItem.Postings, item.Postings[0])
	} else {
		p.idx.ReplaceOrInsert(item)
	}
}

type Entry struct {
	Key      []byte
	Postings []interface{}
}

func (item *Entry) Less(than btree.Item) bool {
	other := than.(*Entry)

	return bytes.Compare(item.Key, other.Key) == -1
}
