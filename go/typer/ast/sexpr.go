package ast

type SExprType int

const (
	SETypeList SExprType = iota
	SETypeAtom
	SETypeTypeName
	SETypeIdentifier
	SETypeNil
)

type SExpr struct {
	Type  SExprType
	Value interface{}
}

func NewList(elems ...[]SExpr) SExpr {
	return newSExpr(SETypeList, elems)
}

func NewAtom(name string) SExpr {
	return newSExpr(SETypeAtom, name)
}

func NewTypeName(name string) SExpr {
	return newSExpr(SETypeTypeName, name)
}

func NewIdentifier(name string) SExpr {
	return newSExpr(SETypeIdentifier, name)
}

var Nil = SExpr{Type: SETypeNil}

func newSExpr(t SExprType, val interface{}) SExpr {
	return SExpr{
		Type:  t,
		Value: val,
	}
}
