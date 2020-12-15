package scanner

import (
	"fmt"
)

type TokenType uint8

const (
	TokParenL TokenType = iota
	TokParenR
	TokIdentifier
	TokAtom
	TokTypeName

	// Special tokens
	TokEOF
	TokIllegal
)

var Nowhere = TokenSource{}

func (t TokenType) String() string {
	switch t {
	case TokParenL:
		return "PAREN-L"
	case TokParenR:
		return "PAREN-R"
	case TokIdentifier:
		return "IDENTIFIER"
	case TokAtom:
		return "ATOM"
	case TokTypeName:
		return "TYPE-NAME"
	case TokEOF:
		return "<EOF>"
	case TokIllegal:
		return "<Illegal>"
	}

	return fmt.Sprintf("Unknown: %d", t)
}

type Token struct {
	Type    TokenType
	Literal string
	Source  TokenSource
}

type TokenSource struct {
	File string
	Line uint
	Col  uint
}

func charToken(t TokenType, l byte) Token {
	return Token{
		Type:    t,
		Literal: string(l),
		Source:  Nowhere,
	}
}
