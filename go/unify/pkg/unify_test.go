package pkg_test

import (
	"fmt"
	"testing"

	"github.com/kendru/darwin/go/unify/pkg"

	"github.com/stretchr/testify/assert"
)

func TestResolveUnique(t *testing.T) {
	idAttr := &Attribute{
		name:     "id",
		isID:     true,
		isUnique: true,
	}
	email := &Attribute{
		name:     "email",
		isUnique: true,
	}
	userID := &Attribute{
		name: "fkey to another user",
	}
	userRef := &Attribute{
		name:     "reference to another user",
		isRef:    true,
		refAttr:  email,
		fkeyAttr: userID,
	}
	name := &Attribute{
		name: "name",
	}
	age := &Attribute{
		name: "age",
	}
	s := NewSchema(idAttr, email, userRef, userID, name, age)

	ur := NewUniqueResolver()

	// Allocate an ID for user 1
	id1 := ur.Allocate([]Lookup{NewLookup(email, "user-1@example.com")})

	// We assume that the user will not provide an ID that has not been seen by the unique resolver.
	patches := []*patch{
		// User 1: second map specifies only unique attribute, not ID
		{
			avs: avs{
				idAttr: id1,
				email:  "user-1@example.com",
			},
		},
		{
			avs: avs{
				email: "user-1@example.com",
				name:  "user 1",
			},
		},
		// User 2: no id
		{
			avs: avs{
				email: "user-2@example.com",
				name:  "user 2",
			},
		},
		{
			avs: avs{
				email: "user-2@example.com",
				age:   "32",
			},
		},
		// User 3: references other user by unique attribute
		{
			avs: avs{
				email:   "user-3@example.com",
				userRef: "user-2@example.com",
			},
		},
	}

	resolveIDs(ur, s, patches)

	fmt.Println(patches)

	assert.Equal(t, id1, patches[0].avs[idAttr])
	assert.Equal(t, id1, patches[1].avs[idAttr])

	assert.Equal(t, patches[2].avs[idAttr], patches[3].avs[idAttr])

	// Reference augmented with foreign key. TODO: Instead specify the value of the foreign key attribute as a lookup,
	// and expand it to the referenced (and maybe newly-allocated) ID.
	assert.Equal(t, patches[4].avs[userID], patches[2].avs[idAttr])
}

type patch struct {
	avs    avs
	entity *pkg.LVar
}

func (p *patch) String() string {
	return fmt.Sprintf("%s", p.avs)
}

func resolveIDs(ur *UniqueResolver, schema *Schema, patches []*patch) {
	idAttr := schema.GetAttribute("id")

	db := pkg.NewDB()

	type workItem struct {
		entity *pkg.LVar
		lookup Lookup
	}
	var workItems []workItem

	for _, patch := range patches {
		patch.entity = db.Fresh()
		for attr, val := range patch.avs {
			if attr.isID {
				// Associate id with entity.
				if err := pkg.Unify(patch.entity, val).Run(db); err != nil {
					panic(fmt.Errorf("error unifying id %s: %w", val, err))
				}
			} else if attr.isUnique {
				// Lookup references same entity.
				workItems = append(workItems, workItem{patch.entity, NewLookup(attr, val)})
			} else if attr.isRef {
				// Lookup references some other entity.
				workItems = append(workItems, workItem{db.Fresh(), NewLookup(attr.refAttr, val)})
			}
		}
	}

	tempIDs := NewTempIDs(db)
	for _, item := range workItems {
		var stmt pkg.LogicStatement
		id := ur.Lookup(item.lookup)
		if id != "" {
			stmt = pkg.Unify(item.entity, id)
		} else {
			stmt = pkg.Unify(item.entity, tempIDs.GetID(item.lookup))
		}

		if err := stmt.Run(db); err != nil {
			panic(fmt.Errorf("Error unifying: %w", err))
		}
	}

	for k, i := range tempIDs.index {
		// Remove current element from remaining list.
		delete(tempIDs.index, k)

		lVar := tempIDs.lvars[i]
		// Create group for unified attributes.
		lookupsForEntity := []Lookup{tempIDs.lookups[i]}
		for otherK, otherI := range tempIDs.index {
			otherLvar := tempIDs.lvars[otherI]
			// If the cast fails, that means that we resolved an ID that the unique resolver didn't know about.
			// We should surface this as an error.
			thisCanonoical := db.Walk(lVar).(*pkg.LVar)
			otherCanonoical := db.Walk(otherLvar).(*pkg.LVar)
			if thisCanonoical.Eq(otherCanonoical) {
				delete(tempIDs.index, otherK)
				lookupsForEntity = append(lookupsForEntity, tempIDs.lookups[otherI])
			}
		}

		id := ur.Allocate(lookupsForEntity)
		if err := pkg.Unify(lVar, id).Run(db); err != nil {
			panic(fmt.Errorf("could not unify newly allocated id for entity: %w", err))
		}
	}

	for _, patch := range patches {
		id := db.Walk(patch.entity)
		patch.avs[idAttr] = id.(string)

		for attr, val := range patch.avs {
			// Replace references with foreign key
			if attr.isRef {
				// delete(patch.avs, attr)
				patch.avs[attr.fkeyAttr] = ur.Lookup(NewLookup(attr.fkeyAttr, val))
			}
		}
	}

	db.Dump()
}

