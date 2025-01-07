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
type CallableFunc = func(args ...Object) (Object, error)
