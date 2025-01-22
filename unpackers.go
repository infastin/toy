package toy

import (
	"unsafe"
)

// StringOrBytes implements Unpacker interface allowing
// functions to accept String and Bytes as an argument.
type StringOrBytes []byte

func (s StringOrBytes) String() string { return unsafe.String(unsafe.SliceData(s), len(s)) }
func (s StringOrBytes) Bytes() []byte  { return s }

func (s *StringOrBytes) Unpack(o Object) error {
	switch o := o.(type) {
	case String:
		*s = StringOrBytes(o)
	case Bytes:
		*s = StringOrBytes(o)
	default:
		return &InvalidValueTypeError{
			Want: "string or bytes",
			Got:  TypeName(o),
		}
	}
	return nil
}

// IntOrFloat implements Unpacker interface allowing
// functions to accept Int and Float as an argument.
type IntOrFloat float64

func (i *IntOrFloat) Unpack(o Object) error {
	switch o := o.(type) {
	case Int:
		*i = IntOrFloat(o)
	case Float:
		*i = IntOrFloat(o)
	default:
		return &InvalidValueTypeError{
			Want: "int or float",
			Got:  TypeName(o),
		}
	}
	return nil
}

// StringOrChar implements Unpacker interface allowing
// functions to accept String and Char as an argument.
type StringOrChar string

func (s *StringOrChar) Unpack(o Object) error {
	switch o := o.(type) {
	case String:
		*s = StringOrChar(o)
	case Char:
		*s = StringOrChar(o)
	default:
		return &InvalidValueTypeError{
			Want: "string or char",
			Got:  TypeName(o),
		}
	}
	return nil
}
