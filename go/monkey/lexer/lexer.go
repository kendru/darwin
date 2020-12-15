package lexer

import (
	"bytes"

	"github.com/kendru/darwin/go/monkey/token"
)

type Lexer struct {
	filename     string
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
	col          int
}

func New(filename, input string) *Lexer {
	l := &Lexer{
		filename: filename,
		input:    input,
	}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	tok.Location = l.currentLocation()

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok.Literal = string(ch) + string(l.ch)
			tok.Type = token.EQ
		} else {
			tok = l.newToken(token.ASSIGN, l.ch)
		}
	case '+':
		if l.peekChar() == '+' {
			ch := l.ch
			l.readChar()
			tok.Literal = string(ch) + string(l.ch)
			tok.Type = token.PLUS_PLUS
		} else {
			tok = l.newToken(token.PLUS, l.ch)
		}
	case '-':
		tok = l.newToken(token.MINUS, l.ch)
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok.Literal = string(ch) + string(l.ch)
			tok.Type = token.NOT_EQ
		} else {
			tok = l.newToken(token.BANG, l.ch)
		}
	case '/':
		tok = l.newToken(token.SLASH, l.ch)
	case '*':
		tok = l.newToken(token.ASTERISK, l.ch)
	case '<':
		tok = l.newToken(token.LT, l.ch)
	case '>':
		tok = l.newToken(token.GT, l.ch)
	case ';':
		tok = l.newToken(token.SEMICOLON, l.ch)
	case ':':
		tok = l.newToken(token.COLON, l.ch)
	case '(':
		tok = l.newToken(token.LPAREN, l.ch)
	case ')':
		tok = l.newToken(token.RPAREN, l.ch)
	case ',':
		tok = l.newToken(token.COMMA, l.ch)
	case '{':
		tok = l.newToken(token.LBRACE, l.ch)
	case '}':
		tok = l.newToken(token.RBRACE, l.ch)
	case '[':
		tok = l.newToken(token.LBRACKET, l.ch)
	case ']':
		tok = l.newToken(token.RBRACKET, l.ch)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		}
		if isDigit(l.ch) {
			tok.Literal = l.readNumber()
			tok.Type = token.INT
			return tok
		}
		tok = l.newToken(token.ILLEGAL, l.ch)
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespace() {
	for isWhitespace(l.ch) {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	return l.readWhile(isLetter)
}

func (l *Lexer) readNumber() string {
	return l.readWhile(isDigit)
}

func (l *Lexer) readWhile(testChar charPredicate) string {
	position := l.position
	for testChar(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]

		// TODO: Advance after reading a character
		if l.ch == '\n' {
			l.line++
			l.col = 0
		} else {
			l.col++
		}
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) readString() string {
	var out bytes.Buffer
	for {
		l.readChar()
		if l.ch == '\\' {
			out.WriteByte(l.readEscape())
			continue
		}
		if l.ch == '"' || l.ch == 0 {
			break
		}
		out.WriteByte(l.ch)
	}

	return out.String()
}

func (l *Lexer) readEscape() (ch byte) {
	l.readChar() // move past backslash
	switch l.ch {
	case '\\':
		ch = '\\'
	case '"':
		ch = '"'
	case 't':
		ch = '\t'
	case 'n':
		ch = '\n'
	default:
		ch = 0
		// TODO: error: invalid escape sequence
	}
	return
}

func (l *Lexer) peekChar() byte {
	if l.readPosition > len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) currentLocation() token.Location {
	return token.Location{
		Filename: l.filename,
		Line:     l.line,
		Column:   l.col,
	}
}

func (l *Lexer) newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch), Location: l.currentLocation()}
}

type charPredicate = func(ch byte) bool

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\n' || ch == '\t' || ch == '\r'
}
