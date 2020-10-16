package table

import "github.com/google/btree"

type indexItem struct {
	key interface{}
	ids []string
}

func newIndexItem(key interface{}, ids []string) *indexItem {
	return &indexItem{key, ids}
}

func newIndexLookup(key interface{}) *indexItem {
	return newIndexItem(key, nil)
}

func (item *indexItem) Less(than btree.Item) bool {
	other := than.(*indexItem)
	switch k := item.key.(type) {
	case int, int8, int16, int32, int64:
		return k.(int64) < other.key.(int64)
	case float32, float64:
		return k.(float64) < other.key.(float64)
	case string:
		return k < other.key.(string)
	}
	return true
}
