package scanner

import "testing"

func TestScan(t *testing.T) {
	input := "(eq? MyType 'my-atom)"

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{TokParenL, "("},
		{TokIdentifier, "eq?"},
		{TokTypeName, "MyType"},
		{TokAtom, "'my-atom"},
		{TokParenR, ")"},
		{TokEOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - token type wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}
