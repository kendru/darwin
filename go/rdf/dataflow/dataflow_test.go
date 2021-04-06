package dataflow

import (
	"testing"

	"github.com/kendru/darwin/go/rdf/tuple"
	"github.com/stretchr/testify/assert"
)

// TODO: Factor these tests into subtests so that we can declare data and
// schema at the top that is shared among all of the test cases.

func TestDataflow(t *testing.T) {
	fruitSchema := NewRowSchema(
		MakeElementDescriptor(tuple.TypeInt64, "id"),
		MakeElementDescriptor(tuple.TypeUnicode, "name"),
	)
	fruitData := []*tuple.Tuple{
		tuple.New(int64(1), "apple"),
		tuple.New(int64(2), "banana"),
		tuple.New(int64(3), "orange"),
		tuple.New(int64(4), "pear"),
	}

	personSchema := NewRowSchema(
		MakeElementDescriptor(tuple.TypeInt64, "id"),
		MakeElementDescriptor(tuple.TypeUnicode, "first_name"),
		MakeElementDescriptor(tuple.TypeUnicode, "last_name"),
		MakeElementDescriptor(tuple.TypeInt64, "age"),
		MakeElementDescriptor(tuple.TypeInt64, "favorite_fruit_id"),
	)
	peopleData := []*tuple.Tuple{
		tuple.New(int64(1), "linda", "anderson", int64(43), int64(4)),
		tuple.New(int64(2), "andriy", "steklov", int64(23), int64(2)),
		tuple.New(int64(3), "ava", "shier", int64(22), int64(3)),
		tuple.New(int64(4), "julian", "gibson", int64(31), int64(3)),
	}

	bookSchema := NewRowSchema(
		MakeElementDescriptor(tuple.TypeInt64, "id"),
		MakeElementDescriptor(tuple.TypeUnicode, "title"),
		MakeElementDescriptor(tuple.TypeInt64, "belongs_to_id"),
	)
	booksData := []*tuple.Tuple{
		tuple.New(int64(1), "All About Stuff", int64(1)),
		tuple.New(int64(2), "Think Big", int64(4)),
		tuple.New(int64(3), "Eat Smart", int64(4)),
	}

	t.Run("slice generator", func(t *testing.T) {
		n := NewSliceGenerator(fruitSchema, fruitData)
		out, err := Collect(n)

		assert.NoError(t, err)
		assert.Equal(t, []*tuple.Tuple{
			tuple.New(int64(1), "apple"),
			tuple.New(int64(2), "banana"),
			tuple.New(int64(3), "orange"),
			tuple.New(int64(4), "pear"),
		}, out)
	})

	t.Run("simple pipeline", func(t *testing.T) {
		var n Node
		n = NewSliceGenerator(fruitSchema, fruitData)
		n = NewFilterNode(n, func(t *tuple.Tuple, s *RowSchema) (bool, error) {
			id, err := t.GetInt64(0)
			if err != nil {
				return false, err
			}

			return id%2 == 0, nil
		})
		n = NewLimitNode(n, 1)

		out, err := Collect(n)
		assert.NoError(t, err)
		assert.Equal(t, []*tuple.Tuple{
			tuple.New(int64(2), "banana"),
		}, out)
	})

	t.Run("inner join - unique key", func(t *testing.T) {
		fruits := NewSliceGenerator(fruitSchema, fruitData)
		people := NewSliceGenerator(personSchema, peopleData)

		// people.favorite_fruit_id = fruits.id
		n := NewInnerJoinNode(people, 4, fruits, 0)

		out, err := Collect(n)
		assert.NoError(t, err)
		assert.Equal(t, []*tuple.Tuple{
			tuple.New(int64(1), "linda", "anderson", int64(43), int64(4), int64(4), "pear"),
			tuple.New(int64(2), "andriy", "steklov", int64(23), int64(2), int64(2), "banana"),
			tuple.New(int64(3), "ava", "shier", int64(22), int64(3), int64(3), "orange"),
			tuple.New(int64(4), "julian", "gibson", int64(31), int64(3), int64(3), "orange"),
		}, out)
	})

	t.Run("inner join - non-unique key", func(t *testing.T) {
		books := NewSliceGenerator(bookSchema, booksData)
		people := NewSliceGenerator(personSchema, peopleData)
		// books.belongs_to_id = people.id
		n := NewInnerJoinNode(books, 2, people, 0)

		out, err := Collect(n)
		assert.NoError(t, err)
		assert.Equal(t, []*tuple.Tuple{
			tuple.New(int64(1), "All About Stuff", int64(1), int64(1), "linda", "anderson", int64(43), int64(4)),
			tuple.New(int64(2), "Think Big", int64(4), int64(4), "julian", "gibson", int64(31), int64(3)),
			tuple.New(int64(3), "Eat Smart", int64(4), int64(4), "julian", "gibson", int64(31), int64(3)),
		}, out)
	})

	t.Run("project/rename", func(t *testing.T) {
		people := NewSliceGenerator(personSchema, peopleData)
		n := NewProjectRenameNode(people,
			MakeProjection("first_name", "name"),
			MakeProjection("age", "age"),
		)

		expectedSchema := NewRowSchema(
			MakeElementDescriptor(tuple.TypeUnicode, "name"),
			MakeElementDescriptor(tuple.TypeInt64, "age"),
		)
		assert.Equal(t, expectedSchema, n.Schema())

		out, err := Collect(n)
		assert.NoError(t, err)
		assert.Equal(t, []*tuple.Tuple{
			tuple.New("linda", int64(43)),
			tuple.New("andriy", int64(23)),
			tuple.New("ava", int64(22)),
			tuple.New("julian", int64(31)),
		}, out)
	})

	t.Run("into document", func(t *testing.T) {
		people := NewSliceGenerator(personSchema, peopleData)
		n := NewIntoDocumentNode(people)

		out, err := Collect(n)
		assert.NoError(t, err)
		assert.Equal(t, []*tuple.Tuple{
			tuple.New(map[string]interface{}{
				"id":                int64(1),
				"first_name":        "linda",
				"last_name":         "anderson",
				"age":               int64(43),
				"favorite_fruit_id": int64(4),
			}),
			tuple.New(map[string]interface{}{
				"id":                int64(2),
				"first_name":        "andriy",
				"last_name":         "steklov",
				"age":               int64(23),
				"favorite_fruit_id": int64(2),
			}),
			tuple.New(map[string]interface{}{
				"id":                int64(3),
				"first_name":        "ava",
				"last_name":         "shier",
				"age":               int64(22),
				"favorite_fruit_id": int64(3),
			}),
			tuple.New(map[string]interface{}{
				"id":                int64(4),
				"first_name":        "julian",
				"last_name":         "gibson",
				"age":               int64(31),
				"favorite_fruit_id": int64(3),
			}),
		}, out)
	})

	t.Run("integration - full query", func(t *testing.T) {
		// SELECT
		//   person.first_name AS fname,
		//   person.last_name AS lname,
		//   book.title AS book,
		//   fruit.name AS favorite_fruit
		// FROM book
		// INNER JOIN person on book.owner_id = person.id
		// INNER JOIN fruit on person.favorite_fruit_id = fruit.id
		// WHERE person.age < 40
		books := NewSliceGenerator(bookSchema, booksData)
		people := NewSliceGenerator(personSchema, peopleData)
		fruit := NewSliceGenerator(fruitSchema, fruitData)

		var n Node
		n = NewInnerJoinNode(books, 2, people, 0)
		n = NewInnerJoinNode(n, 7, fruit, 0)
		n = NewFilterNode(n, func(t *tuple.Tuple, schema *RowSchema) (bool, error) {
			var ageIdx int
			for i, elem := range schema.elements {
				if elem.alias == "age" {
					ageIdx = i
					break
				}
			}
			age, err := t.GetInt64(ageIdx)
			if err != nil {
				return false, err
			}

			return age < 40, nil
		})
		n = NewProjectRenameNode(n,
			MakeProjection("first_name", "fname"),
			MakeProjection("last_name", "lname"),
			MakeProjection("title", "book"),
			MakeProjection("name", "favorite_fruit"),
		)
		n = NewIntoDocumentNode(n)

		out, err := Collect(n)
		assert.NoError(t, err)
		assert.Equal(t, []*tuple.Tuple{
			tuple.New(map[string]interface{}{
				"fname":          "julian",
				"lname":          "gibson",
				"book":           "Think Big",
				"favorite_fruit": "orange",
			}),
			tuple.New(map[string]interface{}{
				"fname":          "julian",
				"lname":          "gibson",
				"book":           "Eat Smart",
				"favorite_fruit": "orange",
			}),
		}, out)
	})
}
