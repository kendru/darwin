package index

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// This file exposes a high-level API that houses convenience methods for
// encoding and decoding index entries, etc. Really, I'm not sure what all
// it should do yet :)

func NewIndexAPI() *IndexAPI {
	return &IndexAPI{}
}

type IndexAPI struct {
}

type DecodedPostingsEntry struct {
	Key    []byte // TODO: Change to Tuple?
	Values []interface{}
}

func (api *IndexAPI) InsertInt64(i Inserter, key []byte, val int64) error {
	buf := bytes.NewBuffer(make([]byte, 0, 8))
	_ = binary.Write(buf, binary.BigEndian, val)

	return InsertOne(i, IndexEntry{
		Key:    key,
		Values: [][]byte{buf.Bytes()},
	})
}

func (api *IndexAPI) ScanInt64(s Scanner, args ScanArgs) (entries []DecodedPostingsEntry, err error) {
	err = ScanAll(s, args, func(entry *IndexEntry) (bool, error) {
		vals := make([]interface{}, len(entry.Values))
		for i, val := range entry.Values {
			var x int64
			if err := binary.Read(bytes.NewReader(val), binary.BigEndian, &x); err != nil {
				return false, fmt.Errorf("error decoding value as int64")
			}
			vals[i] = x
		}
		entries = append(entries, DecodedPostingsEntry{
			Key:    entry.Key,
			Values: vals,
		})

		return true, nil
	})

	return
}
