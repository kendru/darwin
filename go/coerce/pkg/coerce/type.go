package coerce

import (
	"fmt"
	"strings"
)

type TypeExpr interface {
	Name() string
}

type ConcreteType struct {
	name string
}

func (t ConcreteType) Name() string {
	return t.name
}

var (
	Int64   ConcreteType = ConcreteType{"Int64"}
	String               = ConcreteType{"String"}
	Boolean              = ConcreteType{"Boolean"}
	NilType              = ConcreteType{"Nil"}
)

type GenericType struct {
	name       string
	parameters []TypeExpr
}

func NewGenericType(name string, parameters []TypeExpr) GenericType {
	return GenericType{name, parameters}
}

func (t GenericType) Name() string {
	parameterNames := make([]string, len(t.parameters))
	for i, param := range t.parameters {
		parameterNames[i] = param.Name()
	}

	return fmt.Sprintf("%s<%s>", t.name, strings.Join(parameterNames, ", "))
}
