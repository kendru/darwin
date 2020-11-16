package main

import (
	"encoding/json"
	"fmt"

	"github.com/kendru/darwin/go/rdf/database"
)

func main() {
	db := database.New()

	db.CreateIdent("db:id")
	db.CreateIdent("db:schema/type")
	db.CreateIdent("db:type/string")
	db.CreateIdent("db:type/int")
	db.CreateIdent("db:type/ref")

	db.CreateIdent("db:schema/cardinality")
	db.CreateIdent("db:cardinality/one")
	db.CreateIdent("db:cardinality/many")

	nameID := db.CreateIdent("person:name")
	fmt.Printf("ID for person:name: %d\n", nameID)
	db.Observe(db.MustNewFact("person:name", "db:schema/type", db.MustIdent("db:type/string")))
	db.Observe(db.MustNewFact("person:name", "db:schema/cardinality", db.MustIdent("db:cardinality/one")))

	db.CreateIdent("person:likes")
	db.Observe(db.MustNewFact("person:likes", "db:schema/type", db.MustIdent("db:type/string")))
	db.Observe(db.MustNewFact("person:likes", "db:schema/cardinality", db.MustIdent("db:cardinality/many")))

	db.CreateIdent("person:spouse")
	db.Observe(db.MustNewFact("person:spouse", "db:schema/type", db.MustIdent("db:type/ref")))
	db.Observe(db.MustNewFact("person:spouse", "db:schema/cardinality", db.MustIdent("db:cardinality/one")))

	db.Observe(db.MustNewFact(100, "person:name", "Andrew"))
	db.Observe(db.MustNewFact(100, "person:likes", "tacos"))
	db.Observe(db.MustNewFact(100, "person:likes", "nachos"))
	db.Observe(db.MustNewFact(100, "person:spouse", 200))

	db.Observe(db.MustNewFact(200, "person:name", "Diana"))
	db.Observe(db.MustNewFact(200, "person:likes", "popcorn"))
	db.Observe(db.MustNewFact(200, "person:likes", "nachos"))
	db.Observe(db.MustNewFact(200, "person.spouse", 100))

	andrew := database.Fresh()
	diana := database.Fresh()
	_, err := db.Transact(database.TxData{
		Updates: []map[string]interface{}{
			{
				"db:id":         andrew,
				"person:name":   "Andrew",
				"person:likes":  []string{"tacos", "nachos"},
				"person:spouse": diana,
			},
			{
				"db:id":         diana,
				"person:name":   "Diana",
				"person:likes":  []string{"nachos", "popcorn"},
				"person:spouse": andrew,
			},
		},
	})
	panicOnError(err)

	facts, err := db.GetFacts(12)
	panicOnError(err)

	for _, fact := range facts {
		fmt.Printf("%s\n", fact)
	}

	entity, err := db.GetEntity(12)
	panicOnError(err)

	fmt.Printf("Likes: %s\n", entity["person:likes"])

	data, err := json.MarshalIndent(entity, "", "  ")
	panicOnError(err)

	fmt.Println(string(data))
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
