package catalog

import (
	"fmt"
	"strings"

	"github.com/kendru/darwin/go/dbcatalog/internal/datastore"
)

type Catalog struct {
	ds datastore.Datastore
}

func New(opts CatalogOpts) *Catalog {
	return &Catalog{
		ds: opts.Datastore,
	}
}

func (c *Catalog) DefineType(name string, goType string) {
	addType(c.ds, name, goType)
}

func (c *Catalog) DefineAttribute(name string, dbType string) {
	addAttr(c.ds, name, dbType)
}

func (c *Catalog) DefineClass(name string, attributes []string) {
	addClass(c.ds, name, attributes)
}

func (c *Catalog) GetClass(name string) (*Class, error) {
	class, err := c.ds.Get("class:" + name)
	if err != nil {
		return nil, fmt.Errorf("error getting class %q: %w", name, err)
	}

	classProps := class.(map[string]interface{})
	className := classProps["name"].(string)
	attrNames := classProps["attributes"].([]string)

	attributes := make([]*Attribute, len(attrNames))
	for i, attrName := range attrNames {
		if attributes[i], err = c.GetAttribute(attrName); err != nil {
			return nil, err
		}
	}

	return &Class{
		name:       className,
		attributes: attributes,
	}, nil
}

type Class struct {
	name       string
	attributes []*Attribute
}

func (c *Class) String() string {
	var sb strings.Builder

	sb.WriteString("<Class ")
	sb.WriteString(c.name)
	sb.WriteString(":{\n")
	for _, attr := range c.attributes {
		sb.WriteByte('\t')
		sb.WriteString(attr.String())
		sb.WriteByte('\n')
	}
	sb.WriteString("}>")

	return sb.String()
}

func (c *Catalog) GetAttribute(name string) (*Attribute, error) {
	attr, err := c.ds.Get("attr:" + name)
	if err != nil {
		return nil, fmt.Errorf("error getting attribute %q: %w", name, err)
	}

	attrProps := attr.(map[string]string)
	attrName := attrProps["name"]
	typeName := attrProps["type"]

	dbType, err := c.GetDBType(typeName)
	if err != nil {
		return nil, err
	}

	return &Attribute{
		name:   attrName,
		dbType: dbType,
	}, nil
}

func (c *Catalog) GetDBType(name string) (*DBType, error) {
	dbType, err := c.ds.Get("type:" + name)
	if err != nil {
		return nil, fmt.Errorf("error getting type %q: %w", name, err)
	}

	return &DBType{
		name:   name,
		goType: dbType.(map[string]string)["goType"],
	}, nil
}

type CatalogOpts struct {
	Datastore datastore.Datastore
}

type Attribute struct {
	name   string
	dbType *DBType
}

func (a *Attribute) String() string {
	return fmt.Sprintf("<Attr:{name=%q, type=%s}>", a.name, a.dbType)
}

type DBType struct {
	name   string
	goType string
}

func (t *DBType) String() string {
	return fmt.Sprintf("<Type:{name=%q, internalType=%q}>", t.name, t.goType)
}

func BootstrapSchema(ds datastore.Datastore) error {
	// Create list entries.
	ds.Set(":types", []string{})
	ds.Set(":attributes", []string{})
	ds.Set(":classes", []string{})

	tt := []struct {
		name   string
		goType string
	}{
		{
			"int",
			"int64",
		},
		{
			"uint",
			"uint64",
		},
		{
			"text",
			"string",
		},
	}
	for _, t := range tt {
		if err := addType(ds, t.name, t.goType); err != nil {
			return fmt.Errorf("error adding type: %w", err)
		}
	}

	attrs := []struct {
		name   string
		dbType string
	}{
		{
			"class",
			"text",
		},
		{
			"attribute",
			"text",
		},
		{
			"collection",
			"string",
		},
		{
			"id",
			"int",
		},
	}
	for _, a := range attrs {
		if err := addAttr(ds, a.name, a.dbType); err != nil {
			return fmt.Errorf("error adding attribute: %w", err)
		}
	}

	return nil
}

func addClass(ds datastore.Datastore, name string, attributes []string) error {
	if err := appendToList(ds, ":classes", name); err != nil {
		return err
	}

	return ds.Set(fmt.Sprintf("class:%s", name), map[string]interface{}{
		"name":       name,
		"attributes": attributes,
	})
}

func addAttr(ds datastore.Datastore, name, dbType string) error {
	if err := appendToList(ds, ":attributes", name); err != nil {
		return err
	}

	return ds.Set(fmt.Sprintf("attr:%s", name), map[string]string{
		"name": name,
		"type": dbType,
	})
}

func addType(ds datastore.Datastore, name, goType string) error {
	if err := appendToList(ds, ":types", name); err != nil {
		return err
	}

	return ds.Set(fmt.Sprintf("type:%s", name), map[string]string{
		"name":   name,
		"goType": goType,
	})
}

func appendToList(ds datastore.Datastore, listName string, elem string) error {
	listRes, err := ds.Get(listName)
	if err != nil {
		return fmt.Errorf("error getting list %q: %w", listName, err)
	}

	list := append(listRes.([]string), elem)

	if err := ds.Set(listName, list); err != nil {
		return fmt.Errorf("error updating list %q: %w", listName, err)
	}

	return nil
}
