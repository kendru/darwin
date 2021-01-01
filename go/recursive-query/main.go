package main

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/kendru/darwin/go/recursive-query/db"
	"github.com/kendru/darwin/go/recursive-query/lookup"
	"github.com/kendru/darwin/go/recursive-query/table"
)

func main() {
	db := db.NewSimpleDatabase()
	tbl := table.NewTable("people", table.WithPrimaryKey("id"))
	db.RegisterTable(tbl)

	tbl.Insert(table.Row{"id": "1", "name": "Andrew", "age": 30, "spouse": "2"})
	tbl.Insert(table.Row{"id": "2", "name": "Diana", "age": 33, "spouse": "1"})
	tbl.Insert(table.Row{"id": "3", "name": "Audrey", "age": 8})
	tbl.Insert(table.Row{"id": "4", "name": "Arwen", "age": 8})
	tbl.Insert(table.Row{"id": "5", "name": "Jonah", "age": 6})
	tbl.Insert(table.Row{"id": "6", "name": "Abel", "age": 5})

	q := lookup.Query{
		Table: "people",
		ID:    "1",
		Root: lookup.EntityNode{
			Children: []lookup.QueryNode{
				lookup.PropertyNode{Property: "name"},
				lookup.PropertyNode{Property: "age"},
				lookup.ReferenceNode{
					ForeignKey:   "spouse",
					ForeignTable: "people",
					ChildNode: lookup.EntityNode{
						Children: []lookup.QueryNode{
							lookup.PropertyNode{Property: "name"},
							lookup.PropertyNode{Property: "age"},
							lookup.ReferenceNode{
								ForeignKey:   "spouse",
								ForeignTable: "people",
								ChildNode: lookup.EntityNode{
									Children: []lookup.QueryNode{
										lookup.PropertyNode{Property: "name"},
										lookup.PropertyNode{Property: "age"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	log.Println(prettyPrint(q))

	e := lookup.NewExecutor(db)
	res, err := e.Execute(q)
	if err != nil {
		panic(err)
	}

	log.Println(prettyPrint(res))
}

func prettyPrint(o interface{}) string {
	return prettyPrintIndent(o, 0)
}

func prettyPrintIndent(o interface{}, indent int) string {
	var valStr string
	switch v := reflect.ValueOf(o); v.Kind() {
	case reflect.String:
		valStr = fmt.Sprintf("%q", v.String())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		valStr = fmt.Sprintf("%d", v.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		valStr = fmt.Sprintf("%d", v.Uint())

	case reflect.Float32, reflect.Float64:
		valStr = fmt.Sprintf("%.3g", v.Float())

	case reflect.Complex64, reflect.Complex128:
		valStr = fmt.Sprintf("%.2f", v.Complex())

	case reflect.Bool:
		valStr = fmt.Sprintf("%t", v.Bool())

	case reflect.Array, reflect.Slice:
		innerSpaces := strings.Repeat(" ", indent+1)
		ss := make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			ss[i] = prettyPrintIndent(v.Index(i).Interface(), indent+1)
		}
		valStr = fmt.Sprintf("[\n%s\n%s]", strings.Join(ss, ",\n"), innerSpaces)

	case reflect.Map:
		// TODO: Fix map formatting by distinguishing beteween initial and next line spaces
		innerSpaces := strings.Repeat(" ", indent+1)
		var kvs []string
		iter := v.MapRange()
		for iter.Next() {
			kvs = append(kvs, fmt.Sprintf(
				"%s%s: %s",
				innerSpaces,
				prettyPrintIndent(iter.Key().Interface(), indent+1),
				prettyPrintIndent(iter.Value().Interface(), indent+1),
			))
		}
		valStr = fmt.Sprintf(
			"{\n%s\n%s}",
			strings.Join(kvs, ",\n"),
			innerSpaces,
		)

	case reflect.Func:
		valStr = "<func>"

	case reflect.Ptr:
		if v.IsNil() {
			valStr = "<nil>"
		}
		valStr = fmt.Sprintf("*%s", prettyPrintIndent(v.Elem().Interface(), indent+1))

	case reflect.Struct:
		innerSpaces := strings.Repeat(" ", indent+1)
		var kvs []string
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldType := v.Type().Field(i)
			kvs = append(kvs, fmt.Sprintf(
				"%s%s: %s",
				innerSpaces,
				fieldType.Name,
				prettyPrintIndent(field.Interface(), indent+1),
			))
		}
		var out bytes.Buffer
		out.WriteString(v.Type().Name())
		out.WriteString("{\n")
		out.WriteString(strings.Join(kvs, ",\n"))
		out.WriteByte('\n')
		out.WriteString(innerSpaces)
		out.WriteByte('}')

		valStr = out.String()

	default:
		valStr = "<unknown>"
	}

	return strings.Repeat(" ", indent) + valStr
}
