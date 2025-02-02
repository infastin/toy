package toy

import (
	"fmt"
	"slices"

	"github.com/infastin/toy/bytecode"
	"github.com/infastin/toy/token"
)

// functionTypeImpl represents the function type.
type functionTypeImpl struct{}

func (functionTypeImpl) Type() ValueType { return nil }
func (functionTypeImpl) String() string  { return "<function>" }
func (functionTypeImpl) IsFalsy() bool   { return false }
func (functionTypeImpl) Clone() Value    { return functionTypeImpl{} }
func (functionTypeImpl) Name() string    { return "function" }

// FunctionType is the type of all functions.
var FunctionType = functionTypeImpl{}

// BuiltinFunction represents a builtin function provided from Go.
type BuiltinFunction struct {
	name string
	recv Value
	fn   CallableFunc
}

// NewBuiltinFunction creates a new BuiltinFunction.
func NewBuiltinFunction(name string, fn CallableFunc) *BuiltinFunction {
	return &BuiltinFunction{
		name: name,
		recv: nil,
		fn:   fn,
	}
}

func (f *BuiltinFunction) Type() ValueType { return FunctionType }
func (f *BuiltinFunction) String() string  { return fmt.Sprintf("<builtin-function %q>", f.name) }
func (f *BuiltinFunction) IsFalsy() bool   { return false }

func (f *BuiltinFunction) Clone() Value {
	var recv Value
	if f.recv != nil {
		recv = f.recv.Clone()
	}
	return &BuiltinFunction{
		name: f.name,
		recv: recv,
		fn:   f.fn,
	}
}

func (f *BuiltinFunction) Call(r *Runtime, args ...Value) (Value, error) {
	if f.recv != nil {
		args = append([]Value{f.recv}, args...)
	}
	return f.fn(r, args...)
}

func (f *BuiltinFunction) WithReceiver(recv Value) *BuiltinFunction {
	return &BuiltinFunction{
		name: f.name,
		recv: recv,
		fn:   f.fn,
	}
}

// CompiledFunction represents a compiled function.
type CompiledFunction struct {
	receiver      Value
	instructions  []byte
	numLocals     int // number of local variables (including function parameters)
	numParameters int
	numOptionals  int
	varArgs       bool
	sourceMap     map[int]token.Pos
	deferMap      []token.Pos
	free          []*valuePtr
}

func (f *CompiledFunction) Type() ValueType { return FunctionType }
func (f *CompiledFunction) String() string  { return "<compiled-function>" }
func (f *CompiledFunction) IsFalsy() bool   { return false }

func (f *CompiledFunction) Clone() Value {
	return &CompiledFunction{
		instructions:  f.instructions,
		numLocals:     f.numLocals,
		numParameters: f.numParameters,
		numOptionals:  f.numOptionals,
		varArgs:       f.varArgs,
		sourceMap:     f.sourceMap,
		deferMap:      f.deferMap,
		free:          slices.Clone(f.free), // DO NOT Clone() of elements; these are variable pointers
	}
}

func (f *CompiledFunction) Call(r *Runtime, args ...Value) (Value, error) {
	// this method always pauses the execution of the current runtime
	return f.call(r, args, true)
}

func (f *CompiledFunction) call(r *Runtime, args []Value, pause bool) (Value, error) {
	if f.receiver != nil {
		args = append([]Value{f.receiver}, args...)
	}
	numArgs := len(args)

	numRealParams := f.numParameters
	if f.varArgs {
		numRealParams--
	}
	numRequiredParams := numRealParams - f.numOptionals

	if f.numOptionals > 0 && numArgs >= numRequiredParams {
		for i := numArgs; i < numRealParams; i++ {
			args = append(args, Nil)
		}
		numArgs = len(args)
	}

	if f.varArgs && numArgs >= numRealParams {
		// if the function is variadic,
		// roll up all variadic parameters into an array
		varArgs := slices.Clone(args[numRealParams:])
		args = append(args[:numRealParams], NewArray(varArgs))
		numArgs = numRealParams + 1
	}

	if numArgs != f.numParameters {
		if f.varArgs {
			return nil, &WrongNumArgumentsError{
				WantMin: numRequiredParams,
				Got:     numArgs,
			}
		}
		return nil, &WrongNumArgumentsError{
			WantMin: numRequiredParams,
			WantMax: f.numParameters,
			Got:     numArgs,
		}
	}

	// test if it's tail-call
	if !pause && f == r.curFrame.fn { // recursion
		nextOp := r.curInsts[r.ip+1]
		if nextOp == bytecode.OpReturn ||
			(nextOp == bytecode.OpPop && bytecode.OpReturn == r.curInsts[r.ip+2]) {
			copy(r.stack[r.curFrame.basePointer:], args)
			r.ip = -1 // reset IP to beginning of the frame
			return nil, nil
		}
	}

	return r.call(f, args, pause)
}

func (f *CompiledFunction) WithReceiver(recv Value) *CompiledFunction {
	return &CompiledFunction{
		receiver:      recv,
		instructions:  f.instructions,
		numLocals:     f.numLocals,
		numParameters: f.numParameters,
		numOptionals:  f.numOptionals,
		varArgs:       f.varArgs,
		sourceMap:     f.sourceMap,
		deferMap:      f.deferMap,
		free:          slices.Clone(f.free),
	}
}

func (f *CompiledFunction) Instructions() []byte { return slices.Clone(f.instructions) }
func (f *CompiledFunction) NumLocals() int       { return f.numLocals }
func (f *CompiledFunction) NumParameters() int   { return f.numParameters }
func (f *CompiledFunction) NumOptionals() int    { return f.numOptionals }
func (f *CompiledFunction) VarArgs() bool        { return f.varArgs }

// sourcePos returns the source position of the instruction at ip.
func (f *CompiledFunction) sourcePos(ip int) token.Pos {
	for ip >= 0 {
		if p, ok := f.sourceMap[ip]; ok {
			return p
		}
		ip--
	}
	return token.NoPos
}
