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

	// ErrNotConvertable is an error where an Value of some type
	// cannot be converted to another type.
	ErrNotConvertible = errors.New("not convertible")

	// ErrDivisionByZero represents a division by zero error.
	ErrDivisionByZero = errors.New("division by zero")
)

// Exception is a special error type returned by (*Runtime).run()
// when an Exception is thrown by throw keyword.
type Exception struct {
	Value Value
}

func (e *Exception) Error() string {
	return "exception: " + AsString(e.Value)
}

// NewExceptionMsg creates a new exception from a message.
func NewExceptionMsg(msg string) error {
	return &Exception{Value: String(msg)}
}

// NewExceptionMsg creates a new exception from a formatted message.
func NewExceptionMsgf(format string, args ...any) error {
	return &Exception{Value: String(fmt.Sprintf(format, args...))}
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
	Sel  string
	Want string
	Got  string
}

func (e *InvalidValueTypeError) Error() string {
	if e.Sel != "" {
		return fmt.Sprintf("invalid value type for '%s': want '%s', got '%s'",
			e.Sel, e.Want, e.Got)
	}
	return fmt.Sprintf("invalid value type: want '%s', got '%s'",
		e.Want, e.Got)
}

// InvalidArgumentTypeError represents an invalid argument value type error.
type InvalidArgumentTypeError struct {
	Name string
	Sel  string
	Want string
	Got  string
}

func (e *InvalidArgumentTypeError) Error() string {
	return fmt.Sprintf("invalid type for argument '%s%s': want '%s', got '%s'",
		e.Name, e.Sel, e.Want, e.Got)
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
