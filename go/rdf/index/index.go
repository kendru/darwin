package index

import "fmt"

// How should an index expose its properties - e.g. unique vs many?
// Should these be extrinsic properties that the database manages?

type Index interface {
	Inserter
	Scanner
}

type Inserter interface {
	// TODO: Add transaction.
	DoInsert([]IndexEntry) error
}

type IndexEntry struct {
	Key    []byte
	Values [][]byte
}

func Insert(i Inserter, entries []IndexEntry) error {
	return i.DoInsert(entries)
}

func InsertOne(i Inserter, entry IndexEntry) error {
	return Insert(i, []IndexEntry{entry})
}

type Scanner interface {
	// TODO: Add transaction
	DoScan(ScanArgs) ScanIter
	// TODO: Add key-only scan?
}

type ScanArgs struct {
	Prefix []byte
	// TODO: direction, limit, filter, etc.
}

type ScanIter interface {
	Error() error
	Next() bool
	Item() *IndexEntry
  // TODO: Reverse()
}

func Scan(s Scanner, args ScanArgs) ScanIter {
	return s.DoScan(args)
}

type scanVisit func(entry *IndexEntry) (shouldContinue bool, err error)

func ScanAll(s Scanner, args ScanArgs, visit scanVisit) error {
	iter := Scan(s, args)
	if err := iter.Error(); err != nil {
		return fmt.Errorf("Error initializing scanner: %w", err)
	}

	for iter.Next() {
		shouldContinue, err := visit(iter.Item())
		if err != nil {
			return err
		}
		if !shouldContinue {
			break
		}
	}

	if err := iter.Error(); err != nil {
		return fmt.Errorf("Error scanning: %w", err)
	}

	return nil
}
