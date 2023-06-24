package main

import "github.com/kendru/darwin/go/rdf/database"

func main() {
	db := database.NewDefaultDatabase()

	db.CreateIdent("db:id")
	db.CreateIdent("db:schema/type")
	db.CreateIdent("db:type/string")
	db.CreateIdent("db:type/int")
	db.CreateIdent("db:type/ref")
	db.CreateIdent("db:schema/cardinality")
	db.CreateIdent("db:cardinality/one")
	db.CreateIdent("db:cardinality/many")

	dvtSource, _ := db.CreateIdent("dv:type/source")
	dvtHub, _ := db.CreateIdent("dv:type/hub")
	dvtLink, _ := db.CreateIdent("dv:type/link")
	dvtSatellite, _ := db.CreateIdent("dv:type/satellite")

	_, err := db.Transact(database.TxData{
		Updates: []map[string]interface{}{
			{
				"db:id":   database.Fresh(),
				"dv:type": dvtSource,
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
}
