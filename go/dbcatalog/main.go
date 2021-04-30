package main

import (
	"fmt"

	"github.com/kendru/darwin/go/dbcatalog/internal/catalog"
	"github.com/kendru/darwin/go/dbcatalog/internal/datastore"
)

func main() {
	ds := datastore.NewDefault()
	catalog.BootstrapSchema(ds)

	c := catalog.New(catalog.CatalogOpts{
		Datastore: ds,
	})

	attr, err := c.GetAttribute("class")
	ensureNoErr(err, "could not get bootstrapped class")
	fmt.Printf("Got bootstrapped attribute: %s\n", attr)

	c.DefineAttribute("person/name", "text")
	c.DefineAttribute("person/age", "uint")
	c.DefineClass("person", []string{
		"id",
		"person/name",
		"person/age",
	})

	classList, err := ds.Get(":classes")
	ensureNoErr(err, "could not get class list")
	fmt.Println("Classes:")
	for _, class := range classList.([]string) {
		class, err := c.GetClass(class)
		ensureNoErr(err, "could not get class by name")
		fmt.Println(class)
	}
}

func ensureNoErr(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %w", msg, err))
	}
}
