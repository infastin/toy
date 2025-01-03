package tengo

import (
	"slices"

	"github.com/d5/tengo/v2/parser"
)

// BuiltinFunction represents a builtin function.
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
	numResults    int
	sourceMap     map[int]parser.Pos
	free          []*objectPtr
}

func (o *CompiledFunction) Instructions() []byte { return o.instructions }
func (o *CompiledFunction) NumParameters() int   { return o.numParameters }
func (o *CompiledFunction) NumResults() int      { return o.numResults }

// SourcePos returns the source position of the instruction at ip.
func (o *CompiledFunction) SourcePos(ip int) parser.Pos {
	for ip >= 0 {
		if p, ok := o.sourceMap[ip]; ok {
			return p
		}
		ip--
	}
	return parser.NoPos
}

func (o *CompiledFunction) TypeName() string { return "compiled-function" }
func (o *CompiledFunction) String() string   { return "<compiled-function>" }
func (o *CompiledFunction) IsFalsy() bool    { return false }

func (o *CompiledFunction) Copy() Object {
	return &CompiledFunction{
		instructions:  slices.Clone(o.instructions),
		numLocals:     o.numLocals,
		numParameters: o.numParameters,
		varArgs:       o.varArgs,
		free:          slices.Clone(o.free), // DO NOT Copy() of elements; these are variable pointers
	}
}

func (o *CompiledFunction) Call(args ...Object) (Object, error) { return nil, nil }

// UserFunction represents a user function.
type UserFunction struct {
	Name string
	Func CallableFunc
}

func (o *UserFunction) TypeName() string                    { return "user-function:" + o.Name }
func (o *UserFunction) String() string                      { return "<user-function>" }
func (o *UserFunction) IsFalsy() bool                       { return false }
func (o *UserFunction) Copy() Object                        { return &UserFunction{Name: o.Name, Func: o.Func} }
func (o *UserFunction) Call(args ...Object) (Object, error) { return o.Func(args...) }
