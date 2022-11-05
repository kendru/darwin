package inflect

import (
	"strings"
	"unicode"
)

type charClass int

const (
	ccNone charClass = iota
	ccLower
	ccUpper
	ccNumber
	ccPunctuation
)

var Default = Inflect{}

type Inflect struct {
}

func (i Inflect) ToCamelCase(str string) string {
	return camelLike(str, false)
}

func (i Inflect) ToPascalCase(str string) string {
	return camelLike(str, true)
}

func (i Inflect) ToSnakeCase(str string) string {
	return delimitWith(str, '_')
}

func (i Inflect) ToKebabCase(str string) string {
	return delimitWith(str, '-')
}

func camelLike(str string, ucFirst bool) string {
	var sb strings.Builder
	sb.Grow(len(str))
	lastClass := ccNone
	if ucFirst {
		lastClass = ccPunctuation
	}
	for _, char := range str {
		cc := charClassOf(char)

		switch cc {
		case ccLower:
			if lastClass == ccPunctuation {
				char = unicode.ToUpper(char)
			}

		case ccUpper:
			if lastClass == ccNone {
				char = unicode.ToLower(char)
			}

		case ccPunctuation:
			lastClass = cc
			continue
		}

		sb.WriteRune(char)
		lastClass = cc
	}

	return sb.String()
}

func delimitWith(str string, delimiter rune) string {
	var sb strings.Builder
	sb.Grow(len(str))
	lastClass := ccNone
	for _, char := range str {
		cc := charClassOf(char)

		writeDelim := false
		switch cc {
		case ccUpper:
			char = unicode.ToLower(char)
			if lastClass == ccLower || lastClass == ccPunctuation {
				writeDelim = true
			}

		case ccLower:
			if lastClass == ccPunctuation {
				writeDelim = true
			}

		case ccPunctuation:
			lastClass = cc
			continue
		}

		if writeDelim {
			sb.WriteRune(delimiter)
		}
		sb.WriteRune(char)
		lastClass = cc
	}

	return sb.String()
}

func charClassOf(char rune) charClass {
	if unicode.IsUpper(char) {
		return ccUpper
	}
	if unicode.IsLower(char) {
		return ccLower
	}
	if unicode.IsNumber(char) {
		return ccNumber
	}
	return ccPunctuation
}
