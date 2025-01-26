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

	// ErrInvalidOperation represents an error for invalid operator usage.
	ErrInvalidOperation = errors.New("invalid operation")

	// ErrNoSuchField is an error where a field with the given name does not exist.
	ErrNoSuchField = errors.New("no such field")

	// ErrNotConvertable is an error where an Object of some type
	// cannot be converted to another type.
	ErrNotConvertible = errors.New("not convertible")

	// ErrDivisionByZero represents a division by zero error.
	ErrDivisionByZero = errors.New("division by zero")
)

// InvalidValueTypeError represents an invalid index type error.
type InvalidIndexTypeError struct {
	Want string
	Got  string
}

func (e *InvalidIndexTypeError) Error() string {
	return fmt.Sprintf("invalid index type: want '%s', got '%s'",
		e.Want, e.Got)
}

// InvalidKeyTypeError represents an invalid key type error.
type InvalidKeyTypeError struct {
	Want string
	Got  string
}

func (e *InvalidKeyTypeError) Error() string {
	return fmt.Sprintf("invalid key type: want '%s', got '%s'",
		e.Want, e.Got)
}

// InvalidValueTypeError represents an invalid value type error.
type InvalidValueTypeError struct {
	Want string
	Got  string
}

func (e *InvalidValueTypeError) Error() string {
	return fmt.Sprintf("invalid value type: want '%s', got '%s'",
		e.Want, e.Got)
}

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

// UnexpectedArgumentError represents an unexpected argument error.
type UnexpectedArgumentError struct {
	Name string
}

func (e *UnexpectedArgumentError) Error() string {
	return fmt.Sprintf("unexpected argument '%s'", e.Name)
}

// InvalidEntryValueTypeError represents an invalid entry value type error.
type InvalidEntryValueTypeError struct {
	Name string
	Want string
	Got  string
}

func (e *InvalidEntryValueTypeError) Error() string {
	return fmt.Sprintf("invalid type for entry value with key '%s': want '%s', got '%s'",
		e.Name, e.Want, e.Got)
}

// MissingEntryError represents a missing entry error.
type MissingEntryError struct {
	Name string
}

func (e *MissingEntryError) Error() string {
	return fmt.Sprintf("missing entry for '%s'", e.Name)
}

// UnexpectedEntryError represents an unexpected entry error.
type UnexpectedEntryError struct {
	Name string
}

func (e *UnexpectedEntryError) Error() string {
	return fmt.Sprintf("unexpected entry with key '%s'", e.Name)
}
