package pkg

import (
	"fmt"
	"sync"
)

func NewDB() *DB {
	return &DB{
		lVars:  make(map[int]*LVar),
		assocs: make(map[int]interface{}),
	}
}

type DB struct {
	mu     sync.Mutex
	nextID int
	lVars  map[int]*LVar // Is this used?
	assocs map[int]interface{}
}

func (db *DB) Dump() {
	for key, val := range db.assocs {
		var s string
		if lv, ok := val.(*LVar); ok {
			s = fmt.Sprintf("@%d", lv.id)
		} else {
			s = fmt.Sprintf("%v", val)
		}
		fmt.Printf("[%d]: %s (%T)\n", key, s, val)
	}
}

func (db *DB) Fresh() *LVar {
	db.mu.Lock()
	defer db.mu.Unlock()

	id := db.nextID
	lvar := &LVar{id: id}
	db.lVars[id] = lvar
	db.nextID++

	return lvar
}

func (db *DB) Extend(term *LVar, val interface{}) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.assocs[term.id] = val
}

func (db *DB) Walk(v interface{}) interface{} {
	var term *LVar
	var ok bool
	if term, ok = v.(*LVar); !ok {
		return v
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	var result interface{} = v
	for {
		next, ok := db.assocs[term.id]
		if !ok {
			break
		}
		result = next

		if term, ok = next.(*LVar); !ok {
			break
		}
	}

	return result
}
