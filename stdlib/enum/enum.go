package enum

import (
	"fmt"

	"github.com/infastin/toy"
)

type Enum[T toy.Object] struct {
	name     string
	variants map[string]T
	fn       toy.CallableFunc
}

func New[T toy.Object](name string, variants map[string]T, constructor toy.CallableFunc) toy.ObjectType {
	fn := constructor
	if fn == nil {
		fn = func(v *toy.VM, args ...toy.Object) (toy.Object, error) {
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

func (e *Enum[T]) Type() toy.ObjectType { return nil }
func (e *Enum[T]) String() string       { return fmt.Sprintf("<%s>", e.name) }
func (e *Enum[T]) IsFalsy() bool        { return false }

func (e *Enum[T]) Clone() toy.Object {
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

func (e *Enum[T]) Name() string                                           { return e.name }
func (e *Enum[T]) Call(v *toy.VM, args ...toy.Object) (toy.Object, error) { return e.fn(v, args...) }

func (e *Enum[T]) FieldGet(name string) (toy.Object, error) {
	variant, ok := e.variants[name]
	if !ok {
		return nil, toy.ErrNoSuchField
	}
	return variant, nil
}
