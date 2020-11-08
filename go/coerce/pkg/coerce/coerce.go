package coerce

import (
	"fmt"
)

type Coercer interface {
	Coerce(Val) (Val, error)
}

type CoercerFunc func(src Val) (Val, error)

func (c CoercerFunc) Coerce(src Val) (Val, error) {
	return c(src)
}

type coercion struct {
	src TypeExpr
	dst TypeExpr
}

func (c coercion) String() string {
	return fmt.Sprintf("%s -> %s", c.src.Name(), c.dst.Name)
}

var coercers map[coercion]Coercer = make(map[coercion]Coercer)

func RegisterCoercer(src, dst TypeExpr, coercer Coercer) {
	key := coercion{src, dst}
	if _, ok := coercers[key]; ok {
		panic(fmt.Errorf("Cannot redefine coercer: %s", key))
	}
	coercers[key] = coercer
}

func Coerce(src Val, dst TypeExpr) (Val, error) {
	key := coercion{src.Typ, dst}
	coercer, ok := coercers[key]
	if !ok {
		return Nil, fmt.Errorf("No coercion defined for: %s", key)
	}

	// Here we are assuming that src contains a well-typed value
	return coercer.Coerce(src)
}
