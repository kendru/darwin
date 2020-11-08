package database

import (
	"fmt"
	"sync"

	"github.com/kendru/darwin/go/rdf/tuple"

	"github.com/kendru/darwin/go/rdf/postingslist"
)

type Database struct {
	idents map[string]uint64

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

func (d *Database) Observe(f Fact) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.pso.Insert(f.predicateSubjectKey(), f.Object)
	d.spo.Insert(f.subjectPredicateKey(), f.Object)
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

	// FIXME: this does not account for multi-values properties.
	out := make(map[string]interface{})
	for _, fact := range facts {
		out[fact.Predicate] = fact.Object
	}

	return out, nil
}
