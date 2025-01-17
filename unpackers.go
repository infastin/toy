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

func (a *IntOrFloat) Unpack(o Object) error {
	switch o := o.(type) {
	case Int:
		*a = IntOrFloat(o)
	case Float:
		*a = IntOrFloat(o)
	default:
		return &InvalidValueTypeError{
			Want: "int or float",
			Got:  TypeName(o),
		}
	}
	return nil
}
