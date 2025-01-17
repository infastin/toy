package toy

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
type CallableFunc = func(v *VM, args ...Object) (Object, error)

// Variable is a user-defined variable for the script.
type Variable struct {
	name  string
	value Object
}

// NewVariable creates a new Variable.
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
