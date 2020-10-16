package main

import (
	"log"

	"github.com/kendru/darwin/go/recursive-query/db"
	"github.com/kendru/darwin/go/recursive-query/lookup"
	"github.com/kendru/darwin/go/recursive-query/table"
)

func main() {
	db := db.NewSimpleDatabase()

	people := table.NewTable("people", table.WithPrimaryKey("id"))
	db.RegisterTable(people)
	people.Insert(table.Row{
		"id": "1",
		"name": "Andrew",
		"spouse": "2",
	})
	people.Insert(table.Row{
		"id": "2",
		"name": "Diana",
		"spouse": "1",
	})
	people.Insert(table.Row{
		"id": "3",
		"name": "Audrey",
	})


	exec := lookup.NewExecutor(db)

	exec.Push(lookup.FetchEntity{"people", "2"})
	exec.Push(lookup.SelectField{"name"})
	exec.Push(lookup.SelectField{"spouse"})
	exec.Evaluate()

	record, err := exec.Run()
	if err != nil {
		panic(err)
	}

	log.Println(record)
}
