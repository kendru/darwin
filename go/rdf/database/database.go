package database

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/kendru/darwin/go/rdf/tuple"

	"github.com/kendru/darwin/go/rdf/index"
)

const (
	DBIndexSPO DBIndex = iota
	DBIndexPSO
)

type DBIndex int

type Database interface {
	Observe(Fact) error
	Transact(tx TxData) (*TxResult, error)
	NewFact(s interface{}, p string, o interface{}) (Fact, error)
	CreateIdent(ident string) (uint64, error)
	Ident(ident string) (uint64, error)
	SPO() index.Index
	PSO() index.Index
	POS() index.Index
}

type defaultDatabase struct {
	idents map[string]uint64
	maxID  uint64

	mu  sync.Mutex
	spo index.Index
	pso index.Index
	pos index.Index
}

func NewDefaultDatabase() *defaultDatabase {
	return &defaultDatabase{
		idents: make(map[string]uint64),

		spo: index.NewPostingsList(),
		pso: index.NewPostingsList(),
		pos: index.NewPostingsList(), // TODO: only index for specific attributes
	}
}

// TODO: The database should be in control of assigning ids and should disallow the
// client from explicitly setting an ID. This method should probably be private, and
// facts should only be transaced from a higher level API.
func (db *defaultDatabase) NewFact(s interface{}, p string, o interface{}) (Fact, error) {
	var id uint64
	switch v := s.(type) {
	case uint64:
		id = v
	case int:
		id = uint64(v)
	case string:
		var ok bool
		id, ok = db.idents[v]
		if !ok {
			return Fact{}, fmt.Errorf("Unknown ident: %q", v)
		}
	}

	return Fact{
		Subject:   id,
		Predicate: p,
		Object:    o,
	}, nil
}

func MustNewFact(db Database, s interface{}, p string, o interface{}) Fact {
	fact, err := db.NewFact(s, p, o)
	if err != nil {
		panic(fmt.Errorf("Error creating new fact: %w", err))
	}

	return fact
}

func (d *defaultDatabase) nextIDUnsafe() uint64 {
	d.maxID++
	return d.maxID
}

func (d *defaultDatabase) CreateIdent(ident string) (uint64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if id, ok := d.idents[ident]; ok {
		return id, nil
	}

	id := d.nextIDUnsafe()
	d.idents[ident] = id

	return id, nil
}

func (d *defaultDatabase) Ident(ident string) (uint64, error) {
	if id, ok := d.idents[ident]; ok {
		return id, nil
	}
	return 0, fmt.Errorf("Unknown ident: %q", ident)
}

func MustIdent(db Database, ident string) uint64 {
	id, err := db.Ident(ident)
	if err != nil {
		panic(fmt.Errorf("Error looking up ident: %w", err))
	}

	return id
}

func (d *defaultDatabase) Observe(f Fact) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.observeUnsafe(f)
	return nil
}

func (d *defaultDatabase) observeUnsafe(f Fact) {
	// TODO: use custom encoder rather than storing values as a 1-tuple
	o := tuple.New(f.Object).Serialize()

	index.InsertOne(d.spo, index.IndexEntry{
		Key:    f.subjectPredicateKey(),
		Values: [][]byte{o},
	})
	index.InsertOne(d.pso, index.IndexEntry{
		Key:    f.predicateSubjectKey(),
		Values: [][]byte{o},
	})
	index.InsertOne(d.pos, index.IndexEntry{
		Key:    f.predicateObjectKey(),
		Values: [][]byte{tuple.New(f.Subject).Serialize()},
	})
}

type TxData struct {
	Updates []map[string]interface{}
}

func (tx *TxData) getTempIDs() []*TempID {
	tempIDs := make(map[*TempID]struct{})
	for _, update := range tx.Updates {
		for predicate, object := range update {
			id, ok := object.(*TempID)
			if !ok {
				continue
			}

			if predicate == "db:id" {
				id.isAssigned = true
			}

			tempIDs[id] = struct{}{}
		}
	}

	out := make([]*TempID, len(tempIDs))
	var i int
	for tempID := range tempIDs {
		out[i] = tempID
		i++
	}

	return out
}

type TxResult struct {
}

