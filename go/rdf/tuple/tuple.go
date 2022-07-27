package tuple

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

type ElemType int

const (
	TypeUnicode ElemType = iota
	TypeInt64
	TypeUint64
	TypeUnknown
)

func (t ElemType) String() string {
	switch t {
	case TypeUnicode:
		return "string"
	case TypeInt64:
		return "int64"
	case TypeUint64:
		return "uint64"
	case TypeUnknown:
		return "unknown"
	default:
		return "<invalid>"
	}
}

// TODO: Replace `types` with a *Schema reference that contains type definitions, aliases,
// nullability flags, etc.
type Tuple struct {
	elements []interface{}
	types    []ElemType
}

func New(elements ...interface{}) *Tuple {
	t := &Tuple{}

	for _, elem := range elements {
		switch elem.(type) {
		case string:
			t.types = append(t.types, TypeUnicode)
		case int64:
			t.types = append(t.types, TypeInt64)
		case uint64:
			t.types = append(t.types, TypeUint64)
		default:
			t.types = append(t.types, TypeUnknown)
		}
		t.elements = append(t.elements, elem)
	}

	return t
}

func (t *Tuple) Schema() []ElemType {
	// TODO: copy to protect against mutations.
	return t.types
}

func (t *Tuple) Concat(other *Tuple) *Tuple {
	return &Tuple{
		elements: append(append([]interface{}{}, t.elements...), other.elements...),
		types:    append(append([]ElemType{}, t.types...), other.types...),
	}
}

func (t *Tuple) Without(idx int) *Tuple {
	return &Tuple{
		elements: append(append([]interface{}{}, t.elements[:idx]...), t.elements[idx+1:]...),
		types:    append(append([]ElemType{}, t.types[:idx]...), t.types[idx+1:]...),
	}
}

func (t *Tuple) GetUntyped(i int) (interface{}, error) {
	if i > len(t.elements)-1 {
		return "", fmt.Errorf("Index out of range: %d", i)
	}

	return t.elements[i], nil
}

func MustGetUntyped(t *Tuple, i int) interface{} {
	val, err := t.GetUntyped(i)
	if err != nil {
		panic(err)
	}
	return val
}

// TODO: Codegen these.

func (t *Tuple) GetString(i int) (string, error) {
	if i > len(t.elements)-1 {
		return "", fmt.Errorf("Index out of range: %d", i)
	}

	elemType := t.types[i]
	if elemType != TypeUnicode {
		return "", fmt.Errorf("Expected string at %d but got %s", i, elemType)
	}

	return t.elements[i].(string), nil
}

func MustGetString(t *Tuple, i int) string {
	val, err := t.GetString(i)
	if err != nil {
		panic(err)
	}
	return val
}

func (t *Tuple) GetInt64(i int) (int64, error) {
	if i > len(t.elements)-1 {
		return 0, fmt.Errorf("Index out of range: %d", i)
	}

	elemType := t.types[i]
	if elemType != TypeInt64 {
		return 0, fmt.Errorf("Expected int64 at %d but got %s", i, elemType)
	}

	return t.elements[i].(int64), nil
}

func MustGetInt64(t *Tuple, i int) int64 {
	val, err := t.GetInt64(i)
	if err != nil {
		panic(err)
	}
	return val
}

func (t *Tuple) GetUInt64(i int) (uint64, error) {
	if i > len(t.elements)-1 {
		return 0, fmt.Errorf("Index out of range: %d", i)
	}

	elemType := t.types[i]
	if elemType != TypeUint64 {
		return 0, fmt.Errorf("Expected uint64 at %d but got %s", i, elemType)
	}

	return t.elements[i].(uint64), nil
}

func MustGetUInt64(t *Tuple, i int) uint64 {
	val, err := t.GetUInt64(i)
	if err != nil {
		panic(err)
	}
	return val
}

func (t *Tuple) Serialize() []byte {
	var buf bytes.Buffer
	buf.Grow(t.sizeHint())

	for i, typ := range t.types {
		buf.WriteByte(byte(typ))
		elem := t.elements[i]
		switch typ {
		case TypeUnicode:
			buf.Write(escapeNulls([]byte(elem.(string))))
			buf.WriteByte(0)

		case TypeInt64:
			buf.Grow(8)
			_ = binary.Write(&buf, binary.BigEndian, elem.(int64))

		case TypeUint64:
			i := make([]byte, 8)
			binary.BigEndian.PutUint64(i, elem.(uint64))
			buf.Write(i)

		default:
			panic(fmt.Errorf("no serializer defined for %s", typ))
		}
	}

	return buf.Bytes()
}

func Deserialize(buf []byte) (*Tuple, error) {
	t := &Tuple{}

	buflen := len(buf)
	for i := 0; i < buflen; i++ {
		typeTag := ElemType(buf[i])
		t.types = append(t.types, typeTag)
		// Skip past type tag
		i++

		switch typeTag {
		case TypeUnicode:
			var s []byte
			for j := 0; ; j++ {
				ch := buf[i+j]
				if ch == 0x00 {
					if i+j+1 < buflen && buf[i+j+1] == 0xff {
						// The null byte was escaped, so we allow the null byte to be emitted
						s = append(s, ch)
						j++
						continue
					}
					// Advance iterator by characters consumed.
					i += j
					t.elements = append(t.elements, string(s))
					break
				}
				s = append(s, ch)
			}

		case TypeInt64:
			var x int64
			if err := binary.Read(bytes.NewReader(buf[i:i+8]), binary.BigEndian, &x); err != nil {
				return nil, fmt.Errorf("error decoding value as int64")
			}
			t.elements = append(t.elements, x)
			i += 7

		case TypeUint64:
			t.elements = append(t.elements, binary.BigEndian.Uint64(buf[i:i+8]))
			i += 7

		default:
			return nil, fmt.Errorf("Invalid type tag at %d: %x", i, typeTag)
		}
	}

	return t, nil
}

func (t *Tuple) String() string {
	var sb strings.Builder
	sb.WriteByte('(')
	for i, elem := range t.elements {
		typ := t.types[i]
		switch typ {
		case TypeUnicode:
			sb.WriteString(elem.(string))
		case TypeInt64:
			sb.WriteString(fmt.Sprintf("%d", elem))
		default:
			return "<noprint>"
		}
		sb.WriteByte(':')
		sb.WriteString(typ.String())

		if i < len(t.elements)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte(')')

	return sb.String()
}

func (t *Tuple) sizeHint() int {
	size := 0
	for i, typ := range t.types {
		size++ // type tag
		switch typ {
		case TypeUnicode:
			size += len(t.elements[i].(string))
			size++ // delimiter
		case TypeUint64:
			size += 8
		}
	}

	return size
}

// escapeNulls retuns a version of a byte string where all instances of a null
// byte (0x00) are followed by a 0xff byte.
func escapeNulls(in []byte) []byte {
	var buf bytes.Buffer
	buf.Grow(len(in))

	for _, byt := range in {
		buf.WriteByte(byt)
		if byt == 0x00 {
			buf.WriteByte(0xff)
		}
	}

	return buf.Bytes()
}
