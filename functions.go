package toy

import (
	"fmt"
	"slices"

	"github.com/infastin/toy/bytecode"
	"github.com/infastin/toy/token"
)

// functionTypeImpl represents the function type.
type functionTypeImpl struct{}

func (functionTypeImpl) Type() ObjectType { return nil }
func (functionTypeImpl) String() string   { return "<function>" }
func (functionTypeImpl) IsFalsy() bool    { return false }
func (functionTypeImpl) Clone() Object    { return functionTypeImpl{} }
func (functionTypeImpl) Name() string     { return "function" }

// FunctionType is the type of all functions.
var FunctionType = functionTypeImpl{}

// BuiltinFunction represents a builtin function provided from Go.
type BuiltinFunction struct {
	name string
	recv Object
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

func (o *BuiltinFunction) Type() ObjectType { return FunctionType }
func (o *BuiltinFunction) String() string   { return fmt.Sprintf("<builtin-function %q>", o.name) }
func (o *BuiltinFunction) IsFalsy() bool    { return false }

func (o *BuiltinFunction) Clone() Object {
	var recv Object
	if o.recv != nil {
		recv = o.recv.Clone()
	}
	return &BuiltinFunction{
		name: o.name,
		recv: recv,
		fn:   o.fn,
	}
}

func (o *BuiltinFunction) Call(v *VM, args ...Object) (Object, error) {
	if o.recv != nil {
		args = append([]Object{o.recv}, args...)
	}
	return o.fn(v, args...)
}

func (o *BuiltinFunction) WithReceiver(recv Object) *BuiltinFunction {
	return &BuiltinFunction{
		name: o.name,
		recv: recv,
		fn:   o.fn,
	}
}

// CompiledFunction represents a compiled function.
type CompiledFunction struct {
	receiver      Object
	instructions  []byte
	numLocals     int // number of local variables (including function parameters)
	numParameters int
	numOptionals  int
	varArgs       bool
	sourceMap     map[int]token.Pos
	deferMap      []token.Pos
	free          []*objectPtr
}

func (o *CompiledFunction) Type() ObjectType { return FunctionType }
func (o *CompiledFunction) String() string   { return "<compiled-function>" }
func (o *CompiledFunction) IsFalsy() bool    { return false }

func (o *CompiledFunction) Clone() Object {
	return &CompiledFunction{
		instructions:  o.instructions,
		numLocals:     o.numLocals,
		numParameters: o.numParameters,
		numOptionals:  o.numOptionals,
		varArgs:       o.varArgs,
		sourceMap:     o.sourceMap,
		deferMap:      o.deferMap,
		free:          slices.Clone(o.free), // DO NOT Clone() of elements; these are variable pointers
	}
}

func (o *CompiledFunction) Call(v *VM, args ...Object) (Object, error) {
	if o.receiver != nil {
		args = append([]Object{o.receiver}, args...)
	}
	numArgs := len(args)

	numRealParams := o.numParameters
	if o.varArgs {
		numRealParams--
	}
	numRequiredParams := numRealParams - o.numOptionals

	if o.numOptionals > 0 && numArgs >= numRequiredParams {
		for i := numArgs; i < numRealParams; i++ {
			args = append(args, Nil)
		}
		numArgs = len(args)
	}

	if o.varArgs && numArgs >= numRealParams {
		// if the function is variadic,
		// roll up all variadic parameters into an array
		varArgs := slices.Clone(args[numRealParams:])
		args = append(args[:numRealParams], NewArray(varArgs))
		numArgs = numRealParams + 1
	}

	if numArgs != o.numParameters {
		if o.varArgs {
			return nil, &WrongNumArgumentsError{
				WantMin: numRequiredParams,
				Got:     numArgs,
			}
		}
		return nil, &WrongNumArgumentsError{
			WantMin: numRequiredParams,
			WantMax: o.numParameters,
			Got:     numArgs,
		}
	}

	// test if it's tail-call
	if o == v.curFrame.fn { // recursion
		nextOp := v.curInsts[v.ip+1]
		if nextOp == bytecode.OpReturn ||
			(nextOp == bytecode.OpPop && bytecode.OpReturn == v.curInsts[v.ip+2]) {
			copy(v.stack[v.curFrame.basePointer:], args)
			v.ip = -1 // reset IP to beginning of the frame
			return nil, nil
		}
	}
	if v.framesIndex >= MaxFrames {
		return nil, ErrStackOverflow
	}

	// save current call frame
	frame := v.curFrame

	// update call frame
	v.curFrame.ip = v.ip // store current ip before call
	v.curFrame = &v.frames[v.framesIndex]
	v.curFrame.fn = o
	v.curFrame.freeVars = o.free
	v.curFrame.basePointer = v.sp
	v.curInsts = o.instructions
	v.ip = -1
	v.framesIndex++
	copy(v.stack[v.sp:], args)
	v.sp = v.sp + o.numLocals

	if frame.subvm {
		// we should run in the subVM
		// NOTE: we could use proper coroutines
		// that are already in the runtime, but
		// it's not possible to link with them because of the
		// changes to cmd/link in go1.23
		// See: https://github.com/golang/go/issues/67401
		v.run()
		if v.err != nil {
			// subVM closed with an error
			return nil, nil
		}
		// pop result from the stack and return it
		ret := v.stack[v.sp-1]
		v.sp--
		return ret, nil
	}

	return nil, nil
}

func (o *CompiledFunction) WithReceiver(recv Object) *CompiledFunction {
	return &CompiledFunction{
		receiver:      recv,
		instructions:  o.instructions,
		numLocals:     o.numLocals,
		numParameters: o.numParameters,
		numOptionals:  o.numOptionals,
		varArgs:       o.varArgs,
		sourceMap:     o.sourceMap,
		deferMap:      o.deferMap,
		free:          slices.Clone(o.free),
	}
}

func (o *CompiledFunction) Instructions() []byte { return slices.Clone(o.instructions) }
func (o *CompiledFunction) NumLocals() int       { return o.numLocals }
func (o *CompiledFunction) NumParameters() int   { return o.numParameters }
func (o *CompiledFunction) NumOptionals() int    { return o.numOptionals }
func (o *CompiledFunction) VarArgs() bool        { return o.varArgs }

// sourcePos returns the source position of the instruction at ip.
func (o *CompiledFunction) sourcePos(ip int) token.Pos {
	for ip >= 0 {
		if p, ok := o.sourceMap[ip]; ok {
			return p
		}
		ip--
	}
	return token.NoPos
}
