package evaluator

import (
	"fmt"

	"github.com/kendru/darwin/go/monkey/object"
)

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}
			default:
				return newError("argument to `len` not supported. got %s", args[0].Type())
			}
		},
	},

	"head": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `head` must be array. got %s", args[0].Type())
			}

			elems := args[0].(*object.Array).Elements
			if len(elems) == 0 {
				return NULL
			}

			return elems[0]
		},
	},

	"tail": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `tail` must be array. got %s", args[0].Type())
			}

			elems := args[0].(*object.Array).Elements
			length := len(elems)
			if length == 0 {
				return &object.Array{}
			}

			newElems := make([]object.Object, length-1, length-1)
			copy(newElems, elems[1:])

			return &object.Array{Elements: newElems}
		},
	},

	"last": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `last` must be array. got %s", args[0].Type())
			}

			elems := args[0].(*object.Array).Elements
			if len(elems) == 0 {
				return NULL
			}

			return elems[len(elems)-1]
		},
	},

	"push": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `push` must be array. got %s", args[0].Type())
			}

			elems := args[0].(*object.Array).Elements
			length := len(elems)

			newElems := make([]object.Object, length+1, length+1)
			copy(newElems, elems)
			newElems[length] = args[1]

			return &object.Array{Elements: newElems}
		},
	},

	"puts": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return NULL
		},
	},
}
