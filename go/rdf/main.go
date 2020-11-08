package main

import (
	"encoding/json"
	"fmt"

	"github.com/kendru/darwin/go/rdf/database"
)

func main() {
	db := database.New()

	db.Observe(db.MustNewFact(1, "name", "Andrew"))
	db.Observe(db.MustNewFact(1, "likes", "tacos"))
	db.Observe(db.MustNewFact(1, "likes", "nachos"))
	db.Observe(db.MustNewFact(1, "rel:loves", 2))

	db.Observe(db.MustNewFact(2, "name", "Diana"))
	db.Observe(db.MustNewFact(2, "likes", "popcorn"))
	db.Observe(db.MustNewFact(2, "likes", "nachos"))
	db.Observe(db.MustNewFact(2, "rel:loves", 1))

	facts, err := db.GetFacts(1)
	panicOnError(err)

	for _, fact := range facts {
		fmt.Printf("%s\n", fact)
	}

	entity, err := db.GetEntity(1)
	panicOnError(err)

	data, err := json.MarshalIndent(entity, "", "  ")
	panicOnError(err)

	fmt.Println(string(data))
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
