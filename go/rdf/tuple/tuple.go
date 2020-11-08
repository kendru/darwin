package tuple

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type ElemType int

const (
	TypeUnicode ElemType = iota
	TypeUint64
)

func (t ElemType) String() string {
	switch t {
	case TypeUnicode:
		return "string"
	case TypeUint64:
		return "uint64"
	default:
		return "<invalid>"
	}
}

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
		case uint64:
			t.types = append(t.types, TypeUint64)
		default:
			panic(fmt.Errorf("Unsupported type: %t", elem))
		}
		t.elements = append(t.elements, elem)
	}

	return t
}

func (t *Tuple) String(i int) (string, error) {
	if i > len(t.elements)-1 {
		return "", fmt.Errorf("Index out of range: %d", i)
	}

	elemType := t.types[i]
	if elemType != TypeUnicode {
		return "", fmt.Errorf("Expected string at %d but got %s", i, elemType)
	}

	return t.elements[i].(string), nil
}

func (t *Tuple) UInt64(i int) (uint64, error) {
	if i > len(t.elements)-1 {
		return 0, fmt.Errorf("Index out of range: %d", i)
	}

	elemType := t.types[i]
	if elemType != TypeUint64 {
		return 0, fmt.Errorf("Expected uint64 at %d but got %s", i, elemType)
	}

	return t.elements[i].(uint64), nil
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
		case TypeUint64:
			i := make([]byte, 8)
			binary.BigEndian.PutUint64(i, elem.(uint64))
			buf.Write(i)
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

		case TypeUint64:
			t.elements = append(t.elements, binary.BigEndian.Uint64(buf[i:i+8]))
			i += 7
		default:
			return nil, fmt.Errorf("Invalid type tag at %d: %x", i, typeTag)
		}
	}

	return t, nil
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
