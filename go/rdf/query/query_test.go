package query

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kendru/darwin/go/rdf/database"
	"github.com/kendru/darwin/go/rdf/tuple"
)

func TestSimplePattern(t *testing.T) {
	db := database.NewDefaultDatabase()

	// TODO: Explicitly index attributes based on schema. Right now, all values are indexed.

	db.Observe(database.MustNewFact(db, 1, "name", "Fred"))
	db.Observe(database.MustNewFact(db, 1, "show", "The Flinstones"))
	db.Observe(database.MustNewFact(db, 1, "gender", "male"))

	db.Observe(database.MustNewFact(db, 2, "name", "Wilma"))
	db.Observe(database.MustNewFact(db, 2, "show", "The Flinstones"))
	db.Observe(database.MustNewFact(db, 2, "gender", "female"))

	db.Observe(database.MustNewFact(db, 3, "name", "Fred"))
	db.Observe(database.MustNewFact(db, 3, "show", "I Love Lucy"))
	db.Observe(database.MustNewFact(db, 3, "gender", "male"))

	db.Observe(database.MustNewFact(db, 4, "name", "Ethel"))
	db.Observe(database.MustNewFact(db, 4, "show", "I Love Lucy"))
	db.Observe(database.MustNewFact(db, 4, "gender", "female"))

	t.Run("single (SP) -> (O) scan", func(t *testing.T) {
		vVar := Fresh()
		r := NewRule(1, "show", vVar)
		p := NewPattern(vVar)
		q := NewQuery(p, r)

		res, err := q.Execute(db)
		assert.NoError(t, err)

		assert.Equal(t, []*tuple.Tuple{
			tuple.New("The Flinstones"),
		}, res.Rows)
	})

	t.Run("single (S) -> (PO) scan", func(t *testing.T) {
		pVar := Fresh()
		vVar := Fresh()
		r := NewRule(1, pVar, vVar)
		p := NewPattern(pVar, vVar)
		q := NewQuery(p, r)

		res, err := q.Execute(db)
		assert.NoError(t, err)

		assert.Equal(t, []*tuple.Tuple{
			tuple.New("gender", "male"),
			tuple.New("name", "Fred"),
			tuple.New("show", "The Flinstones"),
		}, res.Rows)
	})

	t.Run("single (PO) -> (S) scan", func(t *testing.T) {
		sVar := Fresh()
		r := NewRule(sVar, "name", "Fred")
		p := NewPattern(sVar)
		q := NewQuery(p, r)

		res, err := q.Execute(db)
		assert.NoError(t, err)

		assert.Equal(t, []*tuple.Tuple{
			tuple.New(uint64(1)),
			tuple.New(uint64(3)),
		}, res.Rows)
	})

	t.Run("single (P) -> (SO) scan", func(t *testing.T) {
		sVar := Fresh()
		vVar := Fresh()
		r := NewRule(sVar, "name", vVar)
		p := NewPattern(sVar, vVar)
		q := NewQuery(p, r)

		res, err := q.Execute(db)
		assert.NoError(t, err)

		assert.Equal(t, []*tuple.Tuple{
			tuple.New(uint64(1), "Fred"),
			tuple.New(uint64(2), "Wilma"),
			tuple.New(uint64(3), "Fred"),
			tuple.New(uint64(4), "Ethel"),
		}, res.Rows)
	})

	t.Run("pattern that discards fresh variable", func(t *testing.T) {
		vVar := Fresh()
		r := NewRule(3, Fresh(), vVar)
		p := NewPattern(vVar)
		q := NewQuery(p, r)

		res, err := q.Execute(db)
		assert.NoError(t, err)

		assert.Equal(t, []*tuple.Tuple{
			tuple.New("male"),
			tuple.New("Fred"),
			tuple.New("I Love Lucy"),
		}, res.Rows)
	})

	t.Run("pattern that rearranges fresh variable", func(t *testing.T) {
		pVar := Fresh()
		vVar := Fresh()
		r := NewRule(4, pVar, vVar)
		p := NewPattern(vVar, pVar)
		q := NewQuery(p, r)

		res, err := q.Execute(db)
		assert.NoError(t, err)

		assert.Equal(t, []*tuple.Tuple{
			tuple.New("female", "gender"),
			tuple.New("Ethel", "name"),
			tuple.New("I Love Lucy", "show"),
		}, res.Rows)
	})

	t.Run("single (PS) -> (O) scan", func(t *testing.T) {
		sVar := Fresh()
		vVar := Fresh()
		r := NewRule(sVar, "name", vVar)
		p := NewPattern(sVar, vVar)
		q := NewQuery(p, r)

		res, err := q.Execute(db)
		assert.NoError(t, err)

		assert.Equal(t, []*tuple.Tuple{
			tuple.New(uint64(1), "Fred"),
			tuple.New(uint64(2), "Wilma"),
			tuple.New(uint64(3), "Fred"),
			tuple.New(uint64(4), "Ethel"),
		}, res.Rows)
	})

	t.Run("arbitrary joins", func(t *testing.T) {
		person1 := NamedVar("p1")
		person2 := NamedVar("p2")
		show := NamedVar("show")
		person2Name := NamedVar("p2.name")

		// select p2.name, p2.show
		// from person p1
		// inner join person p2 on p1.show = p2.show
		// where
		//   p1.name = 'Wilma'
		//   AND p2.gender = 'male';
		q := NewQuery(
			NewPattern(person2Name, show),

			// Possible optimization: if we know that some P is unique in a (?s P O) triple,
			// Fetch it in preprocessing, and replace all instances of ?s with the value found.
			NewRule(person1, "name", "Wilma"),
			NewRule(person1, "show", show),
			NewRule(person2, "show", show),
			NewRule(person2, "gender", "male"),
			NewRule(person2, "name", person2Name),
		)
    // SELECT ?person2Name, ?show
    // WHERE {
    //   ?person1 character:name "Wilma" .
    //   ?person1 character:show ?show .
    //   ?person2 character:show ?show .
    //   ?person2 character:gender "male" .
    //   ?person2 character:name ?person2Name .
    // }

		res, err := q.Execute(db)
		assert.NoError(t, err)

		assert.Equal(t, []*tuple.Tuple{
			tuple.New("Fred", "The Flinstones"),
		}, res.Rows)
	})
}

func TestRecursive(t *testing.T) {
  
}