func (d *defaultDatabase) Transact(tx TxData) (*TxResult, error) {
	tempIDs := tx.getTempIDs()

	d.mu.Lock()
	defer d.mu.Unlock()

	// Ensure all tempIDs are assigned within transaction, and allocate ids.
	for _, tempID := range tempIDs {
		if !tempID.isAssigned {
			return nil, errors.New("Unassigned temporary id")
		}
		tempID.ID = d.nextIDUnsafe()
	}

	facts, err := d.getFactsForUpdate(&tx)
	if err != nil {
		return nil, fmt.Errorf("error getting facts for update: %w", err)
	}

	for _, fact := range facts {
		d.observeUnsafe(fact)
	}

	return &TxResult{}, nil
}

func (d *defaultDatabase) getFactsForUpdate(tx *TxData) ([]Fact, error) {
	facts := make([]Fact, 0, len(tx.Updates))

	for _, update := range tx.Updates {
		entityFacts := make([]Fact, 0, len(update))
		var entityID uint64
		for predicate, object := range update {
			if predicate == "db:id" {
				switch v := object.(type) {
				case uint64:
					entityID = v
				case *TempID:
					entityID = v.ID
					entityFacts = append(entityFacts, Fact{
						Subject:   v.ID,
						Predicate: predicate,
						Object:    v.ID,
					})
					continue
				default:
					return nil, errors.New(`"db:id" must be an id or a tempid if specified.`)
				}
			}

			appendAllValues(&entityFacts, predicate, object)
		}

		if entityID == 0 {
			// An anonymous entity is inserted, and no TempID was requested.
			entityID = d.nextIDUnsafe()
		}

		for _, fact := range entityFacts {
			fact.Subject = entityID
			facts = append(facts, fact)
		}
	}

	return facts, nil
}

func appendAllValues(facts *[]Fact, predicate string, object interface{}) {
	if v, ok := object.(*TempID); ok {
		*facts = append(*facts, Fact{
			Subject:   v.ID,
			Predicate: predicate,
			Object:    v.ID,
		})

		return
	}

	switch reflect.TypeOf(object).Kind() {
	case reflect.Slice, reflect.Array:
		s := reflect.ValueOf(object)
		for i := 0; i < s.Len(); i++ {
			*facts = append(*facts, Fact{
				Predicate: predicate,
				Object:    s.Index(i).Interface(),
			})
		}
	default:
		*facts = append(*facts, Fact{
			Predicate: predicate,
			Object:    object,
		})
	}
}

func GetFacts(db Database, s uint64) ([]Fact, error) {
	var facts []Fact

	scanPrefix := tuple.New(s).Serialize()
	if err := index.ScanAll(db.SPO(), index.ScanArgs{
		Prefix: scanPrefix,
	}, func(entry *index.IndexEntry) (bool, error) {
		key, err := tuple.Deserialize(entry.Key)
		if err != nil {
			return false, fmt.Errorf("error deserializing key as tuple: %w", err)
		}

		pred, err := key.GetString(1)
		if err != nil {
			return false, fmt.Errorf("error decoding predicate as string: %w", err)
		}

		for _, val := range entry.Values {
			valTuple, err := tuple.Deserialize(val)
			if err != nil {
				return false, fmt.Errorf("decoding value for %q: %w", pred, err)
			}

			facts = append(facts, Fact{
				Subject:   s,
				Predicate: pred,
				Object:    tuple.MustGetUntyped(valTuple, 0),
			})
		}

		return true, nil
	}); err != nil {
		return nil, err
	}

	return facts, nil
}

func (d *defaultDatabase) SPO() index.Index {
	return d.spo
}

func (d *defaultDatabase) PSO() index.Index {
	return d.pso
}

func (d *defaultDatabase) POS() index.Index {
	return d.pos
}

func GetEntity(d Database, s uint64) (map[string]interface{}, error) {
	facts, err := GetFacts(d, s)
	if err != nil {
		return nil, err
	}

	out := make(map[string]interface{})
	for _, fact := range facts {
		if val, exists := out[fact.Predicate]; exists {
			if reflect.TypeOf(val).Kind() == reflect.Slice {
				s := reflect.ValueOf(val).Interface().([]interface{})
				out[fact.Predicate] = append(s, fact.Object)
			} else {
				out[fact.Predicate] = []interface{}{val, fact.Object}
			}
			continue
		}

		out[fact.Predicate] = fact.Object
	}

	return out, nil
}
