package tengo

import (
	"errors"
	"fmt"
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

// ErrInvalidArgumentType represents an invalid argument value type error.
type ErrInvalidArgumentType struct {
	Name     string
	Expected string
	Found    string
}

func (e *ErrInvalidArgumentType) Error() string {
	return fmt.Sprintf("invalid type for argument %q: want %q, got %q",
		e.Name, e.Expected, e.Found)
}
