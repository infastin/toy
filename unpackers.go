package toy

import (
	"unsafe"
)

// StringOrBytes implements Unpacker interface allowing
// functions to accept String and Bytes as an argument.
type StringOrBytes []byte

func (s StringOrBytes) String() string { return unsafe.String(unsafe.SliceData(s), len(s)) }
func (s StringOrBytes) Bytes() []byte  { return s }

func (s *StringOrBytes) Unpack(v Value) error {
	switch x := v.(type) {
	case String:
		*s = StringOrBytes(x)
	case Bytes:
		*s = StringOrBytes(x)
	default:
		return &InvalidValueTypeError{
			Want: "string or bytes",
			Got:  TypeName(x),
		}
	}
	return nil
}

// IntOrFloat implements Unpacker interface allowing
// functions to accept Int and Float as an argument.
type IntOrFloat float64

func (i *IntOrFloat) Unpack(v Value) error {
	switch x := v.(type) {
	case Int:
		*i = IntOrFloat(x)
	case Float:
		*i = IntOrFloat(x)
	default:
		return &InvalidValueTypeError{
			Want: "int or float",
			Got:  TypeName(x),
		}
	}
	return nil
}

// StringOrChar implements Unpacker interface allowing
// functions to accept String and Char as an argument.
type StringOrChar string

func (s *StringOrChar) Unpack(v Value) error {
	switch x := v.(type) {
	case String:
		*s = StringOrChar(x)
	case Char:
		*s = StringOrChar(x)
	default:
		return &InvalidValueTypeError{
			Want: "string or char",
			Got:  TypeName(x),
		}
	}
	return nil
}
