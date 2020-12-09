package parser

import (
	"github.com/kendru/darwin/go/typer/ast"
	"github.com/kendru/darwin/go/typer/scanner"
)

type Parser struct {
	scanner *scanner.Scanner
	cur     *scanner.Token
	next    *scanner.Token
}

func New(input string) *Parser {
	p := &Parser{
		scanner: scanner.New(input),
	}
	p.advance()
	return p
}

func (p *Parser) Parse() ast.SExpr {
	switch p.cur.Type {
	case scanner.TokEOF:
		return ast.Nil
	case scanner.TokParenL:
		e = p.parseList()
	}

	panic("Extra tokens left after parsing s-expression")
}

func (p *Parser) parseList() ast.SExpr {
	var members []ast.SExpr
	p.advance() // consume left paren
	for next.Type != scanner.TokParenR {
		if next.Type == scanner.TokEOF {
			panic("Expected \")\"")
		}
		members = append(members)
		next = p.scanner.NextToken()
	}
}

func (p *Parser) advance() {
	p.cur = p.scanner.NextToken()
}
