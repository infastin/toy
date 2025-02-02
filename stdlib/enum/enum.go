package enum

import (
	"fmt"

	"github.com/infastin/toy"
)

type Enum[T toy.Value] struct {
	name     string
	variants map[string]T
	fn       toy.CallableFunc
}

func New[T toy.Value](name string, variants map[string]T, constructor toy.CallableFunc) toy.ValueType {
	fn := constructor
	if fn == nil {
		fn = func(v *toy.Runtime, args ...toy.Value) (toy.Value, error) {
			if len(args) != 1 {
				return nil, &toy.WrongNumArgumentsError{
					WantMin: 1,
					WantMax: 1,
					Got:     len(args),
				}
			}
			var res T
			if err := toy.Convert(&res, args[0]); err != nil {
				return nil, err
			}
			return res, nil
		}
	}
	return &Enum[T]{
		name:     name,
		variants: variants,
		fn:       fn,
	}
}

func (e *Enum[T]) Type() toy.ValueType { return nil }
func (e *Enum[T]) String() string      { return fmt.Sprintf("<%s>", e.name) }
func (e *Enum[T]) IsFalsy() bool       { return false }

func (e *Enum[T]) Clone() toy.Value {
	variants := make(map[string]T, len(e.variants))
	for name, variant := range e.variants {
		variants[name] = variant.Clone().(T)
	}
	return &Enum[T]{
		name:     e.name,
		variants: variants,
		fn:       e.fn,
	}
}

func (e *Enum[T]) Name() string                                              { return e.name }
func (e *Enum[T]) Call(v *toy.Runtime, args ...toy.Value) (toy.Value, error) { return e.fn(v, args...) }

func (e *Enum[T]) Property(key toy.Value) (value toy.Value, found bool, err error) {
	keyStr, ok := key.(toy.String)
	if !ok {
		return nil, false, &toy.InvalidKeyTypeError{
			Want: "string",
			Got:  toy.TypeName(key),
		}
	}
	variant, ok := e.variants[string(keyStr)]
	if !ok {
		return toy.Nil, false, nil
	}
	return variant, true, nil
}
