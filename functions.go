package toy

import (
	"slices"

	"github.com/infastin/toy/parser"
)

// BuiltinFunction represents a builtin function provided from Go.
type BuiltinFunction struct {
	Name     string
	Receiver Object
	Func     CallableFunc
}

func (o *BuiltinFunction) TypeName() string { return "builtin-function:" + o.Name }
func (o *BuiltinFunction) String() string   { return "<builtin-function>" }
func (o *BuiltinFunction) IsFalsy() bool    { return false }

func (o *BuiltinFunction) Copy() Object {
	var receiver Object
	if o.Receiver != nil {
		receiver = o.Receiver.Copy()
	}
	return &BuiltinFunction{
		Name:     o.Name,
		Receiver: receiver,
		Func:     o.Func,
	}
}

func (o *BuiltinFunction) Call(v *VM, args ...Object) (Object, error) {
	if o.Receiver != nil {
		args = append([]Object{o.Receiver}, args...)
	}
	return o.Func(v, args...)
}

func (o *BuiltinFunction) WithReceiver(recv Object) *BuiltinFunction {
	return &BuiltinFunction{
		Name:     o.Name,
		Receiver: recv,
		Func:     o.Func,
	}
}

// CompiledFunction represents a compiled function.
type CompiledFunction struct {
	receiver      Object
	instructions  []byte
	numLocals     int // number of local variables (including function parameters)
	numParameters int
	varArgs       bool
	sourceMap     map[int]parser.Pos
	deferMap      []parser.Pos
	free          []*objectPtr
}

func (o *CompiledFunction) Instructions() []byte { return slices.Clone(o.instructions) }
func (o *CompiledFunction) NumLocals() int       { return o.numLocals }
func (o *CompiledFunction) NumParameters() int   { return o.numParameters }
func (o *CompiledFunction) VarArgs() bool        { return o.varArgs }

func (o *CompiledFunction) TypeName() string { return "compiled-function" }
func (o *CompiledFunction) String() string   { return "<compiled-function>" }
func (o *CompiledFunction) IsFalsy() bool    { return false }

func (o *CompiledFunction) Copy() Object {
	return &CompiledFunction{
		instructions:  o.instructions,
		numLocals:     o.numLocals,
		numParameters: o.numParameters,
		varArgs:       o.varArgs,
		sourceMap:     o.sourceMap,
		deferMap:      o.deferMap,
		free:          slices.Clone(o.free), // DO NOT Copy() of elements; these are variable pointers
	}
}

func (o *CompiledFunction) Call(v *VM, args ...Object) (Object, error) {
	if o.receiver != nil {
		args = append([]Object{o.receiver}, args...)
	}
	numArgs := len(args)

	if o.varArgs {
		// if the closure is variadic,
		// roll up all variadic parameters into an array
		numRealArgs := o.numParameters - 1
		numVarArgs := numArgs - numRealArgs
		if numVarArgs >= 0 {
			varArgs := slices.Clone(args[numRealArgs:])
			args = append(args[:numRealArgs], NewArray(varArgs))
			numArgs = numRealArgs + 1
		}
	}

	if numArgs != o.numParameters {
		if o.varArgs {
			return nil, &WrongNumArgumentsError{
				WantMin: o.numParameters - 1,
				Got:     numArgs,
			}
		}
		return nil, &WrongNumArgumentsError{
			WantMin: o.numParameters,
			WantMax: o.numParameters,
			Got:     numArgs,
		}
	}

	// test if it's tail-call
	if o == v.curFrame.fn { // recursion
		nextOp := v.curInsts[v.ip+1]
		if nextOp == parser.OpReturn ||
			(nextOp == parser.OpPop && parser.OpReturn == v.curInsts[v.ip+2]) {
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
	v.curFrame = &(v.frames[v.framesIndex])
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
		varArgs:       o.varArgs,
		sourceMap:     o.sourceMap,
		deferMap:      o.deferMap,
		free:          slices.Clone(o.free),
	}
}

// sourcePos returns the source position of the instruction at ip.
func (o *CompiledFunction) sourcePos(ip int) parser.Pos {
	for ip >= 0 {
		if p, ok := o.sourceMap[ip]; ok {
			return p
		}
		ip--
	}
	return parser.NoPos
}
