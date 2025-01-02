package tengo

// Variable is a user-defined variable for the script.
type Variable struct {
	name  string
	value Object
}

// NewVariable creates a Variable.
func NewVariable(name string, value Object) *Variable {
	return &Variable{name: name, value: value}
}

// Name returns the name of the variable.
func (v *Variable) Name() string {
	return v.name
}

// Value returns an empty interface of the variable value.
func (v *Variable) Value() Object {
	return v.value
}
