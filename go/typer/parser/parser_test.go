package parser

import (
	"testing"

	"github.com/kendru/darwin/go/typer/ast"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input         string
		expectedSExpr ast.SExpr
	}{
		{"", ast.Nil},
		{"()", ast.NewList()},
		{"(my-fn foo)", ast.NewList(ast.NewAtom("my-fn"), ast.NewIdentifier("foo"))},
	}

	for i, tt := range tests {
		p := New(tt.input)
		expr := p.Parse()

		if expr != tt.expectedSExpr {
			t.Fatalf("tests[%d] - parsed expression did not mach", i)
		}
	}
}
