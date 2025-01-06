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

func (o *BuiltinFunction) Call(args ...Object) (Object, error) {
	if o.Receiver != nil {
		args = append([]Object{o.Receiver}, args...)
	}
	return o.Func(args...)
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
	instructions  []byte
	numLocals     int // number of local variables (including function parameters)
	numParameters int
	varArgs       bool
	sourceMap     map[int]parser.Pos
	free          []*objectPtr
}

func (o *CompiledFunction) NumParameters() int { return o.numParameters }
func (o *CompiledFunction) VarArgs() bool      { return o.varArgs }

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
		free:          slices.Clone(o.free), // DO NOT Copy() of elements; these are variable pointers
	}
}

func (o *CompiledFunction) Call(args ...Object) (Object, error) { return nil, nil }

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
