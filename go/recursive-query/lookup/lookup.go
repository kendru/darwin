package lookup

import (
	"fmt"
	"strings"

	"github.com/kendru/darwin/go/recursive-query/db"
	"github.com/kendru/darwin/go/recursive-query/table"
)

type Executor struct {
	db db.Database
}

func NewExecutor(db db.Database) *Executor {
	return &Executor{db: db}
}

// Operations may pop more operations from the stack
func (e *Executor) Execute(q Query) (map[string]interface{}, error) {
	val := make(map[string]interface{})

	entity, err := getEntity(e.db, q.Table, q.ID)
	if entity == nil {
		return nil, nil
	}

	err = e.processEntityNode(q.Root, entity, val)
	return val, err
}

func (e *Executor) processEntityNode(node EntityNode, entity table.Row, out map[string]interface{}) error {
	for _, childNode := range node.Children {
		switch childNode.Type() {
		case QueryNodeProperty:
			n := childNode.(PropertyNode)
			p := n.Property
			out[p] = entity[p]
		case QueryNodeReference:
			n := childNode.(ReferenceNode)
			inner := make(map[string]interface{})
			out[n.ForeignKey] = inner

			fk, ok := entity[n.ForeignKey]
			if !ok {
				return nil
			}

			fkRef, ok := fk.(string)
			if !ok {
				return fmt.Errorf("Property %s not a foreign key: %v", n.ForeignKey, fk)
			}

			nextEntity, err := getEntity(e.db, n.ForeignTable, fkRef)
			if err != nil {
				return err
			}
			if nextEntity == nil {
				return nil
			}

			e.processEntityNode(n.ChildNode, nextEntity, inner)
		default:
			return fmt.Errorf("Unsupported node type: %d", childNode.Type())
		}
	}

	return nil
}

func getEntity(db db.Database, tableName, id string) (map[string]interface{}, error) {
	tbl := db.GetTable(tableName)
	if tbl == nil {
		return nil, fmt.Errorf("No such table: %v",tableName)
	}

	return tbl.Get(id), nil
}

type Query struct {
	// Table and ID both needed because we are using a relational model without
	// global entity IDs. This should change.
	Table string
	ID string
	Root EntityNode
}

func (q Query) String() string {
	return fmt.Sprintf("%s.%s %s", q.Table, q.ID, q.Root)
}

type QueryNodeType int
const (
	QueryNodeEntity QueryNodeType = iota
	QueryNodeProperty
	QueryNodeReference
)

type QueryNode interface {
	fmt.Stringer
	Type() QueryNodeType
}

type EntityNode struct {
	Children []QueryNode
}

func (n EntityNode) Type() QueryNodeType {
	return QueryNodeEntity
}

func (n EntityNode) String() string {
	childStrings := make([]string, len(n.Children))
	for i, child := range n.Children {
		childStrings[i] = child.String()
	}
	return fmt.Sprintf("[%s]", strings.Join(childStrings, " "))
}

type PropertyNode struct {
	Property string
}

func (n PropertyNode) Type() QueryNodeType {
	return QueryNodeProperty
}

func (n PropertyNode) String() string {
	return fmt.Sprintf(".%s", n.Property)
}

// Do we need to distinguish between one and many references?

type ReferenceNode struct {
	ForeignKey string
	ForeignTable string
	ChildNode EntityNode
}

func (n ReferenceNode) Type() QueryNodeType {
	return QueryNodeReference
}

func (n ReferenceNode) String() string {
	return fmt.Sprintf("{%s.%s: %s}", n.ForeignTable, n.ForeignKey, n.ChildNode)
}
