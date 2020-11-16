package database

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/kendru/darwin/go/rdf/tuple"

	"github.com/kendru/darwin/go/rdf/postingslist"
)

type Database struct {
	idents map[string]uint64
	maxID  uint64

	mu  sync.Mutex
	spo *postingslist.PostingsList
	pso *postingslist.PostingsList
}

func New() *Database {
	return &Database{
		idents: make(map[string]uint64),

		spo: postingslist.New(),
		pso: postingslist.New(),
	}
}

// TODO: The database should be in control of assigning ids and should disallow the
// client from explicitly setting an ID. This method should probably be private, and
// facts should only be transaced from a higher level API.
func (db *Database) NewFact(s interface{}, p string, o interface{}) (Fact, error) {
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

func (db *Database) MustNewFact(s interface{}, p string, o interface{}) Fact {
	fact, err := db.NewFact(s, p, o)
	if err != nil {
		panic(fmt.Errorf("Error creating new fact: %w", err))
	}

	return fact
}

func (d *Database) CreateIdent(ident string) uint64 {
	d.mu.Lock()
	defer d.mu.Unlock()

	id := d.nextIDUnsafe()
	d.idents[ident] = id

	return id
}

func (d *Database) nextIDUnsafe() uint64 {
	d.maxID++
	return d.maxID
}

func (d *Database) Ident(ident string) (uint64, error) {
	if id, ok := d.idents[ident]; ok {
		return id, nil
	}
	return 0, fmt.Errorf("Unknown ident: %q", ident)
}

func (db *Database) MustIdent(ident string) uint64 {
	id, err := db.Ident(ident)
	if err != nil {
		panic(fmt.Errorf("Error looking up ident: %w", err))
	}

	return id
}

func (d *Database) Observe(f Fact) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.observeUnsafe(f)
}

func (d *Database) observeUnsafe(f Fact) {
	d.pso.Insert(f.predicateSubjectKey(), f.Object)
	d.spo.Insert(f.subjectPredicateKey(), f.Object)
}

type TxData struct {
	Updates []map[string]interface{}
}

type TxResult struct {
}

func (d *Database) Transact(tx TxData) (*TxResult, error) {
	// Collect fresh ids
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

	d.mu.Lock()
	defer d.mu.Unlock()

	// Ensure all tempIDs are assigned within transaction, and allocate ids.
	for tempID := range tempIDs {
		if !tempID.isAssigned {
			return nil, errors.New("Unassigned temporary id")
		}
		tempID.ID = d.nextIDUnsafe()
	}

	// Convert each update map into facts.
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

			if v, ok := object.(*TempID); ok {
				entityFacts = append(entityFacts, Fact{
					Subject:   v.ID,
					Predicate: predicate,
					Object:    v.ID,
				})
				continue
			}

			switch reflect.TypeOf(object).Kind() {
			case reflect.Slice, reflect.Array:
				s := reflect.ValueOf(object)
				for i := 0; i < s.Len(); i++ {
					entityFacts = append(entityFacts, Fact{
						Predicate: predicate,
						Object:    s.Index(i).Interface(),
					})
				}
			default:
				entityFacts = append(entityFacts, Fact{
					Predicate: predicate,
					Object:    object,
				})
			}
		}

		if entityID == 0 {
			// An anonymous entity is inserted, and no TempID was requested.
			entityID = d.nextIDUnsafe()
		}

		for _, fact := range entityFacts {
			fact.Subject = entityID
			// Observe fact.
			d.observeUnsafe(fact)
		}
	}

	return &TxResult{}, nil
}

func (d *Database) GetFacts(s uint64) ([]Fact, error) {
	var facts []Fact

	prefix := tuple.New(s).Serialize()
	for _, entry := range d.spo.Scan(prefix) {
		key, err := tuple.Deserialize(entry.Key)
		if err != nil {
			return nil, err
		}

		pred, err := key.String(1)
		if err != nil {
			return nil, err
		}

		for _, val := range entry.Postings {
			facts = append(facts, Fact{
				Subject:   s,
				Predicate: pred,
				Object:    val,
			})
		}
	}

	return facts, nil
}

func (d *Database) GetEntity(s uint64) (map[string]interface{}, error) {
	facts, err := d.GetFacts(s)
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
