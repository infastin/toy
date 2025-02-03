package toy

import "github.com/infastin/toy/internal/constraints"

var (
	// GlobalsSize is the maximum number of global variables for a VM.
	GlobalsSize = 1024

	// StackSize is the maximum stack size for a VM.
	StackSize = 2048

	// MaxFrames is the maximum number of function frames for a VM.
	MaxFrames = 1024
)

const (
	// SourceFileExtDefault is the default extension for source files.
	SourceFileExtDefault = ".toy"
)

// CallableFunc is a function signature for the callable functions.
type CallableFunc func(r *Runtime, args ...Value) (Value, error)

// Variable is a user-defined variable for the script.
type Variable struct {
	name  string
	value Value
}

// NewVariable creates a new Variable.
func NewVariable(name string, value Value) *Variable {
	return &Variable{name: name, value: value}
}

// Name returns the name of the variable.
func (v *Variable) Name() string {
	return v.name
}

// Value returns the value of the variable.
func (v *Variable) Value() Value {
	return v.value
}

// AsString converts the given value to string.
func AsString(x Value) string {
	if s, ok := x.(String); ok {
		return string(s)
	}
	if c, ok := x.(Convertible); ok {
		var s String
		if err := c.Convert(&s); err == nil {
			return string(s)
		}
	}
	return string(x.String())
}

// AsString converts the given value to bool.
func AsBool(x Value) bool {
	return !x.IsFalsy()
}

// AsInt converts the given value to an integer.
func AsInt[T constraints.Int | constraints.Uint](v Value) (res T, err error) {
	var i Int
	if err := Convert(&i, v); err != nil {
		return res, err
	}
	return T(i), nil
}

// AsFloat converts the given value to a floating point number.
func AsFloat[T constraints.Float](v Value) (res T, err error) {
	var i Float
	if err := Convert(&i, v); err != nil {
		return res, err
	}
	return T(i), nil
}
