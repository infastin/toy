package toy

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	// ErrStackOverflow is a stack overflow error.
	ErrStackOverflow = errors.New("stack overflow")

	// ErrInvalidIndexType represents an invalid index type.
	ErrInvalidIndexType = errors.New("invalid index type")

	// ErrInvalidIndex represents an invalid index value.
	ErrInvalidIndexValue = errors.New("invalid index")

	// ErrInvalidOperator represents an error for invalid operator usage.
	ErrInvalidOperator = errors.New("invalid operator")

	// ErrBytesLimit represents an error where the size of bytes value exceeds
	// the limit.
	ErrBytesLimit = errors.New("exceeding bytes size limit")

	// ErrStringLimit represents an error where the size of string value
	// exceeds the limit.
	ErrStringLimit = errors.New("exceeding string size limit")

	// ErrNotIndexable is an error where an Object is not indexable.
	ErrNotIndexable = errors.New("not indexable")

	// ErrNoSuchField is an error where an Object doesn't have fields.
	ErrNoFields = errors.New("no fields")

	// ErrNotSliceable is an error where an Object is not sliceable.
	ErrNotSliceable = errors.New("not sliceable")

	// ErrNoSuchField is an error where a field with the given name does not exist.
	ErrNoSuchField = errors.New("no such field")

	// ErrNotHashable is an error where an Object is not hashable.
	ErrNotHashable = errors.New("not hashable")

	// ErrNotConvertable is an error where an Object of some type
	// cannot be converted to another type.
	ErrNotConvertible = errors.New("not convertible")

	// ErrNotImplemented is an error where an Object has not implemented a
	// required method.
	ErrNotImplemented = errors.New("not implemented")
)

// InvalidArgumentTypeError represents an invalid argument value type error.
type InvalidArgumentTypeError struct {
	Name string
	Want string
	Got  string
}

func (e *InvalidArgumentTypeError) Error() string {
	return fmt.Sprintf("invalid type for argument '%s': want '%s', got '%s'",
		e.Name, e.Want, e.Got)
}

// WrongNumArgumentsError represents a wrong number of arguments error.
type WrongNumArgumentsError struct {
	WantMin int
	WantMax int
	Got     int
}

func (e *WrongNumArgumentsError) Error() string {
	var b strings.Builder
	b.WriteString("want")
	if e.WantMin == e.WantMax {
		b.WriteByte(' ')
		b.WriteString(strconv.Itoa(e.WantMax))
	} else if e.Got < e.WantMin {
		b.WriteString(" at least ")
		b.WriteString(strconv.Itoa(e.WantMin))
	} else if e.Got > e.WantMax {
		b.WriteString(" at most ")
		b.WriteString(strconv.Itoa(e.WantMax))
	}
	b.WriteString(" argument(s), got ")
	b.WriteString(strconv.Itoa(e.Got))
	return b.String()
}

// MissingArgumentError represents a missing argument error.
type MissingArgumentError struct {
	Name string
}

func (e *MissingArgumentError) Error() string {
	return fmt.Sprintf("missing argument for '%s'", e.Name)
}
