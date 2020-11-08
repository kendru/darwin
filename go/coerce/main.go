package main

import (
	"fmt"
	"strconv"

	"github.com/kendru/darwin/go/coerce/pkg/coerce"
)

func main() {
	coerce.RegisterCoercer(coerce.Int64, coerce.String, coerce.CoercerFunc(func(src coerce.Val) (coerce.Val, error) {
		return coerce.StringVal(strconv.FormatInt(src.Val.(int64), 10)), nil
	}))

	coerce.RegisterCoercer(coerce.Int64, coerce.Boolean, coerce.CoercerFunc(func(src coerce.Val) (coerce.Val, error) {
		return coerce.BooleanVal(src.Val.(int64) != 0), nil
	}))

	coerce.RegisterCoercer(coerce.String, coerce.Int64, coerce.CoercerFunc(func(src coerce.Val) (coerce.Val, error) {
		val := src.Val.(string)
		intVal, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return coerce.Nil, fmt.Errorf("String not convertable to Int64: %q", val)
		}

		return coerce.Int64Val(intVal), nil
	}))

	x := coerce.Int64Val(int64(3456))
	y, err := coerce.Coerce(x, coerce.String)
	if err != nil {
		panic(err)
	}

	fmt.Printf("x:\t%s\ny:\t%s\n", x, y)

	z, err := coerce.Coerce(x, coerce.Boolean)
	if err != nil {
		panic(err)
	}

	fmt.Printf("x:\t%s\nz:\t%s\n", x, z)

	at := coerce.NewGenericType("Map", []coerce.TypeExpr{
		coerce.String,
		coerce.Int64,
	})
	fmt.Printf("Type: %s\n", at.Name())
}
