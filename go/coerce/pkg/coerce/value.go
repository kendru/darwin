package coerce

import "fmt"

type Val struct {
	Typ TypeExpr
	Val interface{}
}

var Nil = Val{NilType, nil}

func StringVal(v string) Val {
	return Val{String, v}
}

func Int64Val(v int64) Val {
	return Val{Int64, v}
}

func BooleanVal(v bool) Val {
	return Val{Boolean, v}
}

func (v Val) String() string {
	var valStr string
	switch v.Typ {
	case TypeExpr(NilType):
		valStr = "nil"
	case Boolean:
		if v.Val.(bool) {
			valStr = "true"
		} else {
			valStr = "false"
		}
	case Int64:
		valStr = fmt.Sprintf("%d", v.Val.(int64))
	case String:
		valStr = fmt.Sprintf("%q", v.Val.(string))
	default:
		valStr = fmt.Sprintf("<unknown type: %s>", v.Typ)
	}

	return fmt.Sprintf("%s: %s", valStr, v.Typ)
}
