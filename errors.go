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

	// ErrInvalidIndexType represents an invalid index type error.
	ErrInvalidIndexType = errors.New("invalid index type")

	// ErrInvalidIndex represents an invalid index value error.
	ErrInvalidIndex = errors.New("invalid index")

	// ErrInvalidOperator represents an error for invalid operator usage.
	ErrInvalidOperator = errors.New("invalid operator")

	// ErrNoSuchField is an error where a field with the given name does not exist.
	ErrNoSuchField = errors.New("no such field")

	// ErrNotConvertable is an error where an Object of some type
	// cannot be converted to another type.
	ErrNotConvertible = errors.New("not convertible")
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
