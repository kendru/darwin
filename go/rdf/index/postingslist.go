package index

import (
	"bytes"

	"github.com/google/btree"
)

// PostingsList is an inverted (multi-valued) index.
type PostingsList struct {
	idx *btree.BTree
}

func NewPostingsList() *PostingsList {
	return &PostingsList{
		idx: btree.New(2),
	}
}

func (p *PostingsList) DoInsert(entries []IndexEntry) error {
	for _, entry := range entries {
		p.insertItem(&entry)
	}

	return nil
}

func (p *PostingsList) DoScan(args ScanArgs) ScanIter {
	return NewPostingsListIter(p, args)
}

func NewPostingsListIter(p *PostingsList, args ScanArgs) *postingsListIter {
	return &postingsListIter{
		postingsList: p,
		args:         args,
	}
}

type postingsListIter struct {
	postingsList *PostingsList
	args         ScanArgs
	// XXX: Since we are converting from an internal iterator to an external iterator,
	// we either need to buffer elements or block within the callback. We choose the
	// former, although we do not want to use this implementation in production.
	entries []*IndexEntry
	fetched bool
}

func (i *postingsListIter) Error() error {
	return nil
}

func (i *postingsListIter) Next() bool {
	i.ensureFetched()
	return len(i.entries) > 0
}

func (i *postingsListIter) ensureFetched() {
	if i.fetched {
		return
	}

	prefix := i.args.Prefix
	i.postingsList.idx.AscendGreaterOrEqual(&IndexEntry{Key: prefix}, func(item btree.Item) bool {
		entry, ok := item.(*IndexEntry)
		if !ok {
			panic("Found non-index item in index. Check insertion code.")
		}

		if bytes.Compare(prefix, entry.Key[0:len(prefix)]) != 0 {
			return false
		}

		i.entries = append(i.entries, entry)

		return true
	})

	i.fetched = true
}

func (i *postingsListIter) Item() *IndexEntry {
	head, tail := i.entries[0], i.entries[1:]
	i.entries = tail
	return head
}

func (p *PostingsList) insertItem(item *IndexEntry) {
	if existing := p.idx.Get(item); existing != nil {
		existingItem := existing.(*IndexEntry)
		existingItem.Values = append(existingItem.Values, item.Values...)
	} else {
		p.idx.ReplaceOrInsert(item)
	}
}

func (item *IndexEntry) Less(than btree.Item) bool {
	other := than.(*IndexEntry)

	return bytes.Compare(item.Key, other.Key) == -1
}