func NewSchema(attrs ...*Attribute) *Schema {
	s := &Schema{attributes: make(map[string]*Attribute)}
	for _, attr := range attrs {
		s.attributes[attr.name] = attr
	}
	return s
}

type Schema struct {
	attributes map[string]*Attribute
}

func (s *Schema) GetAttribute(name string) *Attribute {
	return s.attributes[name]
}

type Attribute struct {
	name     string
	isID     bool
	isUnique bool
	isRef    bool
	refAttr  *Attribute
	fkeyAttr *Attribute
}

func (a *Attribute) String() string {
	return a.name
}

type avs map[*Attribute]string

func NewTempIDs(db *pkg.DB) *TempIDs {
	return &TempIDs{
		db:    db,
		index: make(map[string]int),
	}
}

// Is this struct necessary, or can unification solve everything?
type TempIDs struct {
	db      *pkg.DB
	index   map[string]int
	lvars   []*pkg.LVar
	lookups []Lookup
}

// GetID fetches the LVar assocaited with this lookup. If no existing lvar is found, we
// associate a new one.
func (t *TempIDs) GetID(l Lookup) *pkg.LVar {
	idx, ok := t.index[l.Hash()]
	if !ok {
		idx = len(t.lvars)
		t.index[l.Hash()] = idx
		t.lvars = append(t.lvars, t.db.Fresh())
		t.lookups = append(t.lookups, l)
	}
	return t.lvars[idx]
}

type UniqueResolver struct {
	ctr int
	// unique hash(attribute, value) -> id
	store map[string]string
}

func NewUniqueResolver() *UniqueResolver {
	return &UniqueResolver{
		store: make(map[string]string),
	}
}

func (ur *UniqueResolver) Allocate(lookups []Lookup) string {
	id := fmt.Sprintf("id:%d", ur.ctr)
	for _, lookup := range lookups {
		ur.store[lookup.Hash()] = id
	}
	ur.ctr++

	return id
}

func (ur *UniqueResolver) Lookup(lookup Lookup) string {
	if id, ok := ur.store[lookup.Hash()]; ok {
		return id
	}
	return ""
}

func NewLookup(attr *Attribute, value string) Lookup {
	return Lookup{
		attr:  attr,
		value: value,
	}
}

type Lookup struct {
	attr  *Attribute
	value string
}

func (l Lookup) Hash() string {
	return fmt.Sprintf("%s:%s", l.attr.name, l.value)
}

func (l Lookup) String() string {
	return fmt.Sprintf("%s = %q", l.attr.name, l.value)
}
