package scanner

type Scanner struct {
	input   string
	pos     int
	readPos int
	ch      byte
}

func New(input string) *Scanner {
	s := &Scanner{
		input: input,
	}
	s.advance()
	return s
}

func (s *Scanner) NextToken() Token {
	s.skipWhitespace()

	var tok Token

	switch s.ch {
	case '(':
		tok = charToken(TokParenL, s.ch)

	case ')':
		tok = charToken(TokParenR, s.ch)

	case '\'':
		tok.Type = TokAtom
		tok.Literal = s.readAtom()
		return tok

	case 0:
		tok.Type = TokEOF
		tok.Literal = ""

	default:
		if isAlphaUpper(s.ch) {
			tok.Type = TokTypeName
			tok.Literal = s.readTypeName()
			return tok
		}

		if isAlphaLower(s.ch) {
			literal := s.readIdent()
			// TODO: Check if keyword/BIF
			tok.Type = TokIdentifier
			tok.Literal = literal
			return tok
		}
	}

	s.advance()

	return tok
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' {
		s.advance()
	}
}

func (s *Scanner) readAtom() string {
	pos := s.pos
	s.advance() // Eat tick mark
	for isAtomChar(s.ch) {
		s.advance()
	}

	return s.input[pos:s.pos]
}

func (s *Scanner) readIdent() string {
	pos := s.pos
	for isIdentChar(s.ch) {
		s.advance()
	}

	return s.input[pos:s.pos]
}

func (s *Scanner) readTypeName() string {
	pos := s.pos
	for isLetter(s.ch) {
		s.advance()
	}

	return s.input[pos:s.pos]
}

func (s *Scanner) advance() {
	if s.readPos >= len(s.input) {
		s.ch = 0
	} else {
		s.ch = s.input[s.readPos]
	}
	s.pos = s.readPos
	s.readPos++
}

func isAtomChar(c byte) bool {
	return c == '-' || isAlphaLower(c)
}

func isIdentChar(c byte) bool {
	return c == '-' || c == '?' || isAlphaLower(c)
}

func isAlphaLower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

func isAlphaUpper(c byte) bool {
	return 'A' <= c && c <= 'Z'
}

func isLetter(c byte) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}
