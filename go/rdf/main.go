package main

import (
	"encoding/json"
	"fmt"

	"github.com/kendru/darwin/go/rdf/database"
)

func main() {
	db := database.NewDefaultDatabase()

	db.CreateIdent("db:id")
	db.CreateIdent("db:schema/type")
	db.CreateIdent("db:type/string")
	db.CreateIdent("db:type/int")
	db.CreateIdent("db:type/ref")

	// db.RegisterCodec()

	db.CreateIdent("db:schema/cardinality")
	db.CreateIdent("db:cardinality/one")
	db.CreateIdent("db:cardinality/many")

	nameID, _ := db.CreateIdent("person:name")
	fmt.Printf("ID for person:name: %d\n", nameID)
	db.Observe(database.MustNewFact(db, "person:name", "db:schema/type", database.MustIdent(db, "db:type/string")))
	db.Observe(database.MustNewFact(db, "person:name", "db:schema/cardinality", database.MustIdent(db, "db:cardinality/one")))

	db.CreateIdent("person:likes")
	db.Observe(database.MustNewFact(db, "person:likes", "db:schema/type", database.MustIdent(db, "db:type/string")))
	db.Observe(database.MustNewFact(db, "person:likes", "db:schema/cardinality", database.MustIdent(db, "db:cardinality/many")))

	db.CreateIdent("person:spouse")
	db.Observe(database.MustNewFact(db, "person:spouse", "db:schema/type", database.MustIdent(db, "db:type/ref")))
	db.Observe(database.MustNewFact(db, "person:spouse", "db:schema/cardinality", database.MustIdent(db, "db:cardinality/one")))

	db.Observe(database.MustNewFact(db, 100, "person:name", "Andrew"))
	db.Observe(database.MustNewFact(db, 100, "person:likes", "tacos"))
	db.Observe(database.MustNewFact(db, 100, "person:likes", "nachos"))
	db.Observe(database.MustNewFact(db, 100, "person:spouse", int64(200)))

	db.Observe(database.MustNewFact(db, 200, "person:name", "Diana"))
	db.Observe(database.MustNewFact(db, 200, "person:likes", "popcorn"))
	db.Observe(database.MustNewFact(db, 200, "person:likes", "nachos"))
	db.Observe(database.MustNewFact(db, 200, "person.spouse", int64(100)))

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

	facts, err := database.GetFacts(db, diana.ID)
	panicOnError(err)

	for _, fact := range facts {
		fmt.Printf("%s\n", fact)
	}

	entity, err := database.GetEntity(db, diana.ID)
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
