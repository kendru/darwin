package db

import "github.com/kendru/darwin/go/recursive-query/table"

type Database interface {
	GetTable(name string) *table.Table
}

func NewSimpleDatabase() *SimpleDatabase {
	return &SimpleDatabase{tables: make(map[string]*table.Table)}
}

type SimpleDatabase struct {
	tables map[string]*table.Table
}

func (db *SimpleDatabase) GetTable(name string) *table.Table {
	return db.tables[name]
}

func (db *SimpleDatabase) RegisterTable(table *table.Table) {
	db.tables[table.Name] = table
}
